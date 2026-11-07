#!/usr/bin/env python3
"""Install the reviewed guard on the existing server 23 Caddy mount."""
import datetime
import ipaddress
import json
import os
import re
from pathlib import Path
import shutil
import subprocess
import urllib.request

BASE = Path('/root/flowai-data/caddy/config/scanguard')
CONFIG = Path('/root/flowai/deploy/Caddyfile.flowai')
SOURCE = Path(__file__).parent
os.umask(0o077)
BASE.mkdir(exist_ok=True, parents=True)
before = CONFIG.read_text()
assert 'scanguard' not in before, 'Already installed; review instead of overwriting'
assert not before.lstrip().startswith('{'), 'Existing global options require manual merge'
response = json.load(urllib.request.urlopen('https://api.cloudflare.com/client/v4/ips', timeout=20))
assert response['success']
cidrs = [str(ipaddress.ip_network(n)) for n in response['result']['ipv4_cidrs'] + response['result']['ipv6_cidrs']]
assert len(cidrs) >= 20
(BASE/'cloudflare-cidrs.txt').write_text('\n'.join(cidrs)+'\n')
for name in ('guard.caddy','ban.py','test_ban.py'):
    shutil.copy2(SOURCE/name,BASE/name)
(BASE/'bans.caddy').write_text('# No active scan bans. Managed by ban.py.\n')
(BASE/'bans.json').write_text('{}')
global_config = '{\n    servers {\n        trusted_proxies static '+ ' '.join(cidrs) +'\n        trusted_proxies_strict\n        client_ip_headers CF-Connecting-IP\n    }\n}\n\n'
after = before
for site in ('aivoza.com, cf.nodx.net', 'us.aivoza.com'):
    pattern = re.escape(site) + r' \{\n(.*?)\n\}'
    match = re.search(pattern, after, re.S)
    assert match
    replacement = (site + ' {\n    import /config/scanguard/guard.caddy\n    route {\n'
                   '        respond @scan_probe "Not Found" 404\n'
                   '        import /config/scanguard/bans.caddy\n' + match.group(1) + '\n    }\n}')
    after = after[:match.start()] + replacement + after[match.end():]
after = global_config + after
candidate = BASE/'candidate.Caddyfile'
candidate.write_text(after)
subprocess.run(['docker','exec','flowai-caddy','caddy','validate','--config','/config/scanguard/candidate.Caddyfile','--adapter','caddyfile'],check=True)
backup = Path('/root/flowai/backups/scanguard-'+datetime.datetime.now().strftime('%Y%m%d-%H%M%S'))
backup.mkdir(parents=True)
shutil.copy2(CONFIG,backup/'Caddyfile.before')
# Preserve bind-mounted inode.
CONFIG.write_text(after)
try:
    subprocess.run(['docker','exec','flowai-caddy','caddy','reload','--config','/etc/caddy/Caddyfile','--adapter','caddyfile'],check=True)
except Exception:
    CONFIG.write_text(before)
    raise
(BASE/'backup-path.txt').write_text(str(backup))
print('STATIC_GUARD_INSTALLED',backup)
