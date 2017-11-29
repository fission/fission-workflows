#!/bin/bash

# Setup, run and teardown of tests.

set -euox pipefail

source $(dirname $0)/utils.sh


if [ ! -f ${HOME}/.kube/config ]
then
    echo "Skipping end to end tests, no cluster credentials"
    exit 0
fi

REPO=gcr.io/fission-ci
BUNDLE_IMAGE=$REPO/fission-workflows-bundle
ENV_IMAGE=$REPO/fission-workflows-env
BUILD_ENV_IMAGE=$REPO/fission-workflows-build-env
TAG=test
ROOT=$(dirname $0)/../..
id=$(generate_test_id)

# Build binaries
bash ${ROOT}/build/build-linux.sh

# Build docker images
bash ${ROOT}/build/docker.sh

# Publish images to gcloud registry
bash ${ROOT}/hack/docker-publish.sh ${REPO} ${TAG}

# Add wfcli to path
cp ${ROOT}/build/wfcli /usr/local/bin/wfcli

# Install fission
fissionHelmId=fission-${id}
clean_tpr_crd_resources
trap "helm_uninstall_fission ${fissionHelmId}" EXIT
if ! helm_install_fission ${fissionHelmId} ; then
	dump_logs $id
	exit 1
fi

dump_system_info

# Deploy workflows
fissionWorkflowsHelmId=fission-${id}
trap "helm_uninstall_fission_workflows ${fissionWorkflowsHelmId}" EXIT
helm_install_fission_workflows ${fissionWorkflowsHelmId}

# Run rests
run_all_tests

if [ ${FAILURES} -ne 0 ]; then
	exit 1
fi