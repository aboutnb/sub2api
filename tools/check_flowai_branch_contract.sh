#!/usr/bin/env bash

# Read-only release gate for the FlowAI fork. It checks semantic anchors rather
# than inferring behavior from commit names.
set -Eeuo pipefail

EXPECTED_BRANCH="main"
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
CHANGELOG="docs/FLOWAI_CHANGELOG.md"
CONTRACT="docs/FLOWAI_BRANCH_CONTRACT.md"
RELEASE_CHECKLIST="docs/FLOWAI_RELEASE_CHECKLIST.md"
GOVERNANCE_MARKER="[flowai-governance]"

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

require_file() {
  local path="$1"
  local label="${2:-$1}"
  if [[ -f "$path" ]]; then
    pass "$label exists"
  else
    fail "$label is missing: $path"
  fi
}

require_text() {
  local path="$1"
  local needle="$2"
  local label="$3"
  if [[ ! -f "$path" ]]; then
    fail "$label cannot be checked; file is missing: $path"
  elif grep -Fq -- "$needle" "$path"; then
    pass "$label"
  else
    fail "$label missing semantic anchor in $path: $needle"
  fi
}

require_first_priority_comparison() {
  local path="$1"
  local function_name="$2"
  local expected="$3"
  local label="$4"
  if [[ ! -f "$path" ]]; then
    fail "$label cannot be checked; file is missing: $path"
    return
  fi
  if awk -v function_name="$function_name" -v expected="$expected" '
    $0 ~ "^func .*" function_name "\\(" { inside = 1; next }
    inside && $0 ~ /^func / { exit !(found && matched) }
    inside && index($0, "Priority") && ($0 ~ />/ || $0 ~ /</) {
      found = 1
      matched = index($0, expected) > 0
      exit !matched
    }
    END { exit !(inside && found && matched) }
  ' "$path"; then
    pass "$label"
  else
    fail "$label did not find the expected first priority comparison in $path"
  fi
}

forbid_text() {
  local path="$1"
  local needle="$2"
  local label="$3"
  if [[ ! -f "$path" ]]; then
    fail "$label cannot be checked; file is missing: $path"
  elif grep -Fq -- "$needle" "$path"; then
    fail "$label found forbidden semantic anchor in $path: $needle"
  else
    pass "$label"
  fi
}

forbid_regex() {
  local path="$1"
  local pattern="$2"
  local label="$3"
  if [[ ! -f "$path" ]]; then
    fail "$label cannot be checked; file is missing: $path"
  elif grep -Eq -- "$pattern" "$path"; then
    fail "$label found forbidden semantic pattern in $path: $pattern"
  else
    pass "$label"
  fi
}

require_marker_section() {
  local path="$1"
  local begin="$2"
  local end="$3"
  local label="$4"
  if [[ ! -f "$path" ]]; then
    fail "$label cannot be checked; file is missing: $path"
    return
  fi
  if awk -v begin="$begin" -v end="$end" '
    index($0, begin) { found_begin = 1; next }
    found_begin && index($0, end) { found_end = 1; exit }
    END { exit !(found_begin && found_end) }
  ' "$path"; then
    pass "$label"
  else
    fail "$label is missing a complete marker section in $path"
  fi
}

extract_marker_section() {
  local path="$1"
  local begin="$2"
  local end="$3"
  awk -v begin="$begin" -v end="$end" '
    index($0, begin) { inside = 1; found_begin = 1; next }
    inside && index($0, end) { found_end = 1; exit }
    inside { print }
    END { if (!found_begin || !found_end) exit 1 }
  ' "$path"
}

ledger_has_commit() {
  local ledger="$1"
  local commit="$2"
  local prefix="${commit:0:9}"
  local short_token='`'"$prefix"'`'
  local full_token='`'"$commit"'`'

  printf '%s\n' "$ledger" | grep -Fq -- "$short_token" && return 0
  printf '%s\n' "$ledger" | grep -Fq -- "$full_token"
}

extract_hash_tokens() {
  local section="$1"
  printf '%s\n' "$section" | grep -Eo '`[0-9a-f]{7,40}`' | tr -d '`' || true
}

