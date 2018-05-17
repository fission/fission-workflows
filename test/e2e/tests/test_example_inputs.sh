#!/usr/bin/env bash

set -euo pipefail

EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name inputs
}
trap cleanup EXIT

fission fn create --name inputs --env workflow --src ${EXAMPLE_DIR}/inputs.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
fission fn test --name inputs -b 'foobar' -H 'HEADER_KEY: HEADER_VAL' --method PUT | grep HEADER_KEY | grep HEADER_VAL | grep PUT | grep foobar