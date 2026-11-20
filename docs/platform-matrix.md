# Platform Matrix

Support is claimed only after target-specific build and behavior evidence exists.

The shared-core and local-service contracts are platform-independent targets, but provider discovery, acceleration, process supervision, credential storage, filesystem isolation, audio, background execution, and service-manager behavior require separate evidence on every supported target.

## Node Process Ownership

Phase 8 gives `apparat` and `apparatd` one logical default node root protected by an OS lock. GUI and headless are alternative owners of the same shared core and SQLite state. Explicit roots remain available for independent nodes.

Starting both artifacts against one node root must fail safely until a separately approved daemon-client mode exists. Multiple intentional nodes on one host require explicit roots, identities, ports, and local inference-service ownership. No platform may claim service-host support until it proves that duplicate processes cannot advertise the same provider endpoint accidentally.

## Steam Deck And Linux GUI

Requires controller-first HUD, readable 1280x800 sizing, keyboard/mouse fallback, audio capture, local storage paths, external WireGuard compatibility, runtime-root `last_run.log`, and generated release artifact under `releases/linux/<arch>/apparat/latest`.

Phase 3 keeps `--smoke-test` non-window for CI and build environments. The Ebitengine run loop is wired behind the `gui` build tag so native GUI validation can be performed on systems with the required desktop libraries.

### Steam Deck persistent install — 2026-11-20

The existing qualified Linux amd64 GUI from delivery commit `dd1594c` was
installed as `/home/deck/.local/bin/apparat` on SteamOS 3.9.5 (x86_64). Installed
and source SHA-256 both equal
`7e2a9de3638342df95b3af66b2b834a52299c021e99836008b493c98243f96c4`.
All ELF dependencies resolved without a system package change. The binary has
mode 755 and its validated desktop entry has mode 644.

Valve's installed `steamos-add-to-steam` helper registered exactly one native
Apparat shortcut in the logged-in Steam account. Steam launch URI
`steam://rungameid/13604856810501570560` opened the Apparat window on Gamescope's
game display. Graceful shutdown and a second launch through the same Steam
entry reused `~/.local/share/apparat/node` and reached runtime ready with no
error-level startup events. The observed X window was 640x480; full 1280x800
layout and physical controls were not qualified by this installation check.
SSH key access works without an agent and sshd is enabled and active. No reboot
or suspend/resume trial was performed. Local hostname discovery timed out, so
the SSH alias uses the owner-confirmed current IP and needs a DHCP reservation
or an address update if it changes.

This proves installation and native launcher startup, not the broader controller,
Steam+X, microphone, workspace, update/migration, suspend or release-support gates.
See the [deployment playbook](../playbooks/how_to_install_apparat_on_steam_deck_and_add_it_to_steam.md).

### Deck UI and Linux update acceptance — 2026-11-20

Canonical delivery `aa37a83b9facf1167ea7a59bea84b1249a27b4ea` qualified and
published the Linux GUI and worker, then authenticated GitHub main read-back
matched their bytes. The previous Android APK and its original input evidence
were preserved because this travelling host lacked its Android prerequisites and
canonical Pixel. No new Android qualification is claimed.

The installed Deck GUI matches published SHA-256
`564a45c9e931b26396a2953f21c20e1e0caffff05674d11c1b4edf0e23e37f28`.
Gamescope now presents a native 1280x800 window at 0,0, compared with the prior
640x480 window. Actual Deck captures show the gallery filling the available
viewport and its centered gear immediately after discovery. Activating that gear
opened embedded settings and Back returned to the gallery. An untargeted PTT
press did not show the removed dictation hint.

The actual Settings updater reported Up To Date against public GitHub. A
controlled trial kept the new updater process running while atomically selecting
the older, previously qualified disk binary from the deployment backup. Check
then downloaded and verified the published binary, Install replaced it atomically
and retained the old `.previous` hash, and Restart re-executed the new binary.
PID 5466 was retained for Steam tracking; fresh runtime ready was recorded at
16:45:34 UTC, with zero startup errors and no staging files left. The same node
root, saved preferences and original Steam entry were retained. This verifies the
real transfer/install/restart path; it is not a claim that the older binary had
an updater. Offline binary/node backup:
`/home/deck/.local/share/apparat/deployment-backups/20261120T163917Z/`.

