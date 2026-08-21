from __future__ import annotations

import argparse
import base64
from dataclasses import dataclass
from html import escape
import json
from pathlib import Path
import re
from typing import Iterable
import xml.etree.ElementTree as ET


LIBRARY_NAME = "railkeeper-workshop-line"
LIBRARY_VERSION = 1
EXPECTED_SYMBOL_COUNT = 94
EXPECTED_ECOS_SYMBOL_COUNT = 86
SVG_NAMESPACE = "http://www.w3.org/2000/svg"
DATA_URL_PREFIX = "data:image/svg+xml;base64,"


@dataclass(frozen=True)
class Palette:
    primary: str
    accent: str


ACTIVE_PALETTE = Palette(primary="#f2f5f6", accent="#a5ec60")
INACTIVE_PALETTE = Palette(primary="#879398", accent="#66736c")
PRINT_PALETTE = Palette(primary="#111111", accent="#1c621b")

ALLOWED_ELEMENTS = {
    "svg",
    "g",
    "path",
    "line",
    "polyline",
    "polygon",
    "rect",
    "circle",
    "ellipse",
}
FORBIDDEN_ELEMENTS = {"script", "text", "image", "foreignObject", "style", "use"}
FORBIDDEN_ATTRIBUTES = {"href", "src", "style"}
SAFE_FILE_NAME = re.compile(r"^[a-z0-9][a-z0-9-]*\.svg$")


@dataclass(frozen=True)
class Symbol:
    key: str
    label: str
    category: str
    description: str
    sort_order: int
    file_name: str
    svg_text: str
    ecos_code: int | None = None


@dataclass(frozen=True)
class Library:
    version: int
    name: str
    symbols: tuple[Symbol, ...]


def local_name(value: str) -> str:
    return value.rsplit("}", 1)[-1]


def validate_svg(svg_text: str, source: str) -> None:
    lowered = svg_text.lower()
    if "<!doctype" in lowered or "<!entity" in lowered:
        raise ValueError(f"{source}: document declarations are not allowed")
    try:
        root = ET.fromstring(svg_text)
    except ET.ParseError as error:
        raise ValueError(f"{source}: invalid SVG: {error}") from error
    if local_name(root.tag) != "svg":
        raise ValueError(f"{source}: root element must be svg")
    if root.attrib.get("viewBox") != "0 0 64 64":
        raise ValueError(f"{source}: viewBox must be 0 0 64 64")

    semantic_roles = 0
    for element in root.iter():
        name = local_name(element.tag)
        if name in FORBIDDEN_ELEMENTS:
            raise ValueError(f"{source}: {name} elements are not allowed")
        if name not in ALLOWED_ELEMENTS:
            raise ValueError(f"{source}: unsupported element {name}")
        role = element.attrib.get("data-rk-role", "")
        if role:
            if role not in {"primary", "accent"}:
                raise ValueError(f"{source}: unsupported data-rk-role {role}")
            semantic_roles += 1
        for raw_attribute, value in element.attrib.items():
            attribute = local_name(raw_attribute)
            if attribute.lower().startswith("on"):
                raise ValueError(f"{source}: {attribute} event handlers are not allowed")
            if attribute in FORBIDDEN_ATTRIBUTES:
                raise ValueError(f"{source}: {attribute} attributes are not allowed")
            if "url(" in value.lower() or value.strip().lower().startswith(("http:", "https:")):
                raise ValueError(f"{source}: external references are not allowed")
    if semantic_roles == 0:
        raise ValueError(f"{source}: at least one data-rk-role is required")


def render_svg(svg_text: str, palette: Palette) -> str:
    root = ET.fromstring(svg_text)
    ET.register_namespace("", SVG_NAMESPACE)
    for element in root.iter():
        role = element.attrib.pop("data-rk-role", "")
        color = palette.primary if role == "primary" else palette.accent if role == "accent" else ""
        if not color:
            continue
        if element.get("stroke") not in (None, "none"):
            element.set("stroke", color)
        if element.get("fill") not in (None, "none"):
            element.set("fill", color)
    return ET.tostring(root, encoding="unicode", short_empty_elements=True)


def encode_svg(svg_text: str) -> str:
    payload = base64.b64encode(svg_text.encode("utf-8")).decode("ascii")
    return DATA_URL_PREFIX + payload


def require_string(item: dict[str, object], field: str, index: int) -> str:
    value = item.get(field)
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"manifest symbol {index}: {field} must be a non-empty string")
    return value.strip()


