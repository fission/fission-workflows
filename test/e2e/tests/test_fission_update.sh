#!/usr/bin/env bash

# Simple test to verify that Fission update works

set -exuo pipefail

FN_NAME=fw-update-test
EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name ${FN_NAME}
}
trap cleanup EXIT

fission fn create --name ${FN_NAME} --env workflow --src $(dirname $0)/resources/hello.wf.yaml
sleep 2
fission fn test --name ${FN_NAME} \
    | tee /dev/tty \
    | grep "hello"

fission fn update --name ${FN_NAME} --src $(dirname $0)/resources/world.wf.yaml
sleep 2
fission fn test --name ${FN_NAME} \
    | tee /dev/tty \
    | grep "world"