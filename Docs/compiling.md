# Compiling

## Requirements
- [glide](http://glide.sh/)

## Compilation
Ensure that all requirements are present, and checkout the repo from github.

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