def require_integer(item: dict[str, object], field: str, index: int) -> int:
    value = item.get(field)
    if isinstance(value, bool) or not isinstance(value, int):
        raise ValueError(f"manifest symbol {index}: {field} must be an integer")
    return value


def load_library(root: Path) -> Library:
    manifest_path = root / "manifest.json"
    try:
        document = json.loads(manifest_path.read_text(encoding="utf-8"))
    except FileNotFoundError as error:
        raise ValueError(f"missing manifest: {manifest_path}") from error
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid manifest JSON: {error}") from error
    if not isinstance(document, dict):
        raise ValueError("manifest root must be an object")
    version = document.get("version")
    name = document.get("library")
    rows = document.get("symbols")
    if version != LIBRARY_VERSION:
        raise ValueError(f"manifest version must be {LIBRARY_VERSION}")
    if name != LIBRARY_NAME:
        raise ValueError(f"manifest library must be {LIBRARY_NAME}")
    if not isinstance(rows, list):
        raise ValueError("manifest symbols must be an array")

    symbols: list[Symbol] = []
    keys: set[str] = set()
    file_names: set[str] = set()
    ecos_codes: set[int] = set()
    for index, raw_item in enumerate(rows):
        if not isinstance(raw_item, dict):
            raise ValueError(f"manifest symbol {index}: entry must be an object")
        item: dict[str, object] = raw_item
        key = require_string(item, "key", index)
        label = require_string(item, "label", index)
        category = require_string(item, "category", index)
        description = require_string(item, "description", index)
        file_name = require_string(item, "file", index)
        sort_order = require_integer(item, "sortOrder", index)
        if key in keys:
            raise ValueError(f"manifest symbol {index}: duplicate key {key}")
        if file_name in file_names:
            raise ValueError(f"manifest symbol {index}: duplicate file {file_name}")
        if not SAFE_FILE_NAME.fullmatch(file_name):
            raise ValueError(f"manifest symbol {index}: unsafe file name {file_name}")

        raw_ecos_code = item.get("ecosCode")
        ecos_code: int | None = None
        if raw_ecos_code is not None:
            if isinstance(raw_ecos_code, bool) or not isinstance(raw_ecos_code, int):
                raise ValueError(f"manifest symbol {index}: ecosCode must be an integer")
            if raw_ecos_code <= 0 or raw_ecos_code in ecos_codes:
                raise ValueError(f"manifest symbol {index}: invalid or duplicate ecosCode {raw_ecos_code}")
            ecos_code = raw_ecos_code
            ecos_codes.add(ecos_code)

        svg_path = root / file_name
        try:
            svg_text = svg_path.read_text(encoding="utf-8")
        except FileNotFoundError as error:
            raise ValueError(f"missing SVG: {svg_path}") from error
        validate_svg(svg_text, file_name)
        symbols.append(
            Symbol(
                key=key,
                label=label,
                category=category,
                description=description,
                sort_order=sort_order,
                file_name=file_name,
                svg_text=svg_text,
                ecos_code=ecos_code,
            )
        )
        keys.add(key)
        file_names.add(file_name)

    return Library(
        version=version,
        name=name,
        symbols=tuple(sorted(symbols, key=lambda symbol: (symbol.sort_order, symbol.key))),
    )


def validate_library_contract(library: Library) -> None:
    if len(library.symbols) != EXPECTED_SYMBOL_COUNT:
        raise ValueError(
            f"library must contain {EXPECTED_SYMBOL_COUNT} symbols, got {len(library.symbols)}"
        )
    ecos_count = sum(symbol.ecos_code is not None for symbol in library.symbols)
    if ecos_count != EXPECTED_ECOS_SYMBOL_COUNT:
        raise ValueError(
            f"library must contain {EXPECTED_ECOS_SYMBOL_COUNT} ECoS mappings, got {ecos_count}"
        )


