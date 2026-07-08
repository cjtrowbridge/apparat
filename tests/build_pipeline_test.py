import unittest
from contextlib import redirect_stdout
from io import StringIO
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

    def test_all_targets_selects_gui_and_headless(self):
        self.assertEqual(build.selected_targets("all"), ("apparat", "apparatd"))

    def test_gui_build_uses_gui_tag(self):
        command = build.build_command("go", "apparat", build.artifact_path("linux", "amd64", "apparat"))
        self.assertIn("-tags", command)
        self.assertIn("gui", command)

    def test_headless_build_does_not_use_gui_tag(self):
        command = build.build_command("go", "apparatd", build.artifact_path("linux", "amd64", "apparatd"))
        self.assertNotIn("-tags", command)

    def test_all_targets_builds_distinct_outputs(self):
        outputs = [build.artifact_path("linux", "amd64", target) for target in build.selected_targets("all")]
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


if __name__ == "__main__":
    unittest.main()
