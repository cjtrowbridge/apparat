---
plan_id: 2026-11-23-21-00-00_owner-agent-harness
title: Build the owner-hosted Go chat harness against the pinned Pi tool protocol
summary: Deliver repo-owned conversations, deterministic commands, durable inference waits and a scoped global PTT session with measured Pi compatibility.
status: current
created_at: 2026-11-23-21-00-00
---

# Owner-hosted agent harness

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Product and implementation boundary

A user on the Deck asks “How is project X going?” and gets its actual progress, decisions and blockers in one lookup. They can direct a project session, leave for the evening, and inspect the result on another device. A future maintainer can clone X and read why the work happened. On a saturated machine a session may wait hours for its first token. That is ordinary operation. The rotating gear means a reply is pending. The adjacent text tells the user whether the job is queued, generating, or awaiting a decision.

Build one Go harness inside the existing GUI and headless application core. Scope determines the available tools. Projects own conversations and effects. A pool schedules inference, and a provider computes a response. Each has a distinct durable identity. Reuse the internal API, existing workspace owner, Task handlers, SQLite, pool queues, REST and authenticated WebSocket. The user authorized starting with the shared core. The first checkpoint covers owner-local create, submit, inspect and complete operations for project and global scopes, durable Markdown and restart verification. UI integration follows separately.

### Shared-core checkpoint

- [x] Expose owner-local create, submit, inspect and complete operations on the shared Runtime. Project and global scopes use an explicitly selected registered filesystem repository. Persist versioned YAML and readable Markdown. Verify concurrent redelivery, mismatched payload rejection, restart with a pending turn, completion replay and external transcript damage. The governed application, domain and project Git package suites passed.
- [ ] Extend the header-based session directory with a tracked index, full-text search and legacy migration. The initial reader loads one requested session and publishes a complete revision per mutation. Measure long-history I/O before using this path for high-frequency updates.

### Provider and interface readiness checkpoint

Implement the owner-local provider handoff, correlated responses, read-only project lookup, session discovery and cancellation on the existing runtime. Verify a real HTTP protocol exchange against a local fixture, duplicate/concurrent dispatch, tool continuation and restart. Existing project/global HUDs and public cluster endpoints remain separate integration milestones.

- [x] Provide durable job claims, callback correlation, digest-checked request recovery, cancellation and an ordered model/tool loop through the shared Runtime.
- [x] Expose paged sessions, owner-scoped catalog queries, a zero-model list command and a terminal conversation client. The non-streaming Chat Completions adapter accepts an explicit endpoint, model and optional bearer credential.
- [x] Verify actual HTTP encoding and tool continuation against a loopback fixture, terminal resume, eight-way duplicate drive, restart, late callbacks, 100 simultaneously pending sessions and paged discovery. Live vendor/model qualification, cluster scheduling, authenticated remote session endpoints and GUI integration remain open.

## What the pinned Pi source actually does

Reference revision `cb1205e4661723c7a8c9fb01acd38b7691f97797` matches the gitlink and local checkout. Keep the pin with every compatibility fixture. The following paths are relative to [the pinned reference](../../third_party/inference/pi). Pi stays a source reference under its MIT license. Preserve the upstream license notice if fixture or documentation text is copied.

