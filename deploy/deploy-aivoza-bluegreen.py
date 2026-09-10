#!/usr/bin/env python3
"""Prepare and promote a digest-pinned app without stopping the serving container.

Run on the production host. Preparation never changes the active Caddy config.
Promotion changes only the two app upstreams, keeps the old container running,
and records rollback state. Secrets are kept in mode-0600 local state files.
"""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import time
import urllib.request


def run(*args, capture=True):
    result = subprocess.run(args, check=True, text=True, stdout=subprocess.PIPE if capture else None)
    return result.stdout.strip() if capture else ''


def inspect(name):
    return json.loads(run('docker', 'inspect', name))[0]


def private_write(path, text):
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, 'w') as stream:
        stream.write(text)
        stream.flush()
        os.fsync(stream.fileno())


def healthy(port):
    with urllib.request.urlopen(f'http://127.0.0.1:{port}/health', timeout=5) as response:
        return response.status == 200 and json.load(response).get('status') == 'ok'


def app_upstreams(config):
    # Restrict replacement to complete reverse_proxy directives; never touch GM.
    return re.findall(r'(?m)^\s*reverse_proxy\s+([^\s{}]+):8080[ \t]*$', config)


def replace_upstreams(config, old, new):
    if app_upstreams(config) != [old, old]:
        raise RuntimeError('Expected exactly two identical app upstreams; review current Caddyfile')
    return re.sub(r'(?m)^(\s*reverse_proxy\s+)' + re.escape(old) + r':8080[ \t]*$',
                  lambda match: match[1] + new + ':8080', config)


def dependency_state():
    names = ['flowai-caddy', 'flowai-postgres', 'flowai-redis', 'flowai-gm', 'flowai-mihomo']
    return {name: inspect(name)['State']['StartedAt'] for name in names}


def wait_healthy(port, seconds=120):
    for _ in range(seconds):
        try:
            if healthy(port):
                return
        except (OSError, ValueError):
            pass
        time.sleep(1)
    raise RuntimeError('Candidate did not become healthy; active app is unchanged')


def prepare(args):
    if not re.fullmatch(r'ghcr\.io/aboutnb/aivoza-sub2api@sha256:[0-9a-f]{64}', args.image or ''):
        raise RuntimeError('Require ghcr.io/aboutnb/aivoza-sub2api@sha256:<digest>')
    if args.state.exists():
        raise RuntimeError('State file already exists; inspect it rather than overwriting rollback data')
    old = inspect('flowai-app')
    if old['State'].get('Health', {}).get('Status') != 'healthy':
        raise RuntimeError('Current flowai-app must be healthy')
    if list(old['NetworkSettings']['Networks']) != ['flowai-network']:
        raise RuntimeError('Unexpected app networks; manual review required')
    caddy = Path('/root/flowai/deploy/Caddyfile.flowai')
    before = caddy.read_text()
    upstreams = app_upstreams(before)
    if len(upstreams) != 2 or len(set(upstreams)) != 1:
        raise RuntimeError('Unexpected Caddy app upstreams')
    run('docker', 'pull', args.image, capture=False)
    image = inspect(args.image)
    revision = image['Config'].get('Labels', {}).get('org.opencontainers.image.revision', '')
    if not re.fullmatch(r'[a-f0-9]{40}', revision):
        raise RuntimeError('Image must contain a full source revision label')
    suffix = revision[:12]
    name = f'flowai-app-next-{suffix}'
    alias = f'aivoza-{suffix}'
    after = replace_upstreams(before, upstreams[0], alias)
    env = dict(item.split('=', 1) for item in old['Config']['Env'] if '=' in item)
    env['AIVOZA_ROLLING_DEPLOY'] = 'true'
    env['SERVER_PORT'] = '8080'
    args.state.parent.mkdir(parents=True, exist_ok=True)
    state = {'old_id': old['Id'], 'old_image': old['Config']['Image'], 'new_image': args.image,
             'revision': revision, 'candidate': name, 'alias': alias, 'port': args.port,
             'caddy_path': str(caddy), 'before': before, 'after': after,
             'dependencies': dependency_state(), 'phase': 'preparing',
             'env_before': Path('/root/flowai/deploy/.env').read_text()}
    private_write(args.state, json.dumps(state, indent=2))
    command = ['docker', 'create', '--name', name, '--restart', 'unless-stopped',
               '--network', 'flowai-network', '--network-alias', alias,
               '--publish', f'127.0.0.1:{args.port}:8080']
    for mount in old['Mounts']:
        if mount['Type'] != 'bind':
            raise RuntimeError('Unexpected non-bind app mount; manual review required')
        command += ['--mount', f'type=bind,src={mount["Source"]},dst={mount["Destination"]}' +
                    (',readonly' if not mount['RW'] else '')]
    host = old['HostConfig']
    for extra in host.get('ExtraHosts') or []:
        command += ['--add-host', extra]
    for value, flag in [(host.get('Memory'), '--memory'), (host.get('MemorySwap'), '--memory-swap'),
                        (host.get('PidsLimit'), '--pids-limit')]:
        if value:
            command += [flag, str(value)]
    if host.get('NanoCpus'):
        command += ['--cpus', str(host['NanoCpus'] / 1_000_000_000)]
    if host.get('ReadonlyRootfs'):
        command += ['--read-only']
    for option in host.get('SecurityOpt') or []:
        command += ['--security-opt', option]
    for capability in host.get('CapDrop') or []:
        command += ['--cap-drop', capability]
    with tempfile.NamedTemporaryFile(mode='w', prefix='aivoza-env-', delete=True) as stream:
        os.chmod(stream.name, 0o600)
        stream.write('\n'.join(f'{key}={value}' for key, value in env.items()) + '\n')
        stream.flush()
        command += ['--env-file', stream.name, args.image]
        run(*command)
    run('docker', 'start', name)
    state['new_id'] = inspect(name)['Id']
    private_write(args.state, json.dumps(state, indent=2))
    wait_healthy(args.port)
    state['phase'] = 'ready'
    private_write(args.state, json.dumps(state, indent=2))
    print(json.dumps({'phase': 'ready', 'candidate': name, 'port': args.port, 'revision': revision}))


