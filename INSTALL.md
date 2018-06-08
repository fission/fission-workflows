# Installation

This document covers installing Fission Workflows.

Fission and Workflows are under active development, which has lead to incompatible versions.
Compatibility chart:

Workflows | Compatible Fission versions 
----------|---------------------------
0.1.x     | 0.3.0 up to 0.6.1
0.2.x     | 0.4.1 up to 0.6.1 
0.3.0     | all (tested on 0.6.0, 0.6.1, and 0.7.2)
0.4.0     | all 
0.5.0     | all 

### Prerequisites

Fission Workflows requires the following to be installed on your host machine:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm](https://github.com/kubernetes/helm)

Additionally, Fission Workflows requires a [Fission](https://github.com/fission/fission) 
deployment on your Kubernetes cluster. If you do not have a Fission deployment, follow
[Fission's installation guide](http://fission.io/docs/0.7.2/install/).

**(Note that Fission Workflows requires Fission 0.4.1 or higher, with the NATS component installed!)**

### Installing Fission Workflows

Fission Workflows is an add-on to Fission. You can install both
Fission and Fission Workflows using helm charts.

Assuming you have a Kubernetes cluster, run the following commands:

```bash
# Add the Fission charts repo
helm repo add fission-charts https://fission.github.io/fission-charts/
helm repo update

# Install Fission 
# This assumes that you do not have a Fission deployment yet, and are installing on a standard Minikube deployment.
# Otherwise see http://fission.io/docs/0.7.2/install/ for more detailed instructions
helm install --wait -n fission-all --namespace fission --set serviceType=NodePort --set analytics=false fission-charts/fission-all --version 0.7.2

# Install Fission Workflows
helm install --wait -n fission-workflows fission-charts/fission-workflows --version 0.5.0
```

### Creating your first workflow

After installing Fission and Workflows, you're ready to run a simple
test workflow.  Clone this repository, and from its root directory, run:

```bash
#
# Add binary environment and create two test functions on your Fission setup:
#
fission env create --name binary --image fission/binary-env
fission function create --name whalesay --env binary --deploy examples/whales/whalesay.sh
fission function create --name fortune --env binary --deploy examples/whales/fortune.sh

#
# Create a workflow that uses those two functions. A workflow is just
# a function that uses the special "workflow" environment.
#
fission function create --name fortunewhale --env workflow --src examples/whales/fortunewhale.wf.yaml

#
# Map an HTTP GET to your new workflow function:
#
fission route create --method GET --url /fortunewhale --function fortunewhale

#
# Invoke the workflow with an HTTP request:
#
curl ${FISSION_ROUTER}/fortunewhale
```

### Workflows client (optional)
To use Fission Workflows there is no need to learn any other client other than the one you already use for function invocation - after all, a workflow is just another function.
However, in many cases it is useful to have more insight in and control over the behaviour of the workflows (for example when developing/debugging workflows).
To get these more capabilities and insight, you can use the `fission-workflows` client.

It has the following features:
- Get insight into workflow and invocations statuses.
- Start, and cancel workflow invocation.
- Perform administrative or debugging actions: for example halting and resuming the engine.
- validating workflow definitions locally.

#### Installation
To install `fission-workflows` either download a version of the binary from the [releases](https://github.com/fission/fission-workflows/releases).
For example, to download and install version 0.5.0,  assuming that you use OS X:
```bash
curl -o fission-workflows https://github.com/fission/fission-workflows/releases/download/0.3.0/fission-workflows-osx
chmod +x ./fission-workflows
sudo mv ./fission-workflows /usr/local/bin
```

Or install the latest, edge version with Go:
```bash
go get -u github.com/fission/fission-workflows/cmd/fission-workflows
```

The `fission-workflows` client uses the `FISSION_URL` environment variable to find the Fission controller server to use as a proxy to the workflow apiserver.
By default `fission-workflows` uses ttp://localhost:31313 to locate the Fission controller.

#### Examples
Get all defined workflows loaded in the workflow engine:
```bash
fission-workflows workflows get
```

Get all workflow invocations:
```bash
fission-workflows invocations get
```

Get a specific task execution in a specific 
```bash
fission-workflows invocations <invocation-id> <task-id>
```

Cancel a workflow invocation
```bash 
fission-workflows invocations cancel <invocation-id>
```
