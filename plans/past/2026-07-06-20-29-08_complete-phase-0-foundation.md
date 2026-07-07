---
plan_id: 2026-07-06-20-29-08_complete-phase-0-foundation
title: Complete Phase 0 Foundation
summary: Finish the Phase 0 repository, Go workspace, dependency, governance, documentation, and verification foundation.
status: past
created_at: 2026-07-06-20-29-08
---

# Complete Phase 0 Foundation

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Complete Phase 0 dependency foundation.
  - [x] 1.1 Pin active Go module dependencies.
    - [x] 1.1.1 Create the root Go module using the approved module path.
    - [x] 1.1.2 Pin `modernc.org/sqlite` in `go.mod` and `go.sum`.
    - [x] 1.1.3 Pin active Ebitengine and EbitenUI dependencies in `go.mod` and `go.sum`.
    - [x] 1.1.4 Decide whether `debugui` is pinned in `go.mod` now or remains source-reference only.
    - [x] 1.1.5 Keep WireGuard, llama.cpp, and whisper.cpp source trees reference-only unless a later approved plan activates adapters.
  - [x] 1.2 Record deferred source admission gates.
    - [x] 1.2.1 Record the Qwen3-TTS admission gate in `ROADMAP.md` and third-party documentation.
    - [x] 1.2.2 Record the Meshtastic admission gate in `ROADMAP.md` and third-party documentation.
    - [x] 1.2.3 Record the Signal gateway admission gate in `ROADMAP.md` and third-party documentation.
    - [x] 1.2.4 Record the BOINC source admission gate in `ROADMAP.md` and third-party documentation.
    - [x] 1.2.5 Record the general alternative-runtime admission gate in `ROADMAP.md` and third-party documentation.
  - [x] 1.3 Record MVP source exclusions.
    - [x] 1.3.1 Record qTox, TokTok qTox, and go-toxcore-c exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.2 Record Tor exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.3 Record WebRTC exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.4 Record curl exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.5 Record OpenSSL and libsodium exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.6 Record that OpenSSL does not supply PGP semantics in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.7 Record Qwen3-ASR exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.8 Record `golang/mobile` source-checkout exclusion in `ROADMAP.md` and third-party documentation.
    - [x] 1.3.9 Record termframe exclusion in `ROADMAP.md` and third-party documentation.

- [x] 2. Establish the Go application workspace.
  - [x] 2.1 Add source layout boundaries.
    - [x] 2.1.1 Create the approved application package skeleton.
    - [x] 2.1.2 Add grouping README files for created application package directories.
    - [x] 2.1.3 Ensure application imports cannot depend on `third_party/salvagecore`.
  - [x] 2.2 Add baseline executable entry points.
    - [x] 2.2.1 Add an approved GUI command stub.
    - [x] 2.2.2 Add an approved headless command stub.
    - [x] 2.2.3 Add an approved shared runtime package stub.
  - [x] 2.3 Add reproducible commands.
    - [x] 2.3.1 Add an approved command runner file if one is desired.
    - [x] 2.3.2 Add formatting command coverage.
    - [x] 2.3.3 Add unit-test command coverage.
    - [x] 2.3.4 Add race-test command coverage.
    - [x] 2.3.5 Add lint command coverage once the linter is approved.
    - [x] 2.3.6 Add dependency-audit command coverage once the audit tool is approved.
  - [x] 2.4 Pin build and development tools.
    - [x] 2.4.1 Record the approved Go version.
    - [x] 2.4.2 Record the approved Ebitengine version.
    - [x] 2.4.3 Record approved lint and audit tool versions if tools are selected for Phase 0.
    - [x] 2.4.4 Keep source-reference submodule pins separate from active Go dependency pins.

- [x] 3. Establish application governance.
  - [x] 3.1 Define module and package boundaries.
    - [x] 3.1.1 Document `cmd/`, `internal/app/`, `internal/domain/`, `internal/adapters/`, and `internal/platform/` boundaries if approved.
    - [x] 3.1.2 Document package dependency direction and forbidden imports.
    - [x] 3.1.3 Document GUI/headless shared-runtime boundaries.
  - [x] 3.2 Define decomposition expectations.
    - [x] 3.2.1 Document file-size and package-size review thresholds.
    - [x] 3.2.2 Document when code should split into subpackages or modules.
  - [x] 3.3 Define logging and redaction requirements.
    - [x] 3.3.1 Document structured logging baseline.
    - [x] 3.3.2 Document sensitive-field redaction baseline.
    - [x] 3.3.3 Document command, event, correlation, and diagnostic logging expectations.
  - [x] 3.4 Define documentation synchronization requirements.
    - [x] 3.4.1 Document when README, ROADMAP, package READMEs, and plans must be updated.
    - [x] 3.4.2 Keep human-facing product material out of agent-facing files and agent-facing policy out of README.

- [x] 4. Verify Phase 0 completion.
  - [x] 4.1 Validate repository state.
    - [x] 4.1.1 Run `git status --short --untracked-files=all`.
    - [x] 4.1.2 Confirm `third_party/salvagecore` remains ignored and absent from staged changes.
    - [x] 4.1.3 Run `git submodule status --recursive`.
  - [x] 4.2 Validate Go workspace.
    - [x] 4.2.1 Run the approved formatting command.
    - [x] 4.2.2 Run the approved unit-test command.
    - [x] 4.2.3 Run the approved race-test command if feasible for the stub workspace.
    - [x] 4.2.4 Run the approved lint command if a linter is selected.
    - [x] 4.2.5 Run the approved dependency-audit command if an audit tool is selected.
  - [x] 4.3 Validate documentation and plans.
    - [x] 4.3.1 Update `ROADMAP.md` Phase 0 statuses to match completed work.
    - [x] 4.3.2 Update `README.md` if user-facing setup, commands, or structure changed.
    - [x] 4.3.3 Update `AGENTS.md` or package README files if agent-facing governance or structure changed.
    - [x] 4.3.4 Append the checkpoint to today's journal.
    - [x] 4.3.5 Regenerate plan indexes.
    - [x] 4.3.6 Check pending downtime reports before final summary.
  - [x] 4.4 Complete checkpoint.
    - [x] 4.4.1 Review the final diff.
    - [x] 4.4.2 Propose a commit message.
    - [x] 4.4.3 Commit after approval.

## Resolved Questions

1. Root module path is `github.com/cjtrowbridge/apparat`.
2. Command stubs are `cmd/apparat` and `cmd/apparatd`.
3. Initial package skeleton is `internal/app`, `internal/domain`, `internal/adapters`, and `internal/platform`.
4. `debugui` remains source-reference only while Apparat stays on stable Ebitengine 2.9.x.
5. Command runner is `Makefile`.
6. Go toolchain baseline is `1.26.4`.
7. Pinned lint and audit tools are `golangci-lint v2.12.2` and `govulncheck v1.5.0`.
