#!/usr/bin/env bash

set -euo pipefail

EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name dump
    fission fn delete --name fissioninputs
}
trap cleanup EXIT
fission fn create --name dump --env workflow --src ${EXAMPLE_DIR}/dump.sh
sleep 5
fission fn create --name fissioninputs --env workflow --src ${EXAMPLE_DIR}/fissioninputs.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
fission fn test --name fissioninputs -b 'foobar' -H 'HEADER_KEY: HEADER_VAL' --method PUT \
    | tee /dev/tty \
    | grep -i Header_Val \
    | grep HEADER_VAL \
    | grep -i PUT \
    | grep -q foobar