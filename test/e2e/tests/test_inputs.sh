#!/usr/bin/env bash

set -exuo pipefail

EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name inputs
}
trap cleanup EXIT

fission fn create --name inputs --env workflow --src ${EXAMPLE_DIR}/inputs.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
fission fn test --name inputs -b 'foobar' -H 'HEADER_KEY: HEADER_VAL' --method PUT \
    | tee /dev/tty \
    | grep -i Header_Val \
    | grep HEADER_VAL \
    | grep -i PUT \
    | grep -q foobar