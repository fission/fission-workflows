#!/usr/bin/env bash

set -ex

#
# Builds all docker images
#
IMAGE_VERSION=0.1.2
IMAGE_ORG=fission

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
rm -f bundle/fission-workflows-bundle
rm -f env/fission-workflows-bundle
rm -f build-env/wfcli
chmod +x fission-workflows-bundle
chmod +x wfcli
yes | cp fission-workflows-bundle env/
yes | cp fission-workflows-bundle bundle/
yes | cp wfcli build-env/

# Build images
docker build --tag="${IMAGE_ORG}/fission-workflows-bundle:${IMAGE_VERSION}" bundle/
docker build --tag="${IMAGE_ORG}/workflow-env:${IMAGE_VERSION}" env/
docker build --tag="${IMAGE_ORG}/workflow-build-env:${IMAGE_VERSION}" build-env/
