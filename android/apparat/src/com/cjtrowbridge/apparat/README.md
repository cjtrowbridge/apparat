# apparat

Host-owned source for android/apparat/src/com/cjtrowbridge/apparat.

## Source inventory

- `MainActivity.java`: Hosts the Android Ebitengine view inside modern system-bar/cutout safe bounds and provides bounded APK-update UI.
- `UpdateApkProvider.java`: Exposes the cached update APK to Android package installation via a content provider.

Android safe-area color comes from the GUI renderer through `BackgroundARGB` in the mobile bridge. The current fixed dark palette is authoritative; runtime theme switching is not implemented. The native parent and navigation background use that color, with light navigation icons and no automatic contrast scrim. Insets and touch bounds are unchanged. Verify matching background pixels and readable buttons in portrait/landscape on an authorized device; system/OEM styling on other devices requires separate evidence.

Updater rejection codes distinguish invalid package, mismatched signer and non-increasing version. The signer is checked before version messaging; none of these paths opens the installer or clears app data.

- `UpdateTransfer.java`: Copies APK bytes with a 15-minute monotonic duration bound, 128 MiB limit and lifecycle cancellation; transport allows 60-second idle reads. Typed timeout/size/cancellation failures retain byte count and elapsed-time evidence without content or URLs. Partial APKs are removed on failure.

Regression verification: compile this helper with `tests/UpdateTransferTest.java` through the build gate and run `com.cjtrowbridge.apparat.UpdateTransferTest`. Controlled streams exercise slow progress beyond the old two-minute cap, idle timeout, duration, size and cancellation. Actual mobile network acceptance remains a separate device check.
- `ProjectDocuments.java`: owns the system folder picker, persisted grants, metadata reads and public clone export.
- `ApkHash.java`: Hashes update packages.

- `WorkspaceDocuments.java`: Resolves relative document names beneath the registered tree, queries children, and provides bounded-scope file descriptors and document operations. The host must support the requested operation.

- `DocumentEditor.java`: Implements Android field prompts and the clipboard bridge for the workspace.

- `VoiceInput.java`: Holds, cancels and finalizes local dictation in the original Android editor selection. Requests microphone permission on the first press. Sliding up locks capture. Sliding left hands the utterance to the global prompt on release and preserves the open editor for return to its tab.

The document and input dialogs anchor the circular PTT control to the lower-right editing area above the keyboard. A listening indicator remains visible alongside partial text until submission. A double tap on PTT or L2 locks dictation in the original selection. A subsequent tap finishes it.
- `AndroidUpdater.java`: Reports the installed version and starts the user-requested signed APK download.

- `ActionVoice.java`: Activity-owned local speech feedback and canvas microphone permission bridge.

- `LanBrowser.java`: browses Apparat service names on the active Wi-Fi network via Android 13+ scoped NSD. Leaving the detail or pausing the Activity releases discovery.
- `NodeKeys.java`: Presents startup key recovery, creation, protected entry and Android document pickers.
- `AndroidKeyProtection.java`: Uses authenticated AES-GCM operations in Android Keystore on eligible Android 11+ devices.
