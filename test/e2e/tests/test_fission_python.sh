#!/usr/bin/env bash


# This test reproduces missing data transfer from or to python functions

set -exuo pipefail

TEST_RESOURCE_DIR=$(dirname $0)/../resources
PY_ENV_IMAGE=fission/python-env
PY_ENV_NAME=test-fission-python-env
PY_FN_NAME=pyecho
WF_NAME=python-test-wf

cleanup() {
    fission fn delete --name ${PY_FN_NAME} || true
    fission fn delete --name ${WF_NAME} || true
    fission env delete --name ${PY_ENV_NAME} || true
}
trap cleanup EXIT

# Deploy and test Python function
fission env create --name ${PY_ENV_NAME} --image ${PY_ENV_IMAGE} --poolsize 1
fission fn create --name ${PY_FN_NAME} --code ${TEST_RESOURCE_DIR}/pyecho.py --env ${PY_ENV_NAME}
fission fn test --name ${PY_FN_NAME} -b 'Hello Python' | tee /dev/tty | grep 'Hello Python'

# Deploy and test Python-based workflow
fission fn create --name ${WF_NAME} --src ${TEST_RESOURCE_DIR}/pyecho.wf.yaml --env workflow
fission fn test --name ${WF_NAME} -b 'Hello Python' | tee /dev/tty | grep 'Hello Python, Bye Python'
