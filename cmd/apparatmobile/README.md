# `apparatmobile`

`cmd/apparatmobile` is the Go package bound into the Android GUI APK through Ebitengine's mobile view pipeline.

It is not a standalone command. During Android app startup, package initialization loads Apparat GUI configuration, initializes the shared `internal/app` runtime, creates the fresh Android `last_run.log`, and registers the HUD game with `mobile.SetGame`.

The Android wrapper activity owns the Java view lifecycle. This package owns only the Go runtime/HUD registration boundary, so Android can render the same HUD model as Debian without depending on the temporary `third_party/salvagecore` reference checkout.

## Source inventory

- `doc.go`: Documents the cmd/apparatmobile package boundary.
- `mobile.go`: Exposes the Android game and updater bridge.

Android safe-area color comes from the GUI renderer through `BackgroundARGB` in the mobile bridge. The current fixed dark palette is authoritative; runtime theme switching is not implemented. The native parent and navigation background use that color, with light navigation icons and no automatic contrast scrim. Insets and touch bounds are unchanged. Verify matching background pixels and readable buttons in portrait/landscape on an authorized device; system/OEM styling on other devices requires separate evidence.
- `projects.go`: bridges user-granted document trees, Git metadata reads and staged public clones into the Android provider.

- `workspace.go`: Opens, lists and updates bounded text documents under an existing persisted grant and supplies the tree to the Git adapter. A failed provider replacement retains the previous document or identifies its recovery file.
- `workspace_fs.go`: Adapts Android document operations and seekable file descriptors to go-git's filesystem interface. Provider errors propagate to the selected project.

- `document.go`: Binds the Android document editor to the selected file and draft session.
- `identity.go`: Bridges the local node identity workflow to private Android prompts and locks the signing session on background.

The Activity supplies its current Android user's files directory before the Go runtime registers the game. Recreating the Activity within one process reuses the same initialized runtime. Each Android profile has its own device identity, database and private storage.