| Pinned source | Behavior to carry forward or adapt |
| --- | --- |
| [AI message types](../../third_party/inference/pi/packages/ai/src/types.ts) | An assistant message carries ordered content blocks. A tool call has a name, ID and JSON object arguments. A tool result has the corresponding call ID, tool name, typed content and error flag. Provider metadata and tool result details have distinct purposes. Preserve enough provider-native continuation state for the next request. |
| [Agent loop](../../third_party/inference/pi/packages/agent/src/agent-loop.ts) | The loop prepares context, streams an assistant response, executes complete calls, appends their results and continues. It supports steering and follow-ups, sequential or parallel batches, argument validation and execution updates. A length-truncated response fails its entire call batch. Use the same visible conversation grammar, with owner-persisted suspension points. |
| [Prompt builder](../../third_party/inference/pi/packages/coding-agent/src/core/system-prompt.ts) | Its default selection is read, bash, edit and write. Visible tools contribute brief descriptions and deduplicated rules. Tagged sections carry tools, rules, skills and working-directory context. Apparat can produce the same recognizable layout once per stable tool selection and supply changing project facts by tool. |
| [Built-in tool definitions](../../third_party/inference/pi/packages/coding-agent/src/core/tools) | The pin specifies one-based reads, multi-replacement edits against the original file, whole-file writes and bounded shell output. Use these definitions and [their tests](../../third_party/inference/pi/packages/coding-agent/test) as the compatibility oracle. |
| [Skill discovery](../../third_party/inference/pi/packages/coding-agent/src/core/skills.ts) | Pi presents name, description and location in an escaped skill listing, omits skills disabled for model invocation, and loads a matching skill on demand. Relative resources belong to the skill directory. The reader and reachable skills must agree. |
| [Cache warmer](../../third_party/inference/pi/packages/coding-agent/src/core/cache-warmer.ts) | Pi considers prior request size, provider cache lifetime and expected savings when refreshing an entry. Apparat should first measure shared cold-prefix prefill and backend-supported disk restore. One durable session need not generate its own warming traffic. |
| [Experimental durable harness](../../third_party/inference/pi/packages/durable/README.md) | The pin already describes persistent model turns, recovery, child tasks and watchers. Its interrupted-model-request retry policy needs adaptation for a cluster where a remote provider may still be computing. Reconcile by job and attempt identity before spending compute twice. |

Compatibility claim for v1: a tested subset of the pinned model-facing tool and skill conventions. Publish the supported subset per scope, backend and model. Pi package installation, JavaScript hooks, terminal UI, MCP, Code Mode and Pi session-file interchange add no capability required by the first two user stories. Revisit a specific interface when an actual consumer needs it.

## The minimum extensible surface

Use one compiled tool registry, one provider adapter contract, ordinary owner-hosted Tasks and declarative skill files. A tool definition contributes a stable name and version, short prompt snippet, usage rules, argument schema, result convention, required scope, effect class and Go handler. Build both the native provider declarations and the prompt's tool list from that definition. Fail startup validation for duplicate names or a declaration without a handler. An authorized feature can register a new typed operation without changing the turn loop. A new data-only skill requires no executable loader. A new provider implements message conversion, stream assembly, capability discovery and recovery. Keep extensions explicit in the composition root.

| Proposed tool | Model-visible contract | Execution boundary |
| --- | --- | --- |
| `projects(project?, needs_input?, cursor?)` | Without a selector, list visible projects. With a name or ID return owner, observation times, availability, sessions, running/completed work, results, blockers, pending input/approvals, recorded next steps and stable drill-down IDs. Page large lists. | One logical read. Use owner records and authorized snapshots. Return ambiguity and staleness explicitly. An offline owner is a valid result. |
| `read(path, offset?, limit?)` | One-based text lines and an explicit continuation when truncated. Read skills, transcript passages and permitted project files. Advertise image blocks only with tested support. | Resolve on the owning device. Limit absolute paths to approved roots and prevent traversal and symlink escapes. Global scope can read its own skill resources. |
| `edit(path, edits[])`, `write(path, content)` | For each edit supply oldText and newText. Match all replacements against the same original file, require unique disjoint matches, apply atomically and report a diff or a recoverable error. Whole-file write creates or replaces permitted files. | Project scope and write authority. Serialize competing edits to a path, check the opening revision, preserve encoding/line endings under the declared contract. Compare BOM, duplicates, partial matches, conflict and cancellation with upstream fixtures. |
| `search_chats(project, query, cursor?)` | Return matching descriptions and body passages, relative paths and stable turn references. | Plain-text/index search over permitted project transcripts. One read can open the original evidence. |
| `run_task(project, task, inputs)` | Use stable IDs from project lookup or a paged Task listing supplied by that lookup. Return the run ID, acceptance or rejection, and a verified outcome when available. | Existing project Task handler, authorization, revision and confirmation rules. The user or an authorized policy supplies approval. A model cannot approve itself. |
| `delegate(project, instruction, session?)` | Start a project conversation or send an instruction to a selected session. Return owner, child session ID, accepted turn ID and a link to follow it. | Same owner API as a direct submission. Check grants again on the child. Start with global-to-project delegation; nesting requires a real use case and bounded fan-out. |
| `bash(command, timeout?)` | Pi's shell contract includes the command string, optional seconds limit and bounded combined output. | Only include this tool in a separately authorized execution scope that can implement its semantics. Current remote workspace shell restrictions remain in force. Named Tasks cover the initial global operations. Scope-specific omissions and text-only reads must be visible in the compatibility report. |

