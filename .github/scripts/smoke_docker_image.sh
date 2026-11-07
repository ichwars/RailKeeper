#!/usr/bin/env bash
set -euo pipefail

image="${1:?Usage: smoke_docker_image.sh IMAGE ARCH}"
expected_arch="${2:?Usage: smoke_docker_image.sh IMAGE ARCH}"
actual_arch="$(docker image inspect "$image" --format '{{.Architecture}}')"
if [[ "$actual_arch" != "$expected_arch" ]]; then
  echo "Expected $expected_arch image, got $actual_arch" >&2
  exit 1
fi

container=''
cleanup() {
  result=$?
  if [[ -n "$container" ]]; then
    if [[ "$result" != 0 ]]; then
      docker logs "$container" >&2 || true
    fi
    docker rm --force --volumes "$container" >/dev/null
  fi
  exit "$result"
}
trap cleanup EXIT

# An anonymous /data volume exercises first-run SQLite migration and seed loading.
container="$(docker run --detach --network none --read-only \
  --cap-drop ALL --security-opt no-new-privileges:true --tmpfs /tmp "$image")"

ready=false
for ((attempt = 0; attempt < 60; attempt++)); do
  if docker exec "$container" railkeeper healthcheck; then
    ready=true
    break
  fi
  if [[ "$(docker inspect "$container" --format '{{.State.Running}}')" != true ]]; then
    echo 'Container stopped before becoming ready' >&2
    exit 1
  fi
  sleep 1
done
if [[ "$ready" != true ]]; then
  echo 'Container did not become ready within 60 seconds' >&2
  exit 1
fi

identity="$(docker exec "$container" railkeeper version)"
expected_go="go$(tr -d '\r\n' < .go-version)"
jq -e --arg platform "linux/$expected_arch" --arg compiler "$expected_go" --arg commit "${GITHUB_SHA:-}" '
  .build.platform == $platform and .build.goVersion == $compiler and
  ($commit == "" or (.build.commit == $commit and (.build.status == "release" or .build.status == "development")))
' <<< "$identity"
mkdir -p output
image_id="$(docker image inspect "$image" --format '{{.Id}}')"
jq --arg imageID "$image_id" '. + {imageID: $imageID}' <<< "$identity" > "output/docker-qualification-$expected_arch.json"

docker exec "$container" wget -q -O - http://127.0.0.1:8080/api/v1/setup/status |
  jq -e '.setupRequired == true'
docker exec "$container" wget -q -O - http://127.0.0.1:8080/ |
  grep -q '<div id="root"></div>'
echo "Docker smoke test passed for linux/$expected_arch"
