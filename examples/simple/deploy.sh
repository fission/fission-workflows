#!/bin/sh

set -x

# Deploy functions
fission env create --name binary-env --image fission/binary-env
fission fn create --name whalesay --env binary-env --code ./whalesay.sh
fission fn create --name fortune --env binary-env --code ./fortune.sh

# Deploy workflow
fission fn create --name fortunewhale --env workflow --code ./fortunewhale.wf.json
