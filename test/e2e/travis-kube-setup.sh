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

# Get kubectl binary
if ! command -v kubectl >/dev/null 2>&1; then
   curl -sLO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
   chmod +x ./kubectl
   mv kubectl ${BIN_DIR}/kubectl
fi
mkdir ${HOME}/.kube

# Get helm binary
if ! command -v helm >/dev/null 2>&1 ; then
    curl -sLO https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz
    tar xzvf helm-*.tar.gz
    chmod +x linux-amd64/helm
    mv linux-amd64/helm ${BIN_DIR}/helm
fi
export PATH=$HOME/testbin:${PATH}

# Get and install gcloud
# https://stackoverflow.com/questions/38762590/how-to-install-google-cloud-sdk-on-travis
#if [ ! -d "${HOME}/google-cloud-sdk/bin" ]
#then
#    rm -rf $HOME/google-cloud-sdk
#    export CLOUDSDK_CORE_DISABLE_PROMPTS=1
#    curl https://sdk.cloud.google.com | bash
#fi
#export PATH=${HOME}/google-cloud-sdk/bin:${PATH}

# get gcloud credentials
#echo ${FISSION_CI_SERVICE_ACCOUNT} | base64 -D - > ${HOME}/gcloud-service-key.json
gcloud auth activate-service-account --key-file <(echo ${FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT} | base64 -d)
unset FISSION_CI_SERVICE_ACCOUNT

# get kube config
# TODO separate cluster for fission-workflows?
gcloud container clusters get-credentials fission-ci-1 --zone us-central1-a --project fission-ci

# ensure we have the gcloud binary
# DEBUG
#echo ${KUBECONFIG_CONTENTS} | base64 -D - > ${HOME}/.kube/config
kubectl version
gcloud version
helm version
# END DEBUG

# does it work?
if [ ! -f ${HOME}/.kube/config ]
then
    echo "Missing kubeconfig"
    exit 1
fi
kubectl get node