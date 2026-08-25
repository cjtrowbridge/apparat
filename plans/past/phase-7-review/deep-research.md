# Major Checkpoint Architecture, Codebase, Governance, and Project-Plan Review — Apparat

## Scope and evidence basis

This report applies the attached review specification to the attached Apparat GitIngest. The review specification explicitly requires reconstruction of the implementation before recommending refactors, treats the roadmap as a proposal rather than truth, requires evidence-calibrated findings, calls for checkpoint hardening where reversal costs are rising, and asks for a complete Draft 2 project plan. 

I distinguish three evidence classes below:

* **Repository fact** — directly supported by Apparat source, tests, governance, plans, journals, or configuration in the GitIngest.
* **Verified upstream fact** — checked against authoritative upstream documentation/source.
* **Inference/recommendation** — my engineering conclusion from those facts.

The GitIngest intentionally excludes recursive third-party source bodies. It does expose the active module versions, dependency roles, and the complete `agentic-pipelines` governance submodule. The separate ignored provenance manifest mentioned in the journal was not attached, so I cannot independently verify every third-party gitlink SHA. Where that matters, I say so rather than substituting a newer arbitrary revision. The repository itself records Ebitengine 2.9.9, EbitenUI's pinned pseudo-version, modernc SQLite's pinned pseudo-version, and Agentic Pipelines at `bdf758c21319b3a2f99d70a9f95417b55e636398`.  

---

# Part I — Executive Assessment

## Overall conclusion

**Apparat is on a fundamentally viable architectural path, but it is not yet safe to treat the existing “foundation” implementation as the foundation for distributed operation. A checkpoint-hardening phase is required before Phase 8 or substantial distributed feature work begins.**

The important distinction is that this does **not** call for a rewrite.

The most encouraging finding is that the repository's *current intended architecture*—especially the rewritten Phase 7—is substantially better than parts of the implementation it is about to replace. Phase 7 correctly proposes:

* one logical node rather than two binary-owned state universes;
* a core that no longer owns the HUD;
* GUI calls directly into a transport-neutral internal application API;
* REST mapping onto that same API rather than inventing parallel rules;
* an exclusive node-runtime lock;
* transactional startup and ordered shutdown;
* injected clocks and ID sources;
* centralized SQLite ownership;
* backup/integrity/restore testing;
* a loopback-only read API before security exists.   

That is mostly the right next move.

The central problem is that **several Phase 3 features were marked complete at a semantic level that the actual implementation does not satisfy**. For example, the roadmap records completed ULID helpers, startup identity consistency classification, signed device profiles, and replay/duplicate tracking.  But:

* the ID generator is not a ULID;
* identity “ready” classification is essentially file-presence classification;
* cluster device records store a `Signature` field but the storage path shown does not verify it;
* replay tracking interprets **every database insertion failure** as “already seen.”

This is more important than any individual bug. It shows that the governance system is currently better at proving that an artifact or mechanism **exists** than proving that the semantic property named by a checklist item **actually holds**.

That is exactly the point at which additional distributed features would make reversal expensive.

## Highest-leverage findings

| #  | Finding                                                                                                                                  | Diagnosis                                                                                                                                                                    | Recommendation                                                                                                               | Priority                                  |
| -- | ---------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| 1  | Phase 7 is directionally correct but needs to become a **foundation-convergence checkpoint** rather than merely the next feature phase.  | The mature architecture is already better than the transitional implementation.                                                                                              | Expand Phase 7 slightly and complete it before Phase 8.                                                                      | **Checkpoint blocker**                    |
| 2  | Roadmap completion has sometimes exceeded implementation evidence.                                                                       | Semantic claims were satisfied by structural/happy-path implementations.                                                                                                     | Add evidence-classification and claim-calibrated completion to product governance.                                           | **Checkpoint blocker**                    |
| 3  | Persistence is not yet a trustworthy architectural boundary.                                                                             | Raw `*sql.DB` leaks upward; tables are initialized outside central migrations; FK configuration is connection-sensitive; replay errors are mishandled; ID contract is wrong. | Consolidate persistence behind one adapter/repository/transaction boundary during Phase 7.                                   | **Checkpoint blocker**                    |
| 4  | Existing identity code is a useful local bootstrap prototype, not a secure Phase-8 identity authority.                                   | File existence is treated as readiness; key envelope metadata is weakly versioned; manifest consistency is not cryptographically established by `Classify`.                  | Explicitly version/migrate it and prohibit its current “ready” status from conferring network trust.                         | **Checkpoint blocker before Phase 8**     |
| 5  | Current `internal/app` has the dependency direction Phase 7 says must disappear.                                                         | The core runtime imports GUI/HUD and owns `hud.Shell`.                                                                                                                       | Move concrete GUI composition outward to `cmd`, keep application use cases presentation-independent.                         | **Checkpoint blocker**                    |
| 6  | The explicit function-comment standard is almost completely unmet.                                                                       | Programmatic inspection found 317 project-owned Go functions/methods; only 2 had immediately preceding explanatory `//` comments.                                            | Adopt the standard now, backfill core code, then require it prospectively; deterministic presence + semantic quality review. | **High**                                  |
| 7  | Current concurrency is mercifully small.                                                                                                 | There is no mature goroutine/service-worker topology yet.                                                                                                                    | Define lifecycle ownership *before* HTTP, provider supervision, queues, and schedulers introduce it.                         | **Major opportunity / blocker if missed** |
| 8  | Historical input-equivalence claims remain stronger than executable behavior.                                                            | Binding data lists A/B/D-pad/menu/etc., but actual gamepad handling presently centers on L1/R1/R2 and `Shell.ApplyAction` rejects several declared actions.                  | Continue Phase-6 reconciliation; close parity while real controls gain backend effects.                                      | **High, but not Phase-7 blocker**         |
| 9  | Android has convincing prototype evidence but not production lifecycle/release readiness.                                                | Wrapper rendering works, but runtime lifetime is poorly owned, `Ready()` can lie, SDK target is 30, and the build relies on local tool patching.                             | Fix runtime ownership now; target-SDK/build-tool hardening before distribution.                                              | **High**                                  |
| 10 | The broad architectural decisions—single ownership, SQLite, external WireGuard first, static provider drivers, no remote shell—are good. | They reduce distributed ambiguity without excessive abstraction.                                                                                                             | Preserve them.                                                                                                               | **Leave alone**                           |

## Health assessment

| Dimension                 | Assessment                                          | Why                                                                                                                      |
| ------------------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Readability               | **Acceptable**                                      | Naming and package intent are generally understandable; functions are mostly short.                                      |
| Human editability         | **Needs improvement**                               | Runtime/persistence boundaries still require understanding multiple layers; comments are nearly absent.                  |
| Documentation             | **Strong, with semantic-overclaim risk**            | Extensive and unusually thoughtful, but historical checkboxes can imply more implementation maturity than exists.        |
| Package cohesion          | **Needs improvement**                               | `internal/app` currently owns HUD state and imports a concrete GUI adapter despite the intended inward dependency model. |
| Dependency isolation      | **Acceptable**                                      | Third-party references are deliberately isolated; Ebitengine build customization is the main exception.                  |
| State ownership           | **High risk / transitional**                        | Mature ownership rules are good, current runtime does not yet implement them consistently.                               |
| Testability               | **Acceptable to strong**                            | Strong HUD/build/governance tests; important persistence/security semantics need better adversarial tests.               |
| Reliability               | **Needs improvement**                               | Replay error collapse, partial startup, identity semantics, and DB configuration need hardening.                         |
| Concurrency safety        | **Acceptable now; high future risk**                | Little concurrency exists today, making this the ideal time to establish ownership.                                      |
| Cross-platform integrity  | **Needs improvement**                               | Real Android work is impressive, but broad supported-platform claims appropriately remain gated.                         |
| Android readiness         | **Prototype strong; production high risk**          | Rendering/install evidence exists; lifecycle, target SDK, release signing, broader devices remain.                       |
| Extensibility             | **Design strong; implementation immature**          | Service/capability/queue abstractions are mostly future architecture.                                                    |
| Observability             | **Acceptable**                                      | `last_run.log`, JSONL, doctor, component events are good foundations.                                                    |
| Trust/security readiness  | **Design strong; implementation not ready**         | Phase-8 documents are much stronger than the current identity implementation.                                            |
| Governance clarity        | **Strong**                                          | Authority hierarchy and planning process are unusually explicit.                                                         |
| Governance enforceability | **Needs improvement**                               | Deterministic gates are strong; semantic evidence calibration is weak.                                                   |
| Planning quality          | **Strong Draft 1**                                  | The Phase-6 rewrite corrected many earlier assumptions.                                                                  |
| Feature readiness         | **Phase 7 ready after amendment; Phase 8+ not yet** | The next phase is correctly aimed at the real fault line.                                                                |

---

# Part II — Repository Governance, Planning, Dependency, and Submodule Inventory

## Canonical project authorities

The repository defines a sensible authority hierarchy:

* `README.md` is the canonical human-facing product, architecture, scope, terminology, and design-decision contract.
* `ROADMAP.md` is the canonical implementation sequence, dependency map, checklist, exit criteria, and open-decision register.
* implementation plans under `plans/` are execution authority for code/schema/build changes.
* journals are append-only historical records.
* `TODO.md` is explicitly only an asynchronous inbox and currently contains no active Apparat tasks.  

