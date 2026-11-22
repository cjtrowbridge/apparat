Linux API hosting uses the embedded linux.py helper invoked by linux_helper.go. It needs system Python 3, python-dbus, PyGObject, Avahi, iproute2 and NetworkManager. The native configuration dialog additionally uses GTK 3. Missing dependencies leave hosting inactive. Grants in local_runtime identify an IPv4 LAN adapter, connection UUID, gateway hardware address and local IP. DHCP changes require a new grant. The listener binds to the exact device and address. Browse uses an explicitly selected interface and transient Avahi records. Verify with two devices and a denied interface.

- `linux.py`: Linux network grants, scoped HTTP and Avahi discovery.
- `linux_helper.go`: Embeds the Linux helper and creates cancellable commands.
