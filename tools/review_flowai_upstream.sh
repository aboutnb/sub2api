#!/usr/bin/env bash

# Read-only pre-merge review for the FlowAI fork.
set -Eeuo pipefail

EXPECTED_BRANCH="sub2api-flowai"
UPSTREAM_REVIEW_ACK="${FLOWAI_UPSTREAM_REVIEW_ACK:-}"
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"

if [[ -z "$ROOT" ]]; then
  printf '[FAIL] not inside a git worktree\n' >&2
  exit 1
fi

cd "$ROOT"

failures=0
warnings=0

pass() {
  printf '[PASS] %s\n' "$1"
}

fail() {
  printf '[FAIL] %s\n' "$1" >&2
  failures=$((failures + 1))
}

warn() {
  printf '[WARN] %s\n' "$1" >&2
  warnings=$((warnings + 1))
}

is_protected_path() {
  case "$1" in
    backend/internal/service/gateway_scheduling.go|\
    backend/internal/service/openai_account_scheduler.go|\
    backend/internal/service/openai_gateway_scheduling.go|\
    backend/internal/service/gemini_messages_compat_service.go|\
    backend/internal/service/batch_image_public.go|\
    backend/internal/repository/account_repo.go|\
    backend/internal/service/concurrency_service.go|\
    backend/internal/handler/gateway_helper.go|\
    backend/internal/service/auth_service.go|\
    backend/internal/service/setting_parse.go|\
    backend/internal/repository/migrations_runner.go|\
    backend/internal/repository/email_broadcast_repo.go|\
    backend/internal/service/email_broadcast_service.go|\
    backend/internal/usdtpayment/*|\
    backend/internal/service/checkin_service.go|\
    backend/internal/service/project_mihomo_service.go|\
    backend/migrations/*.sql|\
    frontend/src/i18n/*|\
    frontend/src/i18n/**/*|\
    deploy/docker-compose.preview.yml|\
    deploy/Caddyfile.flowai|\
    deploy/Dockerfile|\
    deploy/.env.example|\
    deploy/.env.preview.example|\
    deploy/README.md|\
    deploy/config.example.yaml|\
    deploy/deploy-preview-image.sh|\
    .github/workflows/preview-image.yml)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

printf 'FlowAI upstream pre-merge review\n'
printf 'root: %s\n' "$ROOT"

branch="$(git symbolic-ref --quiet --short HEAD || true)"
if [[ -z "$branch" ]]; then
  branch="${GITHUB_REF_NAME:-}"
fi
if [[ "$branch" == "$EXPECTED_BRANCH" ]]; then
  pass "current branch is $EXPECTED_BRANCH"
else
  fail "expected branch $EXPECTED_BRANCH, got ${branch:-detached HEAD}"
fi

if git show-ref --verify --quiet refs/remotes/upstream/main; then
  pass "upstream/main is available"
else
  fail "upstream/main is unavailable; fetch it before reviewing"
fi

if [[ "$failures" -gt 0 ]]; then
  printf 'Upstream pre-merge review stopped: %d failure(s), %d warning(s)\n' "$failures" "$warnings" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain=v1)" ]]; then
  if [[ "${FLOWAI_REQUIRE_CLEAN:-0}" == "1" ]]; then
    fail "worktree is dirty (FLOWAI_REQUIRE_CLEAN=1); preserve local files and review from a clean worktree"
  else
    warn "worktree has local changes; the review is read-only, but merge only after preserving them"
  fi
else
  pass "worktree is clean"
fi

merge_base="$(git merge-base HEAD upstream/main)"
local_head="$(git rev-parse HEAD)"
upstream_head="$(git rev-parse upstream/main)"
printf 'merge-base: %s\n' "$merge_base"
printf 'local HEAD: %s\n' "$local_head"
printf 'upstream/main: %s\n' "$upstream_head"

incoming_count="$(git rev-list --count HEAD..upstream/main)"
local_only_count="$(git rev-list --count upstream/main..HEAD)"
printf 'incoming commits: %s\n' "$incoming_count"
printf 'FlowAI-only commits: %s\n' "$local_only_count"

if [[ "$incoming_count" -eq 0 ]]; then
  pass "upstream/main has no commits not already in the FlowAI branch"
else
  printf '\nIncoming commits:\n'
  git log --format='  %h %s' HEAD..upstream/main
  if [[ "$UPSTREAM_REVIEW_ACK" == "$upstream_head" ]]; then
    pass "upstream review acknowledgement matches $upstream_head"
  elif [[ -z "$UPSTREAM_REVIEW_ACK" ]]; then
    fail "upstream review acknowledgement is required; inspect the report, then rerun with FLOWAI_UPSTREAM_REVIEW_ACK=$upstream_head"
  else
    fail "upstream review acknowledgement does not match upstream/main: expected $upstream_head, got $UPSTREAM_REVIEW_ACK"
  fi
fi

local_paths="$(git diff --name-only "$merge_base" HEAD | sort -u)"
incoming_paths="$(git diff --name-only "$merge_base" upstream/main | sort -u)"
overlap_paths="$(comm -12 \
  <(printf '%s\n' "$local_paths" | sed '/^$/d') \
  <(printf '%s\n' "$incoming_paths" | sed '/^$/d'))"

if [[ -n "$overlap_paths" ]]; then
  warn "upstream and FlowAI both changed these paths; review each one before merging"
  printf '%s\n' "$overlap_paths" | sed 's/^/  /'
else
  pass "no path overlap between upstream-only and FlowAI-only changes"
fi

protected_incoming=""
while IFS= read -r path; do
  [[ -z "$path" ]] && continue
  if is_protected_path "$path"; then
    protected_incoming+="$path\n"
  fi
done <<< "$incoming_paths"

if [[ -n "$protected_incoming" ]]; then
  if [[ "$UPSTREAM_REVIEW_ACK" == "$upstream_head" ]]; then
    pass "protected upstream paths were explicitly acknowledged for review"
  else
    fail "upstream changed FlowAI-protected paths; review the conflict matrix before acknowledging"
  fi
  printf '%b' "$protected_incoming" | sed 's/^/  /'
else
  pass "upstream did not change a FlowAI-protected path"
fi

incoming_migrations="$(git diff --name-status "$merge_base" upstream/main -- 'backend/migrations/*.sql' || true)"
if [[ -n "$incoming_migrations" ]]; then
  warn "upstream added or changed SQL migrations; compare complete filenames and checksums"
  printf '%s\n' "$incoming_migrations" | sed 's/^/  /'
else
  pass "upstream has no migration changes relative to the merge base"
fi

if [[ "$incoming_count" -gt 0 ]]; then
  merge_tree_output=""
  if merge_tree_output="$(git merge-tree --write-tree HEAD upstream/main 2>&1)"; then
    pass "merge simulation has no conflicts"
  else
    fail "merge simulation reports conflicts; do not merge until they are reviewed"
    printf '%s\n' "$merge_tree_output" | sed -n '1,120p' >&2
  fi
else
  pass "merge simulation is unnecessary; upstream is already contained"
fi

if [[ "$failures" -gt 0 ]]; then
  printf 'Upstream pre-merge review failed: %d failure(s), %d warning(s)\n' "$failures" "$warnings" >&2
  exit 1
fi

printf 'Upstream pre-merge review passed: 0 failure(s), %d warning(s)\n' "$warnings"
