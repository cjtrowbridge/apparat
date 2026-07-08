import tempfile
import unittest
from pathlib import Path
from unittest import mock

from scripts import check_directory_docs


class DirectoryDocsTest(unittest.TestCase):
    def test_reports_source_directory_without_readme(self):
        files = [Path("internal/config/config.go")]
        self.assertEqual(check_directory_docs.missing_directory_readmes(Path("/tmp/nope"), files), [Path("internal/config")])

    def test_accepts_source_directory_with_readme(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            directory = root / "internal" / "config"
            directory.mkdir(parents=True)
            (directory / "README.md").write_text("# Config\n")
            files = [Path("internal/config/config.go")]
            self.assertEqual(check_directory_docs.missing_directory_readmes(root, files), [])

    def test_reports_scripts_missing_from_scripts_readme(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / "scripts").mkdir()
            (root / "scripts" / "README.md").write_text("build.py\n")
            files = [Path("scripts/build.py"), Path("scripts/run_artifact.py")]
            self.assertEqual(check_directory_docs.undocumented_scripts(root, files), [Path("scripts/run_artifact.py")])

    def test_reports_script_without_help_usage(self):
        files = [Path("scripts/build.py")]
        completed = mock.Mock(returncode=2, stdout="", stderr="boom")
        with mock.patch("subprocess.run", return_value=completed):
            self.assertEqual(check_directory_docs.scripts_without_help(Path("/tmp/repo"), files), [Path("scripts/build.py")])


if __name__ == "__main__":
    unittest.main()