Start project sessions with the tools they can use, global sessions with catalog, search, delegation and authorized Task operations. Add named destinations to existing action providers in the same feature change. One deterministic command binding gives a HUD action, PTT phrase and typed chat the same behavior. Tool/skill discovery reflects revocation and current scope. A skill can explain a procedure and refer to bundled materials; tool grants, credentials and automation registration stay under owner control.

## System prompt and request assembly

Build the stable prefix in deterministic order for each scope and tool revision. Use the Apparat role and Pi-style sections. Derive the text and the structured function schemas from the same compiled definitions. The following is a proposed excerpt for a global session:

```text
You operate Apparat for the signed-in user. Project owners report work and execute actions.
<tools>
- projects: List visible projects or retrieve one project's work, results and decisions.
- search_chats: Find earlier discussions by description or passage.
- delegate: Submit an instruction to a project session and return its receipt.
</tools>
<rules>
For project status, call projects once with the user's selector. If it is ambiguous, ask which project.
Use observation times in offline reports. Identify any decision the owner is waiting for.
Describe a delegated task as complete only when its owner reports completion.
When earlier reasoning matters, search project chats and open the relevant passage.
</rules>
```

Generate the actual excerpt from active tools (including its declared reader), not a hand-maintained second catalog. Each provider also receives a structured declaration, e.g. the name `projects`, a short description, and an object schema with optional string `project`, boolean `needs_input` and string `cursor`. Reject undeclared fields. Use strict schema mode when verified for the selected model. The model emits a native function call. XML sections teach usage; the executor consumes structured protocol fields. Neither code fences nor text resembling a tool call can trigger execution.

Stable skills may contribute a sorted name, description and logical location. Put changing project instructions and project-specific skill manifests after the common prefix, versioned per checkout. No cluster inventory, user names, timestamps, private transcript bodies or absolute host-specific paths in the reusable shared prefix. Load a skill body only when selected. Treat retrieved documents and tool outputs as data with provenance. On a scope change, regenerate the declared set and revalidate any pending call.

Measure the complete serialized request including schemas, selected history, tool results and completion reserve against the model's context window. Preserve the last complete call/result group across context reduction. Retain older exchanges verbatim in the repo and retrieve by reference. A compacted prompt is a disposable view. Record redacted component sizes, selected IDs and omission reasons. Use the smallest compatible configured context tier and a finite completion budget.

Fingerprint model build, tokenizer/chat template, tool order/schema, prompt bytes, backend settings and cache format. For a capable local backend, prefill and restore only this shared prefix from disk. End every constructed system prompt with the fixed priming instruction for “Ready for prompt.” Replay that exact assistant acknowledgement before live history. Include root AGENTS.md for project sessions. Measure physical cache reuse independently of this required conversation contract. Measure cold, warm and post-restart time to first useful call, actual cached tokens, disk cost and concurrent-session behavior. Hosted caching is provider controlled. A cache miss simply processes the prefix normally.

## Exactly how a turn runs

A user turn is one submitted instruction and its eventual answer or outcome. It may contain several model requests. One request can yield several tool calls. Store these identities separately. The caller's operation ID and the owner's durable receipt allow safe resubmission. Provider call IDs correlate model messages; they are insufficient as authorization or cross-request idempotency keys.

