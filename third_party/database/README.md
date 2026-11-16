# Database Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `modernc-sqlite` | `https://gitlab.com/cznic/sqlite` | `92d72a45bcb65ab33bbe79c8667f2b8b529ae902` (`v1.62.0-1-g92d72a4`) | BSD-3-Clause | Source reference for the cgo-free `modernc.org/sqlite` Go driver | Active root Go dependency at `v1.53.1-0.20260625155647-5d243466fa05` |

## Selection Notes

- The cgo-free driver is preferred initially to reduce cross-compilation and packaging friction on Steam Deck/Linux, Windows, macOS, and Android.
- The source checkout exists for auditing, platform investigation, and reproducible upstream study.
- Apparat should consume the driver through a normal `go.mod` and `go.sum` pin rather than a local `replace` unless a focused debugging plan requires one.
- Source presence does not validate WAL, backup, suspend/resume, filesystem, or performance behavior on any target.
- Driver replacement requires measured compatibility or performance evidence and a migration-aware execution plan.

## 2026-11-16 Source Refresh

Updated to configured upstream heads at the user’s request. Exact revisions above remain pinned by gitlinks; branch tracking was not added. See the [update review](../../plans/past/2026-11-16-11-00-00_update-and-review-submodules.md) for deltas, verification limits, and rollback. Source-reference updates do not change root Go module pins or activate new integrations.
