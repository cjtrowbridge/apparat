# Third-Party Sources

This directory contains pinned upstream source trees used for architecture work, auditing, local replacement, integration, or reproducible study.

A tracked source submodule is not automatically an Apparat build dependency. Go dependencies are activated and pinned separately through `go.mod` and `go.sum`; native services require their own integration and packaging plans.

## Tracked Groups

| Group | Inventory | Role |
| --- | --- | --- |
| Game and HUD | [`game/README.md`](./game/README.md) | Cross-platform runtime, retained HUD controls, and developer diagnostics |
| Database | [`database/README.md`](./database/README.md) | Cgo-free SQLite source and integration reference |
| Networking | [`networking/README.md`](./networking/README.md) | WireGuard implementation, inspection, and configuration references |
| Inference | [`inference/README.md`](./inference/README.md) | Local model-service references |
| Speech | [`speech/README.md`](./speech/README.md) | Local speech-service references |

Each group inventory records the upstream URL, exact gitlink revision, license, purpose, and current relationship to the Apparat build.

`third_party/go.mod` intentionally isolates reference source trees from the root application module. Root commands such as `go test ./...` must not traverse reference checkouts, external upstream tests, GPL reference code, or ignored predecessor material unless a later approved plan explicitly activates an adapter.

## Checkout

Initialize every tracked source and any nested submodules:

```bash
git submodule update --init --recursive
```

The ignored `third_party/salvagecore/` checkout is not a tracked source module. It is an optional temporary predecessor reference documented in the root README and ROADMAP.

## Admission Rules

Add a source tree only when:

- README and ROADMAP identify an approved purpose.
- A focused execution plan names the path, upstream, role, and verification.
- The upstream license has been reviewed for the intended use.
- A specific revision is selected intentionally.
- The relevant group inventory is updated in the same checkpoint.
- The source cannot be represented adequately by a normal package-manager pin or a concise external reference.

Deferred and excluded candidates remain absent until their documented admission gate is satisfied.

## Deferred Admission Gates

These candidates are intentionally absent until a focused architecture and implementation plan proves the correct boundary:

- **Qwen3-TTS:** research runtime, model size, hardware expectations, Python/PyTorch packaging, licensing, service boundaries, streaming behavior, and cross-platform feasibility before adding source.
- **Meshtastic:** select the protobuf/client boundary, payload-size strategy, fragmentation, acknowledgements, store-and-forward behavior, and conformance tests before adding source.
- **Signal gateway:** establish feasibility, account operation, identity mapping, maintenance risk, rate limits, and gateway trust boundaries before selecting an implementation.
- **BOINC:** define the Research tab architecture, validation model, isolation boundary, RPC/client/manager role, and packaging approach before adding BOINC sources.
- **Android mobile tooling:** use the admitted Ebitengine `ebitenmobile` source first; admit `golang/mobile` or an Android wrapper source tree only if Phase 5 proves it is required for a durable APK pipeline independent of the ignored salvagecore checkout.
- **Alternative runtimes:** require a specific workload, adapter contract, license review, platform target, and validation plan before adding any model, speech, artifact, or networking runtime.

## MVP Source Exclusions

These sources remain outside the MVP source set:

- qTox, TokTok qTox, and go-toxcore-c.
- Tor.
- WebRTC until HTTPS REST and event cursors cannot meet a concrete requirement.
- curl, because Go's standard HTTP stack covers the first API.
- OpenSSL and libsodium, because Go-native TLS, signatures, and encrypted-key design cover the first security boundary without cgo.
- OpenSSL as a PGP solution, because OpenSSL does not provide PGP semantics.
- Qwen3-ASR while whisper.cpp is the selected local ASR reference.
- `golang/mobile` as a source checkout while Apparat uses pinned Ebitengine mobile tooling.
- termframe until an interactive headless TUI is approved.

## Update Procedure

For one intentional source update:

1. Create or approve a focused execution plan.
2. Review upstream release notes, security notices, dependency changes, and license changes.
3. Fetch the selected submodule and check out the approved tag or commit.
4. Stage the gitlink so the parent repository records the new revision.
5. Update the group inventory with the exact commit, descriptive tag, rationale, and role changes.
6. Run `git submodule update --init --recursive`.
7. Validate the affected Apparat build, adapter, or reference workflow.
8. Review `git submodule status --recursive`, `.gitmodules`, and the staged diff before committing.

Do not update every submodule opportunistically, and do not set moving branch tracking as a substitute for a pinned gitlink.
