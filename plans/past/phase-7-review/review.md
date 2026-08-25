# Deep Research Prompt: Major Checkpoint Architecture, Codebase, Governance, and Project-Plan Review

## Mission

Perform a comprehensive major-checkpoint review of the attached software repository and its agentic development-governance system.

The primary repository is provided as a GitIngest. The GitIngest contains the application's source code together with the project's governance documents, current plans, future plans, remaining feature definitions, development methodology, agent instructions, workflow rules, completion criteria, and other project-management material.

The repository uses one or more **Git submodules whose complete source is not included in the attached GitIngest because including them would exceed the available input size**.

When the architecture, behavior, API contracts, lifecycle, reliability, or future feature-readiness of a submodule is materially relevant to the analysis, **research that submodule online using its repository URL, module path, commit reference, documentation, source repository, or other identifying information available in the primary repository.**

Do not expect the current/future plans, remaining feature list, methodology, or governance instructions to be separately restated in this prompt. Discover them from the repository and treat them as first-class evidence.

This is **not** a conventional code review. Do not focus primarily on lint issues, isolated style preferences, individual bugs, or generic "clean code" recommendations.

Treat this task as a combined:

* software architecture review;
* maintainability and human-editability audit;
* reliability and structural-integrity assessment;
* distributed-systems and concurrency review;
* cross-platform readiness assessment;
* dependency and submodule architecture review;
* feature-readiness analysis;
* development-governance audit;
* retrospective on the agentic development process;
* architecture and governance gap analysis;
* checkpoint-hardening plan; and
* second-draft project roadmap for the remainder of the project.

The central question is:

> **Given what this system is intended to become, are the current codebase, dependency/submodule architecture, and development-governance framework the right foundation for completing it successfully, and what should change now before the cost of changing those decisions becomes substantially greater?**

The final report must provide detailed, actionable, evidence-based recommendations for both:

1. **the software system itself**, including materially relevant submodules and external dependencies; and
2. **the agentic governance system responsible for producing and modifying it.**

---

# 1. Project Context

This is a **Go application using Ebitengine**.

It targets:

* Windows;
* Linux; and
* Android.

At a high level, the application provides a user interface for clustering and managing a user's personal inference devices, managing projects that use those resources, and connecting with peers for capabilities including communication and shared inference resources.

The detailed current architecture, intended future architecture, current project plan, remaining features, development methodology, and implementation expectations are contained within the attached repository.

**Do not rely on this high-level summary as a substitute for discovering the project's actual intended behavior from its governance and planning documents.**

The repository also contains a substantial **agent-development framework and governance system**.

These materials are part of the system being reviewed.

Do not treat files such as:

* `AGENTS.md`;
* architectural instructions;
* project plans;
* future plans;
* workflow documents;
* task-management documents;
* completion gates;
* validation rules;
* review procedures;
* coding standards;
* agent instructions;
* research methodology;
* retrospective procedures;
* implementation guidance;

as peripheral documentation.

Treat them as a **development-governance system whose output is the codebase itself**.

The review must therefore examine not only whether the code is good, but whether the development system has reliably caused good engineering behavior.

---

# 2. Evidence Model

The attached GitIngest is the primary evidence base for the project-specific review.

It contains multiple kinds of evidence that must be distinguished carefully.

## Implementation evidence

Source code, tests, build files, configuration, platform-specific code, submodule declarations, integration code, and other artifacts reveal:

> **What the primary project actually is today.**

## Governance evidence

Agent instructions, coding standards, architectural rules, workflow documents, completion gates, and similar material reveal:

> **How development is intended to be governed.**

## Planning evidence

Current plans, future plans, feature lists, milestones, task documents, roadmaps, and related files reveal:

> **What the project is intended to become and how development is currently expected to proceed.**

## Methodology evidence

Research procedures, agent methodologies, planning methodology, review methodology, and related instructions reveal:

> **How the project expects agents to reason about and perform development work.**

## External submodule/dependency evidence

For submodules omitted from the GitIngest because of input-size limits, authoritative external sources may be required to determine:

* what the submodule does;
* its public API;
* its internal architecture where materially relevant;
* lifecycle behavior;
* concurrency assumptions;
* platform support;
* persistence or networking behavior;
* failure modes;
* security implications;
* extension mechanisms;
* compatibility constraints;
* whether the primary project uses it correctly.

Do not collapse these categories.

A planning document saying that a feature exists is not proof that it exists.

A governance rule requiring a practice is not proof that the practice has been followed.

A source file showing a pattern is not necessarily proof that the pattern was intentional.

A submodule name or import path is not sufficient evidence for assumptions about that submodule's internals.

Your job is to compare these sources.

---

# 3. Submodule and External Dependency Research Requirement

The GitIngest may contain references to Git submodules without embedding their complete contents.

You are responsible for identifying those submodules and determining whether their details are material to the review.

For each submodule:

1. identify its repository or module identity from `.gitmodules`, Go module metadata, imports, documentation, configuration, or other repository evidence;
2. identify the version, branch, tag, or commit pinned by the primary repository when possible;
3. determine how the primary application depends on it;
4. determine which aspects of the submodule are relevant to the architecture review;
5. research those aspects online from authoritative sources when necessary.

Prefer, in descending order:

1. the exact upstream source repository at the pinned commit/version;
2. official project documentation;
3. official API/package documentation;
4. release notes or changelogs for the relevant version;
5. upstream issue/discussion material only when necessary to understand known architectural or reliability concerns.

Do not substitute generic descriptions, blog posts, or unrelated newer versions when the primary repository pins a specific revision.

If the exact pinned revision cannot be found or inspected, say so explicitly and distinguish verified facts from inference.

---

# 4. Do Not Over-Research Submodules

Submodule research should be **need-driven**.

Do not perform a complete independent audit of every external repository merely because it appears as a submodule.

Research a submodule deeply when its behavior materially affects questions such as:

