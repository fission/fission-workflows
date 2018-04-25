#!/bin/sh

# deploy.sh - deploys all functions and workflows in this directory

set -xe

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
fission fn create --name scopedwhale --env workflow --src ./scopedwhale.wf.yaml
fission fn create --name failwhale --env workflow --src ./failwhale.wf.yaml
fission fn create --name httpwhale --env workflow --src ./httpwhale.wf.yaml
fission fn create --name switchwhale --env workflow --src ./switchwhale.wf.yaml
fission fn create --name foreachwhale --env workflow --src ./foreachwhale.wf.yaml
