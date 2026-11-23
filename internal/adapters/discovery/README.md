Linux API hosting uses the embedded linux.py helper invoked by linux_helper.go. It needs system Python 3, python-dbus, PyGObject, Avahi, iproute2 and NetworkManager. The native configuration dialog additionally uses GTK 3. Missing dependencies leave hosting inactive. Grants in local_runtime identify an IPv4 LAN adapter, connection UUID, gateway hardware address and local IP. DHCP changes require a new grant. The listener binds to the exact device and address. Browse uses an explicitly selected interface and transient Avahi records. Verify with two devices and a denied interface.

- `linux.py`: Linux network grants, scoped HTTP and Avahi discovery.
- `linux_helper.go`: Embeds the Linux helper and creates cancellable commands.

- `browse.go`: Interface-selected IPv4 DNS-SD queries with expiring results for Windows and macOS browsers.

Headless responders notify the owner after a flushed API response wins the SQLite schedule claim. The existing supervisor forwards explicit local check requests. Only the daemon receives update events.
- `resolve.go`: Resolves the selected worker's DNS-SD port on the interface used for browsing before authentication.
- `reconnect.go`: Resolves a saved hostname on its previous local subnet when the last endpoint stops responding. The caller pins the existing worker fingerprint on every retry.
