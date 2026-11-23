"""
Exercise bootstrap and the shared build gate at failure and recovery boundaries.
Mocks process execution so missing prerequisites never compile or install software.
Temporary directories verify check mode preserves absent installation state.
"""
import contextlib
import importlib.util
import io
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

from scripts import build_gate

SPEC = importlib.util.spec_from_file_location('bootstrap_v1', Path('scripts/bootstrap-v1.py'))
bootstrap = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(bootstrap)


class BuildGovernanceTest(unittest.TestCase):
    def test_failing_gate_never_runs_requested_command(self):
        with patch.object(sys, 'argv', ['build_gate.py', '--', 'compiler']), patch.object(
            build_gate.subprocess, 'run', side_effect=subprocess.CalledProcessError(7, 'check')
        ), patch.object(build_gate.subprocess, 'call') as invoke:
            with self.assertRaises(subprocess.CalledProcessError):
                build_gate.main()
            invoke.assert_not_called()

    def test_valid_gate_runs_once_and_preserves_failure_status(self):
        with patch.object(sys, 'argv', ['build_gate.py', '--', 'compiler', 'input']), patch.object(
            build_gate.subprocess, 'run'
        ), patch.object(build_gate.subprocess, 'call', return_value=9) as invoke:
            self.assertEqual(build_gate.main(), 9)
            self.assertEqual(invoke.call_count, 1)
            self.assertEqual(invoke.call_args.args[0], ['compiler', 'input'])

    def test_check_mode_never_copies_or_installs_hooks(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            with patch('scripts.checkout_identity.check'), patch.object(bootstrap, 'ROOT', root), patch.object(sys, 'argv', [
                'bootstrap-v1.py', '--check', '--go-source', '/unavailable'
            ]), patch.object(bootstrap, 'probe', return_value=(False, 'absent')) as probe, patch.object(
                bootstrap.shutil, 'copytree'
            ) as copy, contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(bootstrap.main(), 1)
                copy.assert_not_called()
                self.assertIn('--check', probe.call_args.args[0])
                self.assertEqual(list(root.iterdir()), [])

    def test_existing_local_input_is_not_copied_again(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / '.tools/go1.26.11').mkdir(parents=True)
            with patch('scripts.checkout_identity.check'), patch.object(bootstrap, 'ROOT', root), patch.object(sys, 'argv', [
                'bootstrap-v1.py', '--go-source', '/unavailable'
            ]), patch.object(bootstrap, 'probe', return_value=(True, 'go1.26.11')), patch.object(
                bootstrap.shutil, 'copytree'
            ) as copy, contextlib.redirect_stdout(io.StringIO()):
                bootstrap.main()
                copy.assert_not_called()

    def test_interruption_does_not_invoke_build(self):
        with patch.object(sys, 'argv', ['build_gate.py', '--', 'compiler']), patch.object(
            build_gate.subprocess, 'run', side_effect=KeyboardInterrupt
        ), patch.object(build_gate.subprocess, 'call') as invoke:
            with self.assertRaises(KeyboardInterrupt):
                build_gate.main()
            invoke.assert_not_called()


if __name__ == '__main__':
    unittest.main()
