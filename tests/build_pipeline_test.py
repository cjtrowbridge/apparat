import unittest
from contextlib import redirect_stderr, redirect_stdout
from io import StringIO
from pathlib import Path
from unittest import mock

from scripts import build


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

    def test_android_build_uses_gomobile_and_gui_tag(self):
        command = build.android_build_command(Path("gomobile"), "arm64", build.artifact_path("android", "arm64", "apparat"))
        self.assertEqual(command[:2], ["gomobile", "build"])
        self.assertIn("android/arm64", command)
        self.assertIn("gui", command)
        self.assertIn("./cmd/apparat", command)

    def test_all_targets_builds_distinct_outputs(self):
        outputs = [build.artifact_path("linux", "amd64", target) for target in build.selected_targets("all", "linux")]
        self.assertEqual(
            [path.as_posix().split("/")[-2:] for path in outputs],
            [["apparat", "latest"], ["apparatd", "latest"]],
        )

    def test_print_path_does_not_build(self):
        with mock.patch("subprocess.run") as run:
            with redirect_stdout(StringIO()):
                result = build.main(["--os", "linux", "--arch", "amd64", "--target", "apparat", "--print-path"])
        self.assertEqual(result, 0)
        run.assert_not_called()

    def test_android_print_path_does_not_build(self):
        with mock.patch("subprocess.run") as run:
            output = StringIO()
            with redirect_stdout(output):
                result = build.main(["--os", "android", "--arch", "arm64", "--target", "apparat", "--print-path"])
        self.assertEqual(result, 0)
        run.assert_not_called()
        self.assertTrue(output.getvalue().strip().endswith("releases/android/arm64/apparat/latest.apk"))

    def test_android_headless_is_rejected(self):
        with mock.patch("subprocess.run") as run:
            error = StringIO()
            with redirect_stderr(error):
                result = build.main(["--os", "android", "--arch", "arm64", "--target", "apparatd"])
        self.assertEqual(result, 2)
        run.assert_not_called()
        self.assertIn("GUI `apparat` APK", error.getvalue())

    def test_android_unsupported_arch_is_rejected(self):
        with redirect_stderr(StringIO()):
            result = build.main(["--os", "android", "--arch", "amd64", "--target", "apparat"])
        self.assertEqual(result, 2)

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


if __name__ == "__main__":
    unittest.main()
