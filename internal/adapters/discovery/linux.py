"""Linux bootstrap hosting and multicast browsing through NetworkManager and Avahi.
The owning Apparat process starts this embedded helper with the node database path.
"""
import contextlib
import http.server
import json
import os
import socket
import sqlite3
import subprocess
import sys
import time

TYPE = "_apparat._tcp"
KEY = "bootstrap_grants_v1"


def command(*args):
    return subprocess.check_output(args, text=True, timeout=2).strip()


def profiles():
    result = {}
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


def grants(db):
    row = db.execute("select value from local_runtime where key=?", (KEY,)).fetchone()
    return json.loads(row[0]) if row else []


def save(db, rules):
    with db:
        db.execute("insert or replace into local_runtime values(?,?,0)", (KEY, json.dumps(rules)))


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


def serve(db, rule):
    import dbus
    dev, addr = profiles()[rule]
    index = socket.if_nametoindex(dev)
    def permitted():
        return rule in grants(db) and rule in profiles() and socket.if_nametoindex(dev) == index

    class Handler(http.server.BaseHTTPRequestHandler):
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

        def send_error(self, code, message=None, explain=None):
            self.reply(405 if code == 501 else code, "request_rejected", "Request rejected.")

        def do_GET(self):
            if not permitted():
                return
            if self.path == "/v1":
                self.reply(401, "authentication_required", "Authentication is needed to use the API.")
            else:
                self.reply(404, "not_found", "Resource not found.")

    with http.server.HTTPServer((addr, 0), Handler, bind_and_activate=False) as http:
        http.socket.setsockopt(socket.SOL_SOCKET, socket.SO_BINDTODEVICE, dev.encode()+b"\0")
        http.server_bind()
        http.server_activate()
        http.timeout = 1
        Handler.timeout = 2
        bus, server = avahi()
        group = dbus.Interface(bus.get_object("org.freedesktop.Avahi", server.EntryGroupNew()), "org.freedesktop.Avahi.EntryGroup")
        group.AddService(index, 0, 0, socket.gethostname().split(".")[0], TYPE, "local", "", http.server_port, dbus.Array([], signature="ay"))
        group.Commit()
        while permitted():
            http.handle_request()


def supervise(db, path):
    workers = {}
    try:
        while True:
            rules = set(grants(db))
            try:
                available = profiles() if rules else {}
            except (OSError, subprocess.SubprocessError, ValueError):
                available = {}
            for rule in list(workers):
                if workers[rule].poll() is not None or rule not in rules or rule not in available:
                    workers.pop(rule).terminate()
            for rule in rules.intersection(available).difference(workers):
                workers[rule] = subprocess.Popen([sys.executable,"-c","import sys;SOURCE=sys.argv.pop(1);exec(SOURCE)",SOURCE,"serve",path,rule])
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
    previous = grants(db)
    active = profiles()
    checks = {}
    for rule in sorted(set(previous).union(active)):
        label = rule if rule in active else rule+" (unavailable)"
        checks[rule] = Gtk.CheckButton(label=label)
        checks[rule].set_active(rule in previous)
        box.add(checks[rule])
    dialog.show_all()
    if dialog.run() == Gtk.ResponseType.OK:
        save(db, [key for key, check in checks.items() if check.get_active()])
    dialog.destroy()


def main():
    mode, path, *args = sys.argv[1:]
    if mode == "browse":
        browse(path)
        return
    with contextlib.closing(sqlite3.connect(path)) as db:
        if mode == "configure":
            configure(db)
        elif mode == "serve":
            # Exit if the supervisor dies, including abrupt app shutdown.
            import ctypes
            parent = os.getppid()
            ctypes.CDLL(None).prctl(1, 15, 0, 0, 0)
            if parent != 1 and parent == os.getppid():
                serve(db, args[0])
        elif mode == "supervise":
            supervise(db, path)
        elif mode == "list":
            print(json.dumps({"available":list(profiles()), "enabled":grants(db)}))
        elif mode == "grant" and args[0] in profiles():
            save(db, sorted(set(grants(db)).union(args)))
        elif mode == "disable":
            save(db, [])
        else:
            raise ValueError("Select an available network rule from api-list")
main()
