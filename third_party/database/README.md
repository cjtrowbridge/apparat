# Database Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `modernc-sqlite` | `https://gitlab.com/cznic/sqlite` | `5d243466fa05e8c49a870ead00b79bd4a423f712` (`v1.53.0-3-g5d24346`) | BSD-3-Clause | Source reference for the cgo-free `modernc.org/sqlite` Go driver | Reference only until the root Go module pins the package through `go.mod` |

## Selection Notes

- The cgo-free driver is preferred initially to reduce cross-compilation and packaging friction on Steam Deck/Linux, Windows, macOS, and Android.
- The source checkout exists for auditing, platform investigation, and reproducible upstream study.
- Apparat should consume the driver through a normal `go.mod` and `go.sum` pin rather than a local `replace` unless a focused debugging plan requires one.
- Source presence does not validate WAL, backup, suspend/resume, filesystem, or performance behavior on any target.
- Driver replacement requires measured compatibility or performance evidence and a migration-aware execution plan.
