#!/bin/sh

set -x

# Deploy functions
fission env create --name binary --image fission/binary-env:0.3.0
fission fn create --name whalesay --env binary --deploy ./whalesay.sh
fission fn create --name fortune --env binary --deploy ./fortune.sh

# Deploy workflows
fission fn create --name fortunewhale --env workflow --src ./fortunewhale.wf.yaml
fission fn create --name echowhale --env workflow --src ./echowhale.wf.yaml
fission fn create --name maybewhale --env workflow --src ./maybewhale.wf.yaml
fission fn create --name nestedwhale --env workflow --src ./nestedwhale.wf.yaml
