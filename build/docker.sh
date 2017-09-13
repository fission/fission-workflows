#!/usr/bin/env bash

set -ex

if [ ! -f ./fission-workflow-bundle ]; then
    echo "Executable './fission-workflow-bundle' not found!"
    exit 1;
fi

chmod +x fission-workflow-bundle
yes | cp fission-workflow-bundle bundle/
yes | cp fission-workflow-bundle env/

docker build --tag="fission/fission-workflow-bundle" bundle/
docker build --tag="fission/workflow-env" env/
