# Node identity and API access

The GUI inspects the active signing identity on startup. If keys are missing, the modal explains that a returning user must restore the original private key to retain their cluster connections. Choose **Find existing keys** to open a legacy encrypted device key or a portable recovery copy. **Generate new keys** requires confirmation of the identity change. Cancel leaves signing locked and permits offline work. Existing encrypted keys require local unlock. Cluster claiming remains disabled pending authenticated enrollment.

Apparat verifies a new pair before switching a private `identity/active.json` pointer to its generation. The previously active files remain available for recovery. A surviving valid manifest constrains which key can be restored. The current implementation protects the private key with the existing Argon2id/XChaCha20-Poly1305 format. System enrollment uses a randomly generated secret protected by the selected OS facility. Save a password-protected portable recovery copy when offered, especially with device-bound protection. Moving the active directory alone may not transfer access to its system credential.

The following sections describe the wider Phase 10 design, including additional protection management, authenticated HTTP and enrollment.

## One signing identity per node root

The process that owns the node lock inspects the device manifest and private file before offering setup. Generate a device key only in a new root with both files absent. If setup was interrupted or the files disagree, report recovery required and preserve the evidence. For a valid existing identity, use its recorded format regardless of which executable starts. Opening the GUI on a daemon-created root, restarting, upgrading, or changing protection must preserve the fingerprint. The policy here applies to the node signing key. User identity, cluster CA, TLS leaf and WireGuard keys have separate lifecycles.

| First setup | Default and alternative |
| --- | --- |
| GUI | Ask for and confirm a masked passphrase. Offer an explicit unencrypted option. |
| Headless | Create an owner-only unencrypted key for unattended restart. Offer encrypted setup before first service startup through an interactive terminal with hidden input. When encryption is requested without an available secret, leave setup pending. |
| Existing node | Reuse the current mode. The daemon can run locked and serve a validated public manifest until the local owner unlocks it. |

Show the consequence of unencrypted storage at setup and in Settings. Anyone who reads the private file can impersonate the node. Create the directory and private file with owner-only access from the first write (0700 and 0600 on Unix, restricted ACL on Windows). Fail setup if protection cannot be established. Keep secrets out of SQLite, command arguments, logs and diagnostics.

## Encrypted storage and native unlock

Use a versioned authenticated envelope. Encrypt the signing key with a random data key using XChaCha20-Poly1305. Wrap that data key with a passphrase-derived Argon2id key using a fresh salt and independent nonce. Bind the node identity and format version to the authenticated data. Optional OS unlock protects another copy of the same data key. The password path remains usable after OS credential loss. For legacy files, unlock with the existing passphrase, derive the public key, and compare it with the manifest before migration. For changes of format, verify the staged replacement and its fingerprint before atomic promotion. Reject unknown formats. An incorrect password, cancelled prompt or crash must never generate a new identity.

Offer encryption and passphrase changes locally. Removing encryption requires a successful unlock and explicit confirmation. Retire superseded OS entries after the replacement is durable. Changing local protection cannot revoke copies already made of a plaintext key. An explicit lock or shutdown clears the active signing session and invalidates its tokens. On mobile, relock when the app backgrounds, accounting for the native unlock prompt lifecycle. Remote API requests cannot open an unlock dialog.

| Platform | Proposed integration and fallback |
| --- | --- |
| Android | Offer a Keystore wrapping key requiring user authentication and bind the unwrap operation to BiometricPrompt. Probe device and Android-version capability, including the API 23 floor. Cancellation leaves the node locked. On invalidation offer the Apparat passphrase and local re-enrollment. |
| Linux desktop and Steam Deck | Offer “Use system keyring” when a Secret Service provider is available. It may unlock with the login session. Test desktop and Gaming Mode separately. |
| Linux headless | Support unattended plaintext mode and explicit passphrase setup/unlock at a local terminal. An encrypted daemon starts locked after reboot. |
| Windows | Offer passphrase first, then opt-in user-scoped DPAPI with a restricted file ACL. Label this Windows sign-in. A Windows Hello claim requires a separately verified gesture-bound implementation. |
| macOS | Offer a device-local Keychain item with access controls requiring user presence and Touch ID where available. Verify packaged application access across restart and upgrade. |

All GUI targets retain passphrase unlock if native facilities are absent. Report native support per platform only after exercising it on that platform.

## API hosting policy

