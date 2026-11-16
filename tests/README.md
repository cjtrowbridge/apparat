# Tests

This directory contains Python tests for repository scripts and build/governance behavior.

Run with:

```bash
python3 -m unittest discover -s tests -p '*_test.py'
make test-build
```

Tests here should avoid network access and should not require GUI libraries. Go package tests live next to the Go packages they verify.

## Source inventory

- `build_pipeline_test.py`: Checks build pipeline test behavior with isolated unittest fixtures.
- `code_size_test.py`: Checks code size test behavior with isolated unittest fixtures.
- `directory_docs_test.py`: Checks directory docs test behavior with isolated unittest fixtures.
- `gitingest_artifact_test.py`: Checks gitingest artifact test behavior with isolated unittest fixtures.
- `governance_migration_test.py`: Checks governance migration test behavior with isolated unittest fixtures.

- `build_governance_test.py`: Checks read-only/idempotent bootstrap, failing and interrupted build gates, and single-command status propagation without invoking compilers.

- `mobile_sources_test.py`: Checks isolated Android engine regeneration, unchanged checkout bytes on success/failure, and complete Java template substitution.

- `build_preparation_test.py`: Checks read-only mobile capability inspection, single preparation and failure-before-binding without producing artifacts.

- `android_release_test.py`: Verifies signer, package, version and missing-key publication boundaries.