That separation is good and should remain.

## Agent governance

Root `AGENTS.md` requires agents to read the complete shared Agentic Pipelines baseline and then apply Apparat-specific overrides. It also routes implementation, debugging, risk review, framework updates, and other work to focused playbooks. 

The shared framework contains several high-quality invariants:

* deterministic checks must not impersonate semantic proof;
* model/source data are untrusted;
* mutations require an approved plan;
* documentation and executable contracts must agree;
* retries are bounded and visible;
* failed candidates/evidence are preserved rather than silently normalized;
* plan scope is execution authority. 

Of particular relevance, the framework already says that meaning, equivalence, relevance, and qualitative fitness require semantic or human judgment.  This principle should be imported more explicitly into **product milestone completion**, because the product workflow is where semantic overclaim occurred.

## Application governance

The root rules establish:

* `internal/app` for shared orchestration;
* `internal/domain` for product rules and durable concepts;
* `internal/adapters` for external systems;
* `internal/platform` for OS/platform lifecycle;
* local documentation for code-bearing directories;
* a 400-physical-line source limit;
* structured logging and sensitive-data redaction. 

These are mostly good rules. The problem is primarily compliance/interpretation, not that the rules are conceptually wrong.

## Dependency inventory

Active module versions are explicit in `go.mod`:

* Go 1.26.4;
* Ebitengine 2.9.9;
* EbitenUI pinned pseudo-version;
* `golang.org/x/crypto`;
* modernc SQLite pinned pseudo-version;
* pinned Ebitengine gomobile tooling as an indirect module. 

The root module also replaces Ebitengine with `./third_party/game/ebiten`, meaning the source checkout is not merely documentary—it is active in builds. 

The repository categorizes source references deliberately:

* Ebitengine — active cross-platform runtime source.
* EbitenUI — HUD source/reference plus active module dependency.
* debugui — source reference only.
* modernc SQLite — source reference; actual module dependency is active.
* WireGuard source repositories — future detection/control references, not embedded VPN dependencies.
* llama.cpp — future inference adapter reference.
* whisper.cpp — future ASR reference.
* Agentic Pipelines — governance submodule, not application runtime code. 

This is a good dependency-admission model.

---

# Part III — Current Architecture Reconstruction

## What Apparat actually is today

Apparat currently consists of four major implemented strata.

### 1. Executable/runtime composition

`cmd/apparat` and `cmd/apparatd` load configuration, create an `app.Runtime`, install signal cancellation, invoke smoke/doctor/normal execution paths, and defer runtime close. 

`internal/app.Runtime`, however, is still transitional:

```text
Runtime
 ├─ config
 ├─ hud.Shell
 ├─ *database.DB
 ├─ logger / last_run
 ├─ cluster.Directory
 └─ messaging.Store
```

It directly imports GUI, HUD, persistence, identity, cluster, and messaging packages. 

This means the “application core” presently owns presentation state and concrete adapters.

### 2. GUI/HUD

The GUI is an Ebitengine/EbitenUI retained UI with:

* data-driven tab definitions;
* master/detail layouts;
* scrolling and drag handling;
* mouse/touch handling;
* some controller handling;
* a mock `hud.Shell` and snapshot model.

`Game.Update` currently performs UI updates, presentation-state management, input sampling, tab actions, and status application. It is not currently doing inference, HTTP, or database operations inside the frame loop. 

That is an important positive boundary.

### 3. Local persistence/runtime primitives

`internal/database` opens SQLite and exposes its underlying `*sql.DB`. `internal/cluster` and `internal/messaging` receive that raw handle and initialize their own tables. Runtime initialization opens SQLite, applies the very small Phase-3 migration, then separately calls `cluster.Init` and `messaging.Init`. 

So there is not yet one authoritative migration/persistence boundary.

### 4. Local identity and diagnostics

The identity package includes:

* Ed25519 generation/sign/verify;
* SHA-256 public-key fingerprints;
* Argon2id-derived encryption;
* XChaCha20-Poly1305 encrypted private-key storage;
* manifests;
* repair/rotate helpers;
* a simple state classifier.

The cryptographic primitives are useful, but their orchestration is much less complete than the roadmap status implies.

## Current startup flow

Actual normal startup is approximately:

```text
cmd
 → config.Load
 → app.NewRuntimeWithConfig
 → signal.NotifyContext
 → Runtime.Start
    → Runtime.Initialize
       → EnsureDirectories
       → last_run log
       → JSONL logger
       → database.Open
       → phase3 migration
       → cluster.New(db.SQL).Init
       → messaging.New(db.SQL).Init
       → identity.Classify
       → started=true
    → GUI: gui.RunWithRuntimeInfo
      OR
      headless: wait for context cancellation
```

There is no transactional rollback of already-started components if an intermediate initialization stage fails. `Close` principally closes the database. 

## Current state ownership

Today:

* HUD/navigation state — `hud.Shell` / GUI, but the `Runtime` also owns the shell.
* SQLite — opened by `internal/database`, raw handle distributed to cluster/messaging.
* cluster device profile state — `cluster.Directory`.
* replay/inbox/outbox state — `messaging.Store`.
* identity — partly filesystem-based.
* Android game/runtime — initialized in package `init()`, with lifetime not explicitly tied to Android activity lifecycle. 

This is not yet the mature one-owner model described in the architecture documents.

---

# Part IV — Intended Future State and Current Plan Reconstruction

The mature design is a local-first, owner-authoritative distributed system.

Its most important intended invariants are:

* one authoritative device per Project;
* one authoritative owner per queue;
* Project Tasks owned with the Project;
* workers lease jobs from queue owners rather than becoming queue authorities;
* at-least-once delivery plus stable IDs/idempotency;
* authenticated REST over LAN/external WireGuard;
* signed application envelopes;
* typed service/capability/workload identity;
* local provider endpoints hidden behind Apparat's gateway;
* cached remote data remains non-authoritative;
* no unrestricted remote shell.

The current roadmap sequence after reconciliation is:

1. **Phase 7:** shared core/local persistence/internal API/read-only REST.
2. **Phase 8:** identity, trusted directory, secure REST commands, reusable mock queue.
3. **Phase 9:** discovery/presence/project catalog.
4. **Phase 10:** real Project workspaces/Git/manual Tasks.
5. **Phase 11:** multi-instance inference services/capabilities.
6. **Phase 12:** authoritative queue/leasing/results/artifacts.
7. **Phase 13:** pools/routing/real text generation.
8. **Phase 14:** triggers/automation/workflows.
9. **Phase 15:** speech.
10. **Phase 16:** packaging/release/platform evidence.
11. Post-MVP transport, Comrades, and Research tracks.    

The sequence is broadly correct. Draft 2 below changes its *quality gates and concurrency*, not its overall product philosophy.

---

# Part V — Code Quality, Readability, and Human Editability

## Strengths

### File size discipline works

The 400-line rule is actually producing relatively small implementation units. This is one of the few deterministic maintainability rules here that has clear value and low interpretive ambiguity.

I would retain it.

### Naming is usually domain-oriented

Names such as:

* `Directory`,
* `DeviceProfile`,
* `Message`,
* `Runtime`,
* `Snapshot`,
* `Action`,
* `ServiceID`,
* `CapabilityID`,
* queue owner / Project owner,

are generally readable.

### GUI behavior has been decomposed

The GUI is spread across layout, scrolling, master/detail, theme, builder, debug, and shell files rather than one giant `Game` implementation.

This is appropriate decomposition rather than obvious fragmentation.

## Weaknesses

### Locality of change is poor at the core boundary

Today, changing durable runtime behavior can involve:

* `internal/app`;
* `internal/database`;
* `internal/cluster`;
* `internal/messaging`;
* GUI projections;
* command startup;
* Android composition;
* docs/plans.

Some propagation is legitimate. What is unnecessary is that concrete SQL and presentation ownership still leak into orchestration.

### Hidden semantics in apparently simple functions

Examples include `Seen`:

```go
func (store Store) Seen(ctx context.Context, id string) (bool, error) {
    _, err := store.db.ExecContext(...)
    if err == nil {
        return false, nil
    }
    return true, nil
}
```

The signature implies `(wasAlreadySeen, error)`, but the implementation converts **all** storage errors into a successful duplicate determination. 

That is a human-editability problem as much as a correctness problem: the API contract is stronger than the implementation.

### Runtime is doing too many conceptual jobs

`Runtime` currently means:

* composition;
* startup;
* shutdown;
* persistence ownership;
* logs;
* cluster repositories;
* messaging;
* HUD shell;
* GUI launching.

Phase 7 is right to narrow it.

---

# Part VI — Comment and Documentation Compliance

The review specification says:

> Every function and subroutine should have a human-useful comment at its beginning explaining what it does and why it exists.

It further requires intent, assumptions, state effects, concurrency/lifecycle, cancellation, and errors where relevant—not syntax restatement. 

## Measured result

I programmatically inspected the project-owned Go source under `cmd/` and `internal/` in the supplied GitIngest:

* **49 Go files**
* **317 functions/methods**
* **243 non-test functions/methods**
* only **2** had an immediately preceding explanatory `//` comment under the straightforward source-level test.