Private screenshots are under `.tmp/steam-deck-ui/` (`before.png`, `gallery.png`,
`gallery-settings.png`, `update-current.png`, `update-ready.png`,
`update-installed.png`, `restarted.png`). Canonical Linux GUI evidence is in
`.tmp/build-runs/tmpjmnk869a/`. Package/race tests separately cover failures and
no-op states. Physical controller, microphone and suspend/resume qualification
remain open; this checkpoint does not close the whole Phase 17 support matrix.

## Linux Headless

Requires no Ebitengine initialization, CLI/API/service-manager control, health checks, graceful `SIGINT`/`SIGTERM`, durable storage paths, runtime-root `last_run.log`, and generated release artifact under `releases/linux/<arch>/apparatd/latest`.

Phase 3 headless startup initializes the same config, directory, logging, SQLite, identity-status, cluster-directory, and messaging primitives without importing Ebitengine.

Linux is the first platform for supervising arbitrary localhost inference service instances. Service-host evidence must include multiple same-provider instances, independent health/failure behavior, local credential references, bounded probing, gateway-only remote access, and clean shutdown without taking provider processes hostage.

## Windows

Requires `releases/windows/<arch>/apparat/latest.exe` and `releases/windows/<arch>/apparatd/latest.exe` artifacts, external WireGuard validation, service or tray decisions, certificate store handling, audio and controller validation, and path migration tests.

## macOS

Requires signing/notarization decisions, external WireGuard validation, keychain/certificate handling, app lifecycle, audio permissions, and controller validation.

## Android

Phase 5 adds the first Android build artifact but does not yet claim full Android runtime support.

Current build contract and historical evidence (earlier device results do not validate the refreshed engine):

- Build host: Linux `amd64` development environment.
- Additional Windows evidence: Windows `amd64` produced the canonical signed APK through `python -m scripts.build` with Eclipse Temurin OpenJDK 21 and the pinned repo-local SDK/NDK. A connected Pixel Tablet (Android 16, API 36) accepted a fresh install after removal of a differently signed prior build; `MainActivity` resumed and remained alive. A touch swipe advanced body content while leaving the top tab strip stationary. This is one-device GUI evidence, not full Windows Android runtime support.
- Build command: `python3 scripts/build.py`.
- Preflight command: `python3 scripts/build.py` target report.
- Artifact: `releases/android/arm64/apparat/latest.apk`.
- Package ID: `com.cjtrowbridge.apparat`.
- App label: `Apparat`.
- Native ABI: `arm64-v8a`.
- Permissions: `android.permission.INTERNET` for HTTPS over external WireGuard/local networks, `android.permission.RECORD_AUDIO` for the push-to-talk state path, and temporary Phase 5 `android.permission.REQUEST_INSTALL_PACKAGES` for the dev Updates fieldset button.
- Activity: `com.cjtrowbridge.apparat.MainActivity` with Ebitengine `EbitenView`; the wrapper does not force portrait orientation.
- Toolchain: OpenJDK 21 (Eclipse Temurin preferred; Oracle JDK prohibited), Android platform `android-35`, build-tools `35.0.0`, NDK `27.2.12479018`, and pinned Ebitengine gomobile.
- SDK metadata: `minSdkVersion=23`, `targetSdkVersion=35`, and `platformBuildVersionCode=35`.
- Signing: debug keystore generated under ignored `.tools/android/debug.keystore`; APK verifies with v1, v2, and v3 signature schemes.
- Native page alignment: `lib/arm64-v8a/libapparat.so` LOAD segments align to `0x4000` for 16 KB page-size devices.
- Device validation: installed successfully on connected Pixel device `58051FDCQ002T9`, Android release `16`, SDK `36`; process remained alive after launch, wrote `last_run.log` under app-private storage, rendered the Apparat HUD, and accepted touch tab selection.

