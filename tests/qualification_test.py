"""Check failure isolation and artifact identity at the release boundary.
Use disposable files and injected build/test outcomes, never an actual phone.
"""
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
from scripts import build_evidence as evidence
from scripts.build_orchestrator import BuildPlan


class QualificationTest(unittest.TestCase):
    def test_failed_gui_never_promotes_or_skips_next_target(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            plans = []
            for target in ('apparat','apparatd'):
                dest = root/'releases/linux/amd64'/target/'latest'
                dest.parent.mkdir(parents=True)
                dest.write_text('old')
                plans.append(BuildPlan('linux/amd64/'+target,'linux','amd64',target,dest,True))
            def compile_target(plan,path):
                path.write_text('fresh '+plan.target)
                return 0
            with patch.object(evidence,'fingerprint',return_value='input'), patch.object(evidence,'git',return_value=b'main'), patch.object(evidence.subprocess,'check_output',return_value='go'), patch.object(evidence.artifact_acceptance,'validate',side_effect=[ValueError('GUI not launched'),None]) as validate:
                self.assertEqual(evidence.qualify(root,plans,compile_target),1)
            self.assertEqual(validate.call_count,2)
            for plan in plans:
                self.assertEqual(plan.output.read_text(),'old')

    def test_wrong_checkout_fails_even_with_matching_identity_claim(self):
        from scripts import checkout_identity
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            with patch.object(Path,'home',return_value=root), patch.object(checkout_identity,'identity',return_value=['original','common']):
                checkout_identity.check(root,register=True)
                checkout_identity.check(root)
            with patch.object(Path,'home',return_value=root), patch.object(checkout_identity,'identity',return_value=['copy','other-common']):
                with self.assertRaises(ValueError):
                    checkout_identity.check(root)

    def test_recover_interrupted_promotion(self):
        import json
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            backup = root/'previous'
            backup.write_text('verified')
            (root/'latest').write_text('partial new set')
            journal = root/'promotion.json'
            journal.write_text(json.dumps([['latest',str(backup)]]))
            evidence.restore(root,journal)
            self.assertEqual((root/'latest').read_text(),'verified')
            self.assertFalse(journal.exists())