1. **Capture intent.** At PTT press capture the input role. A focused editor receives local dictation exclusively. On release only the final command transcript dispatches. Cancel discards it. Typed session input joins at submission. Preserve exact navigation, clarification for ambiguous/near commands, and no-match behavior for navigation. Provide an explicit conversational route when the intent is ambiguous. The HUD retains keyboard, D-pad, pointer and narrow/wide behavior.
2. **Accept once.** Resolve project, session and user on the owner, check current grant, assign sequence and turn ID, durably record the user entry and return its receipt. Repeated delivery with matching ID and payload retrieves that receipt. A reused key with different content is a conflict. Serialize conversation order per session. Later user messages may queue as follow-ups. Explicit steering is recorded for the next safe boundary and cancellation fences outstanding work.
3. **Try an exact operation.** Match registered phrases and typed parameter grammars before allocating a model request. Resolve an unambiguous target, recheck eligibility, and use the normal handler. Missing input produces a specific question. For a known status phrase one Go query renders the answer. For a known mutation use its normal confirmation and record the result. These paths use zero agent-model tokens, including classification, narration and description generation.
4. **Prepare a model step.** For a conversational instruction compose the packet described above. Resolve a route supporting the required tools, modalities, budgets, privacy and priority. Persist the model-step ID, packet identity, route and stable logical inference-job ID before submission. Store the private replay material under owner policy. A redacted manifest alone cannot reconstruct a request.
5. **Submit and suspend.** Durably submit the stable model-step ID to its inference pool. A worker proposes against the agreed roster and collects every confirmation before starting. Release the session runner. If acceptance was lost, look up the same job ID or resubmit with that key. If the result arrives before the subscription starts, the owner's durable state supplies it. An immediate cloud reply and an overnight local reply follow this same path.
6. **Assemble the assistant message.** Display optional transient text deltas. Collect tool argument fragments by provider call index or ID, including interleaved calls, and normalize only a completed response. Persist ordered text and tool-call blocks plus the finish reason. Reject incomplete JSON, duplicated or conflicting IDs and a length-truncated call batch before effects. A provider error, user cancellation and a legitimate final answer have distinct dispositions.
7. **Run the tools.** For each complete call, resolve a declared handler, validate arguments and current target, and persist an operation keyed by session, turn, model step and call identity. Recheck authorization and preconditions at execution, including after an approval wait. Bind any approval to the operation, target, inputs and revision. Execute on the owner. Reuse its existing request-ID/revision/receipt mechanism. Start with serial calls in declared order. Parallelize independent reads only if measured latency warrants it, preserving result order.
8. **Correlate every result.** Return one result per call, including unknown tool, denied, invalid, cancelled or skipped. Persist validated outputs and durable result/artifact references. Separate partial success and indeterminate effect from success. A crash after an effect but before receipt delivery triggers reconciliation against the owner. A tool with no safe reconciliation pauses for resolution. Retry budgets cover real transport, validation and provider failures. Queue age and provider silence consume no attempts.
9. **Continue or finish.** Append the complete assistant call message and its matched results in the target provider's native form. Preserve provider-required opaque continuation metadata privately according to route policy. Request another model step only when synthesis or further judgment is useful. A direct Task may finish with its own formatted result. An accepted child session can release the global conversation immediately and send a correlated update later. On a normal text finish, store the answer, advance the session revision once, publish a change hint and service the next queued turn. Bound call count, retries, cost and tokens separately from the unlimited elapsed wait.

One normalized exchange, with bookkeeping fields omitted for readability:

```json
[
  {"role":"user","content":"What still needs my decision in Atlas?"},
  {"role":"assistant","content":[{"type":"toolCall","id":"c1","name":"projects","arguments":{"project":"Atlas","needs_input":true}}]},
  {"role":"toolResult","toolCallId":"c1","toolName":"projects","isError":false,"content":[{"type":"text","text":"Atlas on worker-a is awaiting approval A7 for release run R9. The owner was observed at 18:06 UTC."}]},
  {"role":"assistant","content":[{"type":"text","text":"Atlas is waiting for your approval to publish run R9. Its worker confirmed that at 18:06 UTC."}]}
]
```

