# Platform Matrix

Support is claimed only after target-specific build and behavior evidence exists.

## Steam Deck And Linux GUI

Requires controller-first HUD, readable 1280x800 sizing, keyboard/mouse fallback, audio capture, local storage paths, external WireGuard compatibility, and generated release artifact under `releases/linux/<arch>/latest`.

## Linux Headless

Requires no Ebitengine initialization, CLI/API/service-manager control, health checks, graceful `SIGINT`/`SIGTERM`, durable storage paths, and generated release artifact under `releases/linux/<arch>/latest`.

## Windows

Requires `.exe` artifacts, external WireGuard validation, service or tray decisions, certificate store handling, audio and controller validation, and path migration tests.

## macOS

Requires signing/notarization decisions, external WireGuard validation, keychain/certificate handling, app lifecycle, audio permissions, and controller validation.

## Android

Requires native wrapper, Ebitengine AAR, lifecycle handling, permissions, storage, keyboard/controller/touch, microphone, audio, background behavior, and later VPN-service decisions for app-managed WireGuard.
