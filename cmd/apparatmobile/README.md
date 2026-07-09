# `apparatmobile`

`cmd/apparatmobile` is the Go package bound into the Android GUI APK through Ebitengine's mobile view pipeline.

It is not a standalone command. During Android app startup, package initialization loads Apparat GUI configuration, initializes the shared `internal/app` runtime, creates the fresh Android `last_run.log`, and registers the HUD game with `mobile.SetGame`.

The Android wrapper activity owns the Java view lifecycle. This package owns only the Go runtime/HUD registration boundary, so Android can render the same HUD model as Debian without depending on the temporary `third_party/salvagecore` reference checkout.
