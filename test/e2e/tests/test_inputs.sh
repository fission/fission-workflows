#!/usr/bin/env bash

set -exuo pipefail

FN_NAME=inputs
EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name ${FN_NAME}
}
trap cleanup EXIT

fission fn create --name ${FN_NAME} --env workflow --src ${EXAMPLE_DIR}/inputs.wf.yaml
fission fn test --name ${FN_NAME} -b 'foobar' -H 'HEADER_KEY: HEADER_VAL' -H 'Content-Type: text/plain' --method PUT \
    | tee /dev/tty \
    | grep -i Header_Key \
    | grep HEADER_VAL \
    | grep -i PUT \
    | grep -q foobar