# Installation

This document covers the installation of Fission Workflows.

### Prerequisites

Fission Workflows requires the following to be installed on the host machine:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm](https://github.com/kubernetes/helm)

Additionally, Fission Workflows requires a
[Fission](https://github.com/fission/fission) deployment on your
Kubernetes cluster.  If you do not have a Fission deployment, follow
[Fission's installation guide](http://fission.io/docs/0.3.0/install/).

(Note that Fission Workflows requires Fission 0.3.0 or higher.)

### Installing Fission Workflows

Fission Workflows is an add-on to Fission.  You can install both
Fission and Workflows using helm charts.

Assuming you have a Kubernetes cluster, run the following commands:

```bash

# Add the Fission repo
helm repo add fission-charts https://fission.github.io/fission-charts/
helm repo update

# Install Fission (unless you already have)
helm install --namespace fission --set serviceType=NodePort -n fission-all fission-charts/fission-all --version 0.3.0

# Install Fission Workflows
helm install fission-charts/fission-workflows

```

### Creating your first workflow

After installing Fission and Workflows, you're ready to run a simple
test workflow.

```bash


```
