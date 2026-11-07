#!/usr/bin/env python3
"""Bounded scan-only enforcement. No database access or packet firewall changes."""
import collections
import fcntl
import hashlib
import ipaddress
import json
import os
from pathlib import Path
import subprocess
import time

BASE = Path('/root/flowai-data/caddy/config/scanguard')
WINDOW = 60
THRESHOLD = 10
TTL = 1800
MAX_BANS = 1000


def candidates(records, now, trusted):
    hits = collections.Counter()
    for d in records:
        try:
            if d.get('security_event') != 'scan_probe' or d.get('status') != 404:
                continue
            if not now - WINDOW <= float(d['ts']) <= now + 2:
                continue
            ip = ipaddress.ip_address(d['request']['client_ip'])
            if not ip.is_global or any(ip in net for net in trusted):
                continue
            hits[str(ip)] += 1
        except (ValueError, KeyError, TypeError):
            continue
    return {ip for ip, n in hits.items() if n >= THRESHOLD}


def render(bans):
    if not bans:
        return '# No active scan bans. Managed by ban.py.\n'
    ips = ' '.join(sorted(str(ipaddress.ip_address(ip)) for ip in bans))
    return ('@scan_banned {\n    client_ip ' + ips + '\n'
            '    not path /health /.well-known/*\n}\n'
            'respond @scan_banned "Not Found" 404\n')


def records():
    # Each pass is capped; normal requests never enter these files.
    for path in sorted(BASE.glob('scan*.jsonl'), key=lambda p: p.stat().st_mtime, reverse=True)[:2]:
        if path.stat().st_mtime < time.time() - WINDOW:
            continue
        with path.open('rb') as stream:
            size = path.stat().st_size
            if size > 8 * 1024 * 1024:
                stream.seek(size - 8 * 1024 * 1024)
                stream.readline()
            for raw in stream:
                try:
                    yield json.loads(raw)
                except (ValueError, UnicodeDecodeError):
                    continue


def atomic(path, text):
    tmp = path.with_suffix(path.suffix + '.tmp')
    tmp.write_text(text)
    os.chmod(tmp, 0o600)
    tmp.replace(path)


def run():
    os.umask(0o077)
    with (BASE / 'lock').open('w') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        now = time.time()
        state = BASE / 'bans.json'
        old = json.loads(state.read_text()) if state.exists() else {}
        bans = {ip: expiry for ip, expiry in old.items() if expiry > now}
        trusted = [ipaddress.ip_network(n) for n in (BASE / 'cloudflare-cidrs.txt').read_text().split()]
        for ip in sorted(candidates(records(), now, trusted)):
            if ip not in bans and len(bans) < MAX_BANS:
                bans[ip] = now + TTL
        target = BASE / 'bans.caddy'
        before = target.read_text()
        after = render(bans)
        if after != before:
            atomic(target, after)
            try:
                for action in ('validate', 'reload'):
                    result = subprocess.run(['docker', 'exec', 'flowai-caddy', 'caddy', action,
                        '--config', '/etc/caddy/Caddyfile', '--adapter', 'caddyfile'],
                        capture_output=True, timeout=20)
                    if result.returncode:
                        raise RuntimeError('Caddy ' + action + ' failed')
            except Exception:
                atomic(target, before)
                raise
            print(json.dumps({'event': 'scan_bans_updated', 'active': len(bans),
                'added_hashes': [hashlib.sha256(ip.encode()).hexdigest()[:12] for ip in bans if ip not in old],
                'expired': len(set(old) - set(bans))}))
        atomic(state, json.dumps(bans))


if __name__ == '__main__':
    run()
