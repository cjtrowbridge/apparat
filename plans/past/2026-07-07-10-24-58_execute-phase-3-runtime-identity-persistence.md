---
plan_id: 2026-07-07-10-24-58_execute-phase-3-runtime-identity-persistence
title: Execute Phase 3 Runtime Identity Persistence
summary: Implement Phase 3 local runtime modes, configuration, structured logs, SQLite lifecycle, identity, cluster directory, local messaging primitives, verification, build, commit, and push.
status: past
created_at: 2026-07-07-10-24-58
---

# Execute Phase 3 Runtime Identity Persistence

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Normalize Phase 3 runtime entrypoints.
  - [x] 1.1 Update GUI runtime execution.
    - [?] 1.1.1 Make `cmd/apparat` default execution enter a real Ebitengine run loop.
      - Ebitengine run-loop code is implemented behind the `gui` build tag; native desktop-library validation remains pending. Default builds use the non-window blocking runtime path for CI/headless safety.
    - [x] 1.1.2 Keep `cmd/apparat --smoke-test` as the non-window build and CI verification path.
    - [x] 1.1.3 Keep GUI startup errors visible through structured logging and process exit status.
  - [x] 1.2 Update headless runtime execution.
    - [x] 1.2.1 Keep Ebitengine initialization out of `cmd/apparatd`.
    - [x] 1.2.2 Add explicit `--mode=headless`, `--mode=auto`, and `--mode=gui` parsing where applicable.
    - [x] 1.2.3 Add clean context cancellation and signal handling for headless mode.
  - [x] 1.3 Add startup diagnostics and doctor mode.
    - [x] 1.3.1 Add a doctor command path that validates runtime directories, logging, database, identity, and cluster directory state.
    - [x] 1.3.2 Add startup diagnostics to smoke-test output.

- [x] 2. Add runtime configuration and directories.
  - [x] 2.1 Implement configuration package.
    - [x] 2.1.1 Add config precedence for defaults, environment variables, and explicit CLI flags.
    - [x] 2.1.2 Define runtime root outside the source tree by default.
    - [x] 2.1.3 Define database, logs, identity, cache, artifacts, backups, and recovery paths.
  - [x] 2.2 Add tests for runtime path selection.
    - [x] 2.2.1 Verify default paths do not use the repository source tree.
    - [x] 2.2.2 Verify environment and CLI override precedence.

- [x] 3. Add structured logging.
  - [x] 3.1 Implement append-only JSONL logging.
    - [x] 3.1.1 Add log file opening under the configured logs directory.
    - [x] 3.1.2 Include component, event, device, project, job, task, and correlation fields when supplied.
    - [x] 3.1.3 Add sensitive-value redaction before logging.
    - [x] 3.1.4 Add safe size-based log rotation and retention.
  - [x] 3.2 Add logging tests.
    - [x] 3.2.1 Verify JSONL output.
    - [x] 3.2.2 Verify redaction of tokens, private keys, prompts, model output, and voice markers.
    - [x] 3.2.3 Verify rotation retains bounded files.

- [x] 4. Add SQLite lifecycle.
  - [x] 4.1 Implement database package.
    - [x] 4.1.1 Open, close, ping, and configure SQLite connections.
    - [x] 4.1.2 Enable foreign keys.
    - [x] 4.1.3 Keep WAL opt-in until platform validation is complete.
    - [x] 4.1.4 Add forward migrations with checksums.
    - [x] 4.1.5 Add ULID-like sortable identifiers and UTC millisecond timestamp helpers.
    - [x] 4.1.6 Add repository interfaces that do not leak SQL into the HUD.
    - [x] 4.1.7 Add read-only database diagnostics.
  - [x] 4.2 Add SQLite tests.
    - [x] 4.2.1 Verify open, ping, foreign keys, migrations, and close.
    - [x] 4.2.2 Verify migration checksum mismatch is detected.
    - [x] 4.2.3 Verify read-only diagnostics.

- [x] 5. Add local identity.
  - [x] 5.1 Implement identity package.
    - [x] 5.1.1 Generate and import user Ed25519 identity.
    - [x] 5.1.2 Generate device Ed25519 identity.
    - [x] 5.1.3 Sign device authorization using user identity.
    - [x] 5.1.4 Encrypt private-key files with Argon2id and XChaCha20-Poly1305.
    - [x] 5.1.5 Create public manifests and identity metadata.
    - [x] 5.1.6 Add startup consistency classification.
    - [x] 5.1.7 Add doctor, repair, rotation, revocation, and archived reset primitives.
  - [x] 5.2 Add identity tests.
    - [x] 5.2.1 Verify identity generation and import.
    - [x] 5.2.2 Verify signature and authorization validation.
    - [x] 5.2.3 Verify encrypted private-key round trip and wrong-passphrase failure.
    - [x] 5.2.4 Verify diagnostics and archived reset behavior.

