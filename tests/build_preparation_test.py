"""
Verify capability inspection is read-only and APK execution prepares once.
Mocks external tools so failure coverage does not overwrite release artifacts.
Protects the boundary between eligibility, preparation and binding.
"""
from pathlib import Path
import unittest
from unittest.mock import patch
from scripts import build


class PreparationTest(unittest.TestCase):
    def setUp(self):
        signer = patch.object(build, "ensure_debug_keystore")
        signer.start()
        self.addCleanup(signer.stop)

    def test_inspection_never_builds(self):
        with patch.dict(build.os.environ, {}, clear=True), patch.object(
            build, 'module_cache_dir', return_value=Path('/cache')
        ), patch.object(build, 'check_path'), patch.object(build, 'build_patched_gomobile') as compile:
            self.assertEqual(build.inspect_gomobile('go', []), build.PATCHED_GOMOBILE)
            compile.assert_not_called()

    def test_failed_preparation_never_binds(self):
        toolchain = build.AndroidToolchain(Path('/sdk'), Path('/ndk'), None, Path('/java'), Path('/mobile'), None)
        with patch.object(build, 'resolve_android_toolchain', return_value=(toolchain, [], [])), patch.object(
            build, 'ensure_patched_gomobile', return_value=None
        ) as prepare, patch.object(build.android_wrapper, 'build_wrapper_apk') as bind:
            self.assertEqual(build.build_android('go', 'arm64'), 1)
            prepare.assert_called_once()
            bind.assert_not_called()

    def test_build_prepares_once(self):
        toolchain = build.AndroidToolchain(Path('/sdk'), Path('/ndk'), None, Path('/java'), Path('/mobile'), None)
        with patch.object(build, 'resolve_android_toolchain', return_value=(toolchain, [], [])), patch.object(
            build, 'ensure_patched_gomobile', return_value=toolchain.gomobile
        ) as prepare, patch.object(build.android_wrapper, 'build_wrapper_apk') as bind, patch.object(
            build, 'ensure_debug_keystore'
        ), patch.object(build, 'sign_android_apk'), patch.object(build, 'validate_android_apk'), patch.object(Path, 'mkdir'):
            self.assertEqual(build.build_android('go', 'arm64'), 0)
            prepare.assert_called_once()
            bind.assert_called_once()

    def test_interruption_never_binds(self):
        toolchain = build.AndroidToolchain(Path('/sdk'), Path('/ndk'), None, Path('/java'), Path('/mobile'), None)
        with patch.object(build, 'resolve_android_toolchain', return_value=(toolchain, [], [])), patch.object(
            build, 'ensure_patched_gomobile', side_effect=KeyboardInterrupt
        ), patch.object(build.android_wrapper, 'build_wrapper_apk') as bind:
            with self.assertRaises(KeyboardInterrupt):
                build.build_android('go', 'arm64')
            bind.assert_not_called()