* state ownership;
* concurrency;
* networking;
* lifecycle;
* Ebitengine integration;
* Android behavior;
* device discovery;
* inference dispatch;
* peer communication;
* persistence;
* security/trust boundaries;
* future feature implementation;
* reliability;
* extension boundaries;
* architectural coupling.

If the primary repository wraps a submodule behind a narrow and stable interface and its internals do not materially affect the requested architectural conclusions, a high-level review may be sufficient.

Use research effort proportionally to architectural importance.

---

# 5. Distinguish Project-Owned Code From External Code

Throughout the report, clearly distinguish among:

* **primary-project code**;
* **project-owned submodules**, if any;
* **third-party submodules**;
* **ordinary external dependencies**.

This distinction matters for recommendations.

For example:

* project-owned code can potentially be refactored directly;
* a project-owned submodule may warrant changes in both repositories;
* a third-party dependency should usually be wrapped, configured, upgraded, replaced, or adapted rather than internally rewritten;
* implementation details of an external dependency should not be treated as project governance violations.

Do not attribute shortcomings in third-party code to the project's coding agents unless the real problem is the project's choice, configuration, coupling, or integration of that dependency.

---

# 6. Dependency-Boundary Analysis

For materially important submodules and dependencies, evaluate the quality of the primary project's boundary with them.

Ask:

* Is the external system imported throughout the codebase or localized?
* Does the primary application depend on its concrete internal types unnecessarily?
* Does external state leak directly into UI/domain code?
* Could the dependency be changed or upgraded without widespread modifications?
* Are version-specific assumptions scattered throughout the project?
* Are errors and lifecycle concerns translated cleanly at the boundary?
* Are platform limitations isolated?
* Is external nondeterminism pushed toward a boundary?
* Is the submodule effectively part of the domain model when it should instead be an implementation detail?
* Does the application require a small adapter boundary, or would adding one merely create unnecessary indirection?

Treat dependency isolation as another form of **change-amplification analysis**.

---

# 7. Begin With Repository Discovery

Before evaluating the project, identify the important governance, architecture, planning, methodology, roadmap, dependency, and submodule information contained in the GitIngest.

Construct an inventory of the repository's major informational sources.

Determine:

* which documents define current project state;
* which documents define the intended future state;
* which documents list remaining features;
* which documents define development methodology;
* which documents govern coding agents;
* which documents define architectural standards;
* which documents define testing and validation;
* which documents define completion;
* which documents contain prior architectural decisions;
* which documents describe known technical debt;
* which documents define agent roles;
* which documents appear canonical;
* which documents may be obsolete, duplicated, or contradictory;
* which submodules exist;
* what revision/version of each submodule is expected;
* where each submodule participates in the primary architecture.

When several documents address the same subject, determine their apparent hierarchy and chronology where possible.

Explicitly identify ambiguity if it is unclear which document is authoritative.

---

# 8. Reconstruct Both the Current State and Intended Future State

The review must reconstruct two different models.

## Current-state model

Derived primarily from primary-project implementation evidence and, where necessary, verified external submodule behavior:

> **What actually exists and how it actually works.**

## Intended-future model

Derived primarily from governance, plans, feature definitions, methodology, and architectural guidance:

> **What the project is trying to become.**

Do not prematurely force one model onto the other.

First reconstruct them independently.

Then compare them.

---

# 9. Core Research Sequence

Use the following reasoning sequence.

1. inventory and classify the important repository documents;
2. inventory submodules and major external architectural dependencies;
3. reconstruct the current software architecture;
4. identify where omitted submodule details are material;
5. research those submodules online as necessary;
6. reconstruct the current development-governance architecture;
7. reconstruct the current project plan;
8. reconstruct the intended mature system;
9. identify the remaining features and milestones;
10. reconstruct the prescribed development and research methodology;
11. determine how the software actually behaves;
12. determine what governance actually requires;
13. determine what the current roadmap assumes;
14. identify major architectural commitments already present;
15. identify major dependency/submodule commitments already present;
16. compare governance expectations with implementation outcomes;
17. compare planning assumptions with implementation reality;
18. investigate why meaningful deviations occurred;
19. evaluate architecture against remaining features;
20. identify architecture, dependency, process, planning, and governance gaps;
21. identify decisions whose cost of reversal is increasing;
22. determine the minimum sufficient target architecture;
23. determine the minimum sufficient target governance model;
24. identify checkpoint-hardening work;
25. establish dependencies among architectural work and remaining features;
26. revise implementation sequencing;
27. produce a second-draft project plan derived from the research.

The revised roadmap must be a **consequence of the investigation**, not merely a restatement or polishing of the existing plan.

---

# 10. Evidence Standard

Avoid unsupported generalizations.

For every major finding, identify the supporting evidence as precisely as possible.

Use:

* package;
* file;
* type;
* interface;
* function;
* subsystem;
* governance document;
* planning document;
* methodology document;
* dependency;
* submodule;
* upstream file/symbol where material;
* pinned commit/version where known;
* runtime flow;
* state relationship.

For externally researched claims, identify the authoritative upstream source and relevant version/revision.

For each significant finding, distinguish:

### Observation

What concretely exists?

### Evidence

Where is it found?

### Diagnosis

What condition does the evidence indicate?

### Implication

Why does it matter?

### Recommendation

What, if anything, should change?

### Priority

Classify it as:

* **Checkpoint blocker — fix before substantial feature development continues**
* **High priority — address during the next relevant implementation phase**
* **Medium priority — important but safely deferrable**
* **Low priority / cleanup**
* **Leave alone — current implementation is appropriate and changing it would add unnecessary risk or complexity**

Do not treat an unusual implementation as evidence that it should automatically be replaced.

---

# 11. Reconstruct the Current Software Architecture

Before recommending refactors, reconstruct the current software in detail.

Identify:

