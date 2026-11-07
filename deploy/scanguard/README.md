# Aivoza scan protection — server 23

Installed 2026-11-07 ~20:07–20:12 Asia/Shanghai. This is an ingress-only
configuration release; application revision 9e7f86bf6879 remains unchanged.

## Enforcement

- Cloudflare provider zone **nodx.net** (aivoza.com CNAME uses saas.nodx.net):
  custom rule `9ad5fd4791b34f3182ab526edaa24913`, “Aivoza 自动化漏洞扫描拦截”.
  Restricted to aivoza.com, cf.nodx.net and us.aivoza.com. The last hostname
  currently bypasses Cloudflare, so its protection is enforced at the origin.
- Caddy 2.10.2 imports `/config/scanguard/guard.caddy` in the two application
  sites. Explicit `route` blocks run scan and ban responses before existing
  handles/proxy directives. Known probe paths receive a 9-byte 404.
- `.well-known/`, `/pay/`, `/api/v1/payment/` remain outside probe detection.
  Normal UI routes, model API, login, SSE and WebSocket behavior are preserved.
- Caddy trusts only official Cloudflare IPv4/IPv6 ranges fetched from
  `https://api.cloudflare.com/client/v4/ips`. Only trusted peers can supply
  `CF-Connecting-IP`; direct clients use their network address. CIDRs are an
  installation snapshot and must be refreshed/revalidated when Cloudflare
  updates its ranges. They are never inferred from incoming headers.
- `flowai-scanguard.timer` runs every 10 seconds. Ten logged probe rejections
  from the same public client IP within 60 seconds produce a 30-minute ban.
  Bans affect new requests to the application sites; health and ACME paths
  remain available. Existing streams are not explicitly terminated.
- Excludes private/reserved IPs and Cloudflare proxy ranges. Maximum 1000 active
  IPs; does not modify host packet firewall, application accounts or database.
  Timer processes at most two recent files, each limited to its last 8 MiB.
- Ban changes validate Caddy config before reload, restoring the include on
  failure. Expiry depends on the timer running; investigate failed services
  promptly rather than assuming entries expire inside Caddy itself.

## Files and observability

Host base: `/root/flowai-data/caddy/config/scanguard` (existing Caddy `/config`
mount). `scan.jsonl` stores probe-only logs; authorization/cookies, all headers,
query URI, TLS details and response headers are removed. Path, source, time,
status and duration remain for investigation. Mode 0600; rotation at 10 MiB,
5 archives, maximum age 7 days. `bans.json` contains expiry timestamps;
`bans.caddy` is generated. Journal events expose hashed IPs only.

```sh
systemctl list-timers flowai-scanguard.timer --no-pager
systemctl show flowai-scanguard.service -p Result -p ExecMainStatus
journalctl -u flowai-scanguard.service --since today --no-pager
```

The existing Cloudflare login rate-limit rule consumes the free zone's sole
rate-limit slot and was preserved. The new custom rule blocks scans immediately;
source temporary bans provide the requested 10/60s threshold without buying a
plan or replacing login protection.

## Validation

- Five local unit tests: threshold/window, excluded IPs/forged header fields,
  normal errors, configuration injection, persistent expiry/idempotency/failure.
- Actual isolated Caddy listener inside the existing container: banned IP 404,
  other IP 200, health/ACME exceptions 200. Test listener stopped afterward.
- Production direct-origin spoofed CF-Connecting-IP ignored.
- Public aivoza.com `/.env`, `/.git/config`: Cloudflare 403; us.aivoza.com `/.env`:
  origin 404/9 bytes. Both public `/health`: 200.
- Direct origin main `/help` and both public-settings APIs: 200; empty payment
  webhook request reaches existing application rejection response (no payment
  or order creation). Direct us `/help` stays its previous 404.

## Reproduction and rollback

`install.py` is a guarded installer for the observed unmodified server config;
review before reuse. It refuses already installed/global-options configs.
After copying the directory to the server, run the installer, copy the two
systemd units to `/etc/systemd/system`, daemon-reload and enable the timer.
Do not replace the production file with the repository's old Caddy template.

Original Caddy backup:
`/root/flowai/backups/scanguard-20261107-200720/Caddyfile.before`.

To roll back this ingress release: stop/disable `flowai-scanguard.timer`, copy
that backup **into the existing bind-mounted file inode**, validate and reload
Caddy. If application deployment has occurred since this backup, remove only
the scanguard imports/routes/trusted-proxy additions while preserving the current
app upstream. Restore no old app image or upstream blindly. Disable the specific
Cloudflare custom rule separately; do not change the login rate-limit rule.
Retain logs and backups. Routine manual unban should update `bans.json` and
`bans.caddy` under the guard lock, validate, reload and restart the timer.
