#!/usr/bin/env bash

# Tests the whale examples in the examples/whales directory

set -euo pipefail

WHALES_DIR=${ROOT}/examples/whales
BINARY_ENV_VERSION=latest

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
curl ${FISSION_ROUTER}/fission-function/fortunewhale

# test 2: echowhale - parses body input
echo "Test 2: echowhale"
fission fn create --name echowhale --env workflow --src ${WHALES_DIR}/echowhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -d "Test plz ignore" ${FISSION_ROUTER}/fission-function/fortunewhale

# test 3: metadata - parses metadata (headers, query params...)
echo "Test 3: metadata"
fission fn create --name metadatawhale --env workflow --src ${WHALES_DIR}/metadatawhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl -H "Prefix: The test says:" ${FISSION_ROUTER}/fission-function/metadatawhale

# Test 4: nestedwhale - shows nesting workflows
echo "Test 4: nestedwhale"
fission fn create --name nestedwhale --env workflow --src ${WHALES_DIR}/nestedwhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl ${FISSION_ROUTER}/fission-function/nestedwhale

# Test 5: maybewhale - shows of dynamic tasks
echo "Test 5: maybewhale"
fission fn create --name maybewhale --env workflow --src ${WHALES_DIR}/maybewhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl ${FISSION_ROUTER}/fission-function/maybewhale

# TODO cleanup