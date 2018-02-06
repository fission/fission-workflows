#!/usr/bin/env bash

# Run e2e tests using minikube

set -euo pipefail

. $(dirname $0)/utils.sh
ROOT=$(dirname $0)/../..

#
# Env setup - Minikube
# TODO use test-specific minikube instance?
emph "Setting up Minikube..."
. ${ROOT}/hack/minikube.sh

# Build and run tests
. ${ROOT}/test/e2e/buildtest.sh

# TODO optionally cleanup minikube instance