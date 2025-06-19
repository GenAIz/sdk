#!/bin/bash

REGISTRY=$1

if [ -n "$REGISTRY" ]; then
  TAG_PREFIX="$REGISTRY/genaiz-it"
else
  MODULE=$(go list)
  TAG_PREFIX=${MODULE%%/genaiz-it}
fi

FULL=$(./genaiz-it --version)
VERSION=$(echo "${FULL##genaiz-it version}" | tr -d ' ')
cd ..
docker build -t "$TAG_PREFIX/sdk/genaiz-it:$VERSION" -f ./genaiz-it/Dockerfile .
docker tag "$TAG_PREFIX/sdk/genaiz-it:$VERSION" "$TAG_PREFIX/sdk/genaiz-it:latest"
echo "Successfully tagged $TAG_PREFIX/sdk/genaiz-it:latest"

exit 0