#!/usr/bin/env bash
# Build and push a multi-architecture Docker image (linux/amd64 + linux/arm64).
#
# Usage:
#   ./build.sh                          # build for amd64 + arm64, push to registry
#   IMAGE=myuser/chinasclmdav:v1.0 ./build.sh
#   PUSH=0 ./build.sh                   # build locally without pushing
#   PLATFORMS=linux/amd64 ./build.sh    # single platform
#
# Requires: docker with buildx plugin (docker buildx version).

set -euo pipefail

IMAGE="${IMAGE:-chinasclmdav:latest}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
PUSH="${PUSH:-1}"

echo "==> Building image: ${IMAGE}"
echo "==> Platforms:     ${PLATFORMS}"

# Create a dedicated builder (re-uses existing one if present).
BUILDER="chinasclmdav-builder"
if ! docker buildx inspect "$BUILDER" >/dev/null 2>&1; then
  docker buildx create --name "$BUILDER" --use
fi
docker buildx use "$BUILDER"

ARGS=(build --platform "$PLATFORMS" -t "$IMAGE" --pull --no-cache)

if [ "$PUSH" = "1" ]; then
  ARGS+=(--push)
else
  # --load only works for a single platform.
  if [ "$PLATFORMS" = "linux/amd64" ] || [ "$PLATFORMS" = "linux/arm64" ]; then
    ARGS+=(--load)
  fi
fi

docker buildx "${ARGS[@]}" .

echo "==> Done. Run with:"
echo "    docker run -d --name chinasclmdav -p 8080:8080 -v chinasclmdav-data:/data \\"
echo "      -e CHINASCLMDAV_SEED_USER=vipiu -e CHINASCLMDAV_SEED_EMAIL=vipiu@qq.com -e CHINASCLMDAV_SEED_PASS='your-password' \\"
echo "      ${IMAGE}"
