"""Linux bootstrap hosting and multicast browsing through NetworkManager and Avahi.
The owning Apparat process starts this embedded helper with the node database path.
"""
import contextlib
import hashlib
import http.server
import json
import os
import re
import select
import shutil
import socket
import sqlite3
import subprocess
import sys
import time
import threading

TYPE = "_apparat._tcp"
KEY = "bootstrap_grants_v1"


def command(*args):
    return subprocess.check_output(args, text=True, timeout=2).strip()


def profiles():
    result = {"lo/loopback/local/127.0.0.1": ("lo", "127.0.0.1")}
    if shutil.which("tailscale") and os.path.isdir("/sys/class/net/tailscale0"):
        try:
            status = json.loads(command("tailscale", "status", "--json"))
            tailnet = status["CurrentTailnet"]["MagicDNSSuffix"]
            address = next(a for a in status["Self"]["TailscaleIPs"] if ":" not in a)
            if tailnet and status["BackendState"] == "Running":
                key = "/".join(("tailscale0", tailnet, status["Self"]["ID"], address))
                result[key] = ("tailscale0", address)
        except (KeyError, TypeError, StopIteration, OSError, ValueError, subprocess.SubprocessError):
            pass
    routes = json.loads(command("ip", "-j", "-4", "route", "show", "default"))
    for route in routes:
        dev = route["dev"]
        gw = route.get("gateway")
        if not gw:
            continue
        neighbors = json.loads(command("ip", "-j", "neigh", "show", "to", gw, "dev", dev))
        if not neighbors or not neighbors[0].get("lladdr"):
            continue
        uuid = command("nmcli", "-g", "GENERAL.CON-UUID", "device", "show", dev)
        if not uuid or uuid == "--":
            continue
        pairs = json.loads(command("ip", "-j", "-4", "addr", "show", "dev", dev))
        for addr in pairs[0]["addr_info"]:
            if addr["scope"] == "global":
                key = "/".join((dev, uuid, neighbors[0]["lladdr"], addr["local"]))
                result[key] = (dev, addr["local"])
    return result


def saved(db):
    row = db.execute("select value from local_runtime where key=?", (KEY,)).fetchone()
    return json.loads(row[0]) if row else []


def enabled(db):
    row = db.execute("select value from local_runtime where key='bootstrap_enabled'").fetchone()
    return row[0] == "1" if row else bool(saved(db))


def grants(db):
    return saved(db) if enabled(db) else []


def save(db, rules, on=True):
    with db:
        db.execute("insert or replace into local_runtime values(?,?,0)", (KEY, json.dumps(rules)))
        db.execute("insert or replace into local_runtime values(?,?,0)", ("bootstrap_enabled", "1" if on and rules else "0"))


def avahi():
    import dbus
    bus = dbus.SystemBus()
    return bus, dbus.Interface(bus.get_object("org.freedesktop.Avahi", "/"), "org.freedesktop.Avahi.Server")


def browse(dev):
    from dbus.mainloop.glib import DBusGMainLoop
    from gi.repository import GLib
    DBusGMainLoop(set_as_default=True)
    bus, server = avahi()
    path = server.ServiceBrowserNew(socket.if_nametoindex(dev), -1, TYPE, "local", 0)
    browser = bus.get_object("org.freedesktop.Avahi", path)
    found = set()
    def changed(index, proto, name, kind, domain, flags, add):
        item = (int(proto), str(name)[:100])
        if add and not int(flags) & 8 and len(found) < 128:
            found.add(item)
        else:
            found.discard(item)
        print(json.dumps(sorted({label for _, label in found})), flush=True)
    for event, add in (("ItemNew", True), ("ItemRemove", False)):
        browser.connect_to_signal(event, lambda *args, adding=add: changed(*args, adding))
    print("[]", flush=True)
    GLib.MainLoop().run()


def value(db, name, default):
    row = db.execute("SELECT value FROM local_runtime WHERE key=?", ("worker_update_"+name,)).fetchone()
    return row[0] if row else default


def put(db, name, data):
    db.execute("INSERT OR REPLACE INTO local_runtime VALUES(?,?,0)", ("worker_update_"+name, str(data)))


def worker_idle(db, drain=False):
    if db.execute("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='pool_tasks')").fetchone()[0]:
        if db.execute("""SELECT EXISTS(SELECT 1 FROM pool_tasks WHERE
            (state='running' AND (lease_until>? OR ?=0)) OR (state='queued' AND ?=0))""",
            (int(time.time()*1000), int(drain), int(drain))).fetchone()[0]:
            return False
    return not db.execute("""SELECT EXISTS(SELECT 1 FROM project_entries WHERE kind='run'
        AND json_extract(body,'$.State') IN ('pending','queued','running','assigned'))""").fetchone()[0]


