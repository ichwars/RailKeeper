"""Run the performance regression checks against a synthetic local installation."""

import argparse
import json
import subprocess
import tempfile
from pathlib import Path

from client import setup
from fixture import sample_png
from runtime import Runtime


VEHICLES_WITH_IMAGES = 120
HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]


def seed_inventory(client):
    common = {"manufacturer": "Synthetic", "gauge": "TT"}
    image = sample_png()
    for number in range(1, VEHICLES_WITH_IMAGES + 1):
        vehicle = client.request("POST", "/api/v1/vehicles", {
            **common, "inventoryNumber": f"PERF-{number:04d}", "name": f"Performance vehicle {number:04d}",
            "category": "Lokomotive", "gattung": "Diesellokomotive",
        }, 201)
        client.upload(f"/api/v1/vehicles/{vehicle['id']}/images", "performance.png", image, "image/png",
                      {"isPrimary": "true"})
    client.request("POST", "/api/v1/vehicle-sets", {
        "set": {**common, "name": "Performance set", "category": "Wagen", "gattung": "Reisezugwagen"},
        "members": [{"name": "Performance set A"}, {"name": "Performance set B"}],
    }, 201)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", required=True, type=Path)
    parser.add_argument("--static-dir", type=Path, default=ROOT / "frontend/dist")
    args = parser.parse_args()
    if not args.binary.is_file() or not (args.static_dir / "index.html").is_file():
        parser.error("build the RailKeeper binary and frontend before running the browser checks")

    with tempfile.TemporaryDirectory(prefix="railkeeper-performance-") as temp:
        root = Path(temp)
        with Runtime(args.binary, ROOT / "backend", args.static_dir, root / "data", root / "logs") as runtime:
            credentials = setup(runtime.client, "performance-fixture")
            seed_inventory(runtime.client)
            subprocess.run(["node", str(HERE / "performance.mjs")], check=True, text=True, timeout=120,
                           input=json.dumps({"base": runtime.client.base, "credentials": credentials,
                                             "imageCount": VEHICLES_WITH_IMAGES}))


if __name__ == "__main__":
    main()
