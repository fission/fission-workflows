#!/usr/bin/env bash

set -euo pipefail

EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

cleanup() {
    fission fn delete --name sleepalot
}
trap cleanup EXIT

fission fn create --name sleepalot --env workflow --src ${EXAMPLE_DIR}/sleepalot.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
fission fn test --name sleepalot