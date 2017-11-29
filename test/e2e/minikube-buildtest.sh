#!/usr/bin/env bash

# Run e2e tests using minikube
set -euo pipefail

source $(dirname $0)/utils.sh

REPO=minikube-ci
WORKFLOWS_ENV_IMAGE=${REPO}/workflow-env
WORKFLOWS_BUILD_ENV_IMAGE=${REPO}/workflow-build-env
WORKFLOWS_BUNDLE_IMAGE=${REPO}/fission-workflows-bundle
ROOT=$(dirname $0)/../..
id=$(generate_test_id)
NS=fission
fissionHelmId=fission-test
fissionWorkflowsHelmId=fission-workflows-test
TAG=test

#
# Build
#
prepare-env() {
    # Setup minikube
    bash $ROOT/hack/deploy.sh --nofission
    # TODO check if minikube is empty (and use profiles)
    # TODO use test-specific minikube instance

    eval $(minikube docker-env)
}
# Build binaries
build() {
    bash ${ROOT}/build/build-linux.sh

    # Build docker images
    bash ${ROOT}/build/docker.sh ${REPO} ${TAG}

    # TODO provide wfcli to tests somehow (docker encapsulation?)
}

#
# Setup
#
setup() {
    # Pre-Cleanup
    cleanup

    # TODO termination of namespaces takes time, so we might need to retry helm install
    sleep 5

    # Setup Fission
    # TODO use test specific namespace
    helm_install_fission ${fissionHelmId} ${NS} "serviceType=NodePort,pullPolicy=IfNotPresent,analytics=false"

    # Give fission some time to setup
    # TODO make this a retry
    sleep 30

    # Setup Fission Workflows
    # TODO use test specific namespace
    helm_install_fission_workflows ${fissionWorkflowsHelmId} ${NS} "pullPolicy=IfNotPresent,tag=${TAG},bundleImage=${WORKFLOWS_BUNDLE_IMAGE},envImage=${WORKFLOWS_ENV_IMAGE},buildEnvImage=${WORKFLOWS_BUILD_ENV_IMAGE}"

    # Give fission workflow some time to setup
    sleep 5
}
#
# Test
#
test() {
    export FAILURES=0
    echo "--- Start Tests ---"
    run_all_tests ${ROOT}/test/e2e/tests
    echo "--- End Tests ---"
}
#
# Teardown
#

cleanup() {
    helm_uninstall_release ${fissionWorkflowsHelmId} || true
    helm_uninstall_release ${fissionHelmId} || true

    clean_tpr_crd_resources || true

    rm -f ./fission-workflows-bundle
    rm -f ./wfcli
}

prepare-env

build

setup

test

# TODO ensure cleanup regardless of result
cleanup

echo "minikube-buildtest completed!"