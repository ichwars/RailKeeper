"""Keep executable Go and frontend modules below 500 physical lines."""
import argparse
import json
import re
from pathlib import Path

LIMIT = 499
DICTIONARIES = {"frontend/src/shared/i18n/de.ts", "frontend/src/shared/i18n/en.ts"}
STRING = r'"(?:[^"\\]|\\.)*"'
DICTIONARY = re.compile(
    r"\s*export const (de|en)Translations: Record<string, string> = \{\s*"
    + rf"(?:{STRING}\s*:\s*{STRING}\s*,?\s*)*" + r"\};\s*\Z"
)


def check(root):
    files = sorted((root / "backend").rglob("*.go"))
    files += sorted((root / "frontend/src").rglob("*.ts"))
    files += sorted((root / "frontend/src").rglob("*.tsx"))
    violations, measured, dictionaries = [], [], []
    for file in files:
        relative = file.relative_to(root).as_posix()
        if file.name.endswith(("_test.go", ".test.ts", ".test.tsx", ".d.ts")) or relative.startswith("frontend/src/test/"):
            continue
        source = file.read_text(encoding="utf-8")
        lines = len(source.splitlines())
        if relative in DICTIONARIES:
            if not DICTIONARY.fullmatch(source):
                violations.append(f"{relative}: dictionary exception must contain only literal translations")
            dictionaries.append(relative)
            continue
        measured.append({"path": relative, "lines": lines})
        if lines > LIMIT:
            violations.append(f"{relative}: {lines} lines exceed {LIMIT}")
    return {"limit": LIMIT, "files": measured, "literalDictionaries": dictionaries, "violations": violations}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--report", type=Path)
    args = parser.parse_args()
    result = check(args.root)
    if args.report:
        args.report.write_text(json.dumps(result, indent=2) + "\n")
    if result["violations"]:
        raise SystemExit("\n".join(result["violations"]))
    print(f"{len(result['files'])} executable modules below 500 lines; two literal translation dictionaries checked")


if __name__ == "__main__":
    main()
