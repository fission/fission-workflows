#!/usr/bin/env bash

WHALES_DIR=${ROOT}/examples/whales
BINARY_ENV_VERSION=latest

# Deploy required functions and environments
echo "Deploying required fission functions and environments..."
fission env create --name binary --image fission/binary-env:${BINARY_ENV_VERSION}
fission fn create --name whalesay --env binary --deploy ${WHALES_DIR}/whalesay.sh
fission fn create --name fortune --env binary --deploy ${WHALES_DIR}/fortune.sh

# Ensure that functions are available
retry curl ${FISSION_ROUTER}/fission-function/fortune

fission fn create --name fortunewhale --env workflow --src ${WHALES_DIR}/fortunewhale.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
curl ${FISSION_ROUTER}/fission-function/fortunewhale