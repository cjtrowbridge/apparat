import hashlib
import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest import mock

from scripts import generate_gitingest


class GitIngestArtifactTest(unittest.TestCase):
    def test_manifest_records_digest_hash_and_provenance(self):
        with TemporaryDirectory() as directory:
            output = Path(directory) / "digest.txt"
            output.write_text("project evidence\n", encoding="utf-8")
            provenance = {"commit": "a" * 40, "dirty": False, "submodules": []}
            with mock.patch.object(generate_gitingest, "repository_provenance", return_value=provenance):
                manifest = generate_gitingest.manifest_data(output, False)
            expected_hash = hashlib.sha256(output.read_bytes()).hexdigest()
        self.assertEqual(manifest["repository"], provenance)
        self.assertFalse(manifest["generator"]["include_submodules"])
        self.assertIn("third_party/", manifest["generator"]["excluded_patterns"])
        self.assertEqual(manifest["generator"]["arguments"][-1], str(output))
        self.assertEqual(manifest["artifacts"][0]["sha256"], expected_hash)

    def test_atomic_replace_writes_json_without_leaving_temp_file(self):
        with TemporaryDirectory() as directory:
            output = Path(directory) / "manifest.json"
            generate_gitingest.atomic_replace(output, json.dumps({"schema_version": 1}).encode())
            self.assertEqual(json.loads(output.read_text(encoding="utf-8")), {"schema_version": 1})

    def test_generation_stages_digest_in_ignored_tmp_directory_before_provenance(self):
        with TemporaryDirectory() as directory:
            output = Path(directory) / "digest.txt"
            manifest_path = Path(directory) / "manifest.json"
            observed = {}

            def generate_digest(command, **_kwargs):
                staged_output = Path(command[-1])
                observed["staged_output"] = staged_output
                staged_output.write_text("project evidence\n", encoding="utf-8")

            def provenance():
                staged_output = observed["staged_output"]
                observed["staged_exists_during_provenance"] = staged_output.is_file()
                return {"commit": "a" * 40, "dirty": False, "submodules": []}

            with (
                mock.patch.object(generate_gitingest, "ensure_gitingest"),
                mock.patch.object(generate_gitingest.subprocess, "run", side_effect=generate_digest),
                mock.patch.object(generate_gitingest, "repository_provenance", side_effect=provenance),
            ):
                generate_gitingest.generate(output, manifest_path, False)

        self.assertEqual(observed["staged_output"].parent, generate_gitingest.TEMP_DIR)
        self.assertTrue(observed["staged_exists_during_provenance"])
        self.assertFalse(observed["staged_output"].exists())

    def test_remote_sanitization_removes_user_information(self):
        self.assertEqual(
            generate_gitingest.sanitize_remote("https://token@example.test/owner/repo.git"),
            "https://example.test/owner/repo.git",
        )
        self.assertEqual(
            generate_gitingest.sanitize_remote("git@example.test:owner/repo.git"),
            "example.test:owner/repo.git",
        )


if __name__ == "__main__":
    unittest.main()
