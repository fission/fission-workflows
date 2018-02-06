#!/usr/bin/env bash

# Builds, deploys and tests a Fission Workflow deployment
# This expects a cluster to be present with kubectl and helm clients setup locally.

set -euo pipefail

. $(dirname $0)/utils.sh
ROOT=$(dirname $0)/../..
TEST_SUITE_UID=$(generate_test_id)
DOCKER_REPO=minikube-ci
WORKFLOWS_ENV_IMAGE=${DOCKER_REPO}/workflow-env
WORKFLOWS_BUILD_ENV_IMAGE=${DOCKER_REPO}/workflow-build-env
WORKFLOWS_BUNDLE_IMAGE=${DOCKER_REPO}/fission-workflows-bundle
NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder
fissionHelmId=fission
fissionWorkflowsHelmId=fission-workflows
FISSION_VERSION=0.4.1
TAG=test
TEST_STATUS=0
TEST_LOGFILE_PATH=tests.log


cleanup() {
    echo "Removing Fission and Fission Workflow deployments..."
    helm_uninstall_release ${fissionWorkflowsHelmId} || true
    helm_uninstall_release ${fissionHelmId} || true

    echo "Removing custom resources..."
    clean_tpr_crd_resources || true

    # Trigger deletion of all namespaces before waiting - for concurrency of deletion
    echo "Forcing deletion of namespaces..."
    kubectl delete ns/${NS} > /dev/null 2>&1 # Sometimes it is not deleted by helm delete
    kubectl delete ns/${NS_BUILDER} > /dev/null 2>&1 # Sometimes it is not deleted by helm delete
    kubectl delete ns/${NS_FUNCTION} > /dev/null 2>&1 # Sometimes it is not deleted by helm delete

    # Wait until all namespaces are actually deleted!
    echo "Awaiting deletion of namespaces..."
    retry kubectl delete ns/${NS} 2>&1  | grep -qv "Error from server (Conflict):"
    retry kubectl delete ns/${NS_BUILDER} 2>&1 | grep -qv "Error from server (Conflict):"
    retry kubectl delete ns/${NS_FUNCTION} 2>&1  | grep -qv "Error from server (Conflict):"

    echo "Cleaning up local filesystem..."
    rm -f ./fission-workflows-bundle ./wfcli
    sleep 5
}

print_report() {
    emph "--- Test Report ---"
    cat ${TEST_LOGFILE_PATH} | grep '\(FAILURE\|SUCCESS\).*|'
    emph "--- End Test Report ---"
}

on_exit() {
    # Dump all the logs
    dump_logs ${NS} ${NS_FUNCTION}

    # Ensure teardown after tests finish
    # TODO provide option to not cleanup the test setup after tests (e.g. for further tests)
    emph "Cleaning up cluster..."
    retry cleanup

    # Print a short test report
    print_report

    # Ensure correct exist status
    echo "TEST_STATUS: ${TEST_STATUS}"
    if [ ${TEST_STATUS} -ne 0 ]; then
        exit 1
    fi
}

# Ensure that minikube cluster is cleaned (in case it is an existing cluster)
emph "Cleaning up cluster..."
retry cleanup

# Ensure printing of report
trap on_exit EXIT

#
# Deploy Fission
# TODO use test specific namespace
emph "Deploying Fission: helm chart '${fissionHelmId}' in namespace '${NS}'..."
# Needs to be retried because k8s can still be busy with cleaning up
helm_install_fission ${fissionHelmId} ${NS} ${FISSION_VERSION} "serviceType=NodePort,pullPolicy=IfNotPresent,analytics=false"

# Direct CLI to the deployed cluster
FISSION_URL=http://$(minikube ip):31313
FISSION_ROUTER=$(minikube ip):31314

# Wait for Fission to get ready
sleep 5
retry fission fn list
echo
emph "Fission deployed!"

#
# Build
#
emph "Building binaries..."
bash ${ROOT}/build/build-linux.sh

# Build docker images
emph "Building images..."
bash ${ROOT}/build/docker.sh ${DOCKER_REPO} ${TAG}

#
# Deploy Fission Workflows
# TODO use test specific namespace
emph "Deploying Fission: ${fissionWorkflowsHelmId}"
helm_install_fission_workflows ${fissionWorkflowsHelmId} ${NS} "pullPolicy=IfNotPresent,tag=${TAG},bundleImage=${WORKFLOWS_BUNDLE_IMAGE},envImage=${WORKFLOWS_ENV_IMAGE},buildEnvImage=${WORKFLOWS_BUILD_ENV_IMAGE}"

# Wait for Fission Workflows to get ready
sleep 5
retry wfcli status
echo
emph "Fission Workflows deployed!"

#
# Test
#
emph "--- Start Tests ---"
$(dirname $0)/runtests.sh 2>&1 | tee ${TEST_LOGFILE_PATH}
TEST_STATUS=${PIPESTATUS[0]}
emph "--- End Tests ---"