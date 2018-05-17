#!/bin/sh

# deploy.sh - deploys all functions and workflows in this directory

set -xe

# Deploy workflows
fission fn create --name inputs --env workflow --src ./inputs.wf.yaml
fission fn create --name fibonacci --env workflow --src ./fibonacci.wf.yaml
fission fn create --name sleepalot --env workflow --src ./sleepalot.wf.yaml