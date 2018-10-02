#!/usr/bin/env bash

set -eo pipefail

#
# Builds all docker images. Usage docker.sh [<repo>] [<tag>]
#
BUILD_ROOT=$(dirname $0)
IMAGE_REPO=${1:-fission}
IMAGE_TAG=${2:-latest}
NOBUILD=${3:-false}

# Build bundle images
bundleImage=${IMAGE_REPO}/fission-workflows-bundle
pushd ${BUILD_ROOT}/..
if $NOBUILD ; then
    echo "Using pre-build binaries..."
    if [ ! -f ./fission-workflows-bundle ]; then
        echo "Executable './fission-workflows-bundle' not found!"
        exit 1;
    fi

    if [ ! -f ./fission-workflows ]; then
        echo "Executable './fission-workflows' not found!"
        exit 1;
    fi
fi

echo "Building bundle..."
docker build --tag="${bundleImage}:${IMAGE_TAG}" -f ${BUILD_ROOT}/Dockerfile \
    --no-cache \
    --build-arg NOBUILD="${NOBUILD}" .
popd

# Build bundle-dependent images
echo "Building Fission runtime env..."
docker build --tag="${IMAGE_REPO}/workflow-env:${IMAGE_TAG}" ${BUILD_ROOT}/runtime-env/ \
    --no-cache \
    --build-arg BUNDLE_IMAGE=${bundleImage} \
    --build-arg BUNDLE_TAG=${IMAGE_TAG}
echo "Building Fission build env..."
docker build --tag="${IMAGE_REPO}/workflow-build-env:${IMAGE_TAG}" ${BUILD_ROOT}/build-env/ \
    --no-cache \
    --build-arg BUNDLE_IMAGE=${bundleImage} \
    --build-arg BUNDLE_TAG=${IMAGE_TAG}
echo "Building fission-workflows..."
docker build --tag="${IMAGE_REPO}/fission-workflows-cli:${IMAGE_TAG}" ${BUILD_ROOT}/cli/ \
    --no-cache \
    --build-arg BUNDLE_IMAGE=${bundleImage} \
    --build-arg BUNDLE_TAG=${IMAGE_TAG}

# Remove intermediate images
# docker rmi $(docker images -f "dangling=true" -q)
