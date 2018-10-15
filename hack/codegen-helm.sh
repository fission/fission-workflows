#!/usr/bin/env bash

set -eou pipefail

TARGET=examples/workflows-env.yaml

echo '# An Kubernetes example template of a Fission Workflow deployment as an environment in Fission' > ${TARGET}

helm template --namespace fission --set fission.env.name=workflows charts/fission-workflows/ >> ${TARGET}

echo "Created ${TARGET}"