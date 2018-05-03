# Demo - KubeCon + CloudNativeCon 2018 Europe

## Prerequisites
- fission (CLI)
- kubectl (+ connected cluster)
- wfcli (CLI for Fission Workflows)
- helm

## Usage
- First run `setup.sh` to prepare your cluster.
This will install helm, Fission and Fission Workflows onto it.
For every artefact it verifies if it has been deployed correctly.
- Then you can run `run.sh`, which is a script that runs the commands step-by-step on each enter.
- If you want to do another run it could be best to cleanup the cluster with `cleanup.sh`.
This script removes all artifacts deployed by `run.sh`.
