# End-to-End Tests (e2e)

## Usage
There are several options to run the e2e tests:

- `runtests.sh` will assume a complete deployment of both Fission and Fission Workflows, and will simply run the tests.
- `buildtest.sh` if you have all dependencies (e.g. a connected `kubectl`, `helm` installation) except for a fission cluster.
Buildtest will deploy Fission, build and deploy Fission Workflows, run the tests, and cleanup the deployment.
- `minikube-buildtest.sh` will provision a Kubernetes cluster using Minikube and build, deploy and test Fission Workflows.
- `gcloud-buildtest.sh` will use GKE for the Kubernetes cluster and build, deploy, and test Fission workflows.

## Test Cases
All tests should be placed in the `./tests` directory.
In that directory, the e2e test framework interprets any file with `test_*` as a test case.

A test case is just a regular shell script. 
Depending on the status code of the shell script, the test is either interpreted as SUCCESS (0) or FAILURE (!= 0).

The test framework provides the following setup to each shell script:
```
# Environment Variables
TEST_UID = unique identifier for the test case. Should be used whenever the test modifies the k8s cluster.
ROOT     = root of the project

# PATH
kubectl (CLI)
fission (CLI)
fission-workflows   (CLI)
python3
``` 

#### Conventions
- use `${TEST_UID}` for all naming of objects that are created in the cluster. This prevents collisions between (parallel) tests.