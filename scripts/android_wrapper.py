#!/usr/bin/env python3
"""Assemble the Android GUI APK with Ebitengine's generated mobile view."""
from __future__ import annotations

import shutil
import subprocess
from contextlib import contextmanager
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
JAVA_PACKAGE = "com.cjtrowbridge.apparat"
PREFIX = "apparatmobile"
PREFIX_UPPER = "Apparatmobile"
EBITEN_DISPLAY_INFO_GUARD = """\tif scale <= 0 || cWidth <= 0 || cHeight <= 0 {
\t\treturn 0, 0, 1, false
\t}
"""


def build_wrapper_apk(toolchain, go: str, goarch: str, output: Path, settings: dict[str, str]) -> None:
    if goarch != "arm64":
        raise ValueError("Android wrapper currently supports only arm64")
    work = ROOT / ".tmp" / "android-apparat-wrapper"
    if work.exists():
        shutil.rmtree(work)
    source_dir = work / "src" / "com" / "cjtrowbridge" / "apparat" / PREFIX
    classes_dir = work / "classes"
    dex_dir = work / "dex"
    apk_dir = work / "apk" / "lib" / "arm64-v8a"
    for path in (source_dir, classes_dir, dex_dir, apk_dir):
        path.mkdir(parents=True, exist_ok=True)
    aar = work / "apparat.aar"
    env = wrapper_env(toolchain, settings)
    with patched_ebiten_android_display_info():
        subprocess.run([
            str(toolchain.gomobile), "bind", "-target", "android/arm64", "-androidapi", settings["api"],
            "-javapkg", JAVA_PACKAGE, "-o", str(aar), "-tags", "gui",
            "-ldflags", f"-extldflags=-Wl,-z,max-page-size={settings['page_size']}", "./cmd/apparatmobile",
            "github.com/hajimehoshi/ebiten/v2/mobile/ebitenmobileview",
        ], cwd=ROOT, env=env, check=True)
    write_ebiten_view_sources(source_dir)
    copy_android_wrapper_sources(work / "src")
    run = lambda args: subprocess.run(args, cwd=ROOT, env=env, check=True)
    base_jar = work / "classes-base.jar"
    native_lib = apk_dir / "libgojni.so"
    base_jar.write_bytes(read_zip_member(aar, "classes.jar"))
    native_lib.write_bytes(read_zip_member(aar, "jni/arm64-v8a/libgojni.so"))
    java_files = sorted(str(path) for path in (work / "src").rglob("*.java"))
    android_jar = toolchain.sdk_root / "platforms" / f"android-{settings['api']}" / "android.jar"
    run([str(toolchain.java_home / "bin" / "javac"), "-source", "1.8", "-target", "1.8", "-bootclasspath", str(android_jar), "-classpath", str(base_jar), "-d", str(classes_dir), *java_files])
    wrapper_jar = work / "wrapper.jar"
    run([str(toolchain.java_home / "bin" / "jar"), "cf", str(wrapper_jar), "-C", str(classes_dir), "."])
    build_tools = toolchain.sdk_root / "build-tools" / settings["build_tools"]
    run([str(build_tools / exe("d8")), "--min-api", settings["min_api"], "--classpath", str(android_jar), "--output", str(dex_dir), str(base_jar), str(wrapper_jar)])
    unsigned = work / "unsigned.apk"
    apk = work / "apparat.apk"
    aligned = work / "aligned.apk"
    run([str(build_tools / exe("aapt2")), "link", "-I", str(android_jar), "--manifest", "android/apparat/AndroidManifest.xml", "--min-sdk-version", settings["min_api"], "--target-sdk-version", settings["target_api"], "-o", str(unsigned)])
    shutil.copy2(unsigned, apk)
    run([str(toolchain.java_home / "bin" / "jar"), "uf", str(apk), "-C", str(work / "apk"), "lib"])
    run([str(toolchain.java_home / "bin" / "jar"), "uf", str(apk), "-C", str(dex_dir), "classes.dex"])
    run([str(build_tools / exe("zipalign")), "-p", "-f", "4", str(apk), str(aligned)])
    output.parent.mkdir(parents=True, exist_ok=True)
    run([str(build_tools / exe("apksigner")), "sign", "--ks", str(ROOT / ".tools" / "android" / "debug.keystore"), "--ks-pass", "pass:android", "--key-pass", "pass:android", "--out", str(output), str(aligned)])
    output.with_name(output.name + ".idsig").unlink(missing_ok=True)


