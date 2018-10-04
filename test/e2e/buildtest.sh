#!/usr/bin/env bash

# Builds, deploys and tests a Fission Workflow deployment
# This expects a cluster to be present with kubectl and helm clients setup locally.

set -euo pipefail

. $(dirname $0)/utils.sh

ROOT=$(dirname $0)/../..
TEST_SUITE_UID=$(generate_test_id)
DOCKER_REPO=gcr.io/fission-ci
WORKFLOWS_ENV_IMAGE=${DOCKER_REPO}/workflow-env
WORKFLOWS_BUILD_ENV_IMAGE=${DOCKER_REPO}/workflow-build-env
WORKFLOWS_BUNDLE_IMAGE=${DOCKER_REPO}/fission-workflows-bundle
WORKFLOWS_PROXY_IMAGE=${DOCKER_REPO}/workflows-proxy
TAG=ci-test
NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder
fissionHelmId=fission
fissionWorkflowsHelmId=fission-workflows
TEST_STATUS=0
TEST_LOGFILE_PATH=tests.log
BIN_DIR="${BIN_DIR:-$HOME/testbin}"


print_report() {
    emph "--- Test Report ---"
    if ! cat ${TEST_LOGFILE_PATH} | grep '\(FAILURE\|SUCCESS\).*|' ; then
        echo "No report found."
    fi
    emph "--- End Test Report ---"
}

on_exit() {
    emph "[Buildtest exited]"

    # Dump all the logs
    dump_logs ${NS} ${NS_FUNCTION} ${NS_BUILDER} || true

    # Print a short test report
    print_report

    # Ensure correct exist status
    echo "TEST_STATUS: ${TEST_STATUS}"
    if [ ${TEST_STATUS} -ne 0 ]; then
        exit 1
    fi
}

emph "Starting buildtest..."

trap on_exit EXIT

cleanup_fission_workflows ${fissionWorkflowsHelmId} || true

#
# Build
#
# Build docker images
emph "Building images..."
echo "Repo: ${DOCKER_REPO}, Tag: ${TAG}"
bash ${ROOT}/build/docker.sh ${DOCKER_REPO} ${TAG}

# Ensure cli is in path
emph "Copying fission-workflows to '${BIN_DIR}/fission-workflows'..."
bundleImage=${DOCKER_REPO}/fission-workflows-bundle:${TAG}
bundleContainer=$(docker create ${bundleImage} tail /dev/null)
docker cp ${bundleContainer}:/fission-workflows ${BIN_DIR}/fission-workflows
docker rm -v ${bundleContainer}
fission-workflows -h > /dev/null

# Publish to gcloud
emph "Pushing images to container registry..."
gcloud docker -- push ${WORKFLOWS_ENV_IMAGE}:${TAG}
gcloud docker -- push ${WORKFLOWS_BUILD_ENV_IMAGE}:${TAG}
gcloud docker -- push ${WORKFLOWS_BUNDLE_IMAGE}:${TAG}
gcloud docker -- push ${WORKFLOWS_PROXY_IMAGE}:${TAG}

#
# Deploy Fission Workflows
# TODO use test specific namespace
emph "Deploying Fission Workflows '${fissionWorkflowsHelmId}' to ns '${NS}'..."
helm_install_fission_workflows ${fissionWorkflowsHelmId} ${NS} "pullPolicy=Always,tag=${TAG},bundleImage=${WORKFLOWS_BUNDLE_IMAGE},fission.env.runtimeImage=${WORKFLOWS_PROXY_IMAGE},fission.env.builderImage=${WORKFLOWS_BUILD_ENV_IMAGE}"

# Wait for Fission Workflows to get ready
fission-workflows config
emph "Waiting for Fission Workflows to be ready..."
sleep 5
retry fission-workflows status
echo
emph "Fission Workflows deployed!"

#
# Test
#
retry $(dirname $0)/tests/test_inputs.sh
emph "--- Start Tests ---"
export ROOT
export TEST_SUITE_UID
echo "ROOT: $ROOT"
echo "TEST_SUITE_UID: $TEST_SUITE_UID"
$(dirname $0)/runtests.sh 2>&1 | tee ${TEST_LOGFILE_PATH}
TEST_STATUS=${PIPESTATUS[0]}
emph "--- End Tests ---"
