#!/usr/bin/env bash

# Tests the whale examples in the examples/whales directory

set -euo pipefail

WHALES_DIR=$(dirname $0)/../../../examples/whales
BINARY_ENV_VERSION=latest

# TODO move to util file
# retry function adapted from:
# https://unix.stackexchange.com/questions/82598/how-do-i-write-a-retry-logic-in-script-to-keep-retrying-to-run-it-upto-5-times/82610
function retry {
  local n=1
  local max=5
  local delay=5
  while true; do
    "$@" && break || {
      if [[ ${n} -lt ${max} ]]; then
        ((n++))
        echo "Command '$@' failed. Attempt $n/$max:"
        sleep ${delay};
      else
        >&2 echo "The command has failed after $n attempts."
        exit 1;
      fi
    }
  done
}

# Deploy required functions and environments
echo "Deploying required fission functions and environments..."
fission env create --name binary --image fission/binary-env:${BINARY_ENV_VERSION}
fission fn create --name whalesay --env binary --deploy ${WHALES_DIR}/whalesay.sh
fission fn create --name fortune --env binary --deploy ${WHALES_DIR}/fortune.sh

# Ensure that functions are available
retry curl ${FISSION_ROUTER}/fission-function/fortune

# test 1: fortunewhale - simplest example
echo "Test 1: fortunewhale"
fission fn create --name fortunewhale --env workflow --src ${WHALES_DIR}/fortunewhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -sL ${FISSION_ROUTER}/fission-function/fortunewhale | tee /dev/tty | grep "## ## ## ## ##"

# test 2: echowhale - parses body input
echo "Test 2: echowhale"
fission fn create --name echowhale --env workflow --src ${WHALES_DIR}/echowhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -sL -d "Test plz ignore" ${FISSION_ROUTER}/fission-function/echowhale | tee /dev/tty | grep "Test plz ignore"

# test 3: metadata - parses metadata (headers, query params...)
echo "Test 3: metadata"
fission fn create --name metadatawhale --env workflow --src ${WHALES_DIR}/metadatawhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -sL -H "Prefix: The test says:" ${FISSION_ROUTER}/fission-function/metadatawhale | tee /dev/tty | grep "Prefix: The test says:"

# Test 4: nestedwhale - shows nesting workflows
echo "Test 4: nestedwhale"
fission fn create --name nestedwhale --env workflow --src ${WHALES_DIR}/nestedwhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -sL ${FISSION_ROUTER}/fission-function/nestedwhale | tee /dev/tty | grep "## ## ## ## ##"

# Test 5: maybewhale - shows of dynamic tasks
echo "Test 5: maybewhale"
fission fn create --name maybewhale --env workflow --src ${WHALES_DIR}/maybewhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -sL ${FISSION_ROUTER}/fission-function/maybewhale | tee /dev/tty | grep "## ## ## ## ##"

# TODO cleanup