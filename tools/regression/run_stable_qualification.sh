#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 || ! ${1-} =~ ^v[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
    echo "Usage: bash tools/regression/run_stable_qualification.sh vX.Y.Z NEW_OUTPUT_DIRECTORY" >&2
    exit 2
fi
source_tag=$1
candidate=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
if [[ -n $(git -C "$candidate" status --porcelain --untracked-files=no) ]]; then
    echo "Stable qualification requires a clean candidate commit" >&2
    exit 2
fi
candidate_commit=$(git -C "$candidate" rev-parse HEAD)

mkdir -- "$2"
output=$(cd -- "$2" && pwd)
git clone --quiet --single-branch --depth 1 --branch "$source_tag" \
    https://github.com/ichwars/RailKeeper.git "$output/source"
source_commit=$(git -C "$output/source" rev-parse "refs/tags/$source_tag^{commit}")
test "$(git -C "$output/source" rev-parse HEAD)" = "$source_commit"

build() {
    local checkout=$1 commit=$2 kind=$3 binary=$4
    local go_version
    go_version=$(cat "$checkout/.go-version")
    GOTOOLCHAIN="go$go_version" CGO_ENABLED=0 go -C "$checkout/backend" build \
        -mod=readonly -buildvcs=false -trimpath \
        -ldflags "-X railkeeper/backend/internal/buildinfo.commit=$commit -X railkeeper/backend/internal/buildinfo.kind=$kind -X railkeeper/backend/internal/buildinfo.modified=false" \
        -o "$binary" ./cmd/railkeeper
}

build "$output/source" "$source_commit" release "$output/source-railkeeper"
build "$candidate" "$candidate_commit" development "$output/candidate-railkeeper"

# This qualification exercises the HTTP API. The frontend has its own CI jobs.
mkdir -- "$output/static"
python3 "$candidate/tools/regression/run_baseline.py" \
    --previous-binary "$output/source-railkeeper" --previous-backend "$output/source/backend" \
    --source-tag "$source_tag" --expected-source-commit "$source_commit" \
    --binary "$output/candidate-railkeeper" --expected-commit "$candidate_commit" \
    --static-dir "$output/static" --output "$output/qualification" \
    --catalog-backfill --catalog-authority --decoder-model --decoder-backfill
