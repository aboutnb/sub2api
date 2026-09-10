import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('bluegreen', Path(__file__).parents[1] / 'deploy-aivoza-bluegreen.py')
bluegreen = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bluegreen)

CONFIG = '''aivoza.com, cf.nodx.net {
  handle /pay/* {
    reverse_proxy gm:8000
  }
  reverse_proxy sub2api:8080
}
us.aivoza.com {
  handle @api {
    reverse_proxy sub2api:8080
  }
}
'''


class BlueGreenTests(unittest.TestCase):
    def test_preserves_payment_route_and_newlines(self):
        new = bluegreen.replace_upstreams(CONFIG, 'sub2api', 'aivoza-tested')
        self.assertIn('reverse_proxy gm:8000\n', new)
        self.assertEqual(new.count('reverse_proxy aivoza-tested:8080\n'), 2)
        self.assertEqual(bluegreen.replace_upstreams(new, 'aivoza-tested', 'sub2api'), CONFIG)

    def test_rejects_unexpected_route_before_mutation(self):
        with self.assertRaises(RuntimeError):
            bluegreen.replace_upstreams(CONFIG.replace('gm:8000', 'other:8080'), 'sub2api', 'new')

    def test_rejects_mixed_routes(self):
        with self.assertRaises(RuntimeError):
            bluegreen.replace_upstreams(CONFIG.replace('sub2api:8080', 'changed:8080', 1), 'sub2api', 'new')

    def test_validation_failure_preserves_active_configuration(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'Caddyfile'
            path.write_text(CONFIG)
            def execute(*args, **kwargs):
                if 'validate' in args:
                    raise RuntimeError('invalid config')
                return ''
            with patch.object(bluegreen, 'run', side_effect=execute):
                with self.assertRaises(RuntimeError):
                    bluegreen.load_config('invalid', path)
            self.assertEqual(path.read_text(), CONFIG)

    def test_private_state_is_not_world_readable(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'state.json'
            bluegreen.private_write(path, '{}')
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)


if __name__ == '__main__':
    unittest.main()
