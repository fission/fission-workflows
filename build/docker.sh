#!/usr/bin/env bash

set -ex

#
# Builds all docker images. Usage docker.sh [<repo>] [<tag>]
#
BUILD_ROOT=$(dirname $0)

IMAGE_REPO=$1
if [ -z "$IMAGE_REPO" ]; then
    IMAGE_TAG=fission
fi

IMAGE_TAG=$2
if [ -z "$IMAGE_TAG" ]; then
    IMAGE_TAG=latest
fi

# Check preconditions
if [ ! -f ./fission-workflows-bundle ]; then
    echo "Executable './fission-workflows-bundle' not found!"
    exit 1;
fi

if [ ! -f ./wfcli ]; then
    echo "Executable './wfcli' not found!"
    exit 1;
fi

# Prepare fs
rm -f ${BUILD_ROOT}/bundle/fission-workflows-bundle
rm -f ${BUILD_ROOT}/env/fission-workflows-bundle
rm -f ${BUILD_ROOT}/build-env/wfcli
chmod +x fission-workflows-bundle
chmod +x wfcli
yes | cp fission-workflows-bundle ${BUILD_ROOT}/env/
yes | cp fission-workflows-bundle ${BUILD_ROOT}/bundle/
yes | cp wfcli ${BUILD_ROOT}/build-env/

# Build images
docker build --tag="${IMAGE_REPO}/fission-workflows-bundle:${IMAGE_TAG}" ${BUILD_ROOT}/bundle/
docker build --tag="${IMAGE_REPO}/workflow-env:${IMAGE_TAG}" ${BUILD_ROOT}/env/
docker build --tag="${IMAGE_REPO}/workflow-build-env:${IMAGE_TAG}" ${BUILD_ROOT}/build-env/
