# Aivoza repository workflow

- `main` is the default development, upstream integration, and release branch.
- Reuse the existing `main` checkout/worktree. Do not create a branch or worktree
  for each fix, modification, upstream version, or deployment. Create a topic
  branch only when the user explicitly requests isolated work.
- Before deleting branches, inspect their responsibilities, unique commits,
  open PRs, and local worktrees. Present the exact deletion list and obtain user
  confirmation. Existing confirmation applies only to the agreed list.
- Preserve uncommitted work and untracked data. If the current checkout is dirty
  or historical, inspect `git worktree list` and reuse the existing clean `main`
  worktree instead of resetting files or creating another release branch.
- Read `docs/FLOWAI_CHANGELOG.md`, `docs/FLOWAI_BRANCH_CONTRACT.md`, and
  `docs/FLOWAI_RELEASE_CHECKLIST.md` before merging upstream or deploying.
- Fetch `origin/main` and `upstream/main`, review the incoming changes, then merge
  the reviewed upstream SHA into `main`. Never replace the custom tree with
  upstream files. Update `.github/aivoza-upstream-ref` and the change ledger.
- Preserve priority 1 as highest, negative-concurrency denial, i18n, GM payments,
  recharge bonuses, check-in, email deduplication, and migration checksums.
- Run the required checks, push `main`, and deploy only the GitHub-built
  `ghcr.io/aboutnb/aivoza-sub2api` digest whose CI and security checks succeeded.
  Rerun a failed workflow after diagnosis; do not create empty retry commits.
- Deploy server 23 through Termius with the existing blue-green prepare/promote
  script. Keep serving traffic during candidate startup and route switching.
- Retain only the current application and one previous rollback version.
  After verification and connection draining, remove older app containers and
  their unreferenced images. Never prune volumes, data backups, or unrelated
  services. Preserve the current rollback state, configuration, and image.
