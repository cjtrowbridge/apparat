---
plan_id: 2026-11-20-08-49-00_left-trigger-ptt-feedback
title: Bind left trigger to PTT with matching button feedback
summary: Replace R2 with L2 and show the same pressed PTT appearance for controller, pointer and keyboard holds.
status: current
created_at: 2026-11-20-08-49-00
---

# Bind left trigger to PTT with matching button feedback

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

User explicitly requests L2 instead of R2 and matching on-screen PTT pressed
feedback while the trigger is held. Bind to ROADMAP Phase 9B dictation and
Phase 17 Steam Deck input equivalence. README/controller contract changes to L2;
completed historical roadmap milestones may retain their prior R2 binding with
an explicit supersession note. Preserve Settings last, silent untargeted input,
cancel/release behavior, native editor targeting and exclusive STMPCK game input.

Use the existing standard-gamepad and PTT overlay owners. Track L2 input,
derive the button's pressed appearance from existing held-source flags, and
release trigger state if the first standard controller disappears. Do not add
another capture system or change inference/audio behavior. Existing Linux build,
publication and Deck delivery authorization applies; Android stays at its
previous independent qualification while the user is away from the Pixel.

## Work and acceptance

- [x] 1. Remap controller PTT to standard left trigger, update live binding labels/docs, and share pressed button appearance/visibility across held inputs.
- [ ] 2. Verify matching pointer/trigger feedback, release/disconnect, binding contract and existing GUI regressions; run source/size/docs checks and canonical no-flag Linux build.
- [ ] 3. Update the Deck through its existing Linux updater, preserve node/Steam identity, and record rendered evidence plus any physical controller validation limit. Archive the plan and append the journal.

## Evidence and stopping condition

Initial checkout and bootstrap inspection passed; only the user's TODO edit is
present. Baseline is published `9ce8f1e`, Deck Apparat PID 5466, native 1280x800.
Current code samples FrontBottomRight and starts voice independently of the
pointer-only `pttHeld` visual flag. Change only these owners and their documented
contracts. Stop after qualified Linux delivery and confirmed installed bytes;
distinguish synthetic input/renderer checks from an operator's physical L2 trial.

Initial verification: GUI/HUD race tests pass under Xvfb, source governance
198 files / zero failures, physical 400-line and docs checks pass. The default
binding, live labels and current README/controller contract now use L2. The
button renderer reads shared held-source state, and losing the selected standard
controller cancels its capture. The Deck permits non-root writes to its Steam
virtual Xbox controller (event9), so a bounded synthetic L2/R2 hold can verify the
actual installed input mapping and visible feedback without microphone capture.

### Publication approval boundary

Implementation and GUI/HUD race tests are complete; source, physical line,
documentation, index and whitespace checks pass. Reviewed changes are staged.
The Deck input inspection confirms Apparat reads event9 and both trigger axes
range from 0 to 255, permitting later bounded synthetic held-trigger checks.

Automatic approval review rejected the no-flag build twice. The second review
explicitly stated that the earlier “no just do it” covered the prior Linux
delivery and did not clearly extend to a new commit, publication on main and
Deck deployment for this L2 request. No rejected build was executed. Do not
bypass through an alternate build or delivery path. Ask for explicit build,
publication and Deck installation approval; then run the same canonical pipeline
and complete actual trigger/visual verification. Keep this plan current pending
that approval, and preserve the original user TODO changes unstaged.
