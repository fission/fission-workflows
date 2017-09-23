#!/usr/bin/env bash

set -ex

if [ ! -f ./fission-workflows-bundle ]; then
    echo "Executable './fission-workflows-bundle' not found!"
    exit 1;
fi

rm -f bundle/fission-workflows-bundle
rm -f env/fission-workflows-bundle
chmod +x fission-workflows-bundle
yes | cp fission-workflows-bundle bundle/
yes | cp fission-workflows-bundle env/

docker build --tag="fission/fission-workflows-bundle" bundle/
docker build --tag="fission/workflow-env" env/