That is effectively non-compliance.

This is not just about exported GoDoc. Representative critical routines such as `Runtime.Initialize`, `Runtime.Close`, `messaging.Store.Seen`, and numerous GUI lifecycle routines begin without comments despite having non-obvious state or failure semantics.  

## How to enforce it without creating nonsense comments

I recommend two layers:

**Deterministic gate:** does every project-owned function declaration have an associated comment?

**Semantic review:** does the comment explain intent/contract rather than translate the syntax?

Do **not** build a heuristic that scores comment “quality.” The governance framework already correctly says semantic fitness cannot be proven by a deterministic surrogate. 

## Representative improved comments

For identity classification, the comment should make its present limitation unmistakable:

```go
// Classify reports the filesystem-level availability of the local device
// identity files. It does not decrypt the key, verify that the manifest
// matches that key, validate signatures, or establish network trust.
// Callers must therefore treat StatusReady as local bootstrap evidence only.
func Classify(dir string) Status
```

For replay handling after correcting the implementation:

```go
// Seen atomically records id in the replay ledger and reports whether the
// same identifier was recorded previously. Only the table's uniqueness
// constraint is interpreted as a duplicate; cancellation, locking,
// corruption, and other storage errors are returned to the caller.
func (store Store) Seen(ctx context.Context, id string) (bool, error)
```

For runtime initialization after Phase-7 hardening:

```go
// Initialize transitions the node from constructed to ready by acquiring
// exclusive ownership, starting diagnostics, opening and migrating durable
// state, constructing repositories, and then starting supervised services.
// If any stage fails, already-started stages are unwound in reverse order.
// The method is idempotent only for an already-ready runtime.
func (runtime *Runtime) Initialize(ctx context.Context) error
```

For Ebitengine startup:

```go
// RunWithRuntimeInfo starts the desktop Ebitengine presentation adapter.
// Ebitengine owns the main OS thread while RunGame is active; cancellation
// is therefore forwarded into Game so Update can return ebiten.Termination.
// Durable application shutdown remains owned by the outer node lifecycle.
func RunWithRuntimeInfo(ctx context.Context, info RuntimeInfo) error
```

---

# Part VII — Architecture and Idiomatic Go/SOLID Review

## Single responsibility

**Current:** mixed.

`internal/identity`, `internal/config`, `internal/logging`, and most GUI utility files are cohesive.

`internal/app.Runtime` is not: it combines composition, GUI launching, durable ownership, logging, HUD state, and lifecycle.

**Target:** Phase 7's narrower application orchestration is appropriate.

## Open/closed

The code is not yet meaningfully extensible for inference providers because the provider system is not implemented. The roadmap's proposal for explicit static driver registration is a better fit for Go than a plugin framework. Phase 11 explicitly rejects hidden package-global registration and Go dynamic `.so` plugins. 

I would preserve that.

## Liskov/substitution

Not yet a major issue because few interchangeable runtime implementations exist. It becomes relevant for:

* provider drivers;
* clock/ID test ports;
* persistence transaction interfaces;
* platform locks;
* transports.

Interfaces should be defined where a caller truly consumes interchangeable behavior—not ahead of need.

## Interface segregation

The planned small internal application commands/queries are a good use of interface boundaries. Phase 7 even explicitly says to add abstraction only after a second slice proves a shared contract. 

That is exactly the right anti-overengineering rule.

## Dependency inversion

This is the clearest current violation.

The intended architecture says Ebitengine/EbitenUI belong below the GUI adapter boundary. 

But current `internal/app/runtime.go` imports `internal/adapters/gui` and owns `hud.Shell`. 

That should be reversed:

```text
cmd/apparat
 ├─ constructs app core
 ├─ constructs GUI adapter
 └─ runs both

GUI adapter → app commands/queries
HTTP adapter → app commands/queries
app → domain + narrow outward ports
```

No DI framework is warranted.

---

# Part VIII — Ebitengine, Concurrency, and Lifecycle Review

## Ebitengine loop

The good news is that `Game.Update` presently performs primarily local UI/input work. There is no evidence of database, HTTP, or inference calls in the shown frame path. 

That invariant should be made permanent.

Upstream Ebitengine documentation says game methods execute on the same goroutine, `RunGame` must run on the main thread, and desktop termination is best performed by returning `ebiten.Termination` from `Update`. ([Go Packages][1])

Current `RunWithRuntimeInfo` checks `ctx.Done()` only **before** calling `ebiten.RunGame`. Once the loop starts, the supplied context is not shown being propagated into `Game`. 

### Consequence

A SIGTERM/SIGINT can cancel the outer context without necessarily causing an active desktop Ebitengine loop to exit gracefully.

**Recommendation:** store a cancellation signal/context in `Game`, check it in `Update`, and return `ebiten.Termination`.

## Current concurrency map

One of the most important results of the audit is what **does not exist yet**.

Programmatic inspection of the project-owned Go body found:

* no mature goroutine worker topology;
* no channel-based scheduler;
* no queue worker pool;
* no reconnect loops;
* no inference supervisor;
* no HTTP server lifecycle;
* no long-lived lease heartbeats.

This is an advantage.

The project is about to add all of those concerns, but it has not yet accumulated concurrency debt. Phase 7 is therefore the last cheap moment to standardize:

```text
owner
start
context
state owned
dependencies
error channel/state
shutdown
timeout
restart policy
```

for every long-lived worker.

## Android lifecycle

Android currently initializes the runtime from a Go package `init()` using `context.Background()`. If initialization succeeds, it builds the GUI and calls `mobile.SetGame`. 

Two concerning details:

1. the runtime lifetime is kept alive indirectly rather than explicitly owned by an Android lifecycle object;
2. `Ready()` unconditionally returns `true`, even if initialization returned early after an error. 

This needs correction in Phase 7.

---

# Part IX — Distributed State and Reliability Review

The intended distributed design is substantially stronger than the current implementation.

## Authoritative ownership

The one-owner-per-Project/queue/Task model is the correct simplifying choice.

It avoids having to solve:

* distributed consensus for ordinary project mutations;
* multi-writer queue ordering;
* automatic scheduler election;
* CRDT semantics for all state.

This should be considered an architectural strength, not a limitation to “generalize away.”

## Replay/idempotency

This is currently the most concrete local defect that would become a distributed correctness bug.

`Seen` inserts into `replay_seen`; success means unseen, but **every error** becomes “seen.” 

Under future network operation, a transient database lock or cancellation could therefore cause a legitimate message to be silently suppressed as a duplicate.

That is unacceptable for at-least-once delivery.

## Cluster profile mutation

`Directory.PutDevice`:

1. marshals the supplied profile;
2. upserts it;
3. records a change row. 

There is no visible transaction enclosing both writes, so an interruption can leave a new profile without its associated change entry.

Likewise, no signature verification is shown before persistence.

Before Phase 9 uses directory state for presence/catalog reconciliation, profile updates need:

* validation;
* revision rules;
* transactionality;
* authority source;
* trust state separate from cached observation.

## Failure semantics the mature architecture must preserve

For remote operations:

* transport success ≠ domain application;
* worker completion ≠ authoritative queue completion;
* cached peer state ≠ authorization;
* discovered endpoint ≠ identity;
* duplicate delivery ≠ duplicate logical job;
* lease expiration ≠ cancellation of all possible late worker activity;
* stale capability advertisement ≠ current executability.

The current roadmap understands these distinctions. Phase 8 requires cached records to lose authorization after expiry/revocation; Phase 12 requires owner-created leases/fencing and owner-only authoritative completion.  

Do not simplify those away during implementation.

---

# Part X — Submodule and Dependency Architecture Review

## Ebitengine

**Repository fact:** Apparat declares Ebitengine 2.9.9 and replaces the module with `third_party/game/ebiten`.  

**Verified upstream:** upstream release records identify v2.9.9 and commit `f65118d`. ([GitHub][2])

Because the GitIngest omitted the submodule manifest/gitlink metadata, I cannot establish that the local checkout itself is precisely `f65118d`; I can only establish the intended module version.

**Boundary assessment:** mostly good. Ebitengine is localized to the GUI/mobile/build surfaces.

**Main risk:** the Android pipeline also carries build-time accommodations for gomobile/Ebitengine. This means upgrading Ebitengine is not only a `go.mod` event; it potentially affects:

* local source checkout;
* helper patches;
* Android binding;
* Java wrapper;
* layout/density behavior;
* build tests.

**Recommendation:** keep the source replacement for now because Apparat genuinely uses and studies it, but treat upgrades as a compatibility checkpoint with Android evidence.

## EbitenUI

It is active and pinned. The primary code confines it to GUI code. That is the correct boundary.

Given that the immediate architectural issues are in persistence/lifecycle rather than widget internals, a deeper upstream audit would be disproportionate.

**Leave alone.**

## debugui

The repository explicitly keeps it reference-only because its source revision targets the Ebitengine 2.10-alpha line while Apparat intentionally starts on stable 2.9.x. 

**Leave alone.**

## modernc SQLite

The module is active and the source checkout is reference material. This is a good separation.

The important upstream semantic issue is SQLite itself: foreign-key enforcement must be enabled separately on each database connection. ([SQLite][3])

