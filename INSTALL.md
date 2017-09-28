# Installation

This document covers the installation of fission workflows.

### Prerequisites
Fission Workflows requires the following to be installed on the host machine:
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm](https://github.com/kubernetes/helm)
- [minikube](https://github.com/kubernetes/minikube) (in case of local deployments)

Additionally, Fission Workflows requires a [Fission](https://github.com/fission/fission-workflows) deployment.
If you do not have a deployment yet, follow [Fission's install guide](http://fission.io/docs/v0.2.1/install/) or follow the instructions below.

Note that Fission Workflows requires Fission 0.3.0 or higher.

### Installing Fission Workflow
Fission Workflows is just another Fission environment.
The environment requires only a single additional property `allowedFunctionsPerContainer` to be set to infinite, to ensure that workflows do not require a workflow environment each.
To deploy the environment run install the helm chart:
```bash

# If you haven't add the Fission repo
helm repo add fission-charts https://fission.github.io/fission-charts/
helm repo update

# Optional: Install Fission
helm install --namespace fission --set serviceType=NodePort -n fission-all fission-charts/fission-all --version 0.3.0-rc

# Install Helm package
helm install fission-charts/fission-workflows
```

### Experimental: Automated Installation
There is a deploy script that will manage setting up the deployment as long as the prerequisites are present on the host.
It uses minikube to create the cluster.
```bash
curl -Ls https://raw.githubusercontent.com/fission/fission-workflows/master/hack/deploy.sh | bash
```

You're good to go! Check out the [examples](./examples/). 
