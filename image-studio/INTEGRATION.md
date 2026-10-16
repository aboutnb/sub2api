# Built-in Image Studio

## Architecture

- Upstream: aboutnb/gpt_image_playground at `da4fda85b59ecacc51d6a1e2ef680e3abb9e29b8` (see UPSTREAM.md). MIT license remains in LICENSE and is shipped as LICENSE.txt.
- Vue owns `/image-studio`, site authentication and the fixed sidebar item. React is built separately under `/image-studio-app/`; it does not register a service worker.
- `shared/image-studio.ts` defines the versioned, same-origin postMessage protocol. Only bootstrap, fixed Images operations, prompt optimization and request cancellation are supported. No credentials, arbitrary URLs or headers are passed into React. The parent checks account continuity before sending requests and delivering responses, including account changes in another tab.
- The iframe is trusted same-origin application code, not a sandbox boundary against compromised scripts. JWT never crosses the bridge; normal site XSS protections still matter.
- JWT endpoints: `/api/v1/image-studio/bootstrap` and `/api/v1/image-studio/groups/:group_id/images/{generations,edits,generations/async,edits/async,tasks/:task_id}`.
- New work rechecks feature status, group and model permission, then dispatches through an in-process instance of the existing gateway route chain. It does not call the public domain or bypass gateway middleware.
- Internal keys use purpose `image_studio` and reserved prefix `studio-internal-`. Per-group keys are hidden from ordinary user CRUD and all internal credentials are rejected at public credential entry points. The automatically maintained multi-channel image route is visible in the key list so users can edit its channels; ordinary key creation cannot create this route type. Existing usage attribution is retained with the key name AI 绘图.
- IndexedDB and persisted state are partitioned by server-provided user namespace. No original unpartitioned history is imported. Pending tasks retain their original group/profile; logout stops browser work, not upstream generation.
- Async is selected only when bootstrap reports it enabled. Failed async submissions never retry as synchronous requests.

## Drawing and Prompt Optimization

- Image and optimizer controls each select a channel first, then a model from that channel's current model list. The image channel displays subscription/balance attribution and its multiplier. An explicit optimizer channel prefers `gpt-5.6-sol` only when that model is available there.
- The automatic optimizer uses the saved image smart route, preferring OpenAI. Initial route creation locks the internal key and rechecks stored routes in one transaction; concurrent initialization must not overwrite user edits. Removed or unavailable saved channels never fall back to unrelated public channels.
- Prompt optimization has an independent per-user switch, enabled by default. Turning it off keeps the selected optimizer channel/model for later and skips the optimization request; turning it back on restores that selection. Explicit selections that become unavailable require reselection. Optimization failures stop the task with the actual error; they do not silently proceed to paid image generation.
- Image capabilities, reference count/type/size and mask compatibility are checked before optimization. Task details retain the original prompt, optimized prompt and reported optimizer model, including successful optimization that leaves the text unchanged.
- The submission lock covers local image processing and task persistence. Stopping or deleting an optimizing task prevents the subsequent image submission. Stopping a submitted image task only stops browser waiting; upstream work may continue and incur charges. Retrying a stopped asynchronous task resumes its original task ID instead of submitting another generation.
- Retrying preserves the original image channel/model and any already completed optimization. Local history, drafts and optimizer choices remain isolated by user. Startup failures offer a reload action.

## Build and Local Development

Build Vue before React: Vue clears the shared output directory.

```sh
cd frontend
pnpm install --frozen-lockfile
pnpm build
cd ../image-studio
npm ci --ignore-scripts
npm run build
cd ../backend
CGO_ENABLED=0 go build -tags embed ./cmd/server
```

The root Dockerfile, deploy/Dockerfile, Makefile and release workflow include both builds. No additional production service is needed. For development run the normal backend, Vue on port 5173 and React on port 5174; Vue proxies `/image-studio-app/` to React. Do not use the standalone React URL as the user entry point.

## Rollout and Rollback

1. Back up the production database before deploying migration 239. It adds a default-standard key purpose column and a partial unique index for active internal key ownership.
2. Deploy with AI Image Studio disabled (the default when no setting exists). The independent switch is in admin settings, backed by `/api/v1/admin/image-studio/settings`.
3. Verify authenticated bootstrap for ordinary, subscription and composite groups; aliases, allowlists and channel pricing restrictions must match actual gateway eligibility.
4. Verify `/image-studio-app/` serves the React index and missing assets return 404. Its framing policy permits same-origin framing; other site pages retain existing policy. Check reverse-proxy CSP/X-Frame-Options as well.
5. Object storage must allow browser GET from the site origin. Check image expiry, editing and download CORS. No remote-image proxy is added.
6. Before enabling for users, authorize a small real generation/edit batch and reconcile usage, wallet/subscription deductions, group/user multipliers and failed requests. Repeated task reads must not charge again.
7. Roll back by disabling the switch first. Submitted tasks remain queryable; do not delete internal keys, usage records, history or the additive migration.

## Verification Status

Automated tests and browser checks use mock upstream responses only. Covered surfaces include bridge origin/source/operation checks, credential isolation, upload validation, cancellation, model mappings/composite routes, public internal-key rejection, async ownership, static routing and framing policy. Browser smoke checks cover text generation, reference images, canvas masks, gallery restore, download and theme/language propagation.

On 2026-10-14, the route repository integration tests ran against isolated PostgreSQL/Redis containers and passed: 12 concurrent initializers, preservation of a committed user edit and stale cache, and transactional rollback on a candidate write failure. The broader internal-key integration suite is separate. Real paid generation, end-to-end billing reconciliation, production reverse-proxy headers and object-storage CORS remain deployment acceptance gates. Primary drawing controls are localized; some inherited secondary collection/help/error text is still Chinese.

Upstream development dependencies retain audit findings. The production dependency audit reports moderate findings for dompurify and mermaid in the pinned upstream dependency tree; those Agent-related modules are not imported by the embedded StudioApp entry. Review upstream dependency upgrades separately from the pinned vendor import.