* major packages;
* package responsibilities;
* important structs and types;
* important interfaces;
* subsystem boundaries;
* dependency direction;
* submodule boundaries;
* external-dependency boundaries;
* global state;
* shared mutable state;
* authoritative state ownership;
* persistence mechanisms;
* UI architecture;
* networking architecture;
* inference-provider architecture;
* device representation;
* peer representation;
* project representation;
* scheduling/resource coordination;
* platform-specific components;
* configuration systems;
* initialization paths;
* runtime lifecycle;
* background workers;
* goroutines;
* channels;
* synchronization primitives;
* context propagation;
* cancellation mechanisms;
* shutdown behavior;
* error propagation;
* extension points.

Where architecture is unclear, duplicated, inconsistent, or emergent rather than intentional, explicitly say so.

---

# 12. Reconstruct Important Runtime Flows

Trace representative execution paths.

These should include whichever of the following are present or intended according to the repository:

* application startup;
* configuration loading;
* persistence loading;
* Ebitengine initialization;
* UI interaction resulting in application behavior;
* local device discovery;
* device registration;
* inference-provider discovery;
* model discovery;
* project creation or loading;
* inference-job creation;
* inference dispatch;
* job cancellation;
* device failure;
* device disappearance;
* peer discovery;
* peer connection;
* chat flow;
* peer resource advertisement;
* remote inference dispatch;
* asynchronous state updates;
* persistence changes;
* application shutdown.

When a flow crosses into a submodule, inspect enough of that submodule to understand the material behavior rather than treating the boundary as a black box if its internals affect the conclusion.

Identify which component owns each state transition.

---

# 13. Reconstruct the Intended Mature Architecture and Feature Set

Using the governance and planning documents contained in the repository, determine:

* the mature intended capabilities;
* the remaining user-facing features;
* architectural prerequisites already anticipated;
* planned milestones;
* planned feature ordering;
* expected cross-platform behavior;
* expected peer/network behavior;
* intended resource-sharing semantics;
* intended project-management behavior;
* intended inference-device behavior;
* intended governance/process improvements;
* expected roles of existing submodules;
* whether new features depend strongly on assumptions about those submodules.

Do not assume every plan remains correct.

Treat the current roadmap as a **proposal subject to architectural review**.

---

# 14. Human Readability and Human Editability

Perform a detailed maintainability assessment of the **project-owned code**.

Evaluate:

* naming;
* package organization;
* file organization;
* function size;
* function complexity;
* cognitive load;
* responsibility boundaries;
* duplicated logic;
* hidden dependencies;
* global state;
* mutable state;
* implicit control flow;
* configuration placement;
* public and internal APIs;
* error handling;
* discoverability;
* ease of tracing behavior;
* ease of understanding runtime state;
* ease of safely changing one subsystem.

A core measure should be **locality of change**.

Repeatedly ask:

> **Can a developer modify one behavior without understanding and modifying several unrelated subsystems?**

Include dependency/submodule coupling in this analysis.

---

# 15. Explicit Function-Comment Standard

Evaluate the project-owned codebase against this requirement:

> **Every function and subroutine should have a human-useful comment at its beginning explaining what it does and why it exists.**

For nontrivial functions, this should normally be a detailed multi-line comment.

Where applicable, comments should explain:

* responsibility;
* purpose;
* why the function exists separately;
* assumptions;
* preconditions;
* important invariants;
* non-obvious inputs/outputs;
* externally visible effects;
* state read;
* state mutated;
* concurrency expectations;
* lifecycle;
* cancellation;
* error behavior;
* important implementation decisions;
* relationship to surrounding subsystems.

Comments should communicate **intent and context**, not merely translate the syntax into English.

For trivial accessors or extremely obvious functions, comments may be concise.

Identify:

* undocumented functions;
* inadequate comments;
* comments that merely restate code;
* stale comments;
* misleading comments;
* missing descriptions of side effects;
* missing concurrency/lifecycle documentation;
* complex routines lacking architectural explanation.

Provide representative improved examples for a small number of especially important functions.

Do not treat missing comments in third-party dependencies as project comment-standard violations.

---

# 16. Idiomatic Go and SOLID Analysis

Evaluate relevant SOLID principles through idiomatic Go.

Do not mechanically impose Java/C++ architecture.

Assess:

## Single Responsibility

Are functions, types, files, and packages responsible for coherent concerns?

## Open/Closed

Can clearly anticipated extensions be added without widespread modification of unrelated code?

## Liskov Substitution

Where interchangeable implementations exist, are their contracts coherent?

## Interface Segregation

Are interfaces small, purposeful, and preferably shaped around consumer requirements?

## Dependency Inversion

Are core application behaviors unnecessarily coupled to concrete UI, networking, persistence, platform, inference, or third-party implementations?

Also evaluate:

* composition;
* package cohesion;
* coupling;
* explicit dependencies;
* state ownership;
* separation of pure logic and side effects;
* use of small interfaces;
* dependency injection where it creates useful test seams;
* global/service-locator patterns;
* unnecessary abstractions;
* external-dependency leakage.

**Prefer idiomatic Go and architectural simplicity over mechanical adherence to named patterns.**

---

# 17. Anti-Overengineering Requirement

Do not recommend abstraction merely because abstraction is possible.

Do not recommend:

* an interface for every struct;
* generic service layers;
* factories;
* dependency-injection frameworks;
* plugin systems;
* repositories;
* adapters;
* event buses;
* generalized registries;
* additional packages;
* generalized indirection;

unless evidence shows that they solve an actual problem.

An abstraction should generally justify itself by one or more of:

* eliminating demonstrated coupling;
* isolating platform-specific behavior;
* isolating a volatile external dependency;
* supporting a current or clearly imminent alternate implementation;
* creating a valuable deterministic test seam;
* isolating a trust boundary;
* materially improving reliability;
* reducing change amplification;
* centralizing genuinely duplicated responsibility;
* enabling an important remaining feature.

Optimize for:

> **the minimum sufficient architecture capable of supporting the intended mature system.**

