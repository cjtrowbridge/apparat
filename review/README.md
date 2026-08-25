# Deep-Research Review Artifacts

This directory preserves the supplied deep-research report as advisory input and records the smaller review that determines which of its Phase 7 recommendations are supported by the repository. It is not a second research pipeline and it does not change product behavior, `ROADMAP.md`, or the Phase 7 execution plan by itself.

## Entry points

- `review.md` is the original 58-part review rubric.
- `deep-research.md` is the completed external research report. It is historical input and must not be edited during this review.
- `deep-research-provenance.md` identifies the source snapshot, hashes, and limitations.
- `deep-research-claim-ledger.md` maps the material claims to source locators and verification outcomes.
- `deep-research-decisions.md` records the disposition of verified recommendations.
- `phase-7-amendment-proposal.md` proposes bounded additions to the existing Phase 7 plan; it is not an approved plan amendment.

## Operating rule

The report may suggest a decision, but only the claim ledger's repository or pinned-upstream evidence can support one. An accepted decision remains advisory until the user explicitly approves a revision to `plans/future/2026-07-20-09-04-03_execute-phase-7.md`.

## Provenance and retention

The review's source corpus is the clean GitIngest snapshot documented in `deep-research-provenance.md`. Future reviews must create a new provenance record rather than silently reusing this one after product-source changes. The report itself is preserved byte-for-byte.

## Deterministic comment-presence measurement

The report's comment claim is reproducible only as a narrow source-level measurement: count non-test Go function declarations in `cmd/` and `internal/` whose immediately preceding line begins with `//`. This does not assess comment usefulness, coverage of multi-line declarations, or semantic correctness.

```powershell
$fileCount=0; $functionCount=0; $precededByComment=0
Get-ChildItem cmd,internal -Recurse -File -Filter *.go |
  Where-Object { $_.Name -notlike '*_test.go' } |
  ForEach-Object {
    $fileCount++; $lines=Get-Content $_.FullName
    for($index=0; $index -lt $lines.Count; $index++) {
      if($lines[$index] -match '^\s*func\s+(?:\([^)]*\)\s*)?[A-Za-z_]\w*(?:\[[^]]+\])?\s*\(') {
        $functionCount++
        if($index -gt 0 -and $lines[$index-1] -match '^\s*//') { $precededByComment++ }
      }
    }
  }
[pscustomobject]@{Files=$fileCount;NonTestFunctionDeclarations=$functionCount;ImmediatelyPrecededByLineComment=$precededByComment}
```

For the reviewed snapshot, the command reports 35 files, 243 non-test declarations, and 2 declarations immediately preceded by a line comment. Any future policy must use semantic review to decide whether comments are useful.

## Validation

Review each accepted decision from the decision register against its claim ID and source locator. Confirm that the provenance hashes match their named source artifacts and that the Phase 7 proposal names no changes outside the accepted decisions. Run `python agentic-pipelines/scripts/regenerate_plan_indexes.py --check --repo-root .` after lifecycle changes, and `python scripts/check_directory_docs.py` after adding directory documentation.
