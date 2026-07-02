---
plan_id: 2026-07-01-15-10-29_add-local-salvagecore-reference
title: Add Local Salvagecore Reference
summary: Create an ignored local salvagecore checkout with recursive dependencies and no tracked repository content.
status: past
created_at: 2026-07-01-15-10-29
---

# Add Local Salvagecore Reference

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Add the local salvagecore reference.
  - [x] 1.1 Protect the host repository from accidental tracking.
    - [x] 1.1.1 Add `/third_party/salvagecore/` to the root `.gitignore`.
  - [x] 1.2 Populate the ignored local reference.
    - [x] 1.2.1 Clone `cjtrowbridge/apparat-salvagecore` into `./third_party/salvagecore`.
    - [x] 1.2.2 Initialize and update all nested submodules recursively.
  - [x] 1.3 Record the repository checkpoint.
    - [x] 1.3.1 Append the local reference setup to `journal/2026-07-01.md`.

- [x] 2. Verify tracking boundaries.
  - [x] 2.1 Confirm the local checkout and nested submodules are present.
  - [x] 2.2 Confirm Git ignores `third_party/salvagecore` recursively.
  - [x] 2.3 Confirm no salvagecore file, gitlink, or `.gitmodules` entry is tracked or staged.
  - [x] 2.4 Regenerate plan indexes, review diff integrity, and archive this plan.