validate_ledger_section() {
  local section="$1"
  local kind="$2"
  local tokens
  local duplicates
  local token
  local resolved
  local parent_count

  tokens="$(extract_hash_tokens "$section")"
  if [[ -z "$tokens" ]]; then
    fail "$kind commit ledger has no hash entries"
    return
  fi

  duplicates="$(printf '%s\n' "$tokens" | sort | uniq -d)"
  if [[ -n "$duplicates" ]]; then
    fail "$kind commit ledger contains duplicate hash entries: $duplicates"
  else
    pass "$kind commit ledger has no duplicate hash entries"
  fi

  while IFS= read -r token; do
    [[ -z "$token" ]] && continue
    resolved="$(git rev-parse --verify "${token}^{commit}" 2>/dev/null || true)"
    if [[ -z "$resolved" ]]; then
      fail "$kind commit ledger contains an unknown commit: $token"
      continue
    fi
    parent_count="$(git rev-list --parents -n 1 "$resolved" | awk '{print NF - 1}')"
    if [[ "$kind" == "non-merge" && "$parent_count" -gt 1 ]]; then
      fail "non-merge ledger entry is a merge commit: $token"
    elif [[ "$kind" == "merge" && "$parent_count" -lt 2 ]]; then
      fail "merge ledger entry is not a merge commit: $token"
    fi
  done <<< "$tokens"
}

governance_path_allowed() {
  case "$1" in
    .gitignore|Makefile|.github/workflows/preview-image.yml|.github/workflows/release.yml|.github/workflows/backend-ci.yml|.github/aivoza-upstream-ref|tools/check_flowai_branch_contract.sh|tools/review_flowai_upstream.sh|docs/FLOWAI_*.md|deploy/README.md)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

is_governance_commit() {
  local commit="$1"
  local subject
  local path
  local found_path=0

  subject="$(git show -s --format=%s "$commit")"
  case "$subject" in
    *"$GOVERNANCE_MARKER"*) ;;
    *) return 1 ;;
  esac

  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    found_path=1
    governance_path_allowed "$path" || return 1
  done < <(git diff-tree --no-commit-id --name-only -r "$commit" --)

  [[ "$found_path" -eq 1 ]]
}