Keep typed project/run/approval references in the owner result for drill-down. Provider adapters encode this same exchange through the API they support, e.g. a function-call item and its correlated output, chat tool_calls and tool messages, or Ollama's native arguments and tool-name result. Maintain an internal mapping for backends that omit call IDs. Reject ambiguous associations. An invalid call receives a safe error result such as “Unknown project selector; choose from these authorized matches.” Keep the original assistant call in history and let a bounded subsequent step correct it.

## Scheduling, transport and lifetime

The project owner hosts every project session, including delegated sessions. No product cap on stored or pending sessions. Disk capacity, admission policy and provider slots limit actual simultaneous execution. Use persisted waits and indexed runnable work; avoid one resident model context, timer, polling loop or goroutine per idle session. Apply independent limits per provider so a saturated local model leaves unrelated routes responsive. Preserve the existing interactive Chat Pool priority over Project Pool work. Within eligible work use the queue's stable ordering. Priority applies when capacity opens. Measure starvation under sustained interactive traffic before changing that policy.

The authenticated REST adapter maps create/list/open, submit, fetch-since-revision, inspect, cancel and resolve-approval to the same owner operations used in-process. Page history and catalogs. A WebSocket streams live deltas and coalesced revision hints only to authorized subscribers. Reconnect from a cursor or authoritative snapshot, check grants again, and discard speculative partial text if the owner finalized a different outcome. Disconnecting the HUD detaches the viewer. Multiple devices can watch or submit; the owner sets sequence. Independent project sessions can mutate the checkout concurrently only under existing path/revision and Git operation controls.

Indefinite waiting applies before admission to execution, before the first token, between tokens, and through provider completion. A network handshake, a long-poll, an authentication envelope and a worker lease can all expire independently. Keep the worker's renewable lease alive independent of token production. Separate waiting on the session owner from the provider worker that is actually using compute. After owner restart, query the existing queue job. After worker loss, fence the old attempt and look up any provider job/response ID before authorizing another attempt. A provider offering neither resume nor lookup can leave the outcome indeterminate. Surface that state and provide an explicit retry or cancel. A late or duplicate result cannot resume the same turn twice. Never infer failure, migrate to cloud, retry or declare success from elapsed time alone.

The official rotating gear beside “Queued for Project Pool,” “Generating on worker-a” or “Waiting for your decision” communicates state without implying resource use or a percentage. Show observation age separately from elapsed wait. Offer open source session, inspect decision and cancel as appropriate. Stop rendering animation for hidden sessions and use a static icon for reduced motion. REST reads remain useful throughout a WebSocket outage.

## Repo history, delegation and automation

