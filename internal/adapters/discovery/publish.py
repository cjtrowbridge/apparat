"""Publish Linux worker endpoints through the interface-scoped Avahi entry group.
Claimable workers support browsing. Claimed workers support exact-name lookup.
"""
import os
import re
import struct


def discovery_name():
    name = os.environ.get("APPARAT_DISCOVERY_NAME", "")
    if not name:
        try:
            with open(os.environ["APPARAT_NAME_FILE"]) as source:
                name = source.read(64)
        except (OSError, KeyError):
            return ""
    if not re.fullmatch(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?", name):
        return ""
    if name.lower() in ("localhost", "localhost4", "localhost6", "ip6-localhost", "ip6-loopback", "loopback", "android", "unknown", "none"):
        return ""
    return name


def publish_endpoint(group, server, index, name, port, claimed):
    import dbus
    if claimed:
        # Android NSD resolves the SRV and TXT pair. Only a PTR makes it browsable.
        fqdn = name + "._apparat._tcp.local"
        target = str(server.GetHostNameFqdn()).rstrip(".").encode("ascii")
        wire = b"".join(bytes([len(label)]) + label for label in target.split(b".")) + b"\0"
        srv = struct.pack("!HHH", 0, 0, port) + wire
        for kind, data in ((33, srv), (16, b"\0")):
            group.AddRecord(index, 0, 0, fqdn, 1, kind, 120, dbus.ByteArray(data))
    else:
        group.AddService(index, 0, 0, name, "_apparat._tcp", "local", "", port, dbus.Array([], signature="ay"))
    group.Commit()
