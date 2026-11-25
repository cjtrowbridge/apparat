# `apparatd`

`cmd/apparatd` is the headless worker and service entrypoint.

It loads configuration with binary name `apparatd`, creates the shared application runtime, supports `--smoke-test` and `--doctor`, and runs without initializing Ebitengine.

At startup it prints the selected runtime root and `last_run.log` path before entering the service loop. By default on Linux, the headless runtime root is `~/.local/share/apparat/node` unless `XDG_DATA_HOME`, `APPARAT_RUNTIME_DIR`, or `--runtime-dir` overrides it.

The release pipeline builds this command without the `gui` build tag into:

```text
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

Use this binary for display-free devices, service-manager integration, and headless smoke validation.

## Source inventory

- `main.go`: Parses command options and runs the headless worker.

Phase 8 adds `--demo-services`, `--services`, `--backup`, and `--restore-backup`. GUI/headless are alternative owners of the same locked node root; see root README for migration and recovery. On Linux, starting an unconfigured worker automatically saves a scoped grant for its first eligible LAN and publishes the worker in local discovery. An operator can disable or revoke that grant persistently. Linux headless configuration uses api-list, api-grant RULE, api-revoke RULE and api-disable.

- `updates.go`: Coalesces request and owner signals, verifies the Linux worker release and initiates graceful restart.
- `identity.go`: Provides offline owner setup with `apparatd identity-setup [--runtime-dir DIR]`, a controlling terminal and hidden password entry. Stop the node before using it. A fresh standalone Linux service creates its owner-only unattended identity. An existing database without a key requires an explicit recovery or creation choice.

A returning headless installation with an existing database and missing key waits for operator recovery. After explicitly deciding to generate a new worker key, stop the service, run `apparatd identity-setup --generate-if-missing` as the service account with its normal node root and restart the service. The command holds the node lock, refuses an existing or inconsistent identity and prints the new public fingerprint. The owner-only unencrypted worker key permits unattended startup.

Run `apparatd trust-host` in the service account’s environment to inspect the active public device fingerprint. With one personal fingerprint argument it confirms the intended owner in the worker database. On a claimed worker it only accepts the original owner and worker key. Use the same APPARAT_RUNTIME_DIR as the service. A mismatch or missing active manifest fails without changing ownership.

Use `apparatd workspace-access grant|revoke PROJECT_ID USER_FINGERPRINT SCOPE` on the owner to manage task and terminal grants. Approved membership supplies file reads and cluster administration supplies file writes. Use `workspace-access status PROJECT_ID USER_FINGERPRINT` to inspect effective access. It can run alongside the service. Scopes and required combinations are described in the [workspace guide](../../docs/workspaces.md).

- `chat.go`: interactive owner-local terminal client for the shared chat harness. It registers the selected Git checkout and owns the runtime root for the duration of the conversation. Set APPARAT_RUNTIME_DIR to an unused root when another node owns the default root. The chat endpoint and model are explicit; set APPARAT_CHAT_API_KEY only when required. Each submitted line and consequential tool exchange is recorded in the checkout.

For an Ollama service with a tool-capable model installed, use its Chat Completions endpoint:

```sh
export APPARAT_CHAT_ENDPOINT=http://127.0.0.1:11434/v1/chat/completions
export APPARAT_CHAT_MODEL=your-installed-model
apparatd chat --repo . --topic "Release planning"
```

Enter a line and press Return. Use /quit to detach. List transcripts with `apparatd chat --repo . --list` and reopen one with `apparatd chat --repo . --session SESSION_ID`. Use --global with the checkout chosen to own a new global transcript. Session listing works offline. An uncertain delivery stays pending, and /cancel abandons it locally. A cloud configuration sends conversation history and selected tool results to its configured endpoint. Use an approved provider and project for that route.

- `chat_test.go`: verifies the terminal entrypoint, actual JSON/HTTP tool round trip, session listing and resumed conversation against a local fixture.

The installer calls `user-fingerprint` locally to verify the stored signed user profile and `trust-host --install OWNER_FINGERPRINT` on the worker to reserve its first login. A repeat with the same admin succeeds. A conflicting reservation stops the installer.
