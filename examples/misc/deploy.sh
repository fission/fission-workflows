#!/bin/sh

# deploy.sh - deploys all functions and workflows in this directory

set -x

fission env create --name binary --image fission/binary-env
fission fn create --name dump --env binary --deploy ./dump.sh

# Deploy workflows
fission fn create --name inputs --env workflow --src ./inputs.wf.yaml
fission fn create --name fibonacci --env workflow --src ./fibonacci.wf.yaml
fission fn create --name sleepalot --env workflow --src ./sleepalot.wf.yaml
fission fn create --name fission-inputs --env workflow --src ./fission-inputs.wf.yaml