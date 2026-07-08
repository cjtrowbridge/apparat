---
plan_id: 2026-07-07-10-59-06_split-artifacts-last-run-log-code-size-governance
title: Split Artifacts Last Run Log Code Size Governance
summary: Split GUI/headless release artifacts, add runtime-specific last_run.log diagnostics, and enforce code-file line limits in the build.
status: past
created_at: 2026-07-07-10-59-06
---

# Split Artifacts Last Run Log Code Size Governance

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Correct release artifact layout.
  - [x] 1.1 Update the canonical artifact contract.
    - [x] 1.1.1 Change the release path from `releases/<goos>/<goarch>/latest[.exe]` to `releases/<goos>/<goarch>/<binary>/latest[.exe]`.
    - [x] 1.1.2 Define canonical binary names `apparat` for GUI and `apparatd` for headless.
    - [x] 1.1.3 Document that each binary has separate artifact, runtime, logs, and smoke-run expectations.
    - [x] 1.1.4 Update `README.md`, `ROADMAP.md`, `releases/README.md`, and relevant docs to use the split layout.
  - [x] 1.2 Update build pipeline paths.
    - [x] 1.2.1 Update `scripts/build.py` artifact paths to include target subdirectories.
    - [x] 1.2.2 Add a build mode that can build both `apparat` and `apparatd`.
    - [x] 1.2.3 Keep `--target apparat` and `--target apparatd` for single-target builds.
    - [x] 1.2.4 Update `--print-path` to print the correct target-specific path.
    - [x] 1.2.5 Update `Makefile` so `make build` produces both binaries.
    - [x] 1.2.6 Update `make run-built` to run the GUI artifact explicitly.
    - [x] 1.2.7 Add a `make run-built-headless` target to run the headless artifact explicitly.
  - [x] 1.3 Keep generated artifacts out of Git.
    - [x] 1.3.1 Update `.gitignore` for `releases/*/*/*/latest`.
    - [x] 1.3.2 Update `.gitignore` for `releases/*/*/*/latest.exe`.
    - [x] 1.3.3 Verify no generated GUI or headless binary is staged.

- [x] 2. Add build tests for split artifacts.
  - [x] 2.1 Update Python build tests.
    - [x] 2.1.1 Verify Linux GUI path is `releases/linux/<arch>/apparat/latest`.
    - [x] 2.1.2 Verify Linux headless path is `releases/linux/<arch>/apparatd/latest`.
    - [x] 2.1.3 Verify Windows GUI path is `releases/windows/<arch>/apparat/latest.exe`.
    - [x] 2.1.4 Verify Windows headless path is `releases/windows/<arch>/apparatd/latest.exe`.
    - [x] 2.1.5 Verify multi-target build plans both artifacts without conflating outputs.
  - [?] 2.2 Add smoke-run tests where feasible.
    - [?] 2.2.1 Verify `make build` creates both target-specific artifacts.
    - [?] 2.2.2 Verify GUI smoke test runs from the GUI artifact.
    - [x] 2.2.3 Verify headless smoke test runs from the headless artifact.

