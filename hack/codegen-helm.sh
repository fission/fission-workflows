#!/usr/bin/env bash

set -eou pipefail

TARGET=examples/workflows-env.yaml

helm template \
    --set apiserver=false,env.name=workflows-env \
    charts/fission-workflows/ > ${TARGET}
echo "Created ${TARGET}"