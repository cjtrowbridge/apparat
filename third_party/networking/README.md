# Networking Sources

These WireGuard trees support detection, inspection, platform research, and later control integration. Apparat's MVP assumes WireGuard is configured outside the application and carries authenticated HTTPS REST traffic.

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `wireguard-go` | `https://git.zx2c4.com/wireguard-go` | `ecfc5a8d54462e18e13c72173e2623d16d8e25a0` (`0.0.20250522-1-gecfc5a8`) | MIT | Official userspace WireGuard implementation reference | Source reference; not linked into the MVP |
| `wgctrl-go` | `https://github.com/WireGuard/wgctrl-go.git` | `a9ab2273dd1075ea74b88c76f8757f8b4003fcbf` | MIT | Go API reference for inspecting and later controlling WireGuard devices | Future optional adapter; not an active dependency |
| `wireguard-tools` | `https://git.zx2c4.com/wireguard-tools` | `a998407747005ea7e4e0258d96f105c97241e1d3` (`v1.0.20260223-5-ga998407`) | GPL-2.0 | Linux and Steam Deck configuration and behavior reference | Reference only; Apparat does not link, vendor, or incorporate its GPL implementation into the MVP binary |

## Boundary

- WireGuard supplies network reachability; it is not Apparat's application protocol, user identity, device identity, or authorization model.
- HTTPS, mutual authentication, signed envelopes, replay protection, and application authorization remain mandatory over WireGuard.
- These sources do not prove that Apparat can create or manage tunnels portably.
- App-managed WireGuard is deferred until platform permissions, lifecycle, enrollment, recovery, and failure handling have dedicated designs.
- Signal, Meshtastic, and other future transports must reuse the canonical durable message model rather than inheriting WireGuard-specific domain fields.

## License Note

`wireguard-tools` is distributed under GPL-2.0. Its checkout is retained as an upstream source and behavior reference. Any proposal to copy, modify, link, package, or distribute its implementation as part of Apparat requires a separate license and architecture review.
