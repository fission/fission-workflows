#!/bin/bash

#
# Download kubectl, save kubeconfig, and ensure we can access the test cluster.
# Copied from https://github.com/fission/fission/blob/master/hack/travis-kube-setup.sh
#

set -e

# If we don't have gcloud credentials, bail out of this setup.
if [ -z "$FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT" ]
then
    echo "Skipping tests, no cluster credentials"
    exit 0
fi

BIN_DIR=${HOME}/testbin
HELM_VERSION=v2.7.2

if [ ! -d ${BIN_DIR} ]
then
    mkdir -p ${BIN_DIR}
fi
export PATH=$HOME/testbin:${PATH}

# Get kubectl binary
if ! command -v kubectl >/dev/null 2>&1; then
   curl -sLO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
   chmod +x ./kubectl
   mv -f kubectl ${BIN_DIR}/kubectl
fi
mkdir ${HOME}/.kube

# Get helm binary
if ! command -v helm >/dev/null 2>&1 ; then
    curl -sLO https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz
    tar xzvf helm-*.tar.gz
    chmod +x linux-amd64/helm
    mv -f linux-amd64/helm ${BIN_DIR}/helm
fi

# Get Fission binary
curl -sLo fission https://github.com/fission/fission/releases/download/0.6.0/fission-cli-linux
chmod +x fission
mv -f fission ${BIN_DIR}/fission

# get gcloud credentials
gcloud auth activate-service-account --key-file <(echo ${FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT} | base64 -d)
curl -sLO https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v1.4.3/docker-credential-gcr_linux_amd64-1.4.3.tar.gz
tar xzvf docker-credential-*.tar.gz
mv docker-credential-gcr ${BIN_DIR}/docker-credential-gcr
docker-credential-gcr configure-docker
unset FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT

# get kube config
gcloud container clusters get-credentials fission-workflows-ci-1 --zone us-central1-a --project fission-ci

# ensure we have the binaries
# DEBUG
kubectl version
gcloud version
helm version -c
# fission -v - Do not test because it expects a Fission cluster to be present
# END DEBUG

# does it work?
if [ ! -f ${HOME}/.kube/config ]
then
    echo "Missing kubeconfig"
    exit 1
fi
kubectl get node