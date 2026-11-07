import importlib.util
from pathlib import Path
import ipaddress
import unittest
import tempfile
import json
from unittest.mock import patch
from types import SimpleNamespace
spec = importlib.util.spec_from_file_location('ban', Path(__file__).with_name('ban.py'))
ban = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ban)

class Tests(unittest.TestCase):
    def record(self, ip='8.8.8.8', ts=1000, **kw):
        return dict(security_event='scan_probe', status=404, ts=ts, request={'client_ip': ip}, **kw)
    def test_threshold_and_expiry(self):
        self.assertEqual(ban.candidates([self.record()]*9, 1000, []), set())
        self.assertEqual(ban.candidates([self.record()]*10, 1000, []), {'8.8.8.8'})
        self.assertEqual(ban.candidates([self.record(ts=939)]*10, 1000, []), set())
    def test_untrusted_fields_and_proxy_exclusion(self):
        for ip in ['127.0.0.1','10.0.0.1','invalid','1.1.1.1']:
            self.assertEqual(ban.candidates([self.record(ip)]*10,1000,[ipaddress.ip_network('1.1.1.0/24')]),set())
        self.assertEqual(ban.candidates([{'security_event':'scan_probe','ts':1000,'status':404,'request':{'remote_ip':'8.8.8.8','headers':{'CF-Connecting-IP':['8.8.8.8']}}}]*10,1000,[]),set())
    def test_normal_errors_not_counted(self):
        r=self.record();r.pop('security_event')
        self.assertEqual(ban.candidates([r]*10,1000,[]),set())
    def test_config_injection_rejected(self):
        with self.assertRaises(ValueError):ban.render({'1.1.1.1\nrespond 200':2000})
        self.assertIn('not path /health /.well-known/*',ban.render({'8.8.8.8':2000}))
    def test_persist_expire_and_failed_reload(self):
        with tempfile.TemporaryDirectory() as td:
            base=Path(td)
            (base/'bans.caddy').write_text(ban.render({}))
            (base/'cloudflare-cidrs.txt').write_text('1.1.1.0/24')
            with patch.object(ban,'BASE',base), patch.object(ban.time,'time',return_value=1000), patch.object(ban,'records',return_value=[self.record()]*10), patch.object(ban.subprocess,'run',return_value=SimpleNamespace(returncode=0)) as run:
                ban.run()
                self.assertEqual(json.loads((base/'bans.json').read_text()),{'8.8.8.8':2800})
                self.assertEqual(run.call_count,2)
                run.reset_mock();ban.run();run.assert_not_called()
            with patch.object(ban,'BASE',base), patch.object(ban.time,'time',return_value=2801), patch.object(ban,'records',return_value=[]), patch.object(ban.subprocess,'run',return_value=SimpleNamespace(returncode=0)):
                ban.run()
                self.assertEqual(json.loads((base/'bans.json').read_text()),{})
                self.assertNotIn('client_ip', (base/'bans.caddy').read_text())
            with patch.object(ban,'BASE',base), patch.object(ban.time,'time',return_value=1000), patch.object(ban,'records',return_value=[self.record()]*10), patch.object(ban.subprocess,'run',return_value=SimpleNamespace(returncode=1)):
                with self.assertRaises(RuntimeError):ban.run()
                self.assertEqual(json.loads((base/'bans.json').read_text()),{})
                self.assertNotIn('client_ip', (base/'bans.caddy').read_text())
if __name__ == '__main__':unittest.main()
