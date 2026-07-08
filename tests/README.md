# Tests

This directory contains Python tests for repository scripts and build/governance behavior.

Run with:

```bash
python3 -m unittest discover -s tests -p '*_test.py'
make test-build
```

Tests here should avoid network access and should not require GUI libraries. Go package tests live next to the Go packages they verify.
