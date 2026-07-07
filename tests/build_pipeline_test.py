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
        self.assertEqual(path.as_posix().split("/")[-4:], ["releases", "windows", "amd64", "latest.exe"])

    def test_linux_artifact_uses_latest_without_suffix(self):
        path = build.artifact_path("linux", "arm64", "apparat")
        self.assertEqual(path.as_posix().split("/")[-4:], ["releases", "linux", "arm64", "latest"])

    def test_print_path_does_not_build(self):
        with mock.patch("subprocess.run") as run:
            with redirect_stdout(StringIO()):
                result = build.main(["--os", "linux", "--arch", "amd64", "--print-path"])
        self.assertEqual(result, 0)
        run.assert_not_called()


if __name__ == "__main__":
    unittest.main()
