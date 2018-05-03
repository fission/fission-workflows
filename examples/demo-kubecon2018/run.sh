#!/bin/bash
DEMO_RUN_FAST=1
source util.sh

clear 

# First setup the environment - removed to safe time
#desc "Setup a binary function environment"
#run "fission env create --name binary --image fission/binary-env"

desc "Create fortune function"
run "fission function create --name fortune --env binary --deploy ./fortune.sh"

desc "Test fortune function"
run "fission function test --name fortune"

desc "Create whalesay function"
run "fission function create --name whalesay --env binary --deploy ./whalesay.sh"

desc "Test whalesay function"
run "fission function test --name whalesay -b FaaSinating!"

# Setup the workflow
desc "Workflow overview"
run "cat ./fortunewhale.wf.yaml"

desc "Create fortunewhale workflow"
run "fission function create --name fortunewhale --env workflow --src ./fortunewhale.wf.yaml"

# Test workflow
desc "Run the workflow without input"
run "fission function test --name fortunewhale"

desc "Run the workflow with input"
run "fission fn test --name fortunewhale -b Hello_Copenhagen!"
