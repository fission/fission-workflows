#!/usr/bin/env bash

#
# Generates JSON workflow definitions from YAML workflow definitions in this repo.
#

set -e

wfs=`find . -type f -name "*.wf.yaml"`

while read -r wf; do
    echo ${wf}
    wfparser ${wf}
done <<< "$wfs"
