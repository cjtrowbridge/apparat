import tempfile
import unittest
from pathlib import Path

from scripts import check_code_file_lines


class CodeSizeTest(unittest.TestCase):
    def test_reports_file_over_limit(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            path = root / "main.go"
            path.write_text("package main\n" * 3)
            self.assertEqual(check_code_file_lines.violations(root, 2), [(Path("main.go"), 3)])

    def test_excludes_third_party_and_releases(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            third_party = root / "third_party" / "x.go"
            release = root / "releases" / "linux" / "amd64" / "apparat" / "latest"
            third_party.parent.mkdir(parents=True)
            release.parent.mkdir(parents=True)
            third_party.write_text("package x\n" * 99)
            release.write_text("binary-ish\n" * 99)
            self.assertEqual(check_code_file_lines.violations(root, 1), [])

    def test_main_returns_nonzero_on_violation(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / "tool.py").write_text("print('x')\n" * 4)
            self.assertEqual(check_code_file_lines.main(["--root", str(root), "--limit", "3"]), 1)


if __name__ == "__main__":
    unittest.main()
