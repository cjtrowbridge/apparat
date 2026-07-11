import unittest
from contextlib import redirect_stderr, redirect_stdout
from io import StringIO
from pathlib import Path
from unittest import mock

from scripts import build
from scripts import build_orchestrator
from scripts import android_wrapper


class BuildPipelineTest(unittest.TestCase):
    def test_linux_x86_64_maps_to_go_values(self):
        with mock.patch("platform.system", return_value="Linux"), mock.patch(
            "platform.machine", return_value="x86_64"
        ):
            self.assertEqual(build.host_goos(), "linux")
            self.assertEqual(build.host_goarch(), "amd64")

    def test_windows_artifact_uses_exe_suffix(self):
        path = build.artifact_path("windows", "amd64", "apparat")
        self.assertEqual(path.as_posix().split("/")[-5:], ["releases", "windows", "amd64", "apparat", "latest.exe"])

    def test_windows_headless_artifact_uses_exe_suffix(self):
        path = build.artifact_path("windows", "amd64", "apparatd")
        self.assertEqual(path.as_posix().split("/")[-5:], ["releases", "windows", "amd64", "apparatd", "latest.exe"])

    def test_linux_artifact_uses_latest_without_suffix(self):
        path = build.artifact_path("linux", "arm64", "apparat")
        self.assertEqual(path.as_posix().split("/")[-5:], ["releases", "linux", "arm64", "apparat", "latest"])

    def test_linux_headless_artifact_uses_latest_without_suffix(self):
        path = build.artifact_path("linux", "arm64", "apparatd")
        self.assertEqual(path.as_posix().split("/")[-5:], ["releases", "linux", "arm64", "apparatd", "latest"])

    def test_android_artifact_uses_apk_suffix(self):
        path = build.artifact_path("android", "arm64", "apparat")
        self.assertEqual(path.as_posix().split("/")[-5:], ["releases", "android", "arm64", "apparat", "latest.apk"])

    def test_all_targets_selects_gui_and_headless_for_desktop(self):
        self.assertEqual(build.selected_targets("all", "linux"), ("apparat", "apparatd"))

    def test_all_targets_selects_gui_only_for_android(self):
        self.assertEqual(build.selected_targets("all", "android"), ("apparat",))

    def test_gui_build_uses_gui_tag(self):
        command = build.desktop_build_command("go", "apparat", build.artifact_path("linux", "amd64", "apparat"))
        self.assertIn("-tags", command)
        self.assertIn("gui", command)

    def test_headless_build_does_not_use_gui_tag(self):
        command = build.desktop_build_command("go", "apparatd", build.artifact_path("linux", "amd64", "apparatd"))
        self.assertNotIn("-tags", command)


    def test_android_sdk_metadata_constants_are_modern(self):
        self.assertEqual(build.ANDROID_MIN_API, "23")
        self.assertEqual(build.ANDROID_API, "35")
        self.assertEqual(build.ANDROID_TARGET_API, "30")

    def test_android_badging_expectations_include_sdk_and_abi(self):
        self.assertEqual(
            build.android_badging_expectations("arm64"),
            ("minSdkVersion:'23'", "targetSdkVersion:'30'", "native-code: 'arm64-v8a'"),
        )

    def test_android_manifest_does_not_force_orientation(self):
        manifest = Path("android/apparat/AndroidManifest.xml").read_text(encoding="utf-8")
        self.assertNotIn("android:screenOrientation", manifest)
        self.assertNotIn('android:screenOrientation="landscape"', manifest)
        self.assertIn('android:name=".MainActivity"', manifest)

    def test_android_wrapper_activity_uses_ebiten_view(self):
        activity = Path("android/apparat/src/com/cjtrowbridge/apparat/MainActivity.java").read_text(encoding="utf-8")
        self.assertIn("Apparatmobile.ready()", activity)
        self.assertIn("new EbitenView(this)", activity)

    def test_android_manifest_declares_temporary_update_installer_boundary(self):
        manifest = Path("android/apparat/AndroidManifest.xml").read_text(encoding="utf-8")
        self.assertIn("android.permission.REQUEST_INSTALL_PACKAGES", manifest)
        self.assertIn('android:icon="@drawable/app_icon"', manifest)
        self.assertIn("UpdateApkProvider", manifest)
        self.assertIn("com.cjtrowbridge.apparat.update", manifest)

    def test_android_wrapper_has_temporary_update_button(self):
        activity = Path("android/apparat/src/com/cjtrowbridge/apparat/MainActivity.java").read_text(encoding="utf-8")
        provider = Path("android/apparat/src/com/cjtrowbridge/apparat/UpdateApkProvider.java").read_text(encoding="utf-8")
        mobile = Path("cmd/apparatmobile/mobile.go").read_text(encoding="utf-8")
        self.assertIn("raw.githubusercontent.com/cjtrowbridge/apparat/main/releases/android/arm64/apparat/latest.apk", activity)
        self.assertIn("ACTION_MANAGE_UNKNOWN_APP_SOURCES", activity)
        self.assertIn("canRequestPackageInstalls", activity)
        self.assertIn("setUseCaches(false)", activity)
        self.assertIn("application/vnd.android.package-archive", provider)

    def test_all_targets_builds_distinct_outputs(self):
        outputs = [build.artifact_path("linux", "amd64", target) for target in build.selected_targets("all", "linux")]
        self.assertEqual(
            [path.as_posix().split("/")[-2:] for path in outputs],
            [["apparat", "latest"], ["apparatd", "latest"]],
        )

    def test_build_script_rejects_target_flags(self):
        with self.assertRaises(SystemExit):
            with redirect_stderr(StringIO()):
                build.main(["--target", "apparat"])

    def test_build_plan_reports_android_headless_impossible(self):
        with mock.patch("scripts.build.resolve_android_toolchain", return_value=(None, ["missing sdk"], [])):
            plans = build_orchestrator.build_plans("go")
        headless = [plan for plan in plans if plan.name == "android/arm64/apparatd"][0]
        self.assertFalse(headless.possible)
        self.assertIn("GUI `apparat` APK", headless.reasons[0])

    def test_build_plan_reports_android_prerequisite_failures(self):
        with mock.patch("scripts.build.resolve_android_toolchain", return_value=(None, ["missing sdk"], [])):
            plans = build_orchestrator.build_plans("go")
        android = [plan for plan in plans if plan.name == "android/arm64/apparat"][0]
        self.assertFalse(android.possible)
        self.assertIn("missing sdk", android.reasons)

    def test_main_runs_all_possible_targets(self):
        plans = [
            build_orchestrator.BuildPlan("linux/amd64/apparat", "linux", "amd64", "apparat", Path("out"), True),
            build_orchestrator.BuildPlan(
                "android/arm64/apparat", "android", "arm64", "apparat", Path("apk"), False, ("missing sdk",)
            ),
        ]
        with mock.patch("scripts.build_orchestrator.load_build_environment", return_value=[]), mock.patch(
            "scripts.build_orchestrator.build_plans", return_value=plans
        ), mock.patch("scripts.build_orchestrator.run_plan", return_value=0) as run:
            with redirect_stdout(StringIO()):
                result = build_orchestrator.main([])
        self.assertEqual(result, 0)
        run.assert_called_once_with("go", plans[0])

    def test_android_env_reports_missing_prerequisites(self):
        with mock.patch("scripts.build.default_sdk_root", return_value=Path("/missing-sdk")), mock.patch(
            "scripts.build.default_java_home", return_value=None
        ), mock.patch("scripts.build.ensure_patched_gomobile", return_value=None):
            error = StringIO()
            with redirect_stderr(error):
                result = build.check_android_env("go")
        self.assertEqual(result, 1)
        self.assertIn("Android build environment check failed", error.getvalue())

    def test_android_pipeline_does_not_reference_salvagecore(self):
        text = Path("scripts/build.py").read_text(encoding="utf-8")
        self.assertNotIn("third_party/salvagecore", text)

    def test_android_wrapper_owns_temporary_ebiten_display_guard(self):
        self.assertIn("scale <= 0", android_wrapper.EBITEN_DISPLAY_INFO_GUARD)
        text = Path("scripts/android_wrapper.py").read_text(encoding="utf-8")
        self.assertIn("patched_ebiten_android_display_info", text)


if __name__ == "__main__":
    unittest.main()
