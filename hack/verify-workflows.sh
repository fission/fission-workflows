#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

find_files() {
  find . -not \( \
      \( \
        -wholename '*/vendor/*' \
      \) -prune \
    \) -name '*.wf.yaml'
}

TOOL="fission-workflows validate"
bad_files=$(find_files | xargs ${TOOL})
if [[ -n "${bad_files}" ]]; then
  echo "The following workflows are invalid: "
  echo "${bad_files}"
  exit 1
fi
