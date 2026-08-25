# Deep-Research Provenance and Limitations

## Reviewed source snapshot

| Field | Value |
| --- | --- |
| Advisory report | `review/deep-research.md` |
| Report SHA-256 | `4857d9f9bc395cbea285a972e08c7d960248e602c077150400f84fbb6e2e49a6` |
| GitIngest digest | root `gitingest.txt` |
| Digest SHA-256 | `7166f1bf3bc82173f7360836b643a3be41f56056c89ff7e85ef67356f9316259` |
| Manifest digest SHA-256 | `7166f1bf3bc82173f7360836b643a3be41f56056c89ff7e85ef67356f9316259` |
| Manifest source commit | `8095ee4d2845b31ee2272ac5849a03bf0612f773` |
| Manifest source state | clean (`dirty: false`) |
| GitIngest version | `0.3.1` |
| Manifest generated at | `2026-08-25T00:24:25.392578+00:00` |

The digest and manifest hashes match. The default corpus excludes recursive `third_party/` source bodies, release artifacts, `.tools/`, `.tmp/`, and `.git/`; an inspection found zero `FILE: third_party/` sections. It includes project-owned source, tests, documentation, plans, and governance artifacts.

## Snapshot applicability

The manifest records commit `8095ee4`. Commit `a00b745` that followed changes only the journal, plan indexes, and the archived manifest-cleanliness plan; it contains no product-source or dependency change. The claims in this review therefore use `8095ee4` as their authoritative source snapshot. The review plan's later activation is intentionally not included in that source snapshot.

## Report limitations

1. The report states that it did not receive the provenance manifest when it was produced. Its third-party and submodule assertions must therefore not be treated as independently verified at the exact pinned gitlink revision.
2. The report has only a small set of upstream citations and most repository facts lack path-and-symbol locators. The claim ledger supplies those locators only for the Phase-7-material findings reviewed here.
3. The report is a point-in-time analysis. Each current-state claim below is classified against the frozen snapshot; future source changes require a new provenance record and revalidation.
4. The report's function-comment measurement was not accompanied by its original script. `review/README.md` records the replacement measurement and its deliberately narrow semantics.
5. The report's recommendations are engineering inferences, not an approved product decision. The decision register records their disposition; only an explicitly approved Phase 7 plan amendment may create execution authority.
