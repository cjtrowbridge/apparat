# Security Contract

## Threat Model

Apparat defends against local LAN observers, unauthorized LAN peers, stolen devices, malicious enrolled peers, replay, duplicate delivery, queue abuse, confused deputies, compromised workload services, leaked logs, and lost enrollment material.

It does not treat WireGuard reachability as authorization. HTTPS authentication remains mandatory on trusted LAN and WireGuard networks.

## Identity Model

- User identity: long-lived owner identity.
- Device identity: Apparat device signing key and device record.
- Cluster identity: signed directory and trust policy.
- WireGuard identity: network reachability key bound to, but separate from, Apparat identity.
- TLS identity: mTLS certificate bound to the Apparat device identity.
- Future transport identity: adapter-specific identity mapped to Apparat authorization.

Phase 3 implements local Ed25519 user/device key generation, signatures, public manifests, startup consistency classification, archived reset, and encrypted private-key files using Argon2id-derived XChaCha20-Poly1305 keys.

## X.509 And TLS

The MVP uses one cluster-local X.509 root CA whose certificate fingerprint is included in the out-of-band verified cluster identity. One currently authorized enrollment authority controls issuance under that root during the MVP. Multiple concurrent issuers, delegated intermediates, or threshold issuance require a later explicit hierarchy and conflict-resolution design.

Each device generates its TLS leaf key separately from its Apparat Ed25519 signing key. The device proves possession of both during enrollment. A signed device record binds the Apparat signing public key, TLS leaf public key, certificate serial and fingerprint, WireGuard public key, cluster identity, roles, scopes, validity interval, and status. Peers require both a valid mTLS chain to the cluster root and a current authorized device-record binding; possession of only one key is insufficient.

Mutating operations disable TLS 0-RTT. Rotation creates a new leaf key/serial and signed binding before the old binding is retired. Revocation or lost-device recovery invalidates authorization for both the device record and its TLS certificate even if a cached certificate still chains to the root.

## Enrollment

Enrollment uses an out-of-band QR code or invite containing cluster fingerprint, one-time token, endpoint hints, expiration, and intended role. The enrolling device confirms both sides before trust is recorded.

## Authorization Scopes

Scopes cover projects, queues, services, tasks, settings, enrollment, transports, research, comrades, and diagnostics. Shared compute grants never imply project, file, secret, tool, shell, or administrative access.

Project discovery and project operation are separate permissions. A device may be allowed to see an authorization-filtered project summary in the cluster-wide catalog without receiving file, Git, Task-execution, mutation, secret, or artifact access. Every remote project operation is authenticated and authorized by the device that owns the repository; cached project metadata grants no authority and never provides direct filesystem access.

A Task is an executable Project entrypoint. Manual invocation, schedule management, webhook binding, event binding, owner-local project effects, routed workload submission, approval, and result visibility may require distinct permissions. Webhook possession alone is not project authorization; webhook requests are authenticated, bounded, replay-protected, and mapped to a specifically allowed Task.

Queue permissions distinguish submission, status/result visibility, cancellation, worker membership, claim, heartbeat, and completion. The queue owner revalidates every REST request. A worker receives only a bounded leased task it is authorized and capable of executing; it does not gain queue administration, unrelated job visibility, project access, provider credentials, or other cluster state.

Worker completions must bind the worker identity, queue, job, attempt, lease, fencing token, result schema, artifact hashes, and idempotency identity. The owner rejects expired, superseded, replayed, unauthorized, or mismatched results before they affect authoritative queue state.

## Local Inference Gateway

Remote peers invoke local inference only through authenticated Apparat resources using logical device, service, capability, and model identities. Gateway authorization, queue admission, limits, audit, and routing are evaluated before a provider request. Provider localhost URLs, credential references, tokens, raw failures, prompts, and results are not advertised or disclosed through general health responses.

SQLite stores opaque provider credential references, never credential values in general provider configuration. A secret adapter resolves values from an OS credential store when available or an Apparat-managed encrypted secret file. Every service instance has independent enablement, advertisement, admission, concurrency, cancellation, and failure boundaries so compromise or failure of one endpoint does not implicitly authorize or disable another.

Expired service advertisements are never eligible for routing. Owners refresh the default 120-second advertisement by 60 seconds; stale metadata may remain visible for diagnosis for 24 hours but grants no execution authority.

## Task And Artifact Safety

Task entrypoints execute only versioned allowlisted application actions, Project-scoped filesystem/Git operations, and explicit typed service calls. Definitions contain secret references instead of values. Consequential actions may require configured human approval. Apparat exposes no unrestricted remote shell or arbitrary process/tool endpoint.

Artifact authorization is distinct from Project discovery, job status, or queue membership. Transfers are authenticated, bounded, resumable, and verified by SHA-256 digest before finalization. Queue owners validate the active lease, artifact provenance, ownership, authorization, size, and digest before accepting a worker result. Retention and deletion remain owner-controlled and auditable.

## Audit And Redaction

Audit events record identity changes, enrollment, revocation, auth failures, project advertisement/access/mutation, Task invocation and trigger cause, queue submission/admission/rejection, lease issue/renewal/expiry, cancellation, worker completion acceptance/rejection, policy denial, service health, and configuration changes.

Logs redact private keys, passphrases, tokens, raw voice, prompts, model outputs, chat bodies, file contents, project secrets, and transport credentials by default.
