"""Exercise the embedded Linux responder on a disposable loopback-only grant.
Verify the response status, media type and withdrawal without using a LAN.
"""
import contextlib
import io
import http.client
import importlib.util
import json
import queue
import socket
import sqlite3
import tempfile
import threading
import unittest
from pathlib import Path
from unittest.mock import patch

source = Path(__file__).resolve().parents[1] / "internal/adapters/discovery/linux.py"
spec = importlib.util.spec_from_file_location("linux_discovery", source)
discovery = importlib.util.module_from_spec(spec)
spec.loader.exec_module(discovery)


@unittest.skipUnless(hasattr(socket, "SO_BINDTODEVICE"), "Linux device binding required")
class BootstrapTest(unittest.TestCase):
    def test_scoped_responder_and_revocation(self):
        rule = "lo/loopback/local/127.0.0.1"
        ready = queue.Queue()
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        db = sqlite3.connect(Path(directory.name) / "node.sqlite", check_same_thread=False)
        self.addCleanup(db.close)
        db.execute("create table local_runtime(key text primary key, value text, rev integer)")
        db.execute("create table project_entries(kind text, body text)")
        self.assertEqual(discovery.grants(db), [])
        discovery.save(db, [rule])
        base = discovery.http.server.ThreadingHTTPServer

        class ObservedServer(base):
            def server_activate(self):
                super().server_activate()
                ready.put(self.server_port)

        with patch.object(discovery, "profiles", return_value={rule: ("lo", "127.0.0.1")}), \
             patch.object(discovery.http.server, "ThreadingHTTPServer", ObservedServer), contextlib.redirect_stdout(io.StringIO()) as signals:
            worker = threading.Thread(target=discovery.serve, args=(db, rule, True), daemon=True)
            worker.start()
            try:
                port = ready.get(timeout=4)
                for method, path, status in (("GET", "/v1", 401), ("GET", "/v1/services", 404),
                                            ("POST", "/v1", 405)):
                    client = http.client.HTTPConnection("127.0.0.1", port, timeout=2)
                    try:
                        client.request(method, path)
                        response = client.getresponse()
                        self.assertEqual(response.status, status)
                        self.assertEqual(response.getheader("Content-Type"), "application/json")
                        self.assertIn("error", json.load(response))
                    finally:
                        client.close()
            finally:
                discovery.save(db, [rule], False)
                worker.join(4)
            self.assertEqual(signals.getvalue(), "A")
            self.assertFalse(worker.is_alive())
            with self.assertRaises(ConnectionRefusedError):
                socket.create_connection(("127.0.0.1", port), timeout=2)


class ScheduleTest(unittest.TestCase):
    def test_request_cadence_and_owner_controls(self):
        db = sqlite3.connect(":memory:")
        self.addCleanup(db.close)
        db.execute("create table local_runtime(key text primary key,value text,rev integer)")
        db.execute("create table project_entries(kind text, body text)")
        self.assertTrue(discovery.claim_update(db, now=100))
        self.assertFalse(discovery.claim_update(db, now=3699))
        self.assertTrue(discovery.claim_update(db, now=3700))
        discovery.update_command(db, "result", ["current"])
        self.assertFalse(discovery.claim_update(db, now=90099))
        self.assertTrue(discovery.claim_update(db, now=90100))
        discovery.update_command(db, "disable", [])
        self.assertFalse(discovery.claim_update(db, now=200000))
        discovery.update_command(db, "check", [])
        with db:
            db.execute("""insert into project_entries values('run','{"State":"running"}')""")
        self.assertFalse(discovery.claim_update(db, manual=True, now=90101))
        with db:
            db.execute("delete from project_entries")
        self.assertTrue(discovery.claim_update(db, manual=True, now=90102))
        self.assertFalse(discovery.claim_update(db, manual=True, now=90103))
        discovery.update_command(db, "enable", [])
        self.assertFalse(discovery.claim_update(db, now=90104))