- [x] 3. Add `last_run.log` runtime diagnostics.
  - [x] 3.1 Define last-run logging contract.
    - [x] 3.1.1 Create `last_run.log` in the runtime root or configured log directory at every process start.
    - [x] 3.1.2 Delete or truncate any prior `last_run.log` before new startup diagnostics are written.
    - [x] 3.1.3 Keep append-only durable JSONL logs separate from reset-on-start `last_run.log`.
    - [x] 3.1.4 Include runtime root, binary name, mode, process ID, build target, version/commit if available, OS, architecture, Go version, config source, and command-line flags.
    - [x] 3.1.5 Log every component startup, success, failure, warning, assertion, panic recovery, shutdown signal, and clean/unclean exit state.
    - [x] 3.1.6 Redact secrets, private keys, tokens, passphrases, prompts, model outputs, raw voice data, and sensitive payloads.
  - [x] 3.2 Implement last-run logger.
    - [x] 3.2.1 Add a logging API for reset-on-start human-readable diagnostic logs.
    - [x] 3.2.2 Wire `cmd/apparat` to create GUI-specific `last_run.log`.
    - [x] 3.2.3 Wire `cmd/apparatd` to create headless-specific `last_run.log`.
    - [x] 3.2.4 Ensure GUI and headless runtime roots differ when run from split artifacts unless explicitly overridden.
    - [x] 3.2.5 Capture panic details before process exit.
    - [x] 3.2.6 Capture doctor and smoke-test results in `last_run.log`.
  - [x] 3.3 Add last-run tests.
    - [x] 3.3.1 Verify stale `last_run.log` content is removed at process start.
    - [x] 3.3.2 Verify required startup fields are present.
    - [x] 3.3.3 Verify component success/failure entries are present.
    - [x] 3.3.4 Verify sensitive fields are redacted.
    - [x] 3.3.5 Verify GUI and headless logs are separate under separate runtime roots.

- [x] 4. Add code-size governance.
  - [x] 4.1 Define source-file line-count policy.
    - [x] 4.1.1 Require all code files to be fewer than or equal to 400 lines.
    - [x] 4.1.2 Define code files as tracked source files such as `.go`, `.py`, `.sh`, `.yaml`, `.yml`, and `.json` where applicable.
    - [x] 4.1.3 Exclude generated files, vendored/reference checkouts, `third_party/`, `.tools/`, release artifacts, `.git/`, and plan/journal/README prose.
    - [x] 4.1.4 Require any over-limit file to be decomposed and documented in that directory before the build can pass.
    - [x] 4.1.5 Document the policy in `AGENTS.md`, `README.md` if human-facing, and relevant package READMEs if decomposition happens.
  - [x] 4.2 Implement line-count checker.
    - [x] 4.2.1 Add `scripts/check_code_file_lines.py`.
    - [x] 4.2.2 Count physical lines in included code files.
    - [x] 4.2.3 Report each violating file with line count and limit.
    - [x] 4.2.4 Exit nonzero on any violation.
    - [x] 4.2.5 Support a `--limit` option defaulting to `400`.
  - [x] 4.3 Wire governance into build.
    - [x] 4.3.1 Add `make check-code-size`.
    - [x] 4.3.2 Add code-size checking to `make verify`.
    - [x] 4.3.3 Add tests for the line-count checker.
    - [x] 4.3.4 Fix or decompose any existing over-limit code files before final verification.

- [x] 5. Update runtime and build documentation.
  - [x] 5.1 Update user-facing docs.
    - [x] 5.1.1 Update `README.md` with split artifact paths.
    - [x] 5.1.2 Update `README.md` with GUI/headless run commands.
    - [x] 5.1.3 Update `README.md` with `last_run.log` location and purpose.
    - [x] 5.1.4 Update `README.md` with code-size verification expectations if useful to contributors.
  - [x] 5.2 Update roadmap and architecture docs.
    - [x] 5.2.1 Update `ROADMAP.md` Phase 3 tasks with split artifacts and `last_run.log`.
    - [x] 5.2.2 Update `docs/platform-matrix.md` with target-specific artifact paths.
    - [x] 5.2.3 Update `docs/architecture.md` with separate GUI/headless artifact/runtime/log boundaries.
    - [x] 5.2.4 Update `docs/database.md` or runtime docs only if runtime path semantics change persistence behavior.
  - [x] 5.3 Update operational records.
    - [x] 5.3.1 Append the implementation checkpoint to today's journal.
    - [x] 5.3.2 Regenerate plan indexes.

