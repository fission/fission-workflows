# Compiling

*You only need to do this if you're making Fission Workflows changes; if you're
just deploying Fission Workflows, use the helm chart which points to prebuilt
images by default.*

## Compilation
There are two ways to compiling the environment: locally or in Docker. Regardless of the approach, ensure that your 
environment meets all prerequisite requirements, and checkout the repo from github.

### Local Compilation

#### Requirements
- go >1.8
- docker

```bash
# Build the artifacts: client (fission-workflows) and server (fission-workflows-bundle)
build/build-linux.sh

# Build the docker images (the NOBUILD parameter indicates that Docker should use the artifacts (wfci, 
# fission-workflows-bundle) you just build with build/build-linux.sh)
NOBUILD=y build/docker.sh fission latest
```

### In-Docker Compilation

THe in-Docker compilation approach builds all artifacts in Docker containers. The advantage of this is that builds 
are not affected by the environment differences of your local machine and limits the need to install build tooling on
your machine (except for Docker). This comes at the cost of performance: builds are started completely clean  
every time; dependencies, intermediate build steps and artifacts are not cached. 

While the local compilation is most convenient for developing and testing Fission Workflows, In-Docker compilation 
should be used for building the official/final images. 

```bash
# Build the docker images (the absence of the NOBUILD parameter indicates that Docker should first build the artifacts 
# in a Docker container)
build/docker.sh
```

## Deployment

To deploy your locally compiled version. **As of writing Fission Workflows, requires fission to be installed 
in the fission namespace.**
```bash
helm install --set "tag=latest" --namespace fission charts/fission-workflows
```

## Optional:  CLI

There is an experimental CLI available, called `fission-workflows`.
The intent is to integrate it into the Fission CLI, removing the need for the separate CLI.
```bash
go install github.com/fission/fission-workflows/cmd/fission-workflows/
```
