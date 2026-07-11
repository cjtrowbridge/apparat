#!/usr/bin/env python3
"""No-flag Apparat build target detection and orchestration."""
from __future__ import annotations

import argparse
import importlib.util
import os
import shutil
import sys
from dataclasses import dataclass
from pathlib import Path

try:
    from scripts import build
except ModuleNotFoundError:
    import build


@dataclass(frozen=True)
class BuildPlan:
    name: str
    goos: str
    goarch: str
    target: str
    output: Path
    possible: bool
    reasons: tuple[str, ...] = ()
    warnings: tuple[str, ...] = ()


def load_build_environment() -> list[str]:
    env_file = build.ROOT / "build_environment.py"
    if not env_file.exists():
        return []
    spec = importlib.util.spec_from_file_location("apparat_build_environment", env_file)
    if spec is None or spec.loader is None:
        raise build.BuildError(f"cannot load local build environment file: {env_file}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    notes: list[str] = [f"loaded local build environment: {env_file}"]
    if hasattr(module, "update_environment"):
        updated = module.update_environment(dict(os.environ))
        if updated is None:
            raise build.BuildError("build_environment.update_environment must return a dict")
        os.environ.update({str(key): str(value) for key, value in updated.items()})
        notes.append("applied update_environment")
    if hasattr(module, "build_notes"):
        notes.extend(str(note) for note in module.build_notes())
    return notes


def build_plans(go: str) -> list[BuildPlan]:
    plans: list[BuildPlan] = []
    go_path = go if os.path.isabs(go) or os.sep in go else shutil.which(go)
    go_missing = () if go_path else (f"missing Go executable {go!r} on PATH",)
    try:
        goos = build.host_goos()
        goarch = build.host_goarch()
        for target in build.ALL_TARGETS:
            plans.append(BuildPlan(
                name=f"{goos}/{goarch}/{target}",
                goos=goos,
                goarch=goarch,
                target=target,
                output=build.artifact_path(goos, goarch, target),
                possible=not go_missing,
                reasons=go_missing,
            ))
    except ValueError as error:
        append_unsupported_host_plans(plans, error, go_missing)
    toolchain, failures, warnings = build.resolve_android_toolchain(go)
    android_reasons = tuple(failures)
    if go_missing:
        android_reasons = (*go_missing, *android_reasons)
    plans.append(BuildPlan(
        name="android/arm64/apparat",
        goos="android",
        goarch="arm64",
        target="apparat",
        output=build.artifact_path("android", "arm64", "apparat"),
        possible=not android_reasons and toolchain is not None,
        reasons=android_reasons,
        warnings=tuple(warnings),
    ))
    plans.append(BuildPlan(
        name="android/arm64/apparatd",
        goos="android",
        goarch="arm64",
        target="apparatd",
        output=build.artifact_path("android", "arm64", "apparatd"),
        possible=False,
        reasons=(build.ANDROID_HEADLESS_HELP,),
    ))
    return plans


def append_unsupported_host_plans(plans: list[BuildPlan], error: ValueError, go_missing: tuple[str, ...]) -> None:
    for target in build.ALL_TARGETS:
        plans.append(BuildPlan(
            name=f"host/{target}",
            goos="host",
            goarch="host",
            target=target,
            output=Path(),
            possible=False,
            reasons=(str(error), *go_missing),
        ))


def print_plan_report(plans: list[BuildPlan], notes: list[str]) -> None:
    if notes:
        print("Environment:")
        for note in notes:
            print(f"- {note}")
    print("Build target report:")
    for plan in plans:
        status = "possible" if plan.possible else "impossible"
        print(f"- {plan.name}: {status}")
        if plan.output:
            print(f"  output: {plan.output}")
        for warning in plan.warnings:
            print(f"  warning: {warning}")
        for reason in plan.reasons:
            print(f"  reason: {reason}")


def run_plan(go: str, plan: BuildPlan) -> int:
    if plan.goos == "android":
        return build.build_android(go, plan.goarch)
    return build.build_desktop(go, plan.goos, plan.goarch, (plan.target,))


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Detect Apparat build capabilities, report target status, and build every possible target.",
        epilog=(
            "Canonical invocation: python3 scripts/build.py\n\n"
            "Operational target flags were intentionally removed. Use the no-flag "
            "entry point so it can report possible and impossible targets together."
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("--go", default=os.environ.get("GO", "go"), help=argparse.SUPPRESS)
    args, extras = parser.parse_known_args(argv)
    if extras:
        parser.error("scripts/build.py no longer accepts target flags; run it with no arguments")
    return args


def main(argv: list[str] | None = None) -> int:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    try:
        notes = load_build_environment()
    except build.BuildError as error:
        print(error, file=sys.stderr)
        return 2
    plans = build_plans(args.go)
    print_plan_report(plans, notes)
    possible = [plan for plan in plans if plan.possible]
    if not possible:
        print("No build targets are possible on this machine.", file=sys.stderr)
        return 1
    exit_code = 0
    for plan in possible:
        result = run_plan(args.go, plan)
        if result != 0:
            exit_code = result
    return exit_code


if __name__ == "__main__":
    raise SystemExit(main())