- **Repository contract.** One tracked `.apparat/chats/YYYY-MM-DD-HH-mm-ss.md` per session, UTC with collision-safe suffix. Front matter records stable ID, start time and readable description. Append submitted user turns, visible replies, consequential calls, receipts, decisions and durable relative references. Keep secrets, raw audio and private provider internals out of Git. Show the storage destination when a chat begins. Version under the project's ordinary commit workflow.
- **Exactly one record.** Persist a pending append with a stable entry ID. Serialize owner writes per transcript, check the on-disk revision, write atomically, then mark the append delivered. Restart checks that entry ID before replay. Preserve external changes or surface a conflict. Regenerate a stable, browsable Markdown index from YAML descriptions and relative paths. A derived full-text index can be discarded. Search the source file for a decision mentioned halfway through a different topic. Migrate legacy SQLite chats idempotently with counts and body comparisons. Preserve existing thread addresses or provide a durable mapping.
- **Cross-project conversations.** Global conversations belong to replicated apparatd artifact storage scoped to the authenticated user. Avoid copying a private multi-project exchange into every project. Child work and its reasoning live in the child's repo. Persist parent/child references, grant checks, accepted instruction, current outcome and source link. Parent or child restart can reattach by ID. No duplicate delegation on reconnect.
- **Repeatability.** When a request recurs, ask whether the same trigger, inputs, checks and effects should reliably produce the same outcome. If so, register a versioned project Task and phrase, and expose it through the action catalog. A skill may describe it. An exact phrase invokes it at zero agent-token cost. Where an unexpected failure needs judgment, provide scoped diagnostics, verify the proposed general fix against the failure and normal cases, and promote through the project's normal change workflow. Do not silently install a procedure suggested by model output.
- **Commit-and-push proof.** Resolve owner, project, branch and allowed destination. Record the request before capturing the change set. Follow repository policy, include intended eligible untracked files, run required checks, use a deterministic commit-message default, record commit SHA, push and read back the required remote revision. If clean locally, check for unpushed work. A rejected push preserves its successful local commit. An uncertain acknowledgement reconciles against the remote. Append the receipt after the captured commit boundary so success never recursively commits its own report. Repeat the repaired failure case with zero model calls.

## Delivery and acceptance gates

1. [ ] **Repo conversations and lookup, Phases 9A and 11C.** Extend existing project chat/workspace operations in `internal/app`, Git/filesystem and SQLite adapters. Deliver the transcript, procedural index, migration, search and typed project status. From a fresh clone find an earlier decision using ordinary text tools. Test simultaneous starts, interruption at each append boundary, offline freshness, duplicate project names and current user's visibility.
2. [ ] **Deterministic dispatcher, Phases 9B, 11C and the Phase 15 Task contract.** Bind stable queries and Tasks to typed chat and final non-dictation PTT. Verify the commit/push story in a disposable repository with a local bare remote and a lost acknowledgement. Count calls, including intent classification and narration. Exact status and Task requests must produce zero agent-model calls. Verify editing protection, ambiguity, absent matches, cancel, approval and narrow/wide controls.
3. [ ] **Durable owner protocol, Phases 11C and 13.** Add session REST operations and live updates over the existing authenticated transport. Exercise two devices, two project owners, explicit cancellation, lost acknowledgements, revocation and disconnect during a tool. Use a controlled clock with at least 100 active project sessions on one simulated slow model slot and competing interactive PTT. Advance hours across queue waits and silent generation, restart clients/owner/worker, verify priority and one accepted result per turn. Measure memory, queue reads and idle CPU against the empty baseline.
4. [ ] **First model turn and Pi compatibility, Phases 12–14.** Implement typed declarations, Go registry, request manifests, public skill discovery, provider stream normalization and mock-provider traces. Ship a real cluster Ollama route and configured Cerebras and OpenAI routes via the same contract. Store credentials only on the configured provider host. Enforce authorization, cloud consent, features and deterministic fallback per route. Confirm cancellation and actual provider limitations, including unresumable generation. Bind inference tests to pinned model configuration, hardware, template and backend version.
5. [ ] **Global session, automation and wait UI, Phases 13–16.** Resolve global repo ownership, enable delegation, updates and protected approval handling. Use the official gear in the real session wait view. Ask by PTT, delegate, close the Deck, reopen on another device, inspect/act on a decision and follow the result into the project transcript. Demonstrate one formerly unfamiliar failure converted into verified deterministic handling.
6. [ ] **Conformance, performance and delivery, Phases 14 and 17.** Record golden request/response fixtures from the pin. Test streamed split JSON, unknown calls, duplicate IDs, truncated batches, stale approvals, result ordering, malformed skill metadata and multi-turn recovery. Pair Pi and Apparat on the same model/task/settings for their shared subset. Include a candidate Pi-tuned model and comparison configuration. Count wrong-tool, invalid-argument, wrong-target and wrong-outcome rates as well as end-state success, tokens, retries and latency. Compare warm-up strategies cold, warm and after restart. Report sample counts, variance and unsupported features. Retain optimizations only with a repeatable gain. Run host documentation, source, GUI and governed release checks for the implementation checkpoint.