Prefer evolutionary refactoring.

Do not recommend a greenfield rewrite without overwhelming evidence that incremental repair is economically unsound.

---

# 18. Change-Amplification Analysis

Use future change cost as a primary measure of architectural health.

Based on the project's actual remaining features and future plans discovered in the governance documents, determine what currently must change when extending important capabilities.

At minimum, evaluate relevant versions of:

* adding an inference backend;
* adding an inference device type;
* adding a peer transport;
* replacing or upgrading an important submodule;
* changing an important external dependency;
* adding another supported platform;
* adding a project capability;
* adding a scheduling strategy;
* adding a shared-resource capability;
* changing peer identity;
* adding or changing authentication;
* changing persistence;
* adding an alternate UI or headless interface.

Create a matrix:

| Future change | Current affected areas | Necessary propagation | Unnecessary propagation | Recommended response |
| ------------- | ---------------------- | --------------------- | ----------------------- | -------------------- |

Identify shotgun surgery and dependency lock-in.

---

# 19. Domain/UI Separation

Evaluate whether core application behavior exists independently of Ebitengine presentation code.

Determine whether:

* UI components own domain logic;
* networking directly mutates UI objects;
* persistence is directly tied to widgets;
* resource scheduling occurs inside presentation code;
* peer/device/project state exists independently of its visual representation;
* external dependency types leak directly into UI state;
* the Ebitengine loop owns unrelated application responsibilities.

Use this question as an architectural stress test:

> **Could the core application eventually be controlled through a CLI, headless process, test harness, or alternate UI without reimplementing substantial business logic?**

The application does not necessarily need those interfaces.

Use the question to assess coupling.

---

# 20. Ebitengine-Specific Review

Evaluate the architecture through Ebitengine's runtime constraints.

Pay particular attention to:

* `Update`;
* `Draw`;
* blocking operations;
* networking;
* inference requests;
* filesystem activity;
* long-held locks;
* frame-time-sensitive work;
* communication with goroutines;
* asynchronous UI state changes;
* thread-affinity assumptions;
* graphics/resource lifecycle;
* submodule calls invoked from the frame/update path.

Identify anything that could block or destabilize the update/draw lifecycle.

---

# 21. Go Concurrency and Lifecycle Audit

Map and analyze:

* goroutines;
* channels;
* buffered/unbuffered communication;
* mutexes;
* RWMutexes;
* atomics;
* shared maps;
* shared structs;
* worker pools;
* background loops;
* context propagation;
* cancellation;
* shutdown;
* timeouts;
* retries;
* reconnect loops;
* event delivery;
* possible goroutine leaks;
* deadlock risks;
* race risks;
* stale-state risks.

Include concurrency behavior introduced or controlled by materially relevant submodules where it affects the primary application's lifecycle.

For important long-lived workers determine:

* who creates them;
* who owns them;
* how long they live;
* how they stop;
* how cancellation propagates;
* what state they own;
* what shared state they touch;
* how failure is surfaced;
* what happens when dependencies disappear.

Identify places where an explicit state machine would be safer than loosely coordinated flags, callbacks, or mutable state.

---

# 22. Distributed-Systems Stress Test

Treat the application as an emerging distributed system.

Based on its actual intended future behavior, analyze relevant scenarios such as:

* local device disappearance;
* remote peer disconnection;
* reconnection;
* stale peer state;
* conflicting state;
* disappearing inference workers;
* duplicate messages;
* stale discovery records;
* hung connections;
* canceled jobs;
* network partitions;
* malformed peer data;
* inaccurate capability advertisements;
* references to devices that no longer exist;
* application suspension;
* application resumption with stale state.

Determine:

* authoritative sources of truth;
* state-transition ownership;
* idempotency requirements;
* stale-state handling;
* reconciliation;
* retry policy;
* timeout policy;
* cancellation semantics;
* partial-failure behavior.

If an external submodule implements part of these semantics, inspect that implementation sufficiently to understand what guarantees the primary application can and cannot rely upon.

---

# 23. State-Ownership Analysis

For each important category of mutable state, identify:

* what it represents;
* who owns it;
* who can mutate it;
* who receives views/copies;
* how mutations are synchronized;
* how UI state relates to authoritative state;
* how network state relates to authoritative state;
* how submodule state relates to authoritative application state;
* how persistence relates to runtime state;
* how stale state is detected;
* how conflicts are resolved.

Pay particular attention to state representing:

* devices;
* peers;
* projects;
* capabilities;
* inference jobs;
* network connections;
* resource availability;
* configuration;
* user settings.

Treat unclear ownership as a significant structural risk where justified.

---

# 24. Determinism and Testability

Identify behavior that should be deterministic and easy to test.

Possible examples include:

* state transitions;
* serialization;
* protocol parsing;
* project configuration;
* capability matching;
* resource accounting;
* scheduling decisions;
* registry operations;
* routing;
* persistence transforms.

Contrast this with inherently nondeterministic boundaries such as:

* networking;
* peer discovery;
* hardware availability;
* OS behavior;
* timing;
* inference-provider behavior;
* external submodule behavior.

Assess whether nondeterminism is pushed toward system boundaries while core behavior stays deterministic.

Evaluate:

* unit tests;
* integration tests;
* race testing;
* failure-mode tests;
* platform tests;
* whether tests verify behavior or implementation details;
* missing high-value test seams;
* whether external dependencies can be reliably substituted or simulated during testing.

---

# 25. Cross-Platform Architecture

Perform a dedicated review of Windows, Linux, and Android support.

Identify:

* build tags;
* platform-specific files;
* filesystem assumptions;
* path handling;
* process-spawning assumptions;
* socket/network behavior;
* discovery behavior;
* permissions;
* persistence locations;
* background execution assumptions;
* GPU/inference assumptions;
* Ebitengine platform assumptions;
* OS-specific APIs;
* platform constraints introduced by submodules.

