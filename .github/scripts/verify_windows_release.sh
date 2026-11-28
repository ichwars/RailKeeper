#!/usr/bin/env bash
set -euo pipefail

directory="${1:?Windows evidence directory required}"
source_ref="${2:?Expected source ref required}"
source_sha="${3:?Expected source commit required}"

fail() {
  echo "::error::$*" >&2
  exit 1
}

shopt -s nullglob
packages=("$directory"/RailKeeper-windows-x64-*.zip)
[[ "${#packages[@]}" == 1 && -s "${packages[0]}" ]] || fail "Exactly one nonempty Windows x64 ZIP is required."
package="${packages[0]}"
name="${package##*/}"
if [[ "$source_ref" == refs/tags/* && "$name" != "RailKeeper-windows-x64-${source_ref#refs/tags/}.zip" ]]; then
  fail "Windows package filename does not match the release tag."
fi

digest="$(sha256sum -- "$package")"
printf '%s  %s\n' "${digest%% *}" "$name" |
  cmp -s -- - "$directory/checksums.txt" || fail "Windows ZIP does not match checksums.txt."
[[ -s "$directory/windows-provenance.jsonl" ]] || fail "Windows build attestation is missing."

gh attestation verify "$package" --bundle "$directory/windows-provenance.jsonl" \
  --repo "$GITHUB_REPOSITORY" \
  --signer-workflow "$GITHUB_REPOSITORY/.github/workflows/windows-standalone.yml" \
  --source-ref "$source_ref" --source-digest "$source_sha" --deny-self-hosted-runners