def claim_update(db, manual=False, now=None):
    now = int(time.time()) if now is None else now
    def due():
        if manual:
            return value(db, "manual", "0") == "1"
        return value(db, "enabled", "1") == "1" and now >= int(value(db, "due", "0"))
    if not due():
        return False
    with db:
        db.execute("BEGIN IMMEDIATE")
        if not due() or not worker_idle(db, manual and value(db, "drain_requested", "0") == "1"):
            return False
        put(db, "last", now)
        put(db, "due", now + 3600)
        put(db, "result", "checking")
        if manual:
            put(db, "manual", "0")
    return True


def update_command(db, action, args):
    with db:
        if action == "status":
            print(json.dumps({"enabled": value(db, "enabled", "1") == "1",
                              "last_attempt_unix": int(value(db, "last", "0")),
                              "next_due_unix": int(value(db, "due", "0")),
                              "manual_requested": value(db, "manual", "0") == "1",
                              "result": value(db, "result", "never checked")}))
        elif action in ("enable", "disable"):
            put(db, "enabled", int(action == "enable"))
        elif action == "check":
            put(db, "manual", "1")
        elif action in ("eligible","drain"):
            db.execute("BEGIN IMMEDIATE")
            requested = args == ["manual"] and value(db, "drain_requested", "0") == "1"
            if not worker_idle(db, requested) or (args != ["manual"] and value(db, "enabled", "1") != "1"):
                sys.exit(1)
            if action == "drain":
                put(db, "installing", 1)
        elif action == "result":
            if args[0] != "installed":
                put(db, "installing", 0)
            put(db, "drain_requested", 0)
            put(db, "result", args[0])
            if args[0] in ("current", "installed"):
                put(db, "due", int(value(db, "last", "0")) + 86400)
        else:
            raise ValueError("Unknown update operation")


