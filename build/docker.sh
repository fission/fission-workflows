#!/usr/bin/env bash

set -e
set -x

if [ ! -f ./workflow-engine-bundle ]; then
    echo "Executable './workflow-engine-bundle' not found!"
    exit 1;
fi

docker build --tag="fission/fission-workflow-bundle" .
