# Scripts

This directory contains repository-local automation used by developers and CI-like local checks.

Run scripts from the repository root unless a script explicitly says otherwise. Prefer Makefile targets when they exist because they provide the expected Go cache and tool settings.

## Inventory

- `build.py`: Builds canonical local release artifacts for `apparat` and `apparatd`.
- `check_code_file_lines.py`: Fails when included code files exceed the 400-line limit.
- `check_directory_docs.py`: Fails when source directories or scripts are missing required documentation.
- `regenerate_plan_indexes.py`: Validates plan files and regenerates plan indexes.
- `run_artifact.py`: Runs a built artifact while forwarding arguments after an optional `--`.

## Build Script

Canonical build commands:

```bash
make build
python3 scripts/build.py
python3 scripts/build.py --target apparatd
python3 scripts/build.py --target apparat --print-path
```

`make build` is preferred because it applies repo-local Go cache settings. `python3 build.py` at the repository root is a compatibility wrapper that delegates to `python3 scripts/build.py`.

Outputs:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

The `apparat` target is compiled with the `gui` build tag. On Linux, it requires native desktop development headers used by Ebitengine and GLFW, including `libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libgl1-mesa-dev`, `libxxf86vm-dev`, and `libasound2-dev`.

The `apparatd` target avoids the GUI build tag and is the correct target for headless workers and display-free validation.

## Verification Scripts

Run:

```bash
make check-code-size
make check-docs
make test-build
```

`make verify` runs all three checks along with Go formatting, Go tests, race tests, linting, and vulnerability scanning.

## Failure Modes

- Missing Go modules: rerun through `make build` or allow network access so Go can populate the module cache.
- Missing Linux GUI headers: install the native desktop development packages listed above, then rerun `make build`.
- Missing directory docs: add a `README.md` to the source directory reported by `check_directory_docs.py`.
- Missing script help: add argparse-style `--help` output before considering the script complete.