Determine whether platform-specific behavior is well isolated or scattered through otherwise platform-independent logic.

---

# 26. Android as an Architectural Stress Test

Do not treat successful Android compilation as proof of Android readiness.

Evaluate:

* lifecycle;
* suspend/resume;
* background execution;
* persistent connection assumptions;
* process-spawning assumptions;
* storage;
* permissions;
* local discovery;
* reconnection;
* resource constraints;
* shutdown;
* recovery;
* asynchronous inference work;
* UI lifecycle;
* Android limitations of materially relevant submodules.

Identify desktop assumptions that become fragile on Android.

---

# 27. Reliability and Structural Integrity

Assess long-term reliability.

Review:

* error propagation;
* swallowed errors;
* error context;
* panic behavior;
* retry behavior;
* retry limits;
* timeout behavior;
* fallback behavior;
* partial initialization;
* recovery;
* state corruption;
* persistence integrity;
* startup failure;
* shutdown integrity;
* malformed configuration;
* malformed external data;
* provider failure;
* dependency/submodule failure;
* network failure;
* concurrency failure.

Explicitly identify:

* silent fallback;
* endless retry;
* duplicated recovery logic;
* logging without handling;
* hidden partial success;
* invalid state following failure;
* retries that obscure root causes.

---

# 28. Observability and Human Debuggability

Evaluate whether failures can be understood by developers and users.

Assess:

* structured logging;
* error wrapping;
* subsystem context;
* device identifiers;
* peer identifiers;
* job/request identifiers;
* connection-state visibility;
* inference lifecycle visibility;
* state-transition visibility;
* dependency/submodule error translation;
* user-facing diagnostics;
* recoverable versus fatal errors.

A reliable system should make failure understandable.

---

# 29. Trust and Security Boundaries

Based on the peer and resource-sharing features described by the repository, evaluate structural readiness for trust and security.

Determine whether the architecture can cleanly distinguish among relevant categories such as:

* local application state;
* trusted local devices;
* trusted personal remote devices;
* authenticated peers;
* potentially untrusted peers;
* untrusted network data;
* external dependency/submodule trust boundaries.

Evaluate readiness for:

* identity;
* authentication;
* authorization;
* capability advertisement;
* quotas;
* remote resource consumption;
* project privacy;
* inference privacy;
* malformed messages;
* malicious claims;
* denial-of-service conditions;
* unintended disclosure.

This is not primarily a penetration test.

The question is whether these concerns can be implemented cleanly without later architectural upheaval.

---

# 30. Remaining-Feature Readiness

Discover the remaining feature list from the governance and planning documents.

Then analyze **every remaining feature individually**.

For each feature identify:

### Existing support

What already exists?

### Missing prerequisites

What must exist first?

### Architectural blockers

What current structures make implementation unnecessarily fragile or expensive?

### Dependency/submodule prerequisites

Does the feature rely on behavior from an omitted submodule? If so, verify that behavior online rather than assuming it.

### Change amplification

Which existing areas must change?

### Recommended approach

What is the simplest robust implementation strategy?

### State model

What authoritative state and transitions are required?

### Reliability/failure behavior

What failures should be expected?

### Tests

What deterministic, integration, concurrency, dependency-boundary, or platform tests are needed?

### Cross-platform implications

Especially Android where relevant.

### Trust/security implications

Where relevant.

### Governance implications

What development rules or review criteria should govern implementation?

### Roadmap placement

When should it be implemented and why?

---

# 31. Reconstruct the Development-Governance System

Independently reconstruct how the agentic development framework currently governs work.

Identify documents and mechanisms involving:

* planning;
* task decomposition;
* implementation;
* coding standards;
* architecture;
* dependency/submodule changes;
* testing;
* review;
* validation;
* completion;
* handoff;
* retries;
* failure handling;
* documentation;
* retrospectives;
* agent roles;
* project state;
* research methodology.

Explain the actual intended workflow.

Do not assume a particular pipeline in advance.

---

# 32. Governance as an Engineering System

Treat governance documents as a system that generates engineering behavior.

For each important rule ask:

> **What behavior does this instruction tend to produce when repeatedly followed by coding agents?**

A rule should not be considered successful merely because it sounds reasonable.

Evaluate its outcomes.

---

# 33. Governance Quality Criteria

Evaluate major governance rules according to:

### Correctness

Would following the rule generally produce good engineering behavior?

### Clarity

Could competent agents reasonably interpret it differently?

### Observability

Can compliance actually be determined?

### Enforceability

Does some workflow stage catch violations?

### Consistency

Does it conflict with other guidance?

### Discoverability

Does the relevant agent see it when it matters?

### Cost effectiveness

Does its benefit justify its process cost?

### Durability

Will the rule remain useful as the project evolves?

---

# 34. Classify Governance by Function

Where useful, separate governance into:

## Principles

Values guiding judgment.

## Architectural invariants

Properties that should remain true.

## Procedures

How work should be performed.

## Completion gates

Conditions that must hold before work is declared complete.

Identify where current governance mixes these categories in ways that reduce clarity.

---

# 35. Expectation-vs-Reality Audit

Compare significant governance expectations with actual implementation.

Create a traceability matrix:

| Governance expectation | Defined where | Observed implementation evidence | Compliance | Likely reason for deviation | Recommended response |
| ---------------------- | ------------- | -------------------------------- | ---------- | --------------------------- | -------------------- |

Include important subjects such as:

* function comments;
* readability;
* tests;
* architectural boundaries;
* dependency boundaries;
* concurrency;
* state ownership;
* error handling;
* documentation;
* planning;
* completion;
* review quality.

---

# 36. Planning-vs-Reality Audit

Compare the current project plan against the implementation.

For significant planned items determine:

* whether prerequisites actually exist;
* whether the plan assumes architecture that has not been built;
* whether the plan assumes submodule capabilities that actually exist;
* whether tasks are ordered correctly;
* whether planned work duplicates existing functionality;
* whether supposedly completed work is incomplete;
* whether future work depends on unstable foundations;
* whether omitted architectural work is now necessary;
* whether a planned submodule upgrade/replacement changes the architectural sequence.

