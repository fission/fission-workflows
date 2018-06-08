#!/usr/bin/env bash

set -e

#
# release.sh - Generate all artifacts for a release
#

# fission-workflows
echo "Building linux binaries..."
build/build-linux.sh
mv fission-workflows-bundle fission-workflows-bundle-linux
mv fission-workflows fission-workflows-linux
echo "Building windows binaries..."
build/build-windows.sh
echo "Building osx binaries..."
build/build-osx.sh

# Deployments
echo "Packaging chart..."
helm package ${GOPATH}/src/github.com/fission/fission-workflows/charts/fission-workflows

echo "Done!"
