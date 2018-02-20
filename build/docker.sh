#!/usr/bin/env bash

set -e

#
# Builds all docker images. Usage docker.sh [<repo>] [<tag>]
#
BUILD_ROOT=$(dirname $0)

IMAGE_REPO=$1
if [ -z "$IMAGE_REPO" ]; then
    IMAGE_REPO=fission
fi

IMAGE_TAG=$2
if [ -z "$IMAGE_TAG" ]; then
    IMAGE_TAG=latest
fi

# Build bundle images
bundleImage=${IMAGE_REPO}/fission-workflows-bundle
pushd ${BUILD_ROOT}/..
if [ ! -z "$NOBUILD" ]; then
    if [ ! -f ./fission-workflows-bundle ]; then
        echo "Executable './fission-workflows-bundle' not found!"
        exit 1;
    fi

    if [ ! -f ./wfcli ]; then
        echo "Executable './wfcli' not found!"
        exit 1;
    fi
fi
echo "Building bundle..."
docker build --tag="${bundleImage}:${IMAGE_TAG}" -f ${BUILD_ROOT}/Dockerfile \
    --build-arg NOBUILD="${NOBUILD}" .
popd

# Build bundle-dependent images
echo "Building Fission runtime env..."
docker build --tag="${IMAGE_REPO}/workflow-env:${IMAGE_TAG}" ${BUILD_ROOT}/runtime-env/ \
    --build-arg BUNDLE_IMAGE=${bundleImage} \
    --build-arg BUNDLE_TAG=${IMAGE_TAG}
echo "Building Fission build env..."
docker build --tag="${IMAGE_REPO}/workflow-build-env:${IMAGE_TAG}" ${BUILD_ROOT}/build-env/ \
    --build-arg BUNDLE_IMAGE=${bundleImage} \
    --build-arg BUNDLE_TAG=${IMAGE_TAG}
echo "Building wfcli..."
docker build --tag="${IMAGE_REPO}/wfcli:${IMAGE_TAG}" ${BUILD_ROOT}/wfcli/ \
    --build-arg BUNDLE_IMAGE=${bundleImage} \
    --build-arg BUNDLE_TAG=${IMAGE_TAG}

# Remove intermediate images
# docker rmi $(docker images -f "dangling=true" -q)