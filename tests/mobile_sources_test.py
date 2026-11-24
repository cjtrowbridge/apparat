"""
Verify Android source preparation preserves tracked engine bytes and fails closed.
Uses isolated Git archive fixtures and temporary workspaces, never real devices.
Checks rebuild identity, malformed-source failure and template substitution.
"""
import io
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest.mock import patch

from scripts import android_wrapper, mobile_sources


class MobileSourcesTest(unittest.TestCase):
    def archive(self, content):
        data = io.BytesIO()
        with tarfile.open(fileobj=data, mode='w') as bundle:
            info = tarfile.TarInfo('internal/ui/ui_mobile.go')
            encoded = content.encode()
            info.size = len(encoded)
            bundle.addfile(info, io.BytesIO(encoded))
            keyboard = tarfile.TarInfo('internal/ui/context.go')
            calculation = b'caretBottom := float64(caretBounds.Max.Y)*scale + offsetY'
            keyboard.size = len(calculation)
            bundle.addfile(keyboard, io.BytesIO(calculation))
        return data.getvalue()

    def test_preparation_preserves_checkout_and_replaces_previous_output(self):
        original = 'func display() {\n\tif v == nil {\n\t\treturn 0, 0, 1, false\n\t}\n}\n'
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / 'third_party/game/ebiten/internal/ui/ui_mobile.go'
            source.parent.mkdir(parents=True)
            source.write_text(original)
            archive = self.archive(original)
            with patch.object(mobile_sources, 'ROOT', root), patch.object(
                mobile_sources.subprocess, 'check_output', side_effect=['a' * 40, archive, 'b' * 40, archive]
            ):
                first = mobile_sources.prepare_engine()
                (first / 'stale').write_text('old generated input')
                second = mobile_sources.prepare_engine()
                self.assertFalse((second / 'stale').exists())
                self.assertTrue((second / '.apparat-source').read_text().startswith('b' * 40))
                self.assertEqual(source.read_text(), original)
                self.assertIn('min(float64(c.screenHeight)', (second / 'internal/ui/context.go').read_text())
                self.assertNotEqual((second / 'internal/ui/ui_mobile.go').read_text(), original)

    def test_unknown_source_fails_without_changing_checkout(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / 'third_party/game/ebiten/untouched'
            source.parent.mkdir(parents=True)
            source.write_bytes(b'preserve')
            with patch.object(mobile_sources, 'ROOT', root), patch.object(
                mobile_sources.subprocess, 'check_output', side_effect=['a' * 40, self.archive('unexpected source')]
            ):
                with self.assertRaisesRegex(RuntimeError, 'Unsupported Ebitengine'):
                    mobile_sources.prepare_engine()
            self.assertEqual(source.read_bytes(), b'preserve')

    def test_selected_view_templates_have_no_unexpanded_placeholders(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            android_wrapper.write_ebiten_view_sources(output)
            for source in output.glob('*.java'):
                text = source.read_text()
                self.assertNotIn('$Placeholder_', text)
                self.assertIn('package com.cjtrowbridge.apparat.apparatmobile;', text)


if __name__ == '__main__':
    unittest.main()