def wrapper_env(toolchain, settings: dict[str, str]) -> dict[str, str]:
    env = dict(__import__("os").environ)
    paths = [
        toolchain.java_home / "bin",
        ROOT / ".tools" / "go1.26.4" / "bin",
        ROOT / ".tools" / "bin",
        toolchain.sdk_root / "platform-tools",
        toolchain.sdk_root / "build-tools" / settings["build_tools"],
    ]
    env["JAVA_HOME"] = str(toolchain.java_home)
    env["ANDROID_HOME"] = str(toolchain.sdk_root)
    env["ANDROID_SDK_ROOT"] = str(toolchain.sdk_root)
    env["ANDROID_NDK_HOME"] = str(toolchain.ndk_root)
    env["GOFLAGS"] = "-buildvcs=false"
    env["GOCACHE"] = str(ROOT / ".tmp" / "go-build")
    env["PATH"] = ":".join(str(path) for path in paths) + ":" + env.get("PATH", "")
    return env


def write_ebiten_view_sources(source_dir: Path) -> None:
    replacements = {"{{.JavaPkg}}": JAVA_PACKAGE, "{{.PrefixLower}}": PREFIX, "{{.PrefixUpper}}": PREFIX_UPPER}
    template_dir = ROOT / "third_party" / "game" / "ebiten" / "cmd" / "ebitenmobile" / "_files"
    for name in ("EbitenView.java", "EbitenSurfaceView.java"):
        text = (template_dir / name).read_text(encoding="utf-8")
        for old, new in replacements.items():
            text = text.replace(old, new)
        if name == "EbitenSurfaceView.java":
            # Ensure the surface uses RGBA_8888 for correct alpha compositing.
            text = text.replace(
                "setEGLConfigChooser(8, 8, 8, 8, 0, 0);",
                "setEGLConfigChooser(8, 8, 8, 8, 0, 0);\n        getHolder().setFormat(android.graphics.PixelFormat.RGBA_8888);"
            )
        (source_dir / name).write_text(text, encoding="utf-8")


def copy_android_wrapper_sources(destination: Path) -> None:
    source = ROOT / "android" / "apparat" / "src"
    for path in source.rglob("*.java"):
        target = destination / path.relative_to(source)
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(path, target)


@contextmanager
def patched_ebiten_android_display_info():
    path = ROOT / "third_party" / "game" / "ebiten" / "internal" / "ui" / "ui_android.go"
    original = path.read_text(encoding="utf-8")
    if EBITEN_DISPLAY_INFO_GUARD in original:
        yield
        return
    needle = "\tscale := float64(cScale)\n"
    if needle not in original:
        raise RuntimeError(f"expected Android display-info patch target not found in {path}")
    path.write_text(original.replace(needle, needle + EBITEN_DISPLAY_INFO_GUARD, 1), encoding="utf-8")
    try:
        yield
    finally:
        path.write_text(original, encoding="utf-8")


def read_zip_member(path: Path, member: str) -> bytes:
    import zipfile
    with zipfile.ZipFile(path) as archive:
        return archive.read(member)


def exe(name: str) -> str:
    import platform
    return f"{name}.exe" if platform.system().lower() == "windows" else name


if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Helper module used by scripts/build.py to assemble the Android GUI wrapper APK.")
    parser.parse_args()
    print("Use `python3 scripts/build.py --os android --arch arm64 --target apparat`.")
