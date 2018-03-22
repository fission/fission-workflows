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
NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder
fissionHelmId=fission
fissionWorkflowsHelmId=fission-workflows
FISSION_VERSION=0.6.0
TAG=test
TEST_STATUS=0
TEST_LOGFILE_PATH=tests.log
BIN_DIR="${BIN_DIR:-$HOME/testbin}"

#
# Deploy Fission
# TODO use test specific namespace
emph "Deploying Fission: helm chart '${fissionHelmId}' in namespace '${NS}'..."
# Needs to be retried because k8s can still be busy with cleaning up
# helm_install_fission ${fissionHelmId} ${NS} ${FISSION_VERSION} "serviceType=NodePort,pullPolicy=IfNotPresent,
# analytics=false"
controllerPort=31234
routerPort=31235
helm_install_fission ${fissionHelmId} ${NS} ${FISSION_VERSION} "controllerPort=${controllerPort},routerPort=${routerPort},pullPolicy=Always,analytics=false"

# Direct CLI to the deployed cluster
#set_environment ${NS} ${routerPort}
emph "Fission environment: FISSION_URL: '${FISSION_URL:-EMPTY}' and FISSION_ROUTER: '${FISSION_ROUTER:-EMPTY}'"

# Wait for Fission to get ready
emph "Waiting for fission to be ready..."
sleep 5
retry fission fn list
echo
emph "Fission deployed!"

#
# Build
#
# Build docker images
emph "Building images..."
bash ${ROOT}/build/docker.sh ${DOCKER_REPO} ${TAG}

# Ensure cli is in path
emph "Copying wfcli to '${BIN_DIR}/wfcli'..."
bundleImage=${DOCKER_REPO}/fission-workflows-bundle:${TAG}
bundleContainer=$(docker create ${bundleImage} tail /dev/null)
docker cp ${bundleContainer}:/wfcli ${BIN_DIR}/wfcli
docker rm -v ${bundleContainer}
wfcli -h > /dev/null

# Publish to gcloud
emph "Pushing images to container registry..."
gcloud docker -- push ${WORKFLOWS_ENV_IMAGE}:${TAG}
gcloud docker -- push ${WORKFLOWS_BUILD_ENV_IMAGE}:${TAG}
gcloud docker -- push ${WORKFLOWS_BUNDLE_IMAGE}:${TAG}

#
# Deploy Fission Workflows
# TODO use test specific namespace
emph "Deploying Fission Workflows '${fissionWorkflowsHelmId}' to ns '${NS}'..."
helm_install_fission_workflows ${fissionWorkflowsHelmId} ${NS} "pullPolicy=IfNotPresent,tag=${TAG},bundleImage=${WORKFLOWS_BUNDLE_IMAGE},envImage=${WORKFLOWS_ENV_IMAGE},buildEnvImage=${WORKFLOWS_BUILD_ENV_IMAGE}"

# Wait for Fission Workflows to get ready
wfcli config
emph "Waiting for Fission Workflows to be ready..."
sleep 5
retry wfcli status
echo
emph "Fission Workflows deployed!"

#
# Test
#
emph "--- Start Tests ---"
export ROOT
export TEST_SUITE_UID
echo "ROOT: $ROOT"
echo "TEST_SUITE_UID: $TEST_SUITE_UID"
$(dirname $0)/runtests.sh 2>&1 | tee ${TEST_LOGFILE_PATH}
TEST_STATUS=${PIPESTATUS[0]}
emph "--- End Tests ---"