Known caveats:

- Additional Android device coverage is still pending for safe-area handling, density/readability hardening, keyboard/controller input, portrait and landscape behavior across form factors, and deeper runtime-path validation.
- The current APK uses the wrapper/AAR-style path. Future release-hardening still needs signing, icons, store packaging, additional ABI decisions, and broader Android device validation.
- Android `apparatd` is intentionally unsupported; headless Android work requires a later Termux/service-worker strategy.
- App-managed WireGuard/VPN-service, real microphone capture, broad storage, background execution, release signing, store packaging, signed update manifests, and additional Android ABIs are future work.
- Android local inference-provider supervision is not implied by the GUI APK. Each supported provider, background/lifecycle mode, accelerator path, credential store, and endpoint assumption requires separate validation; Android headless remains out of scope unless a Termux/service-worker strategy is approved.

## 2026-11-16 compatibility checkpoint

The selected stable Ebitengine v2.10.6 and reconciled gomobile tool build Linux amd64 GUI/headless binaries and the Android arm64 APK through the governed no-flag pipeline. `make verify`, GUI tests under Xvfb, module verification, and Linux GUI/headless smoke checks passed. This does not renew Windows/macOS or Steam Deck hardware evidence.

Pixel 4a (sunfish), Android 13/API 33: installed the target-35 APK, launched MainActivity, observed runtime-ready events and a live process, viewed all five HUD tabs with touch selection, swiped the tab strip, rendered portrait/landscape, and resumed after backgrounding. The same process remained alive. Rotation settings were restored. App-scoped logs contain renderer/surface warnings during rotation, but no observed fatal exception. Very short synthetic `adb input tap` events were intermittently missed; 150 ms held touches reached the intended tabs. This is a test-input limitation to investigate if ordinary touch reproduces it, not proof of every gesture working.

The APK SHA-256 was `76f1656aa5f271bcdfc25530d627614c5296d46fee564d89e81ae290572a7b6b`. Local screenshots, logs, build outputs and the validated APK are retained under ignored `.tmp/remediation/`; tracked release artifacts were restored after testing. The startup updater opened Android's unknown-source permission screen; testing returned without granting that permission or installing the downloaded older APK.

Minimum SDK remains 23 (Android 6); target SDK 35 selects newer Android behavior rules rather than raising the installation minimum. Older Android versions and Android 15 edge-to-edge/inset behavior remain untested with these inputs. The APK remains arm64-only. Long labels and portrait debug-footer text clip at the Pixel's size; density, safe-area and broader form-factor hardening remain Phase 5 follow-up work. No full Android platform-support claim follows from this checkpoint.

## Phase 8 acceptance — 2026-11-16

Linux amd64: governed GUI/headless builds, unit/race tests, native GUI tests under Xvfb, lint, docs/source checks and OpenAPI conformance pass. `go list -deps ./cmd/apparatd` contains no Ebitengine, EbitenUI, HUD or GUI adapter. GUI and headless sequentially opened the same root and retained the same three service IDs; contention failed safely. SIGTERM now exits both forms cleanly. Backup/restore into a new root preserves IDs. Windows amd64 and macOS arm64 lock/config/lifecycle packages cross-compile; no native runtime claim follows. Android compilation is exercised by the full signed APK build.

Pixel 4a / Android 13 / API 33: final APK installs over existing data, migrates to `files/apparat/node`, reports actual initialization readiness, and exposes the same three healthy service IDs in Local services and loopback REST (queried through temporary ADB forwarding, then removed). IDs survived repeated APK replacement/process restarts. Portrait/landscape and background/resume retain the live view. Holding PTT then backgrounding yields `voice=idle input=voice-cancelled`, not submission. A 50 ms touch and 150 ms held touch activate controls; instantaneous ADB taps can still be missed. Slow tab drag reaches Settings without changing the selected tab. Faster injected swipes were not uniformly repeatable; physical finger, controller and external-keyboard confirmation remains a Phase 17 device gate, not a claimed pass.