Current Apparat performs one `PRAGMA foreign_keys = ON` through a `database/sql` handle before returning the pool. 

Because `database/sql` may use multiple underlying connections, Phase 7 should guarantee connection initialization semantics rather than assuming the initial statement configures every later pooled connection.

The Phase-7 roadmap already recognizes the need to explicitly define connection limits, busy timeout, foreign-key behavior, retry classes, and transaction behavior. 

## WireGuard references

They are intentionally not application dependencies. The MVP uses externally configured WireGuard as reachability and performs application authentication separately.

This is an excellent boundary.

Do not embed WireGuard now.

## llama.cpp / whisper.cpp

Both are future adapter references. Neither needs a deep audit at this checkpoint.

The right time is:

* llama.cpp: during Phase 11/13 provider-adapter work;
* whisper.cpp: during Phase 15.

The existing “research before adding” policy is appropriate. 

## Agentic Pipelines

Unlike the third-party source references, its source is actually present in the GitIngest.

Its architecture is sophisticated and notably more rigorous about evidence authority than the older Apparat product workflow. The main recommendation is **not** to make ordinary product development run through the whole semantic pipeline system; rather, borrow its evidence-calibration principle for implementation completion.

---

# Part XI — Cross-Platform and Android Readiness

## Linux / Steam Deck

The architecture is fundamentally suitable:

* Ebitengine GUI;
* separate headless executable;
* runtime roots outside source;
* external WireGuard;
* SQLite;
* no cgo requirement for SQLite.

Remaining key evidence gaps:

* full controller-action equivalence;
* desktop GUI signal termination;
* long-running headless service behavior;
* actual secure REST/provider lifecycle;
* packaged service installation.

## Windows

The build pipeline has Windows-specific Android tool resolution and some build-host work, but that is not equivalent to mature Windows Apparat support.

The roadmap correctly delays a support claim until actual packaging/network/input/runtime evidence exists.

## Android

There is meaningful evidence here, not just compilation.

The project records:

* installed Pixel APK;
* process liveness;
* app-private runtime storage;
* rendered HUD;
* touch tab selection;
* Android 16 / SDK 36 test device evidence. 

That is good engineering.

But current metadata still uses `targetSdkVersion=30` while building with platform 35. 

As of August 24, 2026, Google Play's published policy says that starting **August 31, 2026**, new apps and app updates must target Android 16 / API 36; existing apps need at least API 35 for availability to new users on newer Android versions, absent the temporary extension. ([Google Help][4])

That does **not** mean Phase 7 should suddenly become a Play Store project. It means Phase 16 cannot inherit `targetSdkVersion=30` as a release assumption.

### Android stress-test conclusion

Android is already exposing the right architectural questions:

* who owns core lifetime when an Activity stops?
* what can survive process death?
* what is durable versus view state?
* what happens when network becomes unavailable?
* can provider supervision continue?
* when should connections be torn down?
* does PTT continue across focus/lifecycle transitions?

Fixing those through the shared lifecycle architecture is better than adding Android-specific flags later.

---

# Part XII — Security and Trust-Boundary Readiness

## Design readiness: strong

The documented trust model is good:

* Apparat identity separate from WireGuard;
* TLS leaf key separate from Apparat signing key;
* explicit device records bind them;
* mTLS plus current device-record authorization;
* signed envelope;
* scopes;
* replay protection;
* revocation;
* no remote provider endpoint exposure;
* no remote shell.

## Implementation readiness: not yet

### Identity classification

The current classifier exposes only:

* missing;
* ready;
* inconsistent.

The shown implementation considers the existence of key/manifest files sufficient for readiness rather than proving the deep cryptographic consistency described in the architecture. 

### Key envelope

`LoadEncrypted` ignores base64 decoding errors for the salt, nonce, and ciphertext before trying to use them. 

The stored key envelope also does not visibly encode enough self-describing algorithm/KDF/version information for safe long-term migration.

### Manifest

`WriteManifest` writes public identity metadata but the shown manifest is not itself a signed authorization document. `RepairManifest` derives it directly from the private key. 

### Recommendation

Do not discard the crypto primitives.

Instead classify them as:

> **legacy/local bootstrap identity format v0 — not authoritative for remote trust**

Then Phase 8 should implement the documented identity layer as a versioned v1 migration.

That avoids both a rewrite and a dangerous assumption.

---

# Part XIII — Testing, Determinism, Observability, and Debuggability

## Testing strengths

The repository has:

* Go unit tests;
* race tests in `make verify`;
* Python build-pipeline tests;
* code-size tests;
* docs-completeness tests;
* Android build validation;
* GUI behavior tests;
* plan/governance tests.

`make verify` explicitly combines formatting, unit tests, race tests, build tests, code size, docs checks, lint, and vulnerability scanning. 

This is a strong deterministic baseline.

## Testing weaknesses

### Tests often prove mechanism existence rather than contract depth

Examples:

* a database migration test verifies migration/checksum machinery, but does not establish all repository schema is actually centrally migrated;
* identity tests can exercise encryption/manifest handling without establishing that “ready” means cryptographically coherent;
* replay tests can pass duplicate behavior without testing cancellation/database failure differentiation.

This is the same epistemic problem seen in roadmap completion.

## Deterministic core

Phase 7's proposal to inject clock and ID sources is exactly right. 

Do the same for:

* retry schedule calculations;
* service advertisement expiry;
* lease deadlines;
* trigger scheduling;
* route scoring;
* stale-state evaluation.

Avoid mocking the world. Just separate time/randomness from deterministic rules.

## Observability

`last_run.log` is a useful operator-facing idea.

The JSONL logger is also useful.

But current redaction is primarily key-name substring based:

```text
token
private
passphrase
prompt
model_output
voice
secret
```



That is not sufficient once nested request/provider/error structures are introduced.

By Phase 8/11, prefer typed safe logging fields and explicit sensitive-field omission over hoping names contain the right substring.

---

# Part XIV — Change-Amplification Analysis

| Future change                | Current affected areas                                       | Necessary propagation                                                    | Unnecessary propagation today                            | Recommended response                                  |
| ---------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------ | -------------------------------------------------------- | ----------------------------------------------------- |
| Add inference backend        | app composition, future service model, GUI mock data, config | driver package, composition registration, driver config schema, fixtures | provider-specific state leaking into generic HUD/runtime | Phase-11 static driver interface keyed by `ServiceID` |
| Add inference device type    | cluster data, capability view, UI                            | capability advertisement + platform observations                         | provider-specific conditionals                           | keep device identity separate from capabilities       |
| Add peer transport           | currently future only                                        | transport adapter + envelope/conformance                                 | Project/queue/domain rewrites                            | preserve transport-neutral envelope                   |
| Upgrade Ebitengine           | GUI, Android build, gomobile patches                         | GUI/mobile/build evidence                                                | application/domain changes                               | keep Ebitengine behind GUI/platform boundary          |
| Change SQLite implementation | database, cluster, messaging, runtime                        | persistence adapter                                                      | all repositories consuming raw `*sql.DB`                 | eliminate raw SQL handle leakage                      |
| Add platform                 | config, platform paths, GUI/build/package                    | platform adapter/composition/evidence                                    | domain/runtime special cases                             | isolate lifecycle/path APIs                           |
| Add Project capability       | future domain/app/persistence/API/UI                         | one vertical use-case slice                                              | SQL directly from UI or REST                             | use internal command/query API                        |
| Add scheduling strategy      | future Task trigger engine                                   | scheduler/trigger adapter + deterministic evaluator                      | Project authority changes                                | keep Project owner authoritative                      |
| Add shared inference         | authorization + queue admission + policy                     | grants, quotas, queue policy, audit                                      | provider credentials/file access                         | preserve inference-only grant boundary                |
| Change peer identity         | identity/trust/envelope                                      | Phase-8 trust subsystem                                                  | transport IDs in domain entities                         | keep transport bindings separate                      |
| Add authentication           | REST gateway, directory, audit                               | auth adapter + app authorization                                         | GUI-internal HTTP                                        | GUI stays direct in-process                           |
| Change persistence schema    | currently database + package-owned Init methods              | migrations/repositories                                                  | scattered CREATE TABLE calls                             | central migration ownership                           |
| Add alternate UI             | currently core owns HUD                                      | UI adapter + app queries                                                 | reimplement backend or carry HUD core                    | remove HUD from core in Phase 7                       |
| Add headless control         | runtime already separate, but same issue                     | app API + CLI/API adapter                                                | duplicate product rules                                  | same internal use cases                               |
| Add queue recovery           | messaging primitive currently unsafe                         | queue state machine + transactions                                       | generic replay table hacks                               | build explicit queue lifecycle                        |

The largest current shotgun-surgery risks are therefore:

1. raw persistence access;
2. application↔GUI dependency direction;
3. identity semantics;
4. Android/build coupling;
5. checklist semantics being treated as implementation semantics.

---

# Part XV — Remaining-Feature Readiness

## Phase 7 — Shared core/local service slice

**Existing support:** runtime/config/SQLite/logging/HUD plus a strong future plan.

**Missing prerequisites:** none outside the repository; this is the prerequisite phase.

**Architectural blockers:** current HUD ownership, raw DB ownership, separate binary roots, weak lifecycle.

**Recommended approach:** execute Phase 7, but add the hardening amendments in Part XXIV.

