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

The MVP uses a cluster-local certificate authority signed or authorized by the enrollment authority. TLS leaf keys are separate from Apparat device signing keys and are cryptographically bound in signed device records.

Mutating operations disable TLS 0-RTT. Certificates have explicit issuance, expiration, rotation, revocation, and lost-device recovery records.

## Enrollment

Enrollment uses an out-of-band QR code or invite containing cluster fingerprint, one-time token, endpoint hints, expiration, and intended role. The enrolling device confirms both sides before trust is recorded.

## Authorization Scopes

Scopes cover projects, queues, services, tasks, settings, enrollment, transports, research, comrades, and diagnostics. Shared compute grants never imply project, file, secret, tool, shell, or administrative access.

Project discovery and project operation are separate permissions. A device may be allowed to see an authorization-filtered project summary in the cluster-wide catalog without receiving file, Git, Task-execution, mutation, secret, or artifact access. Every remote project operation is authenticated and authorized by the device that owns the repository; cached project metadata grants no authority and never provides direct filesystem access.

A Task is an executable Project entrypoint. Manual invocation, schedule management, webhook binding, event binding, owner-local project effects, routed workload submission, approval, and result visibility may require distinct permissions. Webhook possession alone is not project authorization; webhook requests are authenticated, bounded, replay-protected, and mapped to a specifically allowed Task.

Queue permissions distinguish submission, status/result visibility, cancellation, worker membership, claim, heartbeat, and completion. The queue owner revalidates every REST request. A worker receives only a bounded leased task it is authorized and capable of executing; it does not gain queue administration, unrelated job visibility, project access, provider credentials, or other cluster state.

Worker completions must bind the worker identity, queue, job, attempt, lease, fencing token, result schema, artifact hashes, and idempotency identity. The owner rejects expired, superseded, replayed, unauthorized, or mismatched results before they affect authoritative queue state.

## Audit And Redaction

Audit events record identity changes, enrollment, revocation, auth failures, project advertisement/access/mutation, Task invocation and trigger cause, queue submission/admission/rejection, lease issue/renewal/expiry, cancellation, worker completion acceptance/rejection, policy denial, service health, and configuration changes.

Logs redact private keys, passphrases, tokens, raw voice, prompts, model outputs, chat bodies, file contents, project secrets, and transport credentials by default.