def serve(db, rule, headless=False):
    dev, addr = profiles()[rule]
    index = socket.if_nametoindex(dev)
    database_path = db.execute("pragma database_list").fetchone()[2]
    connections = threading.local()

    def local_db():
        return getattr(connections, "db", db)

    def permitted():
        return rule in grants(local_db()) and rule in profiles() and socket.if_nametoindex(dev) == index

    class Handler(http.server.BaseHTTPRequestHandler):
        def handle(self):
            self.connection.settimeout(3)
            if self.connection.recv(1, socket.MSG_PEEK) != b"\x16":
                return super().handle()
            if not permitted():
                return
            # TLS terminates in the Go worker. Revalidate grants before forwarding
            # each bounded chunk, including after the TLS handshake completes.
            path = database_path
            with socket.socket(socket.AF_UNIX) as backend:
                backend.settimeout(3)
                name = "worker-claims.sock" if headless else "pair-" + hashlib.sha256(rule.encode()).hexdigest()[:16] + ".sock"
                try:
                    backend.connect(os.path.join(os.path.dirname(path), name))
                except OSError:
                    return
                sockets = [self.connection, backend]
                deadline = time.monotonic() + (300 if headless else 90)
                remaining = (64 << 20) if headless else 400000
                while remaining > 0 and time.monotonic() < deadline:
                    ready, _, _ = select.select(sockets, [], [], 0.5)
                    for src in ready:
                        chunk = src.recv(min(16384, remaining))
                        if not chunk or not permitted():
                            return
                        remaining -= len(chunk)
                        dest = backend if src is self.connection else self.connection
                        dest.sendall(chunk)

        def handle_one_request(self):
            if permitted():
                super().handle_one_request()

        def log_message(self, *args):
            pass

        def reply(self, status, code, message):
            body = json.dumps({"error": {"code": code, "message": message}}).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Cache-Control", "no-store")
            if status == 401:
                self.send_header("WWW-Authenticate", "Bearer")
            if status == 405:
                self.send_header("Allow", "GET")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            if self.command != "HEAD":
                self.wfile.write(body)
            self.wfile.flush()
            if headless and claim_update(local_db()):
                print("A", end="", flush=True)

        def send_error(self, code, message=None, explain=None):
            self.reply(405 if code == 501 else code, "request_rejected", "Request rejected.")

        def do_GET(self):
            if not permitted():
                return
            if self.path == "/v1":
                self.reply(401, "authentication_required", "Authentication is needed to use the API.")
            else:
                self.reply(404, "not_found", "Resource not found.")

    class ScopedServer(http.server.ThreadingHTTPServer):
        slots = threading.BoundedSemaphore(16)
        def process_request(self, request, client_address):
            if self.slots.acquire(blocking=False):
                super().process_request(request, client_address)
            else:
                self.shutdown_request(request)

        def process_request_thread(self, request, client_address):
            try:
                with contextlib.closing(sqlite3.connect(database_path)) as connection:
                    connections.db = connection
                    super().process_request_thread(request, client_address)
            finally:
                self.slots.release()

    with ScopedServer((addr, 0), Handler, bind_and_activate=False) as listener:
        listener.socket.setsockopt(socket.SOL_SOCKET, socket.SO_BINDTODEVICE, dev.encode()+b"\0")
        listener.server_bind()
        listener.server_activate()
        listener.timeout = 1
        Handler.timeout = 2
        if dev not in ("lo", "tailscale0"):
            import dbus
            bus, server = avahi()
            group = dbus.Interface(bus.get_object("org.freedesktop.Avahi", server.EntryGroupNew()), "org.freedesktop.Avahi.EntryGroup")
            name = os.environ.get("APPARAT_DISCOVERY_NAME", "")
            if not name:
                try:
                    with open(os.environ["APPARAT_NAME_FILE"]) as source:
                        name = source.read(64)
                except (OSError, KeyError):
                    return
            if not re.fullmatch(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?", name) or name.lower() in ("localhost", "localhost4", "localhost6", "ip6-localhost", "ip6-loopback", "loopback", "android", "unknown", "none"):
                return
            group.AddService(index, 0, 0, name, TYPE, "local", "", listener.server_port, dbus.Array([], signature="ay"))
            group.Commit()
        while permitted():
            listener.handle_request()


def supervise(db, path, *options):
    if options:
        with db:
            put(db, "installing", 0)
            if value(db, "drain_requested", "0") == "1":
                put(db, "manual", 1)
    workers = {}
    try:
        while True:
            if options and value(db, "manual", "0") == "1" and claim_update(db, manual=True):
                print("M", end="", flush=True)
            rules = set(grants(db))
            try:
                available = profiles() if rules else {}
            except (OSError, subprocess.SubprocessError, ValueError):
                available = {}
            for rule in list(workers):
                if workers[rule].poll() is not None or rule not in rules or rule not in available:
                    workers.pop(rule).terminate()
            for rule in rules.intersection(available).difference(workers):
                workers[rule] = subprocess.Popen([sys.executable,"-c","import sys;SOURCE=sys.argv.pop(1);exec(SOURCE)",SOURCE,"serve",path,rule,*options])
            time.sleep(2)
    finally:
        for worker in workers.values():
            worker.terminate()
            worker.wait()


def configure(db):
    import gi
    gi.require_version("Gtk", "3.0")
    from gi.repository import Gtk
    dialog = Gtk.Dialog(title="API configuration")
    dialog.add_button("Cancel", Gtk.ResponseType.CANCEL)
    dialog.add_button("Save", Gtk.ResponseType.OK)
    box = dialog.get_content_area()
    box.add(Gtk.Label(label="Allow API hosting on these networks and interfaces"))
    switch = Gtk.CheckButton(label="Enable API server")
    switch.set_active(enabled(db))
    box.add(switch)
    previous = saved(db)
    active = profiles()
    checks = {}
    for rule in sorted(set(previous).union(active)):
        dev, uuid, gateway, address = rule.split("/")
        label = f"{dev} · {address} · gateway {gateway}"
        if rule not in active:
            label += " (unavailable)"
        checks[rule] = Gtk.CheckButton(label=label)
        checks[rule].set_active(rule in previous)
        box.add(checks[rule])
    dialog.show_all()
    if dialog.run() == Gtk.ResponseType.OK:
        save(db, [key for key, check in checks.items() if check.get_active()], switch.get_active())
    dialog.destroy()


def main():
    mode, path, *args = sys.argv[1:]
    if mode == "browse":
        browse(path)
        return
    with contextlib.closing(sqlite3.connect(path)) as db:
        if mode.startswith("updates-"):
            update_command(db, mode[8:], args)
        elif mode == "configure":
            configure(db)
        elif mode == "serve":
            # Exit if the supervisor dies, including abrupt app shutdown.
            import ctypes
            parent = os.getppid()
            ctypes.CDLL(None).prctl(1, 15, 0, 0, 0)
            if parent != 1 and parent == os.getppid():
                serve(db, args[0], args[1:] == ["headless"])
        elif mode == "supervise":
            supervise(db, path, *args)
        elif mode == "list":
            print(json.dumps({"available":list(profiles()), "granted":saved(db), "enabled":grants(db)}))
        elif mode == "grant" and args[0] in profiles():
            save(db, sorted(set(saved(db)).union(args)))
        elif mode == "disable":
            save(db, saved(db), False)
        elif mode == "revoke":
            save(db, [key for key in saved(db) if key != args[0]], enabled(db))
        else:
            raise ValueError("Select an available network rule from api-list")
if __name__ == "__main__":
    main()
