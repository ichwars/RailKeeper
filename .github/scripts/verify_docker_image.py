"""Check the assembled GHCR index against the native image tests and attestations."""

import argparse
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import urlsplit

ARCHITECTURES = ("amd64", "arm64")
SPDX = "https://spdx.dev/Document"
SLSA = "https://slsa.dev/provenance/v1"


def require(ok, message):
    if not ok:
        raise ValueError(message)


def digest(value):
    require(isinstance(value, str) and re.fullmatch(r"sha256:[0-9a-f]{64}", value),
            f"Invalid OCI digest: {value}")
    return value


def inspect(reference, *options):
    return json.loads(subprocess.check_output(
        ["docker", "buildx", "imagetools", "inspect", reference, *options], text=True,
    ))


def platform_documents(documents):
    result = {}
    for key, value in documents.items():
        platform = "/".join(key.split("/")[:2])
        require(platform not in result, f"Multiple documents for {platform}")
        result[platform] = value
    return result


def source_identity(uri):
    source = urlsplit(uri.removeprefix("git+"))
    return source.scheme, source.netloc.casefold(), source.path.removesuffix(".git").casefold()


def verify(index, qualifications, sbom, provenance, read_manifest, repository, commit):
    """Return the evidence for the exact child manifests that passed the native tests."""
    platforms = {}
    attestations = {}
    for descriptor in index["manifests"]:
        annotations = descriptor.get("annotations", {})
        if annotations.get("vnd.docker.reference.type") == "attestation-manifest":
            subject = digest(annotations.get("vnd.docker.reference.digest"))
            attestations.setdefault(subject, []).append(digest(descriptor["digest"]))
            continue
        platform = descriptor["platform"]
        arch = platform["architecture"]
        require(platform["os"] == "linux" and arch in ARCHITECTURES and arch not in platforms,
                f"Unexpected image platform: {platform}")
        platforms[arch] = digest(descriptor["digest"])

    expected = set(ARCHITECTURES)
    require(platforms.keys() == expected and qualifications.keys() == expected,
            "Expected exactly the amd64 and arm64 images and test reports")
    require(attestations.keys() == set(platforms.values()),
            "Attestation references do not match the runnable platform manifests")

    sbom = platform_documents(sbom)
    provenance = platform_documents(provenance)
    report = {}
    for arch, image_digest in platforms.items():
        platform = f"linux/{arch}"
        tested = qualifications[arch]
        require(tested["platformDigest"] == image_digest, f"{platform} is not the tested manifest")
        require(tested["build"]["platform"] == platform and tested["build"]["commit"] == commit,
                f"{platform} qualification identity does not match the build")
        types = set()
        for attestation_digest in attestations[image_digest]:
            for layer in read_manifest(attestation_digest)["layers"]:
                if layer["mediaType"] == "application/vnd.in-toto+json":
                    types.add(layer.get("annotations", {}).get("in-toto.io/predicate-type"))
        require({SPDX, SLSA} <= types, f"{platform} lacks SBOM or SLSA provenance")
        spdx = sbom.get(platform, {}).get("SPDX", {})
        require(spdx.get("spdxVersion", "").startswith("SPDX-") and spdx.get("packages"),
                f"{platform} lacks a readable SPDX package inventory")
        slsa = provenance.get(platform, {}).get("SLSA", {})
        materials = slsa.get("buildDefinition", {}).get("resolvedDependencies", [])
        require(any(source_identity(material.get("uri", "")) == source_identity(repository)
                    and commit in material.get("digest", {}).values() for material in materials),
                f"{platform} provenance does not name the expected Git source and commit")
        report[platform] = {
            "digest": image_digest,
            "build": tested["build"],
            "buildIndexDigest": digest(tested["indexDigest"]),
            "attestationDigests": attestations[image_digest],
            "predicateTypes": [SPDX, SLSA],
            "sbomPackages": len(spdx["packages"]),
        }
    return report


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("image")
    parser.add_argument("index_digest", type=digest)
    parser.add_argument("qualification", type=Path)
    parser.add_argument("repository")
    parser.add_argument("commit")
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    reference = f"{args.image}@{args.index_digest}"
    index = inspect(reference, "--raw")
    sbom = inspect(reference, "--format", "{{json .SBOM}}")
    provenance = inspect(reference, "--format", "{{json .Provenance}}")
    qualifications = {
        arch: json.loads((args.qualification / f"{arch}.json").read_text())
        for arch in ARCHITECTURES
    }
    report = verify(index, qualifications, sbom, provenance,
                    lambda child: inspect(f"{args.image}@{child}", "--raw"),
                    args.repository, args.commit)
    args.output.mkdir(parents=True, exist_ok=True)
    evidence = {
        "index.json": index,
        "sbom.json": sbom,
        "provenance.json": provenance,
        "report.json": {"image": reference, "commit": args.commit, "platforms": report},
    }
    for filename, document in evidence.items():
        (args.output / filename).write_text(json.dumps(document, indent=2) + "\n")
    print(json.dumps(evidence["report.json"], indent=2))


if __name__ == "__main__":
    main()