Create a planning gap analysis where useful.

---

# 37. Root-Cause Analysis of Repeated Defects

Where similar problems recur, do not simply enumerate local fixes.

Ask:

> **Why did the development system repeatedly permit this outcome?**

Potential causes include:

* missing governance;
* vague governance;
* buried requirements;
* contradictory requirements;
* poor task decomposition;
* planning that omits architecture;
* implementation agents lacking context;
* reviewers lacking context;
* completion gates failing to check important conditions;
* deterministic checks being used for semantic judgments;
* excessive requirements producing workarounds;
* failed runs not producing persistent learning;
* external dependency behavior not being investigated before implementation assumptions were made.

Distinguish:

* local coding mistake;
* repeated implementation-pattern failure;
* architectural failure;
* dependency-integration failure;
* planning failure;
* governance failure.

---

# 38. Governance-Induced Technical Debt

Explicitly search for cases where the governance framework itself may have caused poor outcomes.

Examples may include:

* excessive abstraction;
* premature generalization;
* unnecessary design patterns;
* duplicated validation systems;
* deterministic validation applied to fuzzy semantic problems;
* tests tightly coupled to implementation;
* meaningless compliance comments;
* excessive defensive fallback;
* retry loops hiding root causes;
* overfragmented tasks;
* agents repeatedly correcting symptoms instead of causes;
* agents making unsupported assumptions about external libraries/submodules instead of inspecting them.

For problematic governance recommend whether it should be:

* retained;
* clarified;
* narrowed;
* strengthened;
* relocated;
* replaced;
* removed.

---

# 39. Anti-Bureaucracy Requirement

Do not create a project-wide rule for every bug.

A governance rule is most justified when a problem is:

* recurring;
* architecturally consequential;
* difficult to catch locally;
* likely to recur across tasks;
* preventable with a clear invariant;
* severe enough that prevention is better than ordinary review.

Keep local mistakes local.

The objective is **better governance, not more governance**.

---

# 40. Agent Information Flow

Analyze how important knowledge moves through the agentic workflow.

Determine:

* whether planning agents receive architectural context;
* whether implementation agents must rediscover project expectations;
* whether agents are instructed to inspect relevant external dependencies before making architectural assumptions;
* whether task specifications contain relevant invariants;
* whether reviewers receive original intent;
* whether reviewers understand architectural rationale;
* whether rationale survives handoffs;
* whether implementation discoveries update future planning;
* whether failed runs are analyzed;
* whether repeated failures produce persistent lessons;
* whether completion agents can declare success without revisiting requirements.

Identify information that is lost between stages.

---

# 41. Completion-Gate Audit

Reconstruct what "done" currently means according to the governance documents and actual workflow.

Determine whether tasks can be declared complete while violating:

* functional requirements;
* comment requirements;
* tests;
* architecture;
* dependency assumptions;
* reliability;
* concurrency lifecycle;
* cross-platform requirements;
* documentation;
* cleanup obligations.

Recommend a completion model that is strict without becoming an endless validation bureaucracy.

Explicitly distinguish:

## Deterministic checks

Examples:

* builds;
* tests;
* race tests;
* formatting;
* static analysis;
* missing required structural documentation;
* forbidden dependency direction;
* pinned dependency/version checks.

## Semantic checks

Examples:

* whether an abstraction is warranted;
* whether a comment meaningfully explains intent;
* whether architecture remains coherent;
* whether user intent was satisfied;
* whether a workaround creates unacceptable debt;
* whether an external dependency was understood deeply enough for the design decision being made.

**Do not recommend elaborate deterministic heuristics for inherently semantic judgments.**

---

# 42. End-of-Run Learning

Review whether the governance framework learns effectively from completed or failed development runs.

Determine whether the process asks questions such as:

* What caused failed rounds?
* What information was missing initially?
* What architecture was misunderstood?
* Was an external dependency misunderstood?
* What issue could have been caught earlier?
* Could the first task specification have prevented rework?
* Did the agent solve symptoms instead of causes?
* Should planning templates change?
* Should governance change?
* Was the failure merely local and not worth systematizing?

Recommend a lightweight mechanism for converting meaningful run-level lessons into better future work without excessive retrospective bureaucracy.

---

# 43. Development-Process Retrospective

Produce a retrospective covering:

## What worked well

Practices associated with good outcomes.

## What failed

Rules or workflows that proved ineffective.

## What created friction

Processes causing reruns, confusion, wasted work, or unnecessary complexity.

## What is missing

Important expectations not currently encoded.

## What is obsolete

Rules or plans appropriate to an earlier phase but no longer useful.

## What should be strengthened

Useful standards weakly enforced.

## What should be simplified or removed

Governance whose cost exceeds its value.

## What the framework should learn from this checkpoint

General lessons that should shape future runs.

---

# 44. Architectural Invariants for Future Development

Recommend a **small, high-value set** of architectural invariants for the remainder of the project.

Potential examples, only if supported by evidence, might include:

* UI code does not perform blocking network/inference work.
* Network transports do not directly mutate UI components.
* Domain/application logic does not depend directly on Ebitengine.
* Platform-specific behavior stays behind explicit boundaries.
* Every long-lived goroutine has an owner and shutdown path.
* Mutable state has an authoritative owner.
* External peer input is treated as untrusted.
* Provider differences are represented by capabilities rather than scattered provider-name conditionals.
* Background failures are surfaced rather than silently ignored.
* Volatile third-party/submodule implementation details do not leak unnecessarily through the application.

Do not produce a large ruleset for its own sake.

---

# 45. Decisions Whose Cost of Reversal Is Increasing

This is one of the most important sections.

Identify decisions that are inexpensive to change now but likely to become substantially more expensive after upcoming feature work.