Fixed the portrait footer clipping, selected-tab reveal and live-detail text overlapping PTT in landscape. Existing long **mock selector** labels can still clip in the landscape master pane; broad selector/accessibility/form-factor polish remains Phase 17. Canvas widgets expose no native Android accessibility tree. Native focus/keyboard/back behavior is retained in tests; actual Android screen-reader support is not claimed. Real microphone/audio cancellation belongs to Phase 16.

The user-initiated updater downloaded the configured candidate, rejected it with `unable_verification`, left an empty cache and opened no installer/permission screen. No unknown-source permission was granted and no downloaded downgrade was installed. Success handoff of a production-signed newer release, key rotation and rollback remain Phase 17. Startup does not download updates automatically.

Measurements are scoped observations: 125 SurfaceFlinger presentation intervals on the app's actual SurfaceView were median 16.683 ms, p95 16.704 ms, max 17.161 ms at a 16.667 ms refresh period. The earlier gfxinfo sample covered only Android wrapper frames and is not a comparable Ebitengine frame baseline; no frame-time improvement or energy claim is made. Pixel PSS observations were 151,512 KiB before and 131,911 KiB after one rotation/resume pass, under differing views/lifetimes; these do not establish an optimization effect. Cold Activity launch samples were 527–831 ms, hot resume 93–136 ms (ADB instrumentation included).

Minimum SDK 23, target SDK 35 and arm64 scope are unchanged. Android 6/Android 15 hardware, Windows/macOS native behavior, real controller input, screen readers, production signing and service-manager integration remain Phase 17 acceptance gaps. Go moved from 1.26.4 to security-patched 1.26.11 because the new HTTP surface exposed standard-library advisories; final audit reports zero reachable vulnerabilities and four unused required-module advisories. The exact checkpoint, tests, measurements and review dispositions are recorded in the Phase 8 execution plan; ignored raw evidence is under `.tmp/phase8/`.

## Pixel 10 safe-area correction (2026-11-16)

Pixel 10 Pro XL reports Android 17, physical 1080x2404 and density 390. The target-35 window enforces edge-to-edge; the previous HUD footer appeared under three-button navigation. MainActivity now sizes the engine inside a native parent padded by system-bar/display-cutout insets on API 35+. Portrait uses 161 px above and 117 px below; landscape uses those reservations on the corresponding sides. Final portrait/landscape screenshots, tab touch selection and background/resume passed. Device rotation preferences were restored and install -r preserved app data. The governed Linux GUI/headless and Android build passed.

Final APK SHA-256: `4cfa7a25d14337d1d8373121b5b741f7269b175fd4316b7287a13139ba45f483`; ignored evidence and APK: `.tmp/pixel10-insets/`. Tracked release artifacts were restored after testing. This adds Android 17 three-button-navigation evidence, not Android 15/16, gesture-navigation, keyboard, accessibility or API 23 hardware coverage. Existing long mock selector labels remain a separate Phase 17 limitation.

Pixel 10 color follow-up: cutout and navigation areas now read the renderer background through the mobile bridge. Portrait/landscape screenshot samples match HUD RGB(15,18,28), with white navigation buttons. Evidence `.tmp/pixel10-color/`. Fixed dark palette only; no runtime theme-switching claim.

Pixel 4a icon-tab follow-up: ☭ Comrades, 🔨 Projects, 🖥️ Cluster, 🧪 Research and ⚙️ Settings render as compact 54x54 buttons with 28px symbols. All five fit in portrait/landscape and each 150ms held tap selects only its intended tab. Native GUI tests verify 320px fit and narrower overflow. Embedded Twemoji artwork is font-independent; Comrades retains DejaVu Sans. Evidence `.tmp/emoji-tabs/`; install preserved data and rotation preferences were restored. Pixel 10 was not modified.

2026-11-17 ordinary Apparat: the canonical Linux amd64 GUI was launched in Xvfb and navigated to Projects; the headless executable passed an isolated runtime smoke test. The canonical Android arm64 APK was installed on an authorized Pixel 4a and its installed bytes verified before Projects capture. See releases/verified.json for the qualified set. Arcade Test is a different package, not this release. No Windows or macOS interactive result is claimed.

