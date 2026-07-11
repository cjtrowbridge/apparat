"""Optional local build environment hooks for Apparat.

Copy this file to `build_environment.py` when a machine needs local paths or
environment tweaks before `python3 scripts/build.py` detects build targets.
The copied file is ignored by Git.
"""


def update_environment(env):
    """Return an updated environment dict for the build process."""
    # env["JAVA_HOME"] = "/path/to/jdk-21"
    # env["ANDROID_HOME"] = "/path/to/android-sdk"
    # env["ANDROID_NDK_HOME"] = "/path/to/android-sdk/ndk/27.2.12479018"
    return env


def build_notes():
    """Return optional human-readable notes printed in the build report."""
    return []