**State model:** one node root → one lock → one SQLite authority → internal commands/queries → adapter projections.

**Failure tests:** partial startup at every stage, DB locked/corrupt, migration mismatch, repeated close, killed process, restore test, two competing binaries.

**Cross-platform:** validate lifecycle continuously on Linux/Windows/Android.

**Roadmap placement:** first.

---

## Phase 8 — Secure identity and two-device queue proof

**Existing support:** crypto primitives, local manifests, directory/replay scaffolding, detailed security/API design.

**Missing prerequisites:** Phase-7 persistence/lifecycle; identity-v0 migration.

**Blocker:** current “ready” identity must not be promoted directly into network authority.

**Recommended approach:** introduce versioned identity records, strict validation, TLS/device-signing separation, current signed device binding, authenticated commands, then one mock queue operation.

**State model:** signed device record is authority; discovery/cached endpoint is not.

**Failure tests:** revoked peer, expired certificate, replay, duplicate POST, lost acknowledgement, stale directory, bad key binding, restart during accept/complete.

**Security:** this phase is the trust foundation; no compromises.

**Roadmap placement:** immediately after Phase 7.

---

## Phase 9 — Discovery and Project catalog

**Existing support:** owner model and cluster-directory concept.

**Prerequisites:** Phase 8.

**Blockers:** none if trust directory/revision semantics are correct.

**Approach:** discovery yields endpoint candidates only; trusted signed directory yields authorization; owner Project registry yields authoritative summaries.

**State:** local owner Project rows; signed cached remote summaries with explicit freshness.

**Failures:** stale discovery, key change, owner offline, conflicting revision.

**Tests:** deterministic revision resolution plus multi-device integration.

**Placement:** unchanged.

---

## Phase 10 — Project workspaces, Git, manual Tasks

**Existing support:** UI mockups and detailed ownership contract.

**Prerequisites:** secure owner routing + Project catalog.

**Architectural concern:** Git/filesystem operations must remain constrained application operations, not shell wrappers.

**Approach:** owner-side filesystem and Git ports with explicit path/symlink/race validation; drafts retained independently.

**State:** Git remains source authority; SQLite stores metadata/drafts/transactions/artifacts.

**Failures:** file changed since draft base, Git conflict, owner offline, symlink escape, large/binary file, partial artifact transfer.

**Security:** no arbitrary command construction. The roadmap already requires that. 

**Placement:** can proceed in parallel with parts of Phase 11 after Phase 9/8 foundations.

---

## Phase 11 — Local inference services

**Existing support:** conceptual service/capability taxonomy, mock UI.

**Prerequisites:** Phase-7 service slice and secure gateway.

**Approach:** one driver factory may create many independent `ServiceID` instances; driver kind is never identity.

**State:** desired config, observed health, capabilities, advertisement are separate.

**Failures:** provider timeout, malformed inventory, same-provider one-instance failure, model unload, stale observation.

**Dependency work:** research exact OpenAI-compatible/Ollama/llama.cpp interfaces when each driver is implemented—not before.

**Placement:** parallelizable with Phase 10; must complete before Phase 12/13.

---

## Phase 12 — Queue/leasing/results/artifacts

**Existing support:** only primitive messaging scaffolding; do not treat it as production queue logic.

**Prerequisites:** Phase 8 mock queue, Phase 10 artifacts, Phase 11 capabilities.

**Blockers:** replay semantics must already be fixed.

**Approach:** explicit job/attempt/lease state machine with fencing.

**State:** queue owner alone records authoritative order and terminal result.

**Failures:** duplicate submit, worker dies, lease expires, old worker posts late result, requester loses response, artifact hash mismatch.

**Tests:** property/state-transition tests plus multi-process recovery.

**Placement:** after both Phase 10 artifacts and Phase 11 capabilities.

---

## Phase 13 — Pools/routing/real text inference

**Existing support:** route/capability design and mock UI.

**Prerequisites:** queue + services.

**Approach:** deterministic eligibility filtering followed by explicit ordered fallback.

**State:** route profile contains requirements and ordered destinations; provider health remains observation rather than policy.

**Failures:** stale advertisement, destination disabled between choice/lease, timeout, fallback exhaustion.

**Tests:** deterministic routing matrix.

**Placement:** unchanged.

---

## Phase 14 — Tasks/triggers/automation

**Existing support:** Task concept and UI mockups.

**Prerequisites:** Project entrypoints + queue/routing.

**Approach:** trigger requests a run; Task is the executable identity; Project owner remains run authority.

**State:** binding ≠ Task ≠ run. Persist each separately.

**Failures:** duplicate webhook, clock jump, DST transition, restart mid-step, queued job returns after timeout, approval denied.

**Tests:** fake clock, restart recovery, webhook replay.

**Placement:** after Phase 13.

---

## Phase 15 — Speech

**Existing support:** PTT presentation states only.

**Prerequisites:** service and routing system.

**Approach:** platform capture is transient; explicit STT submission becomes durable work.

**State:** recording buffer remains presentation/platform state; transcription job is core state.

**Failures:** permission revoked, app suspended, input device disappears, oversized capture, route unavailable, cancellation on release race.

**Dependency:** inspect whisper.cpp pinned/current source at implementation time.

**Android:** particularly important lifecycle test.

**Placement:** after routing; parts of capture lifecycle can be tested earlier.

---

## Phase 16 — Packaging/release

**Existing support:** comparatively strong build pipeline and Android evidence.

**Missing:** release signing, modern Android target, broader device lifecycle, Windows/Linux packaging, rollback/update provenance.

**Approach:** continuous evidence throughout earlier phases, culminating here.

**Important:** target API 30 cannot be the final Play-distributed configuration given the August 31, 2026 API-36 requirement. ([Android Developers][5])

**Placement:** final formal release phase, but platform tests run continuously.

---

## Post-MVP Track A — transports/resilience

Correctly deferred. The track explicitly waits for stable identity, envelopes, queues, authorization, and transport conformance. 

No Meshtastic/Signal implementation should be selected yet.

---

## Post-MVP Track B — Comrades/shared inference

Correctly deferred until authorization, routing, queues, and audit are stable. The inference-only default and explicit denial of unrelated Project/file/secret/admin access are strong security choices. 

---

## Post-MVP Track C — BOINC/Research

Correctly deferred. In particular, the roadmap says to decide the BOINC integration boundary **before** selecting submodules. 

That is exactly the dependency discipline this review would recommend.

---

# Part XVI — Governance-System Reconstruction

The current development loop is approximately:

```text
User request
   ↓
read AGENTS + README + ROADMAP
   ↓
select relevant host playbook
   ↓
create/select execution plan
   ↓
approval / active plan
   ↓
implement plan atoms
   ↓
update plan
   ↓
update docs
   ↓
run deterministic verification
   ↓
review diff / journal checkpoint
   ↓
commit/push
```

The shared framework explicitly describes the lifecycle as:

> Prompt → Select/Create Plan → Request approval → Execute approved plan atoms → Plan update → Docs update → Verification. 

This system is strong at:

* scope control;
* preserving user work;
* documentation;
* plan history;
* deterministic verification;
* submodule governance;
* reproducibility.

It is weaker at:

* calibrating the strength of a completion claim to its evidence;
* requiring adversarial/failure evidence for semantic properties;
* preventing a broad plan from checking off “architecture nouns” after thin scaffolding exists.

---

# Part XVII — Governance Expectation-vs-Reality Audit

| Governance expectation                                     | Defined where               | Observed implementation                                                    | Compliance            | Likely reason                                           | Response                            |
| ---------------------------------------------------------- | --------------------------- | -------------------------------------------------------------------------- | --------------------- | ------------------------------------------------------- | ----------------------------------- |
| Product rules in domain; external systems in adapters      | Root `AGENTS.md`            | `internal/app` imports GUI/HUD and owns DB-facing objects                  | **Partial**           | Phase-3 prototype preceded mature seam                  | Fix in Phase 7                      |
| Docs and behavior consistent                               | shared invariants           | Roadmap historically marks deeper semantics complete than code establishes | **Partial**           | documentation tracked intent more closely than evidence | Evidence-level governance           |
| Don't mark plan complete without evidence                  | planning playbook           | Phase-3 broad semantic claims have thin evidence                           | **Partial**           | “corresponding evidence” insufficiently typed           | Require claim-specific evidence     |
| Code ≤400 lines                                            | root governance             | Code appears decomposed and gate is automated                              | **Strong**            | deterministic and easy to enforce                       | Retain                              |
| Directory/script documentation                             | root governance             | Extensive README/script docs                                               | **Strong**            | automated check + visible requirement                   | Retain                              |
| Structured/redacted logs                                   | root governance             | structured log exists; redaction key-name heuristic                        | **Partial**           | simple early implementation                             | strengthen before external payloads |
| Durable replay/duplicate tracking                          | roadmap Phase 3             | `Seen` collapses all DB errors into duplicate                              | **Fail semantically** | happy-path test likely proved table uniqueness only     | fix + failure test                  |
| Startup identity consistency classification                | roadmap Phase 3             | classifier mainly checks files exist                                       | **Fail semantically** | “state classifier exists” mistaken for deep consistency | relabel/migrate                     |
| ULID identifiers                                           | roadmap/database contract   | generator is hex timestamp + random hex                                    | **Fail**              | approximate sortable ID substituted                     | fix before schema proliferation     |
| Controller/input equivalence                               | historical Phase 2          | declared action map exceeds actual action execution                        | **Partial**           | mock/prototype checkpoint overgeneralized               | Phase-6/16 reconciliation           |
| External dependency changes reviewed                       | source/submodule governance | dependency pins and source roles clearly recorded                          | **Strong**            | explicit admission process                              | retain                              |
| Semantic judgment not replaced by deterministic heuristics | Agentic Pipelines           | framework itself is strong here; product milestones less so                | **Partial transfer**  | principle not propagated to host milestone completion   | reuse concept in product planning   |

