"""Require the same pinned Go compiler for local, CI, native and container builds."""
import re
from pathlib import Path

def check(root):
    release = (root / ".go-version").read_text().strip()
    if not re.fullmatch(r"\d+\.\d+\.\d+", release):
        raise SystemExit(".go-version must pin a release compiler")
    minimum = re.search(r"^go (\S+)$", (root / "backend/go.mod").read_text(), re.M)
    if minimum is None or minimum[1] != release:
        raise SystemExit("go.mod and .go-version must declare the same Go compiler")
    docker = (root / "Dockerfile").read_text()
    if f"golang:{release}-alpine AS backend-build" not in docker:
        raise SystemExit("Docker release compiler differs from .go-version")
    for name in ("ci.yml", "windows-standalone.yml", "codeql.yml"):
        workflow = (root / ".github/workflows" / name).read_text()
        if "go-version-file: .go-version" not in workflow:
            raise SystemExit(f"{name} must use the pinned release compiler")
    if "go-version-file: backend/go.mod" not in (root / ".github/workflows/ci.yml").read_text():
        raise SystemExit("the declared minimum compiler must retain its own CI test")
    lint_workflow = (root / ".github/workflows/ci.yml").read_text()
    if "version: v2.14.0" not in lint_workflow:
        raise SystemExit("Go 1.27 builds require the qualified golangci-lint v2.14.0 analyzer")
    return f"Go {release} is identical for module, local, CI, Docker, native builds and lint"


if __name__ == "__main__":
    print(check(Path(__file__).resolve().parents[2]))
