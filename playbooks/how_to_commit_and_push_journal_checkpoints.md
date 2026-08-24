# Playbook: Commit and Push Journal Checkpoints

*Status: Stable*

## Objective

Commit a reviewed journal checkpoint locally and push it only after separate explicit user approval.

## Prerequisites

- Git is available and `origin` is configured.
- The journal and governing plan reflect the completed checkpoint.
- Load `references/conversation_checkpoint_commits.md` for checkpoint scope and summary guidance.

## Procedure

1. Inspect `git status -sb`, the staged diff, and untracked files. Do not assume an untracked file belongs in the commit.
2. Confirm the journal entry, active or archived plan path, checklist deltas, TODO state changes, and verification evidence agree with the actual diff.
3. Regenerate plan indexes when plan files moved or changed.
4. Summarize the exact staged scope and propose a task-scoped commit message.
5. Request explicit approval for the local commit unless the user already directly instructed the agent to commit this completed scope.
6. Commit only the reviewed scope and verify the resulting revision and worktree state.
7. Treat pushing as a separate external action. Push only after explicit user approval, regardless of whether the checkpoint is journal-only or mixed.

## Verification

- The commit contains the reviewed checkpoint and no unrelated files.
- Journal, plan, TODO, and verification evidence agree.
- The worktree is clean or contains only identified unrelated work.
- No push occurred without explicit user approval.