This matrix explains the central governance defect:

> **The repository already has the right epistemic principle, but the product-development layer did not consistently apply it to roadmap completion.**

---

# Part XVIII — Planning-vs-Reality Audit

## What the plan gets right now

The July 19 / Phase-6 rewrite appears to have recognized that earlier completed phases were historical prototype evidence rather than proof of current executable parity.

That reconciliation is a major improvement.

The current Phase-7 plan accurately targets almost every architectural problem discovered independently in this review:

* shared state;
* one node root;
* lock;
* no HUD in core;
* direct GUI internal API;
* loopback REST;
* clock/ID injection;
* transactional startup;
* persistence centralization;
* recovery tests.  

## What Phase 7 still needs

It needs explicit checklist atoms for:

1. **ID contract repair**, not just injection.
2. **migration of package-created Phase-3 tables** into canonical migrations.
3. **replay error semantics**.
4. **legacy identity trust classification/migration boundary**.
5. **desktop Ebitengine context termination**.
6. **Android runtime ownership and truthful readiness**.
7. **function-comment compliance**.
8. **semantic completion-evidence classification**.

Those are small additions with high leverage.

## Ordering revision

I would not radically reorder Phases 8–16.

I would, however, permit **Phase 10 and Phase 11 to run as parallel workstreams** once their shared foundations are stable:

```text
Phase 9 ──→ Project/Git lane ──┐
Phase 8 ──→ Service lane ──────┼→ Phase 12 queue
                               └→ Phase 13 routing
```

That reduces critical-path time without introducing architectural coupling.

---

# Part XIX — Governance-Induced Technical Debt

## 1. Structural verification displaced semantic verification

The process accumulated excellent deterministic checks:

* file presence;
* plan indexes;
* line counts;
* docs inventory;
* unit tests;
* builds;
* lint;
* vuln scanning.

Those are valuable.

But they create a risk: once many green gates exist, a checkpoint can *feel* strongly verified even if the named architectural property was never tested.

Example:

> “startup consistency classification” can pass because a function named `Classify` exists and has tests, even though the function's meaning is much narrower than the roadmap phrase.

## 2. Broad execution plans encouraged shallow horizontal completion

The old Phase-3 plan grouped:

* runtime;
* identity;
* persistence;
* cluster;
* messaging;
* diagnostics;

into one checkpoint.

That makes it tempting to establish a thin slice of each and then mark the conceptual feature complete.

The future plan system is better because Phase 7 is organized around coherent checkpoints and explicit exit criteria.

## 3. UI churn shows an earlier weak feedback loop

The past-plan history contains a burst of HUD recovery/regression plans—scroll-coordinate fixes, layout recovery, scroll-container panic, EbitenUI regressions, settings-first recovery, mobile overflow, drag-selection race. 

Individual bugs are normal.

What matters is that several consecutive plans were correcting assumptions from the previous plan.

The eventual response—EbitenUI consolidation plus visual device evidence—was correct.

The governance lesson should be lightweight:

> after repeated fixes in the same subsystem, require a one-time assumption/root-cause review before the next patch.

Not a new ceremony for every bug.

## 4. Governance itself has recently improved

The August 24 work removed obsolete legacy governance and established one live Agentic Pipelines baseline; the journal notes the next planned implementation is still Phase 7. 

That consolidation should reduce process friction rather than adding another governance layer.

---

# Part XX — Development-Process Retrospective

## What worked well

* Extensive product architecture was written before networking implementation.
* The single-owner distributed model is clear.
* Android was validated on real hardware rather than inferred from compilation.
* Third-party source admission is intentional and documented.
* Source-size and documentation checks are effective.
* Failed GUI approaches were eventually discarded rather than defended.
* The roadmap was explicitly reconciled after implementation reality diverged.
* Agentic Pipelines has unusually thoughtful evidence/authority rules.

## What failed

* Broad checklist wording allowed prototypes to masquerade as semantic completion.
* Identity/persistence primitives were treated as more mature than they are.
* Historical input-parity claims exceeded executable behavior.
* Failure-path testing was weaker than happy-path mechanism testing.
* Runtime/app dependency direction did not initially match the intended architecture.

## What created friction

* repeated UI layout recovery plans;
* duplicate/superseded plans during Android/HUD work;
* build-tool patching around gomobile;
* keeping documentation synchronized with rapidly changing prototype implementation.

## What is missing

* a standard for *strength of evidence* attached to implementation claims;
* explicit function-comment standard;
* lifecycle checklist for long-lived services;
* dependency-source inspection trigger for volatile boundary decisions;
* small post-failure assumption-learning step.

## What is obsolete

The already-removed legacy `agents`/kanban/downtime machinery should stay removed.

Do not recreate it under new names.

## What should be strengthened

* semantic review of completion;
* failure injection;
* state ownership;
* lifecycle tests;
* dependency assumptions;
* comment quality;
* evidence calibration.

## What should be simplified

Do not run ordinary Apparat product work through the full Agentic Pipelines model/rejection-evidence machinery unless the task is actually a reusable semantic pipeline.

Borrow the **principles**, not the bureaucracy.

---

# Part XXI — Recommended Governance Model

I recommend only seven durable changes.

### 1. Claim-calibrated completion

Every nontrivial plan item should indicate the evidence needed:

* structural;
* deterministic unit behavior;
* integration;
* failure/recovery;
* target-platform;
* semantic review.

An item can be marked complete only at the strength its wording claims.

### 2. Evidence-level vocabulary

Use a small vocabulary when useful:

```text
designed
scaffolded
local-tested
integration-tested
target-validated
```

Do not automatically convert “scaffolded” into `[x] implemented` if the roadmap claim describes target behavior.

### 3. Function-comment invariant

Add to Apparat `AGENTS.md`:

> Every project-owned function or subroutine has a human-useful leading comment. Nontrivial functions explain intent, invariants, state effects, lifecycle/concurrency, cancellation, and failure behavior where applicable.

Presence can be deterministic; usefulness is semantic.

### 4. Long-lived-worker contract

Any new goroutine/server/supervisor must document:

```text
owner
start condition
lifetime
state owned
context/cancellation
deadline/retry
failure surfacing
shutdown
restart behavior
```

### 5. External dependency research trigger

Require upstream source inspection when a decision depends on:

* undocumented lifecycle behavior;
* concurrency behavior;
* platform limitation;
* unstable/version-specific API;
* security guarantee;
* extension mechanism.

Do **not** require a deep upstream audit for ordinary narrow API use.

### 6. Lightweight end-of-run learning

When a task had meaningful rework/failure, add three lines to its closeout:

* wrong/missing assumption;
* earlier evidence that could have exposed it;
* durable action: local fix / plan change / governance change / no system change.

### 7. Reviewer verifies the *claim*, not merely the diff

A review should ask:

> “What exact proposition are we declaring true, and what evidence establishes that proposition?”

This is the most important governance change.

---

# Part XXII — Decisions Whose Reversal Cost Is Increasing

## 1. Node ownership and runtime root

**Current:** two binary-oriented runtime defaults still exist conceptually in current implementation.

**Future dependents:** every local service, Project, queue, identity, advertisement.

**Cost inflection:** immediately after Phase 11 starts persisting service instances.

**Recommendation:** settle in Phase 7: one logical node, one root, one writer lock.

---

## 2. Durable ID format

**Current:** roadmap says ULID; implementation generates a different hex format. `NewID` formats a millisecond value and 10 random bytes as hex. 

Canonical ULIDs are 26-character Crockford Base32 strings with a 48-bit millisecond timestamp and 80-bit entropy. ([GitHub][6])

**Future dependents:** Projects, jobs, attempts, leases, services, capabilities, transactions, artifacts, audits.

**Cost inflection:** as soon as Phase-7/8 schema expands.

**Recommendation:** fix now. Either implement actual ULID or revise every contract to name a different identifier. Given how consistently the architecture specifies ULID, use a real ULID.

---

## 3. Identity format

**Current:** useful local cryptographic scaffolding, weak trust semantics.

**Future dependents:** every remote request and ownership assertion.

**Cost inflection:** first enrolled peer.

**Recommendation:** version/migrate before Phase 8.

---

## 4. Application/API dependency direction

**Current:** app imports GUI and owns HUD.

**Future dependents:** REST, CLI, scheduler, alternate UI, Android.

**Cost inflection:** once each adapter starts coding around `Runtime`.

**Recommendation:** reverse now.

---

## 5. Persistence access

**Current:** raw SQL handle escapes and subsystem Init methods create schema.

**Future dependents:** every durable feature.

