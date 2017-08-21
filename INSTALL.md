# Installation

## Setting up Fission (temporary)
As of the moment of writing, the prototype of the Fission Workflow engine has been implemented in Fission using a couple of shortcuts.
In the coming weeks, Fission Workflow will be implemented to fully conform to the Fission Environment API, removing the need for any special modfications to Fission.

```bash
# clone or add a remote to git@github.com:erwinvaneyk/fission.git

# Switch to the branch that contains the Fission Workflow integration
git checkout <remote> fission-workflow-integration

# Follow the [guide on compiling Fission](https://github.com/fission/fission/blob/master/Compiling.md)

# Compile and push the fission-bundle to the local Docker repo
(cd fission-bundle/ && bash ./push.sh)

# Deploy fission (assuming that a cluster is available) including the NATS plugin
kubectl create -f fission.yaml fission-nodeport.yaml fission-nats.yaml

# Setup the Fission env (assuming Minikube)
export FISSION_URL=http://$(minikube ip):31313
export FISSION_ROUTER=$(minikube ip):31314
```

## Installing Fission Workflow
Currently, the only way of deploying Fission Workflow is by [compiling it yourself](./Docs/compiling.md).
The remainder of this section assumes that Fission Workflow has been compiled and is available in the local Docker registry.

```bash
(cd ./build/ && kubectl create -f fission-workflow.yaml)
```

You're good to go! Check out the [examples](./examples/).
