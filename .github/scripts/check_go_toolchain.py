"""Fail when release compilers drift; go.mod separately declares the supported minimum."""
import re
from pathlib import Path

root = Path(__file__).resolve().parents[2]
release = (root / ".go-version").read_text().strip()
if not re.fullmatch(r"\d+\.\d+\.\d+", release):
    raise SystemExit(".go-version must pin a release compiler")
minimum = re.search(r"^go (\S+)$", (root / "backend/go.mod").read_text(), re.M)
if minimum is None or tuple(map(int, minimum[1].split("."))) > tuple(map(int, release.split("."))):
    raise SystemExit("release compiler is older than the declared minimum")
docker = (root / "Dockerfile").read_text()
if f"golang:{release}-alpine AS backend-build" not in docker:
    raise SystemExit("Docker release compiler differs from .go-version")
for name in ("ci.yml", "windows-standalone.yml", "codeql.yml"):
    workflow = (root / ".github/workflows" / name).read_text()
    if "go-version-file: .go-version" not in workflow:
        raise SystemExit(f"{name} must use the pinned release compiler")
if "go-version-file: backend/go.mod" not in (root / ".github/workflows/ci.yml").read_text():
    raise SystemExit("the declared minimum compiler must retain its own CI test")
print(f"Release Go {release}; independently tested minimum Go {minimum[1]}")
