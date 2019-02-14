#!/usr/bin/env bash

# This test checks if a invocation of a failed workflow will not hang indefinitely

set -exuo pipefail

FN_NAME=nonexistentfn
EXAMPLE_DIR=$(dirname $0)/../resources/

cleanup() {
    fission fn delete --name ${FN_NAME}
}
trap cleanup EXIT

fission fn create --name ${FN_NAME} --env workflow --src ${EXAMPLE_DIR}/nonexistentfn.wf.yaml

OUT=$(! fission fn test --name ${FN_NAME} 2>&1)

echo ${OUT} | grep -v "context deadline exceeded"