**Exit evidence.** A known request costs zero agent calls. A model-assisted request uses a measured, declared tool contract. A long wait survives a restart. An owner receipt proves an effect. A fresh clone explains it. Until these scenarios pass on the declared backends and surfaces, compatibility, latency gains and release qualification remain pending.

### Queue and project UI checkpoint (Phases 11C and 13)

Implement project conversation bubbles, priority-ordered hash-service claims and redacted queue-health sampling. Exercise all four priorities and competing claimants. Retire the Personal pool from active routing while preserving historical receipts. Describe the future model-step queue handoff and provider recovery boundary separately from the executable hash service.

### Distributed inference delivery (Phases 13–14, HUD Phase 16)

- [x] Establish an agreed versioned roster, RAM reservations, withdrawal, fencing across restart, per-executor dispatch intent, and authenticated worker-to-worker messages. Exercise delay, duplication and independent restarts.
- [x] Persist global session artifacts and stable turn submissions, replicate to authorized nodes, dispatch project and global jobs through the appropriate pool, and reattach after GUI disconnect.
- [x] Instrument claim/propagation rounds and bounded adaptive rechecks. Replace Local Services with Cluster Status and retain service details on the device page.
- [x] Qualify two independent providers, withheld confirmations, contention, mixed pool priority and live trashcan0 execution. Historical telemetry analysis follows the live status milestone.

### Inference qualification evidence

The two-worker acceptance uses independent SQLite databases, identities, worker listeners and pinned TLS connections. It withholds job admission on one validator, drains eight jobs across the four priorities, checks the counts at the provider boundary, races the contenders, and reopens both ledgers. It submits turns through the authenticated workspace and global APIs, checks owner-side replies and verifies each global Markdown copy. A second signed-in device resumes the conversation through a different worker. Additional faults cover a partition, an interrupted provider dispatch and a transcript exceeding one transport envelope. The live variant completed a Chat Pool request using qwen3.5:4b through a temporary loopback tunnel to trashcan0. The production worker installation and enrolled roster were outside the isolated test. Fixed-roster operation is the delivered scope. Membership handoff and historical telemetry analysis retain separate roadmap entries.

### Cluster hardware scorecard (Phase 14)

- [x] Lead Cluster Status with totals from fresh hardware reports and a measured CPU FP32 score. Deduplicate reports, show measurement coverage and verify narrow and wide presentation.

### Project chat entry (Project chats, Phase 16 HUD)

Open Chat and New conversation directly into an unsaved session. On first send, save the opening sentence as its title. Verify history, retry identity, session model selection and the owner queue path.

### Global planning transcript (Phases 13–14, HUD Phase 16)

Remove the scrolling-demo transcript from new sessions. Verify the empty composer, durable send, pending state and actual worker response.

### Chat submission and sync feedback (Phases 13–14 and 16)

Route Enter in the global and project composers through the Send action. Keep Shift+Enter for line breaks. Show the writer discovery or enrollment blocker when a global submission remains in the originating device's outbox.

### Automatic inference participation and durable intake (Phases 13–14)

Remove separate validator enrollment and hash service assignments from inference. Carry authenticated worker membership through ordinary cluster synchronization, preserve existing claims and global histories across upgrades, and decouple turn receipt from executor selection. Verify singleton upgrade, multiple workers, contention, restart, pool visibility and canonical release.

### Hardware and chat inspection (Phases 14 and 16)

Show a four-column desktop scorecard and an INT8-equivalent CPU plus accelerator estimate with coverage. Share contextual copy and per-step telemetry across both conversation views. Verify partial inventory, tool-call exchange, rates, narrow and wide HUD navigation, then qualify the release. Markdown presentation is a design discussion.

### Persist the Phase 14 CPU baseline

Reuse one successful CPU measurement per node database across inventory refreshes, concurrent startup and restarts. Verify by counting measurements across two database connections and a reopen.
