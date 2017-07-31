#!/usr/bin/env bash

set -e
set -x

if [ ! -f ./workflow-engine ]; then
    echo "Executable './workflow-engine' not found!"
    exit 1;
fi

docker build --tag="fission/fission-workflow" .
