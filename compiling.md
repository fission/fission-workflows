# Compiling

*You only need to do this if you're making Fission Workflows changes; if you're
just deploying Fission Workflows, use the helm chart which points to prebuilt
images by default.*

## Requirements
- go >1.8
- docker
- [glide](http://glide.sh/) package manager

## Compilation
Ensure that your environment meets all prerequisite requirements, and checkout the repo from github.

```bash
# Install dependencies
glide install

cd build
bash ./build-linux.sh

# Optional: Ensure that you target the right docker registry (assuming minikube)
eval $(minikube docker-env)

# Build the docker images
bash ./docker.sh fission latest
```

To deploy your locally compiled version. **As of writing Fission Workflows, requires fission to be installed 
in the fission namespace.**
```bash
helm install --set "tag=latest" --namespace fission charts/fission-workflows
```

### CLI
There is an experimental CLI available, called `wfcli`.
The intent is to integrate it into the Fission CLI, removing the need for the separate CLI.
```bash
go install github.com/fission/fission-workflows/cmd/wfcli/
```
