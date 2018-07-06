#!/usr/bin/env bash

set -euo pipefail

EXAMPLE_DIR=$(dirname $0)/../../../examples/misc

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
    fission fn delete --name dump
    fission fn delete --name fission-inputs
}
trap cleanup EXIT
fission fn create --name dump --env binary --src ${EXAMPLE_DIR}/dump.sh
retry fission fn test --name dump

fission fn create --name fission-inputs --env workflow --src ${EXAMPLE_DIR}/fission-inputs.wf.yaml
sleep 5 # TODO remove this once we can initiate synchronous commands
fission fn test --name fission-inputs -b 'foobar' -H 'HEADER_KEY: HEADER_VAL' --method PUT \
    | tee /dev/tty \
    | grep -i Header_Val \
    | grep HEADER_VAL \
    | grep -i PUT \
    | grep -q foobar