def load_config(config, path):
    # Validate a separate file before changing the bind-mounted inode in place.
    with tempfile.NamedTemporaryFile(mode='w', suffix='.Caddyfile') as stream:
        stream.write(config)
        stream.flush()
        run('docker', 'cp', stream.name, 'flowai-caddy:/tmp/aivoza-candidate.Caddyfile')
    run('docker', 'exec', 'flowai-caddy', 'caddy', 'validate', '--config',
        '/tmp/aivoza-candidate.Caddyfile', '--adapter', 'caddyfile')
    private_write(path, config)
    run('docker', 'exec', 'flowai-caddy', 'caddy', 'reload', '--config',
        '/etc/caddy/Caddyfile', '--adapter', 'caddyfile')


def promote(args):
    state = json.loads(args.state.read_text())
    if state['phase'] != 'ready':
        raise RuntimeError('Candidate must be ready; never rerun an incomplete promotion blindly')
    if inspect('flowai-app')['Id'] != state['old_id']:
        raise RuntimeError('Active app changed since preparation')
    if inspect(state['candidate'])['Id'] != state['new_id']:
        raise RuntimeError('Candidate changed since preparation')
    if dependency_state() != state['dependencies']:
        raise RuntimeError('Dependencies changed since preparation')
    path = Path(state['caddy_path'])
    if path.read_text() != state['before']:
        raise RuntimeError('Caddyfile changed since preparation')
    wait_healthy(state['port'], 10)
    try:
        load_config(state['after'], path)
    except Exception:
        load_config(state['before'], path)
        raise
    state['phase'] = 'routed'
    private_write(args.state, json.dumps(state, indent=2))
    rollback_name = 'flowai-app-rollback-' + state['old_id'][:12]
    run('docker', 'rename', state['old_id'], rollback_name)
    run('docker', 'rename', state['new_id'], 'flowai-app')
    # Persist the new immutable image/port for later administrative operations.
    env_path = Path('/root/flowai/deploy/.env')
    env = env_path.read_text()
    for key, value in [('SUB2API_IMAGE', state['new_image']), ('SERVER_PORT', str(state['port'])),
                       ('AIVOZA_ROLLING_DEPLOY', 'true')]:
        line = f'{key}={value}'
        if re.search(rf'(?m)^{key}=', env):
            env = re.sub(rf'(?m)^{key}=.*$', lambda _: line, env)
        else:
            env += '\n' + line + '\n'
    private_write(env_path, env)
    state.update(phase='promoted', rollback_container=rollback_name)
    private_write(args.state, json.dumps(state, indent=2))
    print(json.dumps({'phase': state['phase'], 'image': state['new_image'],
                      'rollback_container': rollback_name, 'old_container_still_running': True}))


def rollback(args):
    state = json.loads(args.state.read_text())
    if state['phase'] not in ('routed', 'promoted'):
        raise RuntimeError('No promoted route to roll back')
    old = inspect(state['old_id'])
    if old['State'].get('Health', {}).get('Status') != 'healthy':
        raise RuntimeError('Rollback container must still be healthy')
    path = Path(state['caddy_path'])
    if path.read_text() != state['after']:
        raise RuntimeError('Caddyfile changed after promotion; manual review required')
    load_config(state['before'], path)
    if inspect('flowai-app')['Id'] == state['new_id']:
        run('docker', 'rename', state['new_id'], state['candidate'])
        run('docker', 'rename', state['old_id'], 'flowai-app')
    private_write(Path('/root/flowai/deploy/.env'), state['env_before'])
    state['phase'] = 'rolled_back'
    private_write(args.state, json.dumps(state, indent=2))
    print('Rolled back route; both app containers remain running')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action', choices=['prepare', 'promote', 'rollback'])
    parser.add_argument('--image')
    parser.add_argument('--port', type=int, default=3001)
    parser.add_argument('--state', type=Path, required=True)
    args = parser.parse_args()
    globals()[args.action](args)


if __name__ == '__main__':
    main()