- [x] 6. Add local cluster directory.
  - [x] 6.1 Implement cluster directory package.
    - [x] 6.1.1 Store signed device profiles.
    - [x] 6.1.2 Store roles, permissions, endpoints, certificate fingerprints, WireGuard keys, and typed workload capabilities.
    - [x] 6.1.3 Store capability runtime/provider, models or research projects, modalities, limits, hardware, queue eligibility, health, and policy constraints.
    - [x] 6.1.4 Store last-seen and reachability state.
    - [x] 6.1.5 Add change feeds and sync cursors.
  - [x] 6.2 Add cluster directory tests.
    - [x] 6.2.1 Verify signed profile persistence.
    - [x] 6.2.2 Verify typed capability persistence.
    - [x] 6.2.3 Verify change feed cursor progression.

- [x] 7. Add durable local messaging primitives.
  - [x] 7.1 Implement messaging package.
    - [x] 7.1.1 Add outbox repository.
    - [x] 7.1.2 Add inbox repository.
    - [x] 7.1.3 Add replay and duplicate tracking.
    - [x] 7.1.4 Add event cursor state.
    - [x] 7.1.5 Add bounded retry scheduling.
  - [x] 7.2 Add messaging tests.
    - [x] 7.2.1 Verify outbox and inbox persistence.
    - [x] 7.2.2 Verify duplicate and replay rejection.
    - [x] 7.2.3 Verify bounded retry schedule progression.

- [x] 8. Integrate runtime startup.
  - [x] 8.1 Wire runtime components together.
    - [x] 8.1.1 Initialize config, directories, logging, SQLite, identity status, cluster directory, and messaging repositories in shared runtime startup.
    - [x] 8.1.2 Ensure GUI and headless modes share local runtime behavior.
    - [x] 8.1.3 Ensure headless mode does not import or initialize Ebitengine.
  - [x] 8.2 Add startup tests.
    - [x] 8.2.1 Verify GUI smoke path initializes shared runtime without opening a window.
    - [x] 8.2.2 Verify headless smoke path initializes shared runtime without Ebitengine.
    - [x] 8.2.3 Verify doctor reports healthy local runtime after initialization.

- [x] 9. Update documentation and roadmap.
  - [x] 9.1 Update human-facing docs.
    - [x] 9.1.1 Update `README.md` with Phase 3 runtime behavior, doctor mode, runtime paths, and local persistence notes.
    - [x] 9.1.2 Update `docs/database.md` with implemented migration and diagnostic decisions.
    - [x] 9.1.3 Update `docs/security.md` with implemented local identity and encrypted-key decisions.
    - [x] 9.1.4 Update `docs/platform-matrix.md` with GUI run-loop and headless runtime behavior.
  - [x] 9.2 Update operational docs.
    - [x] 9.2.1 Update package README files if package boundaries change.
    - [x] 9.2.2 Mark Phase 3 `ROADMAP.md` checklist items complete only after implementation and verification evidence exists.
    - [x] 9.2.3 Append today's journal checkpoint.

- [x] 10. Verify, build, commit, and push.
  - [x] 10.1 Run required verification.
    - [x] 10.1.1 Run formatting.
    - [x] 10.1.2 Run unit tests.
    - [x] 10.1.3 Run race tests if feasible.
    - [x] 10.1.4 Run build-pipeline tests.
    - [x] 10.1.5 Run lint.
    - [x] 10.1.6 Run vulnerability audit.
    - [x] 10.1.7 Run `make build` and `make run-built`.
  - [x] 10.2 Check repository safety.
    - [x] 10.2.1 Confirm generated binaries remain ignored.
    - [x] 10.2.2 Confirm `third_party/salvagecore` is absent from staged changes.
    - [x] 10.2.3 Check pending downtime reports.
    - [x] 10.2.4 Regenerate plan indexes.
  - [x] 10.3 Complete checkpoint.
    - [x] 10.3.1 Review final diff.
    - [x] 10.3.2 Commit if all required checks pass.
    - [x] 10.3.3 Push the passing checkpoint to `origin`.

## Approval Notes

- The user requested Phase 3 execution on 2026-07-07.
- The two currently uncommitted documentation edits to `ROADMAP.md` and `journal/2026-07-07.md` are treated as pre-existing Phase 3 clarification context and will be included in the Phase 3 checkpoint unless the user asks to split them.
- This plan uses Go standard crypto plus `golang.org/x/crypto` for Argon2id and XChaCha20-Poly1305 because Phase 3 explicitly requires those algorithms.
