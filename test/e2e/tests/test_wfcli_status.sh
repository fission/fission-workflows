#!/bin/bash

# Test if `wfcli status` correctly shows status of a cluster

set -euo pipefail

wfcli status

# Check if hitting a non-existing cluster results in an error
if $(wfcli --url http://127.0.0.1:1337 status 2> /dev/null) ; then
    echo "wfcli status indicated healthy for non-existing cluster"
    exit 1
fi