def sql_quote(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def symbol_metadata(symbol: Symbol) -> dict[str, object]:
    active = render_svg(symbol.svg_text, ACTIVE_PALETTE)
    inactive = render_svg(symbol.svg_text, INACTIVE_PALETTE)
    printed = render_svg(symbol.svg_text, PRINT_PALETTE)
    metadata: dict[str, object] = {
        "category": symbol.category,
        "description": symbol.description,
        "library": LIBRARY_NAME,
        "libraryVersion": LIBRARY_VERSION,
        "imageMime": "image/svg+xml",
        "imageData": encode_svg(printed),
        "activeImageData": encode_svg(active),
        "inactiveImageData": encode_svg(inactive),
    }
    if symbol.ecos_code is not None:
        metadata["ecosCode"] = symbol.ecos_code
    return metadata


def build_sql(library: Library) -> str:
    lines = [
        "-- Generated by tools/build_function_symbols.py. Do not edit by hand.",
        "-- RailKeeper Werkstatt-Linie function symbol library.",
        "",
    ]
    for symbol in library.symbols:
        metadata_json = json.dumps(
            symbol_metadata(symbol), ensure_ascii=False, separators=(",", ":")
        )
        lines.extend(
            [
                "INSERT INTO master_data_entries(",
                "  id, type, key, label, active, sort_order, source_url, metadata_json,",
                "  created_at, updated_at, origin",
                ") VALUES(",
                f"  {sql_quote('symbols:' + symbol.key)}, 'symbols', {sql_quote(symbol.key)},",
                f"  {sql_quote(symbol.label)}, 1, {symbol.sort_order}, '', {sql_quote(metadata_json)},",
                "  datetime('now'), datetime('now'), 'bundled'",
                ")",
                "ON CONFLICT(type, key) DO UPDATE SET",
                "  source_url='',",
                "  metadata_json=excluded.metadata_json,",
                "  updated_at=datetime('now'),",
                "  origin='bundled';",
                "",
            ]
        )
    return "\n".join(lines)


def write_atomic(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(content, encoding="utf-8", newline="\n")
    temporary.replace(path)


def icon_data_url(symbol: Symbol, palette: Palette) -> str:
    return encode_svg(render_svg(symbol.svg_text, palette))


def contact_sheet(library: Library) -> str:
    columns = 6
    cell_width = 210
    cell_height = 126
    rows = (len(library.symbols) + columns - 1) // columns
    width = columns * cell_width + 40
    height = rows * cell_height + 100
    items: list[str] = []
    for index, symbol in enumerate(library.symbols):
        column = index % columns
        row = index // columns
        x = 20 + column * cell_width
        y = 70 + row * cell_height
        image_data = icon_data_url(symbol, ACTIVE_PALETTE)
        items.append(
            f'<g transform="translate({x} {y})">'
            '<rect width="198" height="114" rx="8" fill="#111719" stroke="#2c393f"/>'
            f'<image href="{image_data}" x="10" y="10" width="52" height="52"/>'
            f'<image href="{image_data}" x="76" y="14" width="32" height="32"/>'
            f'<image href="{image_data}" x="120" y="18" width="24" height="24"/>'
            f'<image href="{image_data}" x="158" y="21" width="19" height="19"/>'
            f'<text x="10" y="82" fill="#f2f5f6" font-family="Segoe UI,sans-serif" '
            f'font-size="12">{escape(symbol.label)}</text>'
            f'<text x="10" y="101" fill="#879398" font-family="Consolas,monospace" '
            f'font-size="9">{escape(symbol.key)}</text>'
            "</g>"
        )
    return (
        f'<svg xmlns="{SVG_NAMESPACE}" width="{width}" height="{height}" '
        f'viewBox="0 0 {width} {height}">'
        '<rect width="100%" height="100%" fill="#0c1215"/>'
        '<text x="20" y="38" fill="#f2f5f6" font-family="Segoe UI,sans-serif" '
        'font-size="24" font-weight="700">RailKeeper Werkstatt-Linie</text>'
        + "".join(items)
        + "</svg>"
    )


def library_root(value: str | None) -> Path:
    return Path(value or "assets/function-symbols/workshop-line")


def parse_arguments(arguments: Iterable[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Build RailKeeper function symbols")
    parser.add_argument("--root", help="symbol source directory")
    parser.add_argument("--check", action="store_true", help="validate the complete library")
    parser.add_argument("--write-migration", type=Path, help="write generated migration SQL")
    parser.add_argument("--contact-sheet", type=Path, help="write SVG contact sheet")
    options = parser.parse_args(arguments)
    if not (options.check or options.write_migration or options.contact_sheet):
        parser.error("choose --check, --write-migration, or --contact-sheet")
    return options


def main(arguments: Iterable[str] | None = None) -> int:
    options = parse_arguments(arguments)
    library = load_library(library_root(options.root))
    validate_library_contract(library)
    if options.write_migration:
        write_atomic(options.write_migration, build_sql(library))
        print(f"wrote migration for {len(library.symbols)} bundled symbols")
    if options.contact_sheet:
        write_atomic(options.contact_sheet, contact_sheet(library))
        print(f"wrote contact sheet for {len(library.symbols)} bundled symbols")
    if options.check:
        ecos_count = sum(symbol.ecos_code is not None for symbol in library.symbols)
        print(f"validated {len(library.symbols)} symbols ({ecos_count} ECoS mappings, 8 fallbacks)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