Use the repository's actual future plans to identify these.

Potential examples may involve:

* device identity;
* peer identity;
* capability representation;
* project state;
* persistence;
* inference-provider boundaries;
* scheduler boundaries;
* resource-sharing semantics;
* trust model;
* state ownership;
* concurrency lifecycle;
* UI/domain coupling;
* transport abstraction;
* cancellation semantics;
* dependency/submodule selection;
* coupling to specific submodule APIs;
* version/protocol compatibility assumptions.

For each provide:

### Current state

### Future features depending on it

### Why it matters

### Consequence of leaving it unchanged

### Point where reversal becomes more expensive

### Recommendation

---

# 46. Recommended Target Architecture

Based on the evidence, describe the recommended **evolutionary target architecture**.

Do not produce an idealized greenfield design.

Explain how the current system should evolve.

For each significant proposed boundary describe:

* responsibility;
* state ownership;
* dependency direction;
* interface surface;
* why it exists;
* what current issue it resolves;
* what future feature it supports;
* what dependency/submodule concern it isolates, if any;
* why its added complexity is justified.

Only introduce abstractions that earn their cost.

---

# 47. Migration Strategy

For each substantial architecture change explain:

* current state;
* target state;
* incremental migration path;
* prerequisite tests;
* compatibility considerations;
* dependency/submodule version implications;
* temporary adapters or shims if useful;
* feature work that should wait;
* feature work that can proceed concurrently;
* regression risk;
* point where old architecture can be removed.

Prefer evolutionary change over large rewrites.

---

# 48. Checkpoint-Hardening Phase

Determine whether a dedicated checkpoint-hardening phase should happen before major feature work continues.

This phase should contain only changes whose delay would:

* compound technical debt;
* lock in poor architecture;
* deepen dependency/submodule coupling;
* reduce reliability;
* cause future features to depend on unstable assumptions;
* substantially increase reversal cost.

Separate recommendations into:

## Must fix before substantial feature development resumes

## Fix while implementing the next related feature

## Can safely defer

## Intentionally leave unchanged

Do not turn checkpoint hardening into unlimited cleanup.

---

# 49. Governance Remediation

For significant engineering findings, determine whether governance should change to prevent recurrence.

Where justified specify:

### Current governance gap

### Proposed governance change

### Where it should live

### Which agent/stage should enforce it

### Verification type

* deterministic;
* semantic;
* mixed.

Where appropriate, consider whether governance should explicitly require agents to inspect the source/documentation of important external dependencies rather than infer their behavior from names or APIs.

Do not add governance when ordinary code review is sufficient.

---

# 50. Synchronized Engineering and Governance Roadmap

Create a roadmap showing how technical improvements and process improvements reinforce each other.

Use a structure like:

| Engineering change | Governance/process change | Why paired | Prerequisites | Enables |
| ------------------ | ------------------------- | ---------- | ------------- | ------- |

This roadmap should explain not only how to fix current weaknesses, but how to prevent them from recurring during the remainder of development.

---

# 51. Second-Draft Project Plan

Using the current project plan discovered in the repository as Draft 1, produce a **complete Draft 2** for the remainder of the project.

Do not merely polish or reorganize the existing plan.

The new plan must be **derived from the architecture, dependency, and governance review**.

It should incorporate:

* checkpoint-hardening work;
* architectural prerequisites;
* dependency/submodule changes where required;
* governance remediation;
* remaining features;
* feature dependencies;
* concurrency/reliability work;
* security/trust groundwork where necessary;
* testing;
* documentation;
* observability;
* Windows/Linux/Android validation;
* integration milestones;
* completion criteria.

For each major phase include:

### Goal

### Why this phase occurs now

### Prerequisites

### Architectural work

### Dependency/submodule work

Where applicable.

### User-facing feature work

### Governance/process changes

### Tests and validation

### Cross-platform validation

### Reliability/failure validation

### Exit criteria

### What becomes possible afterward

The ordering should be based on dependencies discovered through the research.

---

# 52. Prioritization Framework

Prioritize recommendations according to:

* architectural leverage;
* reliability impact;
* future reversal cost;
* reduction in change amplification;
* dependency/submodule lock-in reduction;
* number of remaining features affected;
* implementation cost;
* regression risk;
* recurrence;
* governance preventability.

Do not elevate cosmetic cleanup to the same level as foundational architectural changes.

---

# 53. Suggested Health Assessment

Where useful, summarize major dimensions using qualitative ratings such as:

* Strong
* Acceptable
* Needs improvement
* High risk
* Critical

Potential dimensions include:

* readability;
* human editability;
* documentation;
* package cohesion;
* coupling;
* dependency isolation;
* submodule integration quality;
* state ownership;
* testability;
* reliability;
* concurrency safety;
* cross-platform integrity;
* Android readiness;
* extensibility;
* observability;
* trust/security readiness;
* governance clarity;
* governance enforceability;
* planning quality;
* feature readiness.

Every rating must be supported by explanation.

Do not imply false numerical precision.

---

# 54. Final Report Structure

Produce the final report in this order.

## Part I — Executive Assessment

Overall project health, strongest aspects, highest risks, readiness for continued development, whether checkpoint hardening is required, and the 5–10 highest-leverage findings.

## Part II — Repository Governance, Planning, Dependency, and Submodule Inventory

Identify the important governance, methodology, roadmap, feature, architectural, submodule, and dependency sources and explain their roles.

## Part III — Current Architecture Reconstruction

Describe what the software actually is today, including materially important submodule boundaries.

## Part IV — Intended Future State and Current Plan Reconstruction

Explain what the project is trying to become, what remains to be built, and how the current plan expects development to proceed.

## Part V — Code Quality, Readability, and Human Editability

Detailed maintainability assessment.

## Part VI — Comment and Documentation Compliance

Evaluate the explicit function-comment requirement in project-owned code.

## Part VII — Architecture and Idiomatic Go/SOLID Review

