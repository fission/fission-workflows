# Compiling

## Requirements
- go >1.8
- docker
- [glide](http://glide.sh/) package manager

## Compilation
Ensure that all requirements are met, and checkout the repo from github.

```bash
# Install dependencies
glide install

cd build
bash ./build-linux.sh

# Optional: Ensure that you target the right docker registry (assuming minikube)
eval $(minikube docker-env)

# Build the docker image
bash ./docker.sh
```

### CLI
There is an experimental CLI available, called `wfcli`.
The intent is to integrate it into the Fission CLI, removing the need for the separate CLI.
```bash
go install github.com/fission/fission-workflows/cmd/wfcli/
```
