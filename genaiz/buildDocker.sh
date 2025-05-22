#!/bin/bash

REGISTRY=$1

if [ -n "$REGISTRY" ]; then
  TAG_PREFIX="$REGISTRY/genaiz"
else
  MODULE=$(go list)
  TAG_PREFIX=${MODULE%%/genaiz}
fi

FULL=$(./genaiz --version)
VERSION=$(echo "${FULL##genaiz version}" | tr -d ' ')
docker build -t "$TAG_PREFIX/sdk/genaiz:$VERSION" .
docker tag "$TAG_PREFIX/sdk/genaiz:$VERSION" "$TAG_PREFIX/sdk/genaiz:latest"
echo "Successfully tagged $TAG_PREFIX/sdk/genaiz:latest"

exit 0