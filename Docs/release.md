# Release Process

This document describes the process of releasing a new version of this repository.
Replace `$VERSION` with the target version. In the future a full script will be provided 
to automate this, in which case this document serves as documentation.

## 1. Prepare the codebase
1. From an up-to-date master, create a new branch: `git checkout -b $VERSION`
2. Replace all references in the codebase to old version to new version.
3. Add compatibility documentation to `INSTALL.md`.
4. Run all code-generation scripts: `make generate` or run all `hack/codegen-*` scripts.
5. Generate the version package: `make version VERSION=$VERSION`
6. Generate a changelog: `make changelog GITHUB_TOKEN=$GITHUB_TOKEN VERSION=$VERSION`
7. Create a PR of the changes originated from the prior steps.
8. Merge PR after CI succeeds.
9. Fetch and checkout new master

## 2. Prepare the artifacts
1. Build the images `build/docker.sh fission $VERSION`
2. Build the release artifacts (binaries and helm chart): `hack/release.sh` 

## 3. Publish the release
1. Publish images to docker under `$VERSION`: `hack/docker-publish.sh fission $VERSION`
2. Publish images to docker under `latest`: `hack/docker-publish.sh fission latest`
3. Add Helm chart to the `fission/fission-charts` repo. Follow instructions there.
4. Create a release on Github, and include:
  - Changelog of the version
  - All generated artifacts (except for the docker images)
5. Update the docs with new version: `https://github.com/fission/docs.fission.io`

## 4. Verify the release
1. Prepare a new Kubernetes cluster 
2. Install the latest version of Fission
3. Install the `$VERSION` of Fission Workflows
4. Run the end-to-end tests on the cluster: `test/e2e/runtests.sh`