- [?] 6. Verify, build, commit, and push.
  - [?] 6.1 Run required checks.
    - [x] 6.1.1 Run `make fmt`.
    - [x] 6.1.2 Run `make test`.
    - [x] 6.1.3 Run `make test-race`.
    - [x] 6.1.4 Run `make test-build`.
    - [x] 6.1.5 Run `make check-code-size`.
    - [x] 6.1.6 Run `make lint`.
    - [x] 6.1.7 Run `make audit`.
    - [?] 6.1.8 Run `make build`.
    - [?] 6.1.9 Run `make run-built`.
    - [x] 6.1.10 Run `make run-built-headless`.
  - [?] 6.2 Validate artifacts and logs.
    - [?] 6.2.1 Confirm `releases/<goos>/<goarch>/apparat/latest[.exe]` exists and is ignored.
    - [x] 6.2.2 Confirm `releases/<goos>/<goarch>/apparatd/latest[.exe]` exists and is ignored.
    - [?] 6.2.3 Confirm GUI smoke run writes GUI `last_run.log`.
    - [x] 6.2.4 Confirm headless smoke run writes headless `last_run.log`.
  - [x] 6.3 Complete checkpoint.
    - [x] 6.3.1 Check pending downtime reports.
    - [x] 6.3.2 Confirm `third_party/salvagecore` is absent from staged changes.
    - [x] 6.3.3 Review final diff.
    - [x] 6.3.4 Commit if all checks pass.
    - [x] 6.3.5 Push the passing checkpoint to `origin`.

- [x] 7. Add documentation completeness governance.
  - [x] 7.1 Document build operations.
    - [x] 7.1.1 Expand the root `README.md` with build prerequisites, commands, outputs, smoke tests, native GUI dependencies, and troubleshooting.
    - [x] 7.1.2 Add `scripts/README.md` with script inventory, canonical invocations, side effects, and failure modes.
    - [x] 7.1.3 Add README inventories for current code and test directories that contain tracked source files.
  - [x] 7.2 Improve script self-documentation.
    - [x] 7.2.1 Make `scripts/build.py --help` explain prerequisites, canonical invocation, outputs, and common dependency failures.
    - [x] 7.2.2 Add a root `build.py` compatibility wrapper so `python3 build.py` explains and forwards to the canonical script.
  - [x] 7.3 Add automated documentation checks.
    - [x] 7.3.1 Add `scripts/check_directory_docs.py`.
    - [x] 7.3.2 Require tracked code/script/test directories to have a local `README.md`.
    - [x] 7.3.3 Require every tracked Python script in `scripts/` to be mentioned in `scripts/README.md`.
    - [x] 7.3.4 Require executable Python scripts to respond to `--help`.
    - [x] 7.3.5 Wire the check into `make verify`.
    - [x] 7.3.6 Add tests for the documentation checker.
  - [x] 7.4 Record the governance decision.
    - [x] 7.4.1 Update `AGENTS.md` with the every-file-or-feature documentation rule.
    - [x] 7.4.2 Update `ROADMAP.md` with the documentation completeness gate.
    - [x] 7.4.3 Append the checkpoint to today's journal.

## Approval Notes

- This plan intentionally treats the current immediate-exit behavior as a build-artifact/runtime-boundary problem: the GUI and headless binaries must no longer overwrite the same `latest` artifact.
- Generated binaries remain ignored; other devices should pull source and run the build pipeline to produce local artifacts.
- `last_run.log` is reset on every process start for immediate debugging; append-only JSONL logs remain the durable audit/debug stream.
- The 400-line policy is a build gate for code files, not prose documentation or operational logs.

## Execution Notes

- `make verify` passed after split artifacts, `last_run.log`, and code-size governance were implemented.
- `scripts/build.py --target apparatd` produced `releases/linux/amd64/apparatd/latest`, and `make run-built-headless` passed using `/tmp/apparat-run-built/apparatd`.
- `make build` and the GUI artifact checks are validation-blocked on this host because Ebitengine's Linux GUI build requires native desktop headers; compilation currently fails on missing `X11/Xlib.h`.
- Attempted `sudo apt-get update && sudo apt-get install -y libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libgl1-mesa-dev libxxf86vm-dev libasound2-dev`, but sudo requires an interactive password in this session.
