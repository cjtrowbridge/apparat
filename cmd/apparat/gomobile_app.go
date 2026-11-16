//go:build android

// Connects the Android entrypoint to the GUI lifecycle.
// Dependencies: github.com/ebitengine/gomobile/app.
package main

import "github.com/ebitengine/gomobile/app"

var _ = app.Main