**Cost inflection:** Phase 8 migrations onward.

**Recommendation:** centralize before adding schemas.

---

## 6. Worker lifecycle convention

**Current:** largely not implemented.

**Future dependents:** HTTP server, service health, inference execution, queues, trigger scheduler, audio, transport retries.

**Cost inflection:** Phase 8/11.

**Recommendation:** define in Phase 7 while there are almost no workers.

---

## 7. Capability identity

**Current:** mostly design/mock state.

**Future dependents:** service discovery, routing, queues, Comrades, Research.

**Cost inflection:** Phase 11.

**Recommendation:** preserve the documented separation:

```text
WorkloadClass
DriverKind
ServiceID
CapabilityID
```

Do not collapse them.

---

## 8. Android build/runtime lifecycle

**Current:** working wrapper but bespoke build accommodations and weak core lifecycle ownership.

**Future dependents:** speech, background service behavior, upgrades.

**Cost inflection:** once providers/queues exist on Android.

**Recommendation:** core lifecycle now; distribution/toolchain modernization by Phase 16.

---

# Part XXIII — Recommended Target Architecture

The minimum sufficient target architecture is not substantially different from the current Phase-7 vision.

```text
                    ┌────────────────────┐
                    │      domain        │
                    │ values/invariants  │
                    └─────────▲──────────┘
                              │
                    ┌─────────┴──────────┐
                    │    application     │
                    │ commands / queries │
                    │ orchestration      │
                    └────▲─────▲────▲────┘
                         │     │    │
          ┌──────────────┘     │    └───────────────┐
          │                    │                    │
 ┌────────┴───────┐   ┌────────┴───────┐   ┌────────┴────────┐
 │ GUI adapter    │   │ HTTP adapter   │   │ CLI / triggers │
 │ Ebiten/EbitenUI│   │ REST DTO/auth  │   │ later          │
 └────────────────┘   └────────────────┘   └─────────────────┘

 application outward ports
          │
 ┌────────┴────────┐
 │ persistence     │ SQLite
 │ provider drivers│ Ollama/etc
 │ artifact/Git    │
 └─────────────────┘

 platform
 ├─ node lock
 ├─ paths
 ├─ signals
 ├─ Android lifecycle
 └─ OS credential storage
```

## Why each boundary earns its cost

### Application boundary

Exists because GUI, REST, Task triggers, and tests will genuinely share behavior.

### Persistence adapter

Exists because raw SQL leakage is already causing migration/config/error-semantic problems.

### GUI adapter

Exists because Ebitengine has real lifecycle/thread constraints and the product explicitly needs headless operation.

### Provider driver boundary

Exists because several real provider kinds are already planned.

### Platform boundary

Exists because Android/desktop lifecycle and lock/path semantics genuinely differ.

No generic event bus, DI container, dynamic plugin framework, “repository for every struct,” or microservice split is needed.

---

# Part XXIV — Checkpoint-Hardening Plan

## Must fix before substantial feature development resumes

### H1 — Amend Phase 7 into Foundation Convergence

Keep its current scope but add the following explicit atoms.

### H2 — Repair ID semantics

* choose real ULID;
* inject ID source;
* test lexical/time behavior as required;
* migrate any persisted development IDs only if they matter.

### H3 — Consolidate SQLite authority

* all owned schema through migrations;
* no subsystem `CREATE TABLE` at runtime outside migrations;
* no raw `*sql.DB` above persistence adapter;
* connection/foreign-key policy defined correctly;
* typed transactions;
* busy/cancel/error classification;
* backup/integrity/restore tests.

### H4 — Correct replay semantics

Only uniqueness violation means duplicate. Return all other errors.

### H5 — Establish transactional runtime lifecycle

* `starting → ready → stopping → stopped/failed`;
* reverse unwind;
* idempotent `Close`;
* node lock;
* supervised worker ownership.

### H6 — Fix presentation lifecycle

* app core no HUD shell;
* `cmd` performs composition;
* Ebitengine receives cancellation and returns `ebiten.Termination`;
* Android owns runtime lifetime explicitly;
* Android `Ready()` reflects real initialization state.

### H7 — Quarantine identity v0 from trust

* explicitly classify current files as local bootstrap identity;
* strict decode/length checks;
* version the encrypted envelope;
* document Phase-8 migration;
* current `StatusReady` grants no remote authority.

### H8 — Establish comment standard

Backfill all project-owned functions. Prioritize:

1. runtime;
2. persistence;
3. identity;
4. cluster/messaging;
5. GUI lifecycle/input;
6. tests/helpers.

### H9 — Change completion governance

Add evidence class and semantic claim review before Phase 7 can be closed.

---

## Fix while implementing the next related feature

* schema-aware/typed log redaction — Phase 7/8.
* full v1 identity/cert lifecycle — Phase 8.
* complete controller/action parity — as controls become real, with final Phase-16 gate.
* provider error normalization — Phase 11.
* complete artifact transaction semantics — Phase 10/12.
* target-SDK modernization — Android validation lane before Phase 16 release.

---

## Can safely defer

* WAL;
* daemon-client GUI mode;
* CRDT;
* scheduler election;
* dynamic load/cost routing;
* Signal;
* Meshtastic;
* BOINC;
* app-managed WireGuard;
* out-of-process provider plugins;
* Android headless APK.

---

## Intentionally leave unchanged

* Go;
* SQLite;
* Ebitengine/EbitenUI;
* external-WireGuard-first strategy;
* one authoritative owner;
* static provider registration;
* REST before WebSockets;
* no unrestricted shell;
* Git remains Project-file authority;
* local provider endpoints remain private.

---

# Part XXV — Migration and Implementation Strategy

## Step 1 — Preserve behavior while changing ownership

Move composition, not functionality:

* keep existing GUI;
* keep existing DB contents;
* expose existing read state behind initial application queries;
* have GUI read those queries;
* remove `hud.Shell` from core only once equivalent GUI-local state exists.

## Step 2 — Introduce new persistence adapter around existing database

Do not rewrite the schema all at once.

1. take ownership of opening/migrations;
2. migrate cluster/messaging schema creation;
3. wrap current operations;
4. add transactions/error classification;
5. stop exporting raw SQL;
6. delete old database package path when unused.

## Step 3 — Replace weak primitives before new consumers exist

Correct:

* ID generation;
* replay;
* identity envelope parsing;
* runtime state.

This is cheap now because few features depend on them.

## Step 4 — Make GUI/headless consumers of the same application use cases

GUI direct calls.

Loopback REST separate adapter.

No GUI-over-HTTP.

## Step 5 — Add security only after the local semantics are trustworthy

Phase 8 then becomes security added around a known durable application core instead of security being used to stabilize an unstable core.

---

# Part XXVI — Synchronized Engineering/Governance Roadmap

| Engineering change           | Governance/process change                           | Why paired                                            | Enables                 |
| ---------------------------- | --------------------------------------------------- | ----------------------------------------------------- | ----------------------- |
| Correct persistence boundary | require claim-specific persistence/failure evidence | prevents “table exists = durability solved”           | every future schema     |
| Correct ID contract          | exact-contract verification                         | prevents approximate protocol primitives              | stable external IDs     |
| Runtime lifecycle            | worker ownership checklist                          | every future goroutine inherits a stop/error model    | REST/providers/queues   |
| Remove HUD from core         | architecture-boundary semantic review               | prevents adapters re-entering core                    | CLI/REST/headless       |
| Identity-v0 quarantine       | security claim evidence                             | prevents local file state becoming trust accidentally | secure Phase 8          |
| Replay fix                   | failure-path evidence category                      | duplicate semantics are correctness-critical          | at-least-once messaging |
| Function comments            | comment-presence gate + semantic review             | makes state/lifecycle intent human-visible            | safer future editing    |
| Phase-8 secure queue         | security/failure matrix required                    | distributed success needs adverse-case proof          | discovery               |
| Service manager              | dependency-source inspection when needed            | avoids provider-name assumptions                      | Phase 11/13             |
| Android lifecycle            | target-device evidence category                     | compilation cannot prove lifecycle                    | speech/release          |
| Platform release             | support-claim evidence                              | prevents “builds = supported”                         | honest distribution     |

---

# Part XXVII — Second-Draft Project Plan

## Draft 2 Phase A — Foundation Convergence

### Goal

Turn the current prototype foundation into the single trustworthy local node substrate on which all remaining features depend.

### Why now

Every subsequent phase multiplies persisted identities, background workers, and externally visible contracts.

### Prerequisites

Current Phase-6 reconciliation.

### Architectural work

* one node root and lock;
* true ULID contract;
* application dependency direction;
* no HUD in core;
* canonical SQLite adapter/migrations;
* transactional lifecycle;
* replay correction;
* clock/ID ports;
* direct GUI commands/queries;
* read-only loopback REST;
* versioned local identity-v0 handling;
* truthful Android lifecycle/readiness.

### Dependency work

Verify Ebitengine lifecycle at current version; no new runtime dependency.

### User-facing work

Routing/Cluster view begins consuming real local durable service read models.

### Governance changes

Introduce claim-calibrated evidence and function-comment rule.

### Tests

* migration;
* lock contention;
* partial startup;
* repeated close;
* DB failure;
* replay non-duplicate error;
* backup/restore;
* deterministic ID/clock;
* GUI/API same-source projections.

