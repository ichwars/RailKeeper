import base64
import json
from pathlib import Path
import tempfile
import unittest
import xml.etree.ElementTree as ET

from tools.build_function_symbols import (
    ACTIVE_PALETTE,
    PRINT_PALETTE,
    build_sql,
    encode_svg,
    load_library,
    render_svg,
    validate_svg,
)


def geometry_signature(svg_text: str) -> str:
    root = ET.fromstring(svg_text)
    for element in root.iter():
        element.attrib.pop("stroke", None)
        element.attrib.pop("fill", None)
        element.attrib.pop("data-rk-role", None)
    return ET.tostring(root, encoding="unicode", short_empty_elements=True)


class FunctionSymbolGeneratorTest(unittest.TestCase):
    def test_rejects_unsafe_svg_content(self) -> None:
        unsafe = (
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">'
            "<script>alert(1)</script></svg>"
        )
        with self.assertRaisesRegex(ValueError, "script"):
            validate_svg(unsafe, "unsafe.svg")

    def test_rejects_wrong_view_box_and_event_handlers(self) -> None:
        wrong_view_box = (
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">'
            '<path d="M0 0h1"/></svg>'
        )
        with self.assertRaisesRegex(ValueError, "viewBox"):
            validate_svg(wrong_view_box, "small.svg")

        event_handler = (
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">'
            '<path onclick="alert(1)" d="M0 0h1"/></svg>'
        )
        with self.assertRaisesRegex(ValueError, "onclick"):
            validate_svg(event_handler, "event.svg")

    def test_renders_geometry_identically_for_all_palettes(self) -> None:
        source = (
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">'
            '<path data-rk-role="primary" fill="none" stroke="#111111" d="M8 32h48"/>'
            '<path data-rk-role="accent" fill="none" stroke="#419310" d="M32 8v48"/>'
            "</svg>"
        )
        active = render_svg(source, ACTIVE_PALETTE)
        printed = render_svg(source, PRINT_PALETTE)
        self.assertIn("#f2f5f6", active)
        self.assertIn("#111111", printed)
        self.assertEqual(geometry_signature(active), geometry_signature(printed))

    def test_data_url_is_base64_svg(self) -> None:
        svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"/>'
        value = encode_svg(svg)
        self.assertTrue(value.startswith("data:image/svg+xml;base64,"))
        self.assertEqual(base64.b64decode(value.split(",", 1)[1]).decode("utf-8"), svg)

    def test_loads_library_and_builds_deterministic_sql(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "light.svg").write_text(
                '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">'
                '<path data-rk-role="primary" fill="none" stroke="#111111" '
                'stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round" '
                'd="M8 32h48"/></svg>',
                encoding="utf-8",
            )
            manifest = {
                "version": 1,
                "library": "railkeeper-workshop-line",
                "symbols": [
                    {
                        "key": "light",
                        "label": "Licht",
                        "category": "Licht",
                        "description": "RailKeeper Funktionssymbol: Licht.",
                        "sortOrder": 10,
                        "file": "light.svg",
                    }
                ],
            }
            (root / "manifest.json").write_text(
                json.dumps(manifest, ensure_ascii=False), encoding="utf-8"
            )

            library = load_library(root)
            first = build_sql(library)
            second = build_sql(library)

        self.assertEqual(first, second)
        self.assertIn("railkeeper-workshop-line", first)
        self.assertIn("symbols:light", first)
        self.assertIn("ON CONFLICT(type, key) DO UPDATE", first)


if __name__ == "__main__":
    unittest.main()
