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

cleanup() {
    echo "cleaning up..."
    set +e
    fission fn delete --name whalesay
    fission fn delete --name fortune
    fission fn delete --name fortunewhale
    fission fn delete --name echowhale
    fission fn delete --name metadatawhale
    fission fn delete --name nestedwhale
    fission fn delete --name maybewhale
    fission fn delete --name failwhale
    fission fn delete --name foreachwhale
    fission fn delete --name switchwhale
    fission fn delete --name whilewhale
    fission fn delete --name httpwhale
    fission fn delete --name scopedwhale
}
trap cleanup EXIT

# Deploy required functions and environments
echo "Deploying required fission functions and environments..."
fission env create --name binary --image fission/binary-env:${BINARY_ENV_VERSION} || true
fission fn create --name whalesay --env binary --deploy ${WHALES_DIR}/whalesay.sh
fission fn create --name fortune --env binary --deploy ${WHALES_DIR}/fortune.sh

# Ensure that functions are available
retry fission fn test --name fortune > /dev/null

# test 1: fortunewhale - simplest example
echo "[Test 1]: fortunewhale"
fission fn create --name fortunewhale --env workflow --src ${WHALES_DIR}/fortunewhale.wf.yaml
fission fn test --name fortunewhale | tee /dev/tty | grep -q "## ## ## ## ##"

# test 2: echowhale - parses body input
echo "[Test 2]: echowhale"
fission fn create --name echowhale --env workflow --src ${WHALES_DIR}/echowhale.wf.yaml
fission fn test --name echowhale -b "Test plz ignore" -H "Content-Type: text/plain" | tee /dev/tty | grep -q "Test plz ignore"

# test 3: metadata - parses metadata (headers, query params...)
echo "[Test 3]: metadatawhale"
fission fn create --name metadatawhale --env workflow --src ${WHALES_DIR}/metadatawhale.wf.yaml
fission fn test --name metadatawhale -H "Prefix: The test says:" | tee /dev/tty | grep -q "The test says:"

# Test 4: nestedwhale - shows nesting workflows
echo "[Test 4]: nestedwhale"
fission fn create --name nestedwhale --env workflow --src ${WHALES_DIR}/nestedwhale.wf.yaml
fission fn test --name nestedwhale | tee /dev/tty | grep -q "## ## ## ## ##"

# Test 5: maybewhale - shows of dynamic tasks
echo "[Test 5]: maybewhale"
fission fn create --name maybewhale --env workflow --src ${WHALES_DIR}/maybewhale.wf.yaml
fission fn test --name maybewhale | tee /dev/tty | grep -q "## ## ## ## ##"

echo "[Test 6]: failwhale"
fission fn create --name failwhale --env workflow --src ${WHALES_DIR}/failwhale.wf.yaml
fission fn test --name failwhale | tee /dev/tty | grep -q "all has failed"

echo "[Test 7]: foreachwhale"
fission fn create --name foreachwhale --env workflow --src ${WHALES_DIR}/foreachwhale.wf.yaml
fission fn test --name foreachwhale | tee /dev/tty | grep -q "[10,20,30,40,50]"

echo "[Test 8]: switchwhale"
fission fn create --name switchwhale --env workflow --src ${WHALES_DIR}/switchwhale.wf.yaml
fission fn test --name switchwhale -b 'hello' | tee /dev/tty | grep -q "world"
fission fn test --name switchwhale -b 'foo' | tee /dev/tty | grep -q "bar"
fission fn test --name switchwhale -b 'acme' | tee /dev/tty | grep -q "right..."

echo "[Test 9]: whilewhale"
fission fn create --name whilewhale --env workflow --src ${WHALES_DIR}/whilewhale.wf.yaml
fission fn test --name whilewhale | tee /dev/tty | grep -q "5"

echo "[Test 10]: httpwhale"
fission fn create --name httpwhale --env workflow --src ${WHALES_DIR}/httpwhale.wf.yaml
fission fn test --name httpwhale | tee /dev/tty | grep -q "## ## ## ## ##"

echo "[Test 11]: scopedwhale"
fission fn create --name scopedwhale --env workflow --src ${WHALES_DIR}/scopedwhale.wf.yaml
fission fn test --name scopedwhale | tee /dev/tty | grep -q "## ## ## ## ##"