### Cross-platform

Linux + Windows build tests + Android device lifecycle smoke.

### Failure validation

Kill/restart, DB lock, corrupt backup, unavailable paths.

### Exit

Phase-7 existing criteria plus H1–H9 above.

### Enables

Every remaining phase.

---

## Draft 2 Phase B — Secure Two-Device Authority

### Goal

Prove one real authenticated distributed state transition.

### Why now

Trust must exist before discovery and remote state gain meaning.

### Prerequisites

Phase A.

### Architecture

* identity v1;
* enrollment;
* root CA/mTLS;
* signed device records;
* authorization scopes;
* signed envelopes;
* replay/idempotency;
* trusted-device directory;
* authenticated REST commands;
* reusable mock queue.

### Dependency work

Standard Go TLS/crypto only unless a concrete gap appears.

### User-facing

Enroll second device; submit/observe echo/mock task.

### Governance

Any security completion claim requires negative tests.

### Tests

revocation, expiry, mismatched key, replay, lost response, restart, duplicate POST.

### Cross-platform

Linux-to-Linux first; then at least Android-client secure read/command evidence where feasible.

### Exit

One logical job survives disconnect/restart and cannot double-complete.

### Enables

Discovery and remote Projects.

---

## Draft 2 Phase C — Discovery and Project Ownership

### Goal

Build a truthful cluster-wide Project catalog.

### Prerequisites

Phase B.

### Architecture

* endpoint discovery as suggestion;
* signed/revisioned trusted records;
* owner-local Project registry;
* stale cached summary model.

### User features

Every authorized Project appears with owner/freshness/offline status.

### Tests

partition, stale record, key change, conflicting revision.

### Exit

No cached/discovered state can confer authority.

### Enables

Real workspaces.

---

## Draft 2 Phase D1 — Project Workspace Lane

### Goal

Real owner-local and owner-remote Project operations.

### Architecture

* constrained filesystem;
* Git service;
* drafts;
* transactions;
* artifacts;
* manual Task entrypoints.

### Security

No shell; path/symlink scope validation.

### Tests

conflicts, concurrent file change, remote owner unavailable, artifact interruption.

### Exit

Local/remote safe Git and one manual Task work.

---

## Draft 2 Phase D2 — Local Service Lane

**Runs substantially in parallel with D1 after shared prerequisites.**

### Goal

Manage several independent local inference service instances.

### Architecture

* workload-class registry;
* `DriverKind`;
* `ServiceID`;
* `CapabilityID`;
* desired/observed state separation;
* static driver factories;
* supervised instances;
* safe advertisements.

### Dependency research

Start with OpenAI-compatible interface. Inspect Ollama/llama.cpp upstream only as those drivers are admitted.

### Tests

two same-provider instances, one failing without poisoning another, restart, stale capability.

### Exit

Several real/mock services coexist independently.

---

## Draft 2 Phase E — Owner-Authoritative Queue and Artifact Recovery

### Goal

Make distributed work durable and duplicate-safe.

### Prerequisites

D1 artifacts + D2 capabilities + secure authority.

### Architecture

* queue definitions;
* jobs;
* attempts;
* leases;
* fencing;
* heartbeat;
* cancellation;
* result validation;
* artifact transfer/recovery.

### Tests

late worker, worker death, requester restart, owner restart, duplicate completion, corrupt artifact.

### Exit

Exactly one authoritative logical completion under all tested retry/recovery cases.

---

## Draft 2 Phase F — Pools, Routing, and First Real Inference

### Goal

Deliver one real text-generation request through the complete architecture.

### Architecture

* pool membership;
* deterministic eligibility;
* route profiles;
* ordered fallback;
* real text driver.

### Tests

incompatible model, stale service, service dies after selection, fallback.

### Exit

Routing explains every inclusion/exclusion and real inference survives retry/restart.

---

## Draft 2 Phase G — Durable Automation

### Goal

Run the same Project Task manually or via authorized triggers.

### Architecture

* separate trigger bindings;
* scheduler;
* authenticated webhook adapter;
* internal events;
* approvals;
* durable workflow state.

### Tests

DST/timezone, duplicate webhook, restart mid-step, approval denial, stuck inference.

### Exit

One logical run remains idempotent through triggers and restart.

---

## Draft 2 Phase H — Voice and Audio Lifecycle

### Goal

Convert PTT from mock UI state to real routed STT/TTS.

### Architecture

* platform capture adapter;
* transient bounded audio;
* STT/TTS service drivers;
* queue/routing integration;
* explicit privacy retention.

### Dependency

Research exact whisper.cpp integration at this phase.

### Cross-platform

Steam Deck, Debian, Android permissions/suspend/resume.

### Tests

cancel/release race, permission denial, app pause, device removal, oversized capture.

### Exit

Editable transcription and independent speech output are reliable and private.

---

## Draft 2 Phase I — Platform and Release Hardening

### Goal

Turn evidence-producing development artifacts into honestly supported products.

### Linux/Steam Deck

* package/service integration;
* controller parity;
* microphone/audio;
* update/rollback;
* WireGuard/LAN.

### Windows

* GUI/headless packaging;
* service lifecycle;
* external WireGuard;
* input/audio;
* update/rollback.

### Android

* modern target API;
* release signing;
* remove/contain fragile tool patching;
* full lifecycle;
* safe areas/density;
* keyboard/controller/touch;
* portrait/landscape;
* update provenance.

### Governance

Support claim requires target-platform evidence—not compilation.

### Exit

Every claimed platform has independently reproduced installation, startup, persistence, networking, security, input, audio where applicable, shutdown/recovery, update, and rollback evidence.

---

## Draft 2 Post-MVP Tracks

### Track A — Alternative transports and resilience

Only after stable common transport conformance.

### Track B — Comrades and shared inference

Only after secure queues/routing/audit.

### Track C — Research/BOINC/gameplay

Only after resource policy, scheduling, isolation, audit, and packaging are mature.

These remain independent tracks rather than being forced into the MVP critical path.

---

# Final synthesis

If development simply continued from the current checkpoint **without modifying the current architecture/governance assumptions**, the most expensive future problems would not be the GUI or ordinary code cleanliness.

They would be:

1. **persisted identity and ID contracts becoming externally visible before they are correct;**
2. **raw SQLite ownership spreading into every future subsystem;**
3. **HTTP/providers/queues/schedulers creating long-lived concurrency before ownership and shutdown semantics are established;**
4. **the current local identity prototype accidentally becoming the basis of network trust;**
5. **GUI, REST, and headless behavior developing parallel application rules because the internal API boundary was not completed first;**
6. **historical checklist completion continuing to outrun the strength of actual implementation evidence;**
7. **human comprehensibility degrading sharply as distributed state is added to an almost entirely uncommented codebase.**

Those are cheap to change **now** and progressively expensive after Phases 8–14.

The exact sequence I recommend is therefore:

```text
1. Amend Phase 7 into the Foundation-Convergence checkpoint.
        ↓
2. Fix IDs, persistence ownership, replay semantics,
   runtime lifecycle, GUI dependency direction, and Android lifecycle.
        ↓
3. Explicitly quarantine/migrate the current identity format;
   do not let present "ready" status imply remote trust.
        ↓
4. Add the function-comment standard and claim-calibrated
   completion evidence to the governance workflow.
        ↓
5. Close Phase 7 only with restart/failure/restore/platform evidence.
        ↓
6. Implement Phase 8 secure two-device authority.
        ↓
7. Build discovery/Project ownership.
        ↓
8. Develop Project workspaces and local service drivers in parallel.
        ↓
9. Join them at the authoritative queue/artifact layer.
        ↓
10. Add routing and real inference.
        ↓
11. Add durable automation.
        ↓
12. Add speech.
        ↓
13. Culminate continuous platform testing in release hardening.
        ↓
14. Only then activate Comrades, alternate transports,
    BOINC/Research, failover, or dynamic routing.
```

The architectural conclusion is therefore **not** “Apparat needs to be redesigned.” It is almost the opposite:

> **The repository has now designed a considerably better architecture than the one its early prototype actually implements. The correct move is to make Phase 7 the point where implementation finally converges on that design, correct the handful of foundational semantic mismatches before they become protocol history, and strengthen governance so future completion claims prove the property they name rather than merely the presence of a mechanism.**

Once that is done, the remaining roadmap is unusually well positioned for incremental development. The high-value choices—single authority, transport-neutral application semantics, typed capabilities, local provider isolation, SQLite durability, external WireGuard first, deterministic routing before optimization, and no unrestricted remote execution—should be preserved rather than revisited. 

[1]: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2 "https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2"
[2]: https://github.com/hajimehoshi/ebiten/releases "https://github.com/hajimehoshi/ebiten/releases"
[3]: https://www.sqlite.org/foreignkeys.html "https://www.sqlite.org/foreignkeys.html"
[4]: https://support.google.com/googleplay/android-developer/answer/11926878?hl=en "https://support.google.com/googleplay/android-developer/answer/11926878?hl=en"
[5]: https://developer.android.com/google/play/requirements/target-sdk "https://developer.android.com/google/play/requirements/target-sdk"
[6]: https://github.com/ulid/spec "https://github.com/ulid/spec"
