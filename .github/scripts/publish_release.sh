#!/usr/bin/env bash
set -euo pipefail

repo="$GITHUB_REPOSITORY"

fail() {
  echo "::error::$*" >&2
  exit 1
}

current_main() {
  gh api "repos/$repo/git/ref/heads/main" --jq '.object.sha'
}

main_sha="$(current_main)"
if [[ "$RELEASE_SHA" != "$main_sha" ]]; then
  echo "::notice::Skipping a superseded main commit."
  exit 0
fi

ci_result="$(gh api --method GET "repos/$repo/actions/workflows/ci.yml/runs" \
  -f branch=main -f event=push -f head_sha="$RELEASE_SHA" \
  --jq '.workflow_runs[0].conclusion')"
[[ "$ci_result" == success ]] || fail "The current main commit has no successful CI run."

version="$(sed -nE 's/^[[:space:]]*version[[:space:]]*=[[:space:]]*"([^"]+)".*/\1/p' \
  backend/cmd/railkeeper/main.go)"
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]] ||
  fail "Invalid runtime release version."
tag="v$version"
notes="docs/releases/$tag.md"

[[ -f "$notes" && "$(head -n 1 "$notes")" == "# RailKeeper $tag" ]] ||
  fail "Matching release notes are required."
grep -Fxq '## Deutsch' "$notes" && grep -Fxq '## English' "$notes" ||
  fail "German and English release notes are required."

releases="$(gh api "repos/$repo/releases?per_page=100" --paginate --jq '.[] | [.tag_name, .draft] | @tsv')"
if grep -Fxq -- "$tag"$'\tfalse' <<< "$releases"; then
  echo "::notice::Release $tag already exists."
  exit 0
fi
if grep -Fxq -- "$tag"$'\ttrue' <<< "$releases"; then
  fail "A draft for $tag already exists. Inspect it before retrying publication."
fi

tag_ref="refs/tags/$tag"
tag_refs="$(gh api "repos/$repo/git/matching-refs/tags/$tag" --jq '.[].ref')"
if ! grep -Fxq -- "$tag_ref" <<< "$tag_refs"; then
  # GITHUB_TOKEN tag writes do not emit push workflows. Dispatch the build explicitly.
  gh api --method POST "repos/$repo/git/refs" -f ref="$tag_ref" -f sha="$RELEASE_SHA" >/dev/null
fi

check_tag() {
  local tagged_sha
  tagged_sha="$(gh api "repos/$repo/commits/$tag" --jq '.sha')"
  [[ "$tagged_sha" == "$RELEASE_SHA" ]] || fail "The release tag points to another commit."
}
check_tag

run_id="$(gh api --method POST "repos/$repo/actions/workflows/windows-standalone.yml/dispatches" \
  -H 'X-GitHub-Api-Version: 2026-03-10' -f ref="$tag" --jq '.workflow_run_id')"
[[ "$run_id" =~ ^[0-9]+$ ]] || fail "Windows workflow dispatch did not return a run ID."
echo "Windows package validation: https://github.com/$repo/actions/runs/$run_id"

# Allow the native Windows qualification and fail closed on cancellation or timeout.
deadline=$((SECONDS + 55 * 60))
while true; do
  run="$(gh api "repos/$repo/actions/runs/$run_id" \
    --jq '[.event, .head_branch, .head_sha, .status, (.conclusion // "pending")] | @tsv')"
  IFS=$'\t' read -r event ref commit status conclusion <<< "$run"
  [[ "$event" == workflow_dispatch && "$ref" == "$tag" && "$commit" == "$RELEASE_SHA" ]] ||
    fail "Windows workflow did not run against $tag at $RELEASE_SHA."
  if [[ "$status" == completed ]]; then
    [[ "$conclusion" == success ]] || fail "Windows workflow finished with $conclusion."
    break
  fi
  ((SECONDS < deadline)) || fail "Timed out waiting for the verified Windows package."
  sleep 30
done

evidence_dir="$(mktemp -d)"
trap 'rm -rf -- "$evidence_dir"' EXIT
gh run download "$run_id" --repo "$repo" --name railkeeper-windows-verified --dir "$evidence_dir"
bash .github/scripts/verify_windows_release.sh "$evidence_dir" "$tag_ref" "$RELEASE_SHA"

main_sha="$(current_main)"
if [[ "$RELEASE_SHA" != "$main_sha" ]]; then
  echo "::notice::Skipping a superseded main commit."
  exit 0
fi
check_tag

release_flags=(--latest)
if [[ "$version" == *-* ]]; then
  release_flags=(--prerelease --latest=false)
fi

# gh stages the attachments in a draft before publishing it.
gh release create "$tag" "$evidence_dir/RailKeeper-windows-x64-$tag.zip" "$evidence_dir/checksums.txt" \
  --repo "$repo" --verify-tag --title "RailKeeper $tag" --notes-file "$notes" "${release_flags[@]}"
gh workflow run docker-image.yml --repo "$repo" --ref "$tag"
