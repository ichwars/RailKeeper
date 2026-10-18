#!/usr/bin/env bash
set -euo pipefail

# v1.7.0, verified against the release asset digest. No remote installer execution.
destination="${1:?Usage: install_strix.sh DESTINATION}"
mkdir -p "$destination"
archive="$(mktemp)"
trap 'rm -f "$archive"' EXIT
curl --fail --location --retry 3 --output "$archive" \
  https://github.com/usestrix/strix/releases/download/v1.7.0/strix-1.7.0-linux-x86_64.tar.gz
printf '%s  %s\n' b8cd7e84d63cbeff6726301a9e08fddedf40fc6a4edcdb4edda8a853b3312de6 "$archive" | sha256sum --check
tar -xzf "$archive" -C "$destination"
mv "$destination/strix-1.7.0-linux-x86_64" "$destination/strix"
"$destination/strix" --version
