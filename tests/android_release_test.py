"""
Verify Android publication version and signer boundaries without device effects.
Uses controlled metadata and temporary missing keys; never creates signing material.
"""
from pathlib import Path
import unittest
from scripts.android_release import PACKAGE, validate_candidate, validate_keystore


class AndroidReleaseTest(unittest.TestCase):
    def test_newer_installed_signer_accepts_old_channel_transition(self):
        validate_candidate(
            {'package': PACKAGE, 'version': 2, 'signers': ['installed']},
            {'package': PACKAGE, 'version': 1, 'signers': ['old-channel']}, 'installed')

    def test_rejects_same_or_older_versions(self):
        for version in (1, 2):
            with self.subTest(version=version), self.assertRaises(ValueError):
                validate_candidate(
                    {'package': PACKAGE, 'version': version, 'signers': ['installed']},
                    {'package': PACKAGE, 'version': 2, 'signers': ['installed']}, 'installed')

    def test_rejects_wrong_or_extra_signers(self):
        for signers in ([], ['other'], ['installed', 'other']):
            with self.subTest(signers=signers), self.assertRaises(ValueError):
                validate_candidate(
                    {'package': PACKAGE, 'version': 2, 'signers': signers},
                    {'package': PACKAGE, 'version': 1, 'signers': ['installed']}, 'installed')

    def test_rejects_wrong_package(self):
        with self.assertRaises(ValueError):
            validate_candidate(
                {'package': 'other.app', 'version': 2, 'signers': ['installed']},
                {'package': PACKAGE, 'version': 1, 'signers': ['installed']}, 'installed')

    def test_missing_key_is_not_regenerated(self):
        import tempfile
        with tempfile.TemporaryDirectory() as directory:
            key = Path(directory) / 'missing.keystore'
            with self.assertRaisesRegex(ValueError, 'restore the existing'):
                validate_keystore(Path(directory), key)
            self.assertFalse(key.exists())