# GitHub adds a merge envelope after the reviewed PR commit. It introduces no
# code only when the base is already included and the candidate tree is exact.
is_reviewed_pr_merge() {
  local commit="$1" subject parents first second
  subject="$(git show -s --format=%s "$commit")"
  [[ "$subject" =~ ^Merge\ pull\ request\ \#[0-9]+\ from\ aboutnb/(feature|fix|sync|migrate)/ ]] || return 1
  parents="$(git show -s --format=%P "$commit")"
  [[ "$(wc -w <<< "$parents" | tr -d ' ')" == "2" ]] || return 1
  first="${parents%% *}"
  second="${parents##* }"
  git merge-base --is-ancestor "$first" "$second" || return 1
  git diff --quiet "$commit" "$second"
}

check_commit_ledger() {
  local merge_base="$1"
  local nonmerge_ledger=""
  local merge_ledger=""
  local commit
  local total_nonmerge=0
  local total_merge=0
  local missing_nonmerge=0
  local missing_merge=0
  local governance_exempt=0
  local prefix

  if ! nonmerge_ledger="$(extract_marker_section "$CHANGELOG" \
    '<!-- FLOWAI_LEDGER_NON_MERGE_BEGIN -->' \
    '<!-- FLOWAI_LEDGER_NON_MERGE_END -->')"; then
    fail "could not read the non-merge commit ledger"
  fi
  if ! merge_ledger="$(extract_marker_section "$CHANGELOG" \
    '<!-- FLOWAI_LEDGER_MERGE_BEGIN -->' \
    '<!-- FLOWAI_LEDGER_MERGE_END -->')"; then
    fail "could not read the merge commit ledger"
  fi

  if [[ -n "$nonmerge_ledger" ]]; then
    validate_ledger_section "$nonmerge_ledger" "non-merge"
  fi
  if [[ -n "$merge_ledger" ]]; then
    validate_ledger_section "$merge_ledger" "merge"
  fi

  while IFS= read -r commit; do
    [[ -z "$commit" ]] && continue
    total_nonmerge=$((total_nonmerge + 1))
    prefix="${commit:0:9}"
    if is_governance_commit "$commit"; then
      governance_exempt=$((governance_exempt + 1))
    elif ! ledger_has_commit "$nonmerge_ledger" "$commit"; then
      fail "non-merge commit is missing from $CHANGELOG: $prefix $(git show -s --format=%s "$commit")"
      missing_nonmerge=$((missing_nonmerge + 1))
    fi
  done < <(git rev-list --no-merges "$merge_base..HEAD")

  while IFS= read -r commit; do
    [[ -z "$commit" ]] && continue
    total_merge=$((total_merge + 1))
    prefix="${commit:0:9}"
    if ! ledger_has_commit "$merge_ledger" "$commit" && ! is_reviewed_pr_merge "$commit"; then
      fail "merge commit is missing from $CHANGELOG: $prefix $(git show -s --format=%s "$commit")"
      missing_merge=$((missing_merge + 1))
    fi
  done < <(git rev-list --merges "$merge_base..HEAD")

  if [[ "$missing_nonmerge" -eq 0 ]]; then
    pass "non-merge ledger covers $total_nonmerge current commits ($governance_exempt governance exemption(s))"
  fi
  if [[ "$missing_merge" -eq 0 ]]; then
    pass "merge ledger covers $total_merge current merge commits"
  fi
}

check_migration_ledger() {
  local merge_base="$1"
  local migration_ledger=""
  local path
  local added=0
  local modified=0
  local deleted=0
  local missing=0
  local modified_without_record=0
  local migration_record

  if ! migration_ledger="$(extract_marker_section "$CHANGELOG" \
    '<!-- FLOWAI_MIGRATION_LEDGER_BEGIN -->' \
    '<!-- FLOWAI_MIGRATION_LEDGER_END -->')"; then
    fail "could not read the migration ledger"
    return
  fi

  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    [[ "$path" == *.sql ]] || continue

    if ! git cat-file -e "HEAD:$path" 2>/dev/null; then
      deleted=$((deleted + 1))
      fail "migration was deleted from the FlowAI branch: $path"
      continue
    fi

    if git cat-file -e "$merge_base:$path" 2>/dev/null; then
      modified=$((modified + 1))
      migration_record="$(printf '%s\n' "$migration_ledger" | grep -F -- "\`$path\`" | head -n 1 || true)"
      if [[ -z "$migration_record" ]] ||
        ! printf '%s\n' "$migration_record" | grep -Eiq 'checksum|兼容|immutable|不可编辑'; then
        fail "existing migration changed without an explicit checksum/compatibility record: $path"
        modified_without_record=$((modified_without_record + 1))
      fi
    else
      added=$((added + 1))
    fi

    if printf '%s\n' "$migration_ledger" | grep -Fq -- "\`$path\`"; then
      :
    else
      fail "changed migration is missing from $CHANGELOG: $path"
      missing=$((missing + 1))
    fi
  done < <(git diff --name-only "$merge_base" HEAD -- 'backend/migrations/*.sql')

  if [[ "$missing" -eq 0 && "$deleted" -eq 0 && "$modified_without_record" -eq 0 ]]; then
    pass "migration ledger covers $added added and $modified modified SQL migration(s)"
  fi
}

printf 'FlowAI branch contract check\n'
printf 'root: %s\n' "$ROOT"

for path in "$CHANGELOG" "$CONTRACT" "$RELEASE_CHECKLIST"; do
  require_file "$path" "FlowAI governance document"
done
require_marker_section "$CHANGELOG" \
  '<!-- FLOWAI_MIGRATION_LEDGER_BEGIN -->' \
  '<!-- FLOWAI_MIGRATION_LEDGER_END -->' \
  'migration ledger markers'
require_marker_section "$CHANGELOG" \
  '<!-- FLOWAI_LEDGER_NON_MERGE_BEGIN -->' \
  '<!-- FLOWAI_LEDGER_NON_MERGE_END -->' \
  'non-merge commit ledger markers'
require_marker_section "$CHANGELOG" \
  '<!-- FLOWAI_LEDGER_MERGE_BEGIN -->' \
  '<!-- FLOWAI_LEDGER_MERGE_END -->' \
  'merge commit ledger markers'
require_text "$CHANGELOG" \
  '每次合并上游或发布前' \
  'changelog requires a pre-merge and pre-release review'
require_text "$CONTRACT" \
  '只修改代码、不更新本文件或变更台账的提交，不能作为 FlowAI 发布候选。' \
  'contract requires documentation for behavior changes'
require_text "$RELEASE_CHECKLIST" \
  'FLOWAI_CHANGELOG.md' \
  'release checklist points to the change ledger'
require_text "$RELEASE_CHECKLIST" \
  'make review-flowai-upstream' \
  'release checklist points to the upstream pre-merge review'

# Branch and ancestry are hard gates. GitHub Actions checks out a detached ref,
# so use GITHUB_REF_NAME there while still rejecting an unrelated local ref.
branch="$(git symbolic-ref --quiet --short HEAD || true)"
if [[ -z "$branch" ]]; then
  branch="${GITHUB_HEAD_REF:-${GITHUB_REF_NAME:-}}"
fi
case "$branch" in
  main|feature/*|fix/*|sync/*|migrate/*)
    pass "Aivoza mainline or reviewed candidate branch: $branch" ;;
  *) fail "expected main or feature/fix/sync/migrate candidate, got ${branch:-detached HEAD}" ;;
esac
if [[ -n "${GITHUB_BASE_REF:-}" && "$GITHUB_BASE_REF" != "main" ]]; then
  fail "Aivoza pull requests must target main"
fi

merge_base=""
if git show-ref --verify --quiet refs/remotes/upstream/main; then
  pass "upstream/main is available"
  merge_base="$(git merge-base HEAD upstream/main 2>/dev/null || true)"
  if git merge-base --is-ancestor upstream/main HEAD; then
    pass "HEAD contains upstream/main"
  else
    fail "HEAD is behind or unrelated to upstream/main; review and merge upstream before release"
  fi
  if [[ -n "$merge_base" ]]; then
    pass "merge base is $merge_base"
  else
    fail "HEAD and upstream/main have no common merge base"
  fi
else
  fail "upstream/main is unavailable; fetch it before running the release gate"
fi

version="$(tr -d '[:space:]' < backend/cmd/server/VERSION 2>/dev/null || true)"
if [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  pass "version is $version"
else
  fail "backend/cmd/server/VERSION is not a semantic version: ${version:-empty}"
fi

if git diff --check && git diff --cached --check; then
  pass "working tree diff has no whitespace errors"
else
  fail "working tree diff contains whitespace errors"
fi

if [[ "${FLOWAI_REQUIRE_CLEAN:-0}" == "1" ]]; then
  if [[ -z "$(git status --porcelain=v1)" ]]; then
    pass "worktree is clean"
  else
    fail "worktree is dirty (FLOWAI_REQUIRE_CLEAN=1); review staged, unstaged, and untracked files"
  fi
else
  if [[ -z "$(git status --porcelain=v1)" ]]; then
    pass "worktree is clean"
  else
    warn "worktree has staged, unstaged, or untracked changes; review them before publishing"
  fi
fi

# Account priority: 1 is highest and lower account values win. These checks
# deliberately do not reject ASC globally because group-member priority also
# intentionally uses ASC.
require_text backend/internal/service/gateway_scheduling.go \
  'acc.account.Priority < minPriority' \
  'gateway layered selection keeps minimum account priority'
require_text backend/internal/service/gateway_scheduling.go \
  'return a.account.Priority < b.account.Priority' \
  'gateway account sort uses ascending priority'
require_text backend/internal/service/gateway_scheduling.go \
  'return a.Priority < b.Priority' \
  'gateway priority-only sort uses ascending priority'
require_text backend/internal/service/openai_account_scheduler.go \
  'return left.account.Priority < right.account.Priority' \
  'OpenAI scheduler tie-break uses lower priority value'
require_text backend/internal/service/openai_account_scheduler.go \
  'priorityFactor = 1 - float64(item.priority-minPriority)/float64(maxPriority-minPriority)' \
  'OpenAI scheduler score rewards lower priority value'
require_text backend/internal/service/openai_account_scheduler.go \
  'priorityFactor = 1 - float64(candidate.priority-minPriority)/float64(maxPriority-minPriority)' \
  'OpenAI scheduler snapshot rewards lower priority value'
require_text backend/internal/service/openai_gateway_scheduling.go \
  'Higher priority (lower value)' \
  'legacy OpenAI scheduler documents lower priority value'
require_text backend/internal/service/gemini_messages_compat_service.go \
  'candidate.Priority < current.Priority' \
  'Gemini scheduler uses lower priority value'
require_first_priority_comparison backend/internal/service/openai_gateway_scheduling.go \
  'isBetterAccount' 'if candidate.Priority < current.Priority' \
  'legacy OpenAI selector checks lower priority value first'
require_first_priority_comparison backend/internal/service/gemini_messages_compat_service.go \
  'isBetterGeminiAccount' 'if candidate.Priority < current.Priority' \
  'Gemini selector checks lower priority value first'
require_text backend/internal/service/batch_image_public.go \
  'accounts[i].Priority < accounts[j].Priority' \
  'batch image scheduler uses lower priority value'
require_text backend/internal/repository/account_repo.go \
  'Order(dbent.Asc(dbaccount.FieldPriority)' \
  'account repository uses ascending account priority'
require_text backend/internal/repository/account_repo.go \
  'a.priority ASC' \
  'group query uses ascending account priority'
require_text frontend/src/i18n/locales/zh/admin/accounts.ts \
  '1 为最高优先级，数值越小的账号优先使用' \
  'Chinese account priority copy'
require_text frontend/src/i18n/locales/en/admin/accounts.ts \
  'Priority 1 is highest; lower-value accounts are used first' \
  'English account priority copy'
require_text frontend/src/i18n/locales/zh/admin/overview.ts \
  '1 为最高优先级，数值越小越优先，用于账号调度' \
  'Chinese scheduling priority copy'
require_text frontend/src/i18n/locales/en/admin/overview.ts \
  'Priority 1 is highest; lower values are preferred for account scheduling' \
  'English scheduling priority copy'

# Reverse checks catch the exact upstream regressions that previously passed
# compilation while silently changing account scheduling semantics.
forbid_regex backend/internal/service/gateway_scheduling.go \
  'filterByMaxPriority|a\.account\.Priority[[:space:]]*>[[:space:]]*b\.account\.Priority|a\.Priority[[:space:]]*>[[:space:]]*b\.Priority' \
  'gateway code contains no larger-value-first implementation'
forbid_regex backend/internal/service/openai_account_scheduler.go \
  'priorityFactor[[:space:]]*=[[:space:]]*float64\((item|candidate)\.priority-minPriority\)[[:space:]]*/' \
  'OpenAI score contains no larger-value-first priority factor'
forbid_regex backend/internal/service/openai_gateway_scheduling.go \
  'Higher priority \(larger value\)|优先级更高（数值更大）|a\.account\.Priority[[:space:]]*>[[:space:]]*b\.account\.Priority' \
  'legacy OpenAI scheduler contains no larger-value priority rule'
forbid_regex backend/internal/service/batch_image_public.go \
  'accounts\[i\]\.Priority[[:space:]]*>[[:space:]]*accounts\[j\]\.Priority' \
  'batch image scheduler contains no descending account priority sort'
forbid_regex backend/internal/repository/account_repo.go \
  'a\.priority[[:space:]]+DESC' \
  'account SQL contains no descending account priority sort'
forbid_regex frontend/src/i18n/locales/zh/admin/accounts.ts \
  '数值越大优先级越高|优先级越大.*优先|数值越大.*优先使用' \
  'Chinese account copy contains no larger-value-first rule'
forbid_regex frontend/src/i18n/locales/en/admin/accounts.ts \
  'higher value.*priority|higher number.*priority|larger.*priority|Higher value accounts are used first' \
  'English account copy contains no larger-value-first rule'

# Gemini and the legacy OpenAI selector contain both sides of a comparison in
# their guard clauses. The positive anchors above check the lower-value branch;
# these checks retain the legitimate inverse branch used to reject worse candidates.
require_text backend/internal/service/openai_gateway_scheduling.go \
  'if candidate.Priority > current.Priority' \
  'legacy OpenAI selector has an inverse priority guard'
require_text backend/internal/service/gemini_messages_compat_service.go \
  'if candidate.Priority > current.Priority' \
  'Gemini selector has an inverse priority guard'

# User concurrency and risk-signup admission.
require_text backend/internal/service/concurrency_service.go \
  'if maxConcurrency < 0' \
  'negative user/account concurrency is denied before Redis'
require_text backend/internal/service/concurrency_service.go \
  'if maxConcurrency == 0' \
  'zero concurrency keeps unlimited behavior'
require_text backend/internal/handler/gateway_helper.go \
  'if maxConcurrency < 0' \
  'negative concurrency is rejected before the wait queue'
require_text backend/internal/service/concurrency_service_test.go \
  'TestAcquireUserSlot_DenyAll' \
  'deny-all concurrency regression test'
require_text backend/internal/handler/gateway_helper_hotpath_test.go \
  'TestAcquireUserSlotWithWait_DenyAllDoesNotQueue' \
  'deny-all wait-queue regression test'
require_text backend/internal/service/setting_parse.go \
  'func normalizeUserConcurrency(value int) int' \
  'invalid configured concurrency has a normalization function'
require_text backend/internal/handler/admin/setting_handler_update.go \
  'req.DefaultConcurrency < -1' \
  'default concurrency rejects/normalizes values below -1'
require_text backend/internal/service/admin_user.go \
  'concurrency must be -1 or greater' \
  'admin user paths reject values below -1'
require_text backend/internal/service/auth_service.go \
  'return -1' \
  'risk signup starts with deny-all concurrency'
require_text backend/internal/service/signup_risk_grant_test.go \
  'TestSignupRiskGrantWritesUnlimitedConcurrencyAfterApproval' \
  'risk signup applies the grant only after approval'

# Required FlowAI feature, migration, i18n, and deployment surfaces.
for path in \
  backend/migrations/185_auth_ip_bans.sql \
  backend/migrations/191_daily_checkin.sql \
  backend/migrations/202_checkin_unrecharged_reward_policy.sql \
  backend/migrations/203_add_email_broadcast_tasks.sql \
  backend/migrations/204_generalize_email_broadcasts.sql \
  backend/migrations/205_add_invoice_applications.sql \
  backend/migrations/206_add_usdt_payments.sql \
  backend/migrations/208_signup_risk_grant_guard.sql \
  backend/internal/repository/email_broadcast_repo.go \
  backend/internal/service/email_broadcast_service.go \
  backend/internal/service/checkin_service.go \
  backend/internal/service/project_mihomo_service.go \
  frontend/src/i18n/index.ts \
  docs/DAILY_CHECK_IN_PRD.md \
  docs/PAYMENT.md \
  docs/PAYMENT_CN.md \
  BEPUSDT_USDT_INTEGRATION_DESIGN.md \
  deploy/README.md \
  deploy/docker-compose.preview.yml \
  deploy/deploy-preview-image.sh \
  tools/review_flowai_upstream.sh \
  .github/workflows/preview-image.yml; do
  require_file "$path"
done

require_text backend/internal/repository/email_broadcast_repo.go \
  'FOR UPDATE OF r, t SKIP LOCKED' \
  'email broadcast claim is row-locked and deduplicated'
require_text backend/internal/service/email_broadcast_service.go \
  'an ambiguous broadcast delivery automatically.' \
  'email broadcast ambiguous sends are not auto-retried'
require_text backend/internal/server/routes/payment.go \
  'invoices := authenticated.Group("/invoices")' \
  'invoice routes remain registered'
require_text backend/internal/server/routes/user.go \
  'user.POST("/checkin", h.Checkin.CheckIn)' \
  'user check-in route remains registered'
require_text backend/internal/repository/migrations_runner.go \
  'migrationChecksumCompatibilityRules' \
  'migration checksum compatibility is explicit and allowlisted'
require_text backend/internal/repository/migrations_runner.go \
  'migration %s checksum mismatch' \
  'migration checksum mismatch remains a hard error'
require_text deploy/docker-compose.preview.yml \
  'GATEWAY_IMAGE_CONCURRENCY_ENABLED' \
  'image concurrency configuration remains in the preview compose contract'
require_text deploy/docker-compose.preview.yml \
  'GATEWAY_IMAGE_CONCURRENCY_MAX_CONCURRENT_REQUESTS' \
  'image concurrency limit remains separate from user concurrency'
require_text deploy/README.md \
  'Use an immutable commit tag for every' \
  'deployment documentation requires immutable release tags'
require_text deploy/README.md \
  'does not build on the server or remove PostgreSQL' \
  'deployment documentation preserves server data and prebuilt images'

if [[ -n "${merge_base:-}" ]]; then
  check_commit_ledger "$merge_base"
  check_migration_ledger "$merge_base"
fi

if [[ "$failures" -gt 0 ]]; then
  printf 'Contract check failed: %d failure(s), %d warning(s)\n' "$failures" "$warnings" >&2
  exit 1
fi

printf 'Contract check passed: 0 failure(s), %d warning(s)\n' "$warnings"
