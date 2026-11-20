# Linux self-update adapter

- `client.go`: Reads the public Apparat development channel, pins main to one
  commit, validates accepted ordinary target metadata and hashes files.
- `install.go`: Downloads to the executable directory, bounds bytes, verifies
  SHA-256 and native ELF, keeps `.previous` and replaces the binary atomically.

Only the fixed GitHub HTTPS channel is used in production. Tests inject HTTP
clients/endpoints locally. Metadata is bounded to 2 MiB, binaries to 256 MiB;
requests cancel with the GUI and have a 15-minute overall limit and 30-second
response-header deadline. There is one UI-owned operation at a time. No credentials
or raw server errors enter UI text or logs. A non-writable directory, changed
installed binary, corrupt download or unsupported architecture fails before
promotion. Candidate files are removed on cancellation/close.

The published digest establishes consistency with the HTTPS development channel,
not an independent release signature or monotonic version policy. GitHub account
compromise and schema downgrade prevention require Phase 17 release hardening.
Binary rollback uses `.previous` only with compatible node state or a verified
offline backup. Steam continues to use the original executable path.

Verification: governed Go tests exercise metadata pinning, corruption, target
rejection, cancellation, atomic promotion and preservation of installed bytes.

- `client_test.go`: Tests pinned metadata, native bytes, corrupt/cancelled downloads, changed files, backups and atomic promotion.
- `limits_test.go`: Rejects oversized transfers without payload allocation and verifies installed-byte preservation and staging cleanup.
