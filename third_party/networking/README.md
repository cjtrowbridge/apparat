# Networking Sources

These WireGuard trees support detection, inspection, platform research, and later control integration. Apparat's MVP assumes WireGuard is configured outside the application and carries authenticated HTTPS REST traffic.

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `wireguard-go` | `https://git.zx2c4.com/wireguard-go` | `b3bc704ef0ee21a36298601a9e1e5ff90377ac52` (`0.0.20250522-3-gb3bc704`) | MIT | Official userspace WireGuard implementation reference | Source reference; not linked into the MVP |
| `wgctrl-go` | `https://github.com/WireGuard/wgctrl-go.git` | `a9ab2273dd1075ea74b88c76f8757f8b4003fcbf` | MIT | Go API reference for inspecting and later controlling WireGuard devices | Future optional adapter; not an active dependency |
| `wireguard-tools` | `https://git.zx2c4.com/wireguard-tools` | `ba8556ce45c230906041a1a04450b7b638057917` (`v1.0.20260223-9-gba8556c`) | GPL-2.0 | Linux and Steam Deck configuration and behavior reference | Reference only; Apparat does not link, vendor, or incorporate its GPL implementation into the MVP binary |

## Boundary

- WireGuard supplies network reachability; it is not Apparat's application protocol, user identity, device identity, or authorization model.
- HTTPS, mutual authentication, signed envelopes, replay protection, and application authorization remain mandatory over WireGuard.
- These sources do not prove that Apparat can create or manage tunnels portably.
- App-managed WireGuard is deferred until platform permissions, lifecycle, enrollment, recovery, and failure handling have dedicated designs.
- Signal, Meshtastic, and other future transports must reuse the canonical durable message model rather than inheriting WireGuard-specific domain fields.

## License Note

`wireguard-tools` is distributed under GPL-2.0. Its checkout is retained as an upstream source and behavior reference. Any proposal to copy, modify, link, package, or distribute its implementation as part of Apparat requires a separate license and architecture review.

## 2026-11-16 Source Refresh

Updated to configured upstream heads at the user’s request. Exact revisions above remain pinned by gitlinks; branch tracking was not added. See the [update review](../../plans/past/2026-11-16-11-00-00_update-and-review-submodules.md) for deltas, verification limits, and rollback. Source-reference updates do not change root Go module pins or activate new integrations.