Evaluate cohesion, coupling, responsibility, interfaces, dependency direction, abstraction quality, and external dependency isolation.

## Part VIII — Ebitengine, Concurrency, and Lifecycle Review

Analyze runtime-loop constraints, goroutines, cancellation, worker ownership, and state coordination.

## Part IX — Distributed State and Reliability Review

Analyze peer/device failure, stale state, partitions, retries, reconciliation, and structural integrity.

## Part X — Submodule and Dependency Architecture Review

For materially important submodules, summarize verified upstream behavior, how the project uses them, risks, coupling, compatibility assumptions, and whether current boundaries are appropriate.

## Part XI — Cross-Platform and Android Readiness

Evaluate Windows/Linux/Android assumptions, including dependency-specific constraints.

## Part XII — Security and Trust-Boundary Readiness

Evaluate readiness for peer interaction and shared inference.

## Part XIII — Testing, Determinism, Observability, and Debuggability

Assess test design, deterministic core logic, dependency seams, logging, diagnostics, and failure visibility.

## Part XIV — Change-Amplification Analysis

Include the future-change impact matrix.

## Part XV — Remaining-Feature Readiness

Analyze every remaining feature.

## Part XVI — Governance-System Reconstruction

Explain how the agentic development process currently operates.

## Part XVII — Governance Expectation-vs-Reality Audit

Include the governance traceability matrix.

## Part XVIII — Planning-vs-Reality Audit

Compare the current roadmap with actual architecture, implementation, and dependency capabilities.

## Part XIX — Governance-Induced Technical Debt

Identify where governance itself may have caused poor engineering outcomes.

## Part XX — Development-Process Retrospective

Explain what worked, failed, caused friction, became obsolete, or should change.

## Part XXI — Recommended Governance Model

Recommend revised principles, architectural invariants, procedures, dependency-research expectations, information flow, review standards, completion gates, and end-of-run learning.

## Part XXII — Decisions Whose Reversal Cost Is Increasing

Identify decisions that should be settled before upcoming features lock them in.

## Part XXIII — Recommended Target Architecture

Describe the minimum sufficient evolutionary architecture.

## Part XXIV — Checkpoint-Hardening Plan

Identify what should happen now.

## Part XXV — Migration and Implementation Strategy

Explain how to reach the target architecture incrementally.

## Part XXVI — Synchronized Engineering/Governance Roadmap

Show how code and process changes reinforce each other.

## Part XXVII — Second-Draft Project Plan

Provide the complete revised project plan for finishing the project.

---

# 55. Failure Modes to Avoid

Do not:

* return generic clean-code advice;
* mechanically enforce SOLID;
* recommend interfaces everywhere;
* equate abstraction with good architecture;
* recommend a rewrite without strong evidence;
* equate comment quantity with comment quality;
* merely paraphrase syntax in comments;
* assume compilation equals cross-platform readiness;
* assume desktop behavior transfers directly to Android;
* treat distributed failures as unusual edge cases;
* assume local and remote resources share the same trust model;
* create a governance rule for every bug;
* treat every bug as a governance failure;
* use deterministic heuristics for inherently semantic judgments;
* recommend broad cleanup without explaining what it enables;
* assume the current roadmap is correct;
* simply rewrite the current roadmap more elegantly;
* assume a feature exists because a planning document lists it;
* assume a governance rule was followed because it exists;
* assume a submodule's behavior from its name or API surface;
* analyze the wrong/latest version of a submodule when the project pins another revision;
* perform a full audit of irrelevant submodules;
* attribute third-party coding-style deficiencies to project agents;
* preserve an implementation merely because it already exists;
* redesign an implementation merely because another design looks cleaner;
* make major claims without repository or authoritative upstream evidence.

---

# 56. Desired Engineering Philosophy

Optimize for a future codebase that is:

* maximally understandable by humans;
* straightforward to edit safely;
* explicit rather than clever;
* modular without unnecessary fragmentation;
* idiomatic Go;
* appropriate for Ebitengine;
* resilient under partial failure;
* easy to debug;
* testable;
* observable;
* portable across Windows, Linux, and Android;
* appropriately insulated from volatile external dependencies;
* appropriate for distributed resource coordination;
* capable of safely evolving toward remote peer interaction;
* governed by clear, useful, enforceable expectations.

The goal is not theoretical architectural perfection.

The goal is:

> **the simplest architecture and development process capable of supporting the intended mature application reliably over the long term.**

---

# 57. Final Synthesis Question

Conclude the report by directly answering:

> **If development continued immediately from the current checkpoint using the current architecture, current dependency/submodule boundaries, current project plan, and current governance rules, what problems are most likely to become substantially harder or more expensive to fix later, why, and what exact sequence of technical and governance changes should be made now to avoid that outcome?**

The answer should integrate:

* repository reality;
* verified behavior of materially relevant submodules;
* intended future state;
* current plan;
* code quality;
* architecture;
* dependency boundaries;
* reliability;
* concurrency;
* distributed-state concerns;
* cross-platform concerns;
* remaining features;
* governance effectiveness;
* process failures;
* target architecture;
* checkpoint hardening;
* the revised project plan.

---

# 58. Repository Input

The attached GitIngest contains the complete **primary repository** and all available governance/planning materials.

Git submodules may be represented only by their declarations, references, integration code, or identifying metadata rather than by their full source contents.

**When information about an omitted submodule is necessary to answer the research questions accurately, locate and inspect the appropriate upstream repository, documentation, and—where possible—the exact revision used by this project.**

Treat the primary repository as authoritative for:

* project intent;
* governance;
* planning;
* project-owned implementation;
* how submodules are integrated.

Treat authoritative upstream sources as evidence for:

* omitted submodule implementation;
* submodule API contracts;
* version-specific behavior;
* upstream platform/lifecycle/reliability characteristics.

Clearly distinguish verified upstream facts, primary-repository facts, and your own inferences.
