#!/usr/bin/env python3
"""Generate a project-focused GitIngest digest and provenance manifest."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import subprocess
import sys
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from urllib.parse import urlsplit, urlunsplit

ROOT = Path(__file__).resolve().parents[1]
VENV = ROOT / ".tools" / "gitingest-venv"
OUTPUT = ROOT / "gitingest.txt"
MANIFEST = ROOT / "gitingest.manifest.json"
GITINGEST_VERSION = "0.3.1"
CORE_EXCLUDES = ("releases/", ".tools/", ".tmp/", ".git/")
TEMP_DIR = ROOT / ".tmp"


class GitIngestError(RuntimeError):
    """A user-facing GitIngest setup or generation failure."""


def venv_python() -> Path:
    return VENV / ("Scripts/python.exe" if sys.platform == "win32" else "bin/python")


def install_command() -> list[str]:
    return [str(venv_python()), "-m", "pip", "install", "--upgrade", f"gitingest=={GITINGEST_VERSION}"]


def generate_command(output: Path = OUTPUT, include_submodules: bool = False) -> list[str]:
    command = [str(venv_python()), "-m", "gitingest", str(ROOT)]
    excludes = CORE_EXCLUDES if include_submodules else (*CORE_EXCLUDES, "third_party/")
    for pattern in excludes:
        command.extend(["--exclude-pattern", pattern])
    if include_submodules:
        command.append("--include-submodules")
    command.extend(["--output", str(output)])
    return command


def installed_version() -> str | None:
    if not venv_python().is_file():
        return None
    command = [str(venv_python()), "-c", "from importlib.metadata import version; print(version('gitingest'))"]
    try:
        result = subprocess.run(command, cwd=ROOT, check=True, capture_output=True, text=True)
    except (OSError, subprocess.CalledProcessError):
        return None
    return result.stdout.strip()


def ensure_gitingest() -> None:
    if venv_python().is_file() and installed_version() == GITINGEST_VERSION:
        return
    try:
        subprocess.run([sys.executable, "-m", "venv", str(VENV)], cwd=ROOT, check=True)
        subprocess.run(install_command(), cwd=ROOT, check=True)
    except (OSError, subprocess.CalledProcessError) as error:
        raise GitIngestError(
            "could not install pinned GitIngest into .tools/gitingest-venv; "
            "ensure Python can create virtual environments and allow network access, then rerun the generator"
        ) from error
    if installed_version() != GITINGEST_VERSION:
        raise GitIngestError(f"GitIngest installation did not produce version {GITINGEST_VERSION}")


def run_git(*args: str) -> str:
    try:
        result = subprocess.run(["git", *args], cwd=ROOT, check=True, capture_output=True, text=True)
    except (OSError, subprocess.CalledProcessError) as error:
        raise GitIngestError(f"could not collect Git provenance with `git {' '.join(args)}`") from error
    return result.stdout.strip()


def sanitize_remote(remote: str) -> str:
    parsed = urlsplit(remote)
    if parsed.scheme and parsed.netloc:
        host = parsed.hostname or ""
        if parsed.port:
            host = f"{host}:{parsed.port}"
        return urlunsplit((parsed.scheme, host, parsed.path, parsed.query, parsed.fragment))
    if "@" in remote and ":" in remote:
        return remote.split("@", 1)[1]
    return remote


def submodule_urls() -> dict[str, str]:
    paths = run_git("config", "--file", ".gitmodules", "--get-regexp", r"^submodule\..*\.path$")
    names = {line.split(maxsplit=1)[0][10:-5]: line.split(maxsplit=1)[1] for line in paths.splitlines() if line}
    urls = run_git("config", "--file", ".gitmodules", "--get-regexp", r"^submodule\..*\.url$")
    result = {}
    for line in urls.splitlines():
        if line:
            key, url = line.split(maxsplit=1)
            name = key[10:-4]
            if name in names:
                result[names[name]] = sanitize_remote(url)
    return result


def submodules() -> list[dict[str, str | None]]:
    urls = submodule_urls()
    entries = []
    for line in run_git("submodule", "status", "--recursive").splitlines():
        if line:
            marker = line[0]
            revision, path, *_ = line[1:].split()
            entries.append({"path": path, "revision": revision, "status": marker, "url": urls.get(path)})
    return entries


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def relative_path(path: Path) -> str:
    try:
        return path.resolve().relative_to(ROOT).as_posix()
    except ValueError:
        return str(path)


def repository_provenance() -> dict[str, object]:
    remotes = []
    for name in run_git("remote").splitlines():
        remotes.append({"name": name, "url": sanitize_remote(run_git("remote", "get-url", name))})
    return {
        "commit": run_git("rev-parse", "HEAD"),
        "branch": run_git("branch", "--show-current") or None,
        "dirty": bool(run_git("status", "--porcelain=v1")),
        "remotes": remotes,
        "submodules": submodules(),
        "module_manifests": [
            {"path": name, "sha256": sha256_file(ROOT / name)}
            for name in ("go.mod", "go.sum")
            if (ROOT / name).is_file()
        ],
    }


def manifest_data(output: Path, include_submodules: bool) -> dict[str, object]:
    excludes = CORE_EXCLUDES if include_submodules else (*CORE_EXCLUDES, "third_party/")
    arguments = ["."]
    for pattern in excludes:
        arguments.extend(["--exclude-pattern", pattern])
    if include_submodules:
        arguments.append("--include-submodules")
    arguments.extend(["--output", relative_path(output)])
    return {
        "schema_version": 1,
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "repository": repository_provenance(),
        "generator": {
            "name": "gitingest",
            "version": GITINGEST_VERSION,
            "include_submodules": include_submodules,
            "excluded_patterns": list(excludes),
            "arguments": arguments,
        },
        "artifacts": [
            {"path": relative_path(output), "sha256": sha256_file(output), "purpose": "project-focused digest"}
        ],
    }


def atomic_replace(path: Path, content: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=path.parent, delete=False) as temporary:
        temporary.write(content)
        temporary_path = Path(temporary.name)
    try:
        os.replace(temporary_path, path)
    except OSError:
        temporary_path.unlink(missing_ok=True)
        raise


def generate(output: Path, manifest: Path, include_submodules: bool) -> Path:
    ensure_gitingest()
    TEMP_DIR.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=TEMP_DIR, prefix="gitingest-", suffix=".txt", delete=False) as temporary:
        temporary_output = Path(temporary.name)
    try:
        subprocess.run(generate_command(temporary_output, include_submodules), cwd=ROOT, check=True)
        if not temporary_output.is_file():
            raise GitIngestError("GitIngest completed without producing a digest")
        atomic_replace(output, temporary_output.read_bytes())
        content = json.dumps(manifest_data(output, include_submodules), indent=2, sort_keys=True).encode() + b"\n"
        atomic_replace(manifest, content)
    except (OSError, subprocess.CalledProcessError) as error:
        raise GitIngestError("could not generate the research digest; ensure GitIngest is available and rerun") from error
    finally:
        temporary_output.unlink(missing_ok=True)
    return output


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--include-submodules", action="store_true", help="include third-party submodule source bodies")
    parser.add_argument("--output", type=Path, default=OUTPUT, help="digest output path (default: gitingest.txt)")
    parser.add_argument("--manifest", type=Path, default=MANIFEST, help="provenance sidecar path")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    try:
        output = generate(args.output.resolve(), args.manifest.resolve(), args.include_submodules)
    except GitIngestError as error:
        print(error, file=sys.stderr)
        return 1
    print(f"GitIngest digest: {output}")
    print(f"GitIngest manifest: {args.manifest.resolve()}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