## Phase 5 closeout on the Pixel 4a

On 2026-11-17 system ADB launched the installed ordinary Apparat package twice from a stopped process. Both launches reported COLD and succeeded (522 ms and 446 ms). Process IDs changed from 26001 to 26171. The app-private last_run.log advanced from 1,739 bytes to 1,738 and then 1,739 with a new modification time on each launch. This agrees with the truncate-on-start behavior in internal/logging. The app UID had write and traversal access to data, logs, identity, cache, artifacts, backups and recovery under files/apparat/node. All seven directories were mode 0700, and data/apparat.db was nonempty and mode 0600. These checks cover path placement and permissions. Backup restoration and recovery behavior have separate acceptance tests.

The Linux checkout also provides the absent-reference build proof. The third_party directory's last entry change was at 13:23 local time on 2026-11-17, and it contains no salvagecore entry. Canonical run tmpj5ljtmh8 began at 15:55 in that checkout and built and exercised the Linux GUI, Linux daemon and Android APK. Its delivery record identifies commit 90a2202. The Android build reference regression test passed again during closeout. No checkout relocation was needed.

## Phase 9 local repositories

Linux registered vibe-modeling through the desktop folder chooser and cloned the public fixture and private Apparat fixture with one system GitHub login. On the Pixel 4a, an existing Documents checkout, an app-private clone and a Documents clone retained distinct IDs after an offline cold start. External task edits appeared after restart. Folder reselection recovered a moved checkout. Two live Linux accounts with disjoint private access and explicit Android grant revocation remain unverified.

## Phase 9A offline local workspace

Linux amd64 ran the Projects GUI in Xvfb inside a network namespace with only loopback. On Pixel 4a, Wi-Fi and mobile data were disabled for the workflow. Both platforms edited a selected file, displayed its diff, staged and committed it, ran the manual snapshot Task, then reopened history and a verified artifact after a cold start. Pixel testing used both app-private storage and a Documents tree selected through the system picker. Both checkouts began at `10af86a5014f`. Native Git independently verified the exported objects, index and working files.

| Pixel checkout | Commit | SHA-256 of edited file and snapshot |
| --- | --- | --- |
| App-private | `ff19bfe9f9b5` | `b43fd9eed3f08852b374f01444faa5d890568b1a16d7b71c3c0234e0d40af571` |
| Documents | `f050c46b8f0b` | `e110a7edd5f8de0b6e903a9f0693c0eda0469ba199c8ad24ea98638765107baa` |

For Phase 11A, repeat with two registrations at the same revision. Stage a distinct edit in the second checkout before using the first. Inspect both indexes and tips independently. Change the first file outside Apparat with a draft pending, restart, revise against its new base, retry and check for a single effect. The app integration tests exercise these recovery and isolation checks. Physical Steam Deck controller acceptance remains open.

### Workspace usability baseline

The redesigned GUI opened a registered checkout directly and edited a multiline document offline on Linux (GTK 4) and Pixel 4a (Android IME). Both platforms staged and committed the reviewed file and ran a snapshot Task. Both Pixel storage types reopened the verified output after a cold start. Independent Git object checks and byte comparisons matched working tree, index, HEAD and artifact. The app-private test commit was `77f73f7`, and the picker-granted Documents commit was `e4ee288`. The focused app test covers a second outside edit between comparison and rebase. An uncoached first-use observation and physical Steam Deck test are outstanding.

On 0.1.24, the Pixel pasted a multiline selection, exercised Undo and closed the editor with the checkout untouched. After an independent edit and cold launch, it recovered the draft, displayed the conflicting checkout, and saved only after explicit review. A pending message with a named run attachment reopened in the same conversation offline. Linux committed the intended file at `a29a095`, then reopened the verified Task output after a cold start in a compact window. The canonical release qualified the Linux GUI, Linux daemon and Android APK. Local screenshots are under `.tmp/phase9a-device/`.