Inbound hosting and outbound client access have independent lifecycles. All GUI platforms start client-only. The Linux headless service can start without user interaction, but its initial allowlist is also empty. With no matching affirmative grant there is no HTTP or HTTPS socket, public-key response, authentication message or API service advertisement. This includes IPv4 and IPv6 loopback. The in-process GUI and owner-only local administration socket remain usable to configure an inactive server.

**Settings → API configuration** opens a local modal. Show a server enable switch (initially off), current networks grouped with their interfaces, and any saved rules for unavailable networks. Each rule names both the network and interface, shows the concrete addresses that would be served, and starts unchecked. Explain that a Tailscale tunnel and home Wi-Fi require separate selections. Save applies the chosen rules; Cancel discards edits. Enabling requires at least one explicit rule. Show actual listening addresses separately from saved permission, with a reason when a rule is unavailable or blocked on authentication prerequisites. Turning hosting off closes all listeners and stops active API connections and streams. Give Linux headless operators equivalent local CLI operations to list, grant, revoke, enable, disable and inspect effective bindings. No remote API access is needed for setup.

Persist node-local hosting intent and approved pairs in SQLite. For existing installations without a saved policy, migrate to disabled with zero grants. A legacy listen address or existing cluster membership supplies no approval. The runtime owns one policy for bootstrap, authenticated routes, diagnostics, and future provider streams. Before accepting or advertising traffic, require hosting enabled, an exact match to an approved network **and** interface, and eligibility for that transport. Never bind wildcards. Scope ingress to the approved interface as well as the local address, using platform-supported socket or network controls. A local IP, adapter name, SSID or private subnet alone cannot establish the approved network. Use a stable platform network/profile or tunnel identity and verified interface association; when the platform cannot distinguish or enforce the pair reliably, leave it inactive and explain why. Treat loopback as its own local-only pair.

Reconcile on startup, resume, VPN changes, roaming, address changes and policy edits. Invalidate affected sockets, accepted connections, long-lived streams and advertisements when permission or network identity is lost. Revalidate the actual ingress path when admitting connections and suppress serving while network state is uncertain. Failures must never broaden the binding. A network grant only selects where the API can listen; cluster authentication and operation-specific authorization still control who may use it. Future LLM serving on Windows or Android follows these same gates.

## Local network discovery milestone

Use DNS-SD over link-local mDNS with service type `_apparat._tcp.local.`. Publish an instance label, service port and protocol version after the explicitly granted bootstrap listener is ready. Minimize TXT data. Withdraw the advertisement and stop serving on grant removal, loss of the network or shutdown. This narrow bootstrap exception precedes enrollment; the full diagnostic handler and authenticated network operations retain their later gates.

Opening Local network discovery starts a cancellable browse on a user-selected local network. Leaving the screen stops it. Browsing requires neither inbound API activation nor advertising this client. Show searching, empty, permission-denied, unavailable and discovered states. Associate results with their receiving interface, expire stale entries, suppress self and bound untrusted labels and result counts. No automatic URL fetch or membership write. Link-local multicast can be unavailable across VPNs or isolated Wi-Fi; report that condition without widening the search or requesting router port mappings.

## Caller credentials and network boundaries

On loopback, the operator can issue a short-lived opaque bearer token after local unlock. For an unencrypted daemon, an owner-restricted Unix socket checks the connecting account before issuing a token. Use an equivalently user-restricted named pipe on Windows if headless administration is supported there. An encrypted node additionally requires an unlocked local owner session. Generate tokens with a cryptographic random source, retain only a verifier, expiry and local read scope in memory, and accept tokens only in the Authorization header on loopback. Expire them on timeout, lock or restart. The ordinary HTTP listener cannot mint a token.

A public key advertises an identity. During cluster enrollment the owner verifies its fingerprint through a trusted channel, such as scanning an invitation on the other device. Phase 10 subsequently authorizes enrolled peers through TLS 1.3 mutual authentication plus a current device record and scopes. Keep the TLS leaf distinct from the node signing key. If first-contact lookup is enabled on LAN/WireGuard, expose only the bootstrap response and public identity on a limited HTTPS listener. The protected peer listener requires a client certificate at handshake. Reuse the same public-key projection on both. Loopback tokens have no authority on either network listener. Keep protected network routes disabled until enrollment, revocation, limits and audit meet the roadmap gates.
