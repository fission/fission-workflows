#!/usr/bin/env bash

set -eou pipefail

TARGET=examples/workflows-env.yaml

echo '# Example of the configuration of the workflow engine as a Fission Environment\n\n/' > ${TARGET}

helm template \
    --set apiserver=false,env.name=workflows-env \
    charts/fission-workflows/ | grep -v '^---$' >> ${TARGET}

echo "Created ${TARGET}"