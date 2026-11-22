"""Exercise the embedded Linux responder on a disposable loopback-only grant.
Verify the response status, media type and withdrawal without using a LAN.
"""
import http.client
import importlib.util
import json
import queue
import socket
import sqlite3
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
        db = sqlite3.connect(":memory:", check_same_thread=False)
        self.addCleanup(db.close)
        db.execute("create table local_runtime(key text primary key, value text, rev integer)")
        self.assertEqual(discovery.grants(db), [])
        discovery.save(db, [rule])
        base = discovery.http.server.HTTPServer

        class ObservedServer(base):
            def server_activate(self):
                super().server_activate()
                ready.put(self.server_port)

        with patch.object(discovery, "profiles", return_value={rule: ("lo", "127.0.0.1")}), \
             patch.object(discovery.http.server, "HTTPServer", ObservedServer):
            worker = threading.Thread(target=discovery.serve, args=(db, rule), daemon=True)
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
            self.assertFalse(worker.is_alive())
            with self.assertRaises(ConnectionRefusedError):
                socket.create_connection(("127.0.0.1", port), timeout=2)
