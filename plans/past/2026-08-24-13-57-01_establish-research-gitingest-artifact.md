---
plan_id: 2026-08-24-13-57-01_establish-research-gitingest-artifact
title: Establish Research GitIngest Artifact
summary: Replace the full-submodule GitIngest default with a project-focused digest and reproducibility manifest, then regenerate and validate the current research snapshot.
status: past
created_at: 2026-08-24-13-57-01
---

# Establish Research GitIngest Artifact

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## 1. Define reproducible project-focused artifact behavior

- [x] Pin GitIngest to the locally validated 0.3.1 release rather than an open-ended version range.
- [x] Keep `gitingest.txt` as the default ignored output while adding an ignored `gitingest.manifest.json` sidecar.
- [x] Generate the default corpus without recursive submodule source and exclude generated release artifacts and tool/cache directories.
- [x] Offer an explicit `--include-submodules` compatibility option rather than including submodules by default.
- [x] Create the manifest from deterministic local Git state: root commit, branch, dirty state, sanitized remotes, generator version and arguments, recursive submodule status, module manifests, artifact paths, and SHA-256 hashes.
- [x] Write the digest and manifest atomically so a failed generation does not replace a previous usable snapshot.

## 2. Document and test the research artifact contract

- [x] Update generator help and `scripts/README.md` with the default corpus boundary, output paths, provenance fields, explicit submodule option, prerequisites, side effects, and failure handling.
- [x] Add focused tests for pinned installation, default filtering, opt-in submodule behavior, manifest content, and output hashing without requiring a network or a real GitIngest execution.

## 3. Regenerate and verify the current snapshot

- [x] Promote this plan to `plans/current/` immediately before implementation edits.
- [x] Run focused Python tests, code-size and documentation checks, plan-index validation, and generator help validation.
- [x] Generate the default digest and manifest from the current repository snapshot.
- [x] Verify the digest does not contain recursive third-party source, preserves project-owned evidence, is materially smaller than the prior 189 MB artifact, and hashes match the manifest.
- [x] Review the diff and working tree for generated artifacts, secrets, unrelated changes, and untracked research output.
- [x] Record the checkpoint in today's journal, archive the plan, refresh indexes, and commit locally without pushing.
