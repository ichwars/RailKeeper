"""Check fixed-scope, cross-package Go statement coverage without double-counting blocks."""
import argparse
import json
from pathlib import Path


def summarize(profile, scope):
    lines = profile.splitlines()
    if not lines or lines[0] not in ("mode: set", "mode: count", "mode: atomic"):
        raise ValueError("invalid Go coverage mode")
    blocks = {}
    for line in lines[1:]:
        block, statements, count = line.rsplit(" ", 2)
        statements, count = int(statements), int(count)
        file = block.rsplit(":", 1)[0]
        if not file.startswith(scope) or statements < 0 or count < 0:
            raise ValueError("coverage block is outside the agreed scope or has invalid counts")
        if block in blocks and blocks[block][0] != statements:
            raise ValueError("inconsistent statement count for the same block")
        blocks[block] = (statements, bool(count) or blocks.get(block, (0, False))[1])
    modules = {}
    for block, (statements, covered) in blocks.items():
        module = block.rsplit(":", 1)[0][len(scope):].rsplit("/", 1)[0]
        value = modules.setdefault(module, {"statements": 0, "covered": 0})
        value["statements"] += statements
        value["covered"] += statements if covered else 0
    total = {"statements": sum(v["statements"] for v in modules.values()),
             "covered": sum(v["covered"] for v in modules.values())}
    for item in [total, *modules.values()]:
        item["percent"] = 100 * item["covered"] / item["statements"] if item["statements"] else 0
    return {"scope": scope, "total": total, "modules": modules}


def failures(report, baseline):
    problems = []
    if report["total"]["percent"] < baseline["total"]:
        problems.append(f"total coverage is below {baseline['total']}%")
    for module, minimum in baseline["modules"].items():
        value = report["modules"].get(module)
        if value is None:
            problems.append(f"required module {module} is missing from coverage")
        elif value["percent"] < minimum:
            problems.append(f"{module}: {value['percent']:.2f}% is below {minimum}%")
    return problems


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--profile", type=Path, required=True)
    parser.add_argument("--baseline", type=Path, default=Path(__file__).with_name("go-coverage-baseline.json"))
    parser.add_argument("--report", type=Path, required=True)
    args = parser.parse_args()
    baseline = json.loads(args.baseline.read_text())
    report = summarize(args.profile.read_text(), baseline["scope"])
    problems = failures(report, baseline)
    args.report.write_text(json.dumps({**report, "passed": not problems, "failures": problems}, indent=2) + "\n")
    print(f"Backend statement coverage: {report['total']['percent']:.2f}% across {len(report['modules'])} modules")
    if problems:
        raise SystemExit("\n".join(problems))


if __name__ == "__main__":
    main()
