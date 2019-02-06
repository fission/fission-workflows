#!/usr/bin/env bash

set -exuo pipefail


FN_NAME=inputs
EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name ${FN_NAME}
}
trap cleanup EXIT

fission fn create --name ${FN_NAME} --env workflow --src ${EXAMPLE_DIR}/inputs.wf.yaml
fission fn test --name ${FN_NAME} -H 'Content-Type: application/x-www-form-urlencoded' --method POST \
    -b 'foo=bar&acme=123' \
    | tee /dev/tty \
    | grep "\"foo\":\"bar\"" \
    | grep -q "\"acme\":\"123\""
