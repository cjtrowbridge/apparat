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

## X.509 And TLS

The MVP uses a cluster-local certificate authority signed or authorized by the enrollment authority. TLS leaf keys are separate from Apparat device signing keys and are cryptographically bound in signed device records.

Mutating operations disable TLS 0-RTT. Certificates have explicit issuance, expiration, rotation, revocation, and lost-device recovery records.

## Enrollment

Enrollment uses an out-of-band QR code or invite containing cluster fingerprint, one-time token, endpoint hints, expiration, and intended role. The enrolling device confirms both sides before trust is recorded.

## Authorization Scopes

Scopes cover projects, queues, services, tasks, settings, enrollment, transports, research, comrades, and diagnostics. Shared compute grants never imply project, file, secret, tool, shell, or administrative access.

## Audit And Redaction

Audit events record identity changes, enrollment, revocation, auth failures, queue admission, cancellation, policy denial, task execution, service health, and configuration changes.

Logs redact private keys, passphrases, tokens, raw voice, prompts, model outputs, chat bodies, file contents, project secrets, and transport credentials by default.
