#!/bin/bash

REGISTRY=$1

if [ -n "$REGISTRY" ]; then
  TAG_PREFIX="$REGISTRY/genaiz-oauth"
else
  MODULE=$(go list)
  TAG_PREFIX=${MODULE%%/genaiz-oauth}
fi

FULL=$(./genaiz-oauth --version)
VERSION=$(echo "${FULL##genaiz-oauth version}" | tr -d ' ')
cd ..
docker build -t "$TAG_PREFIX/sdk/genaiz-oauth:$VERSION" -f ./genaiz-oauth/Dockerfile .
docker tag "$TAG_PREFIX/sdk/genaiz-oauth:$VERSION" "$TAG_PREFIX/sdk/genaiz-oauth:latest"
echo "Successfully tagged $TAG_PREFIX/sdk/genaiz-oauth:latest"

exit 0