#!/bin/sh

# deploy.sh - deploys all functions and workflows in this directory

set -x

# Deploy functions
fission env create --name binary --image fission/binary-env
fission fn create --name whalesay --env binary --deploy ./whalesay.sh
fission fn create --name fortune --env binary --deploy ./fortune.sh

# Deploy workflows
fission fn create --name fortunewhale --env workflow --src ./fortunewhale.wf.yaml
fission fn create --name echowhale --env workflow --src ./echowhale.wf.yaml
fission fn create --name maybewhale --env workflow --src ./maybewhale.wf.yaml
fission fn create --name nestedwhale --env workflow --src ./nestedwhale.wf.yaml
fission fn create --name metadatawhale --env workflow --src ./metadatawhale.wf.yaml
