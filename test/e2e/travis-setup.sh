#!/bin/bash

set -euo pipefail

. $(dirname $0)/utils.sh

BIN_DIR=/tmp/fission-workflow-ci/bin
HELM_VERSION=2.8.2
KUBECTL_VERSION=1.9.6
FISSION_VERSION=0.6.0
fissionHelmId=fission
NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder

if [ ! -d ${BIN_DIR} ]
then
    mkdir -p ${BIN_DIR}
fi
export PATH=${BIN_DIR}:${PATH}

# Get kubectl binary
if ! kubectl version 2>/dev/null | grep ${KUBECTL_VERSION} >/dev/null; then
   emph "Installing kubectl ${KUBECTL_VERSION}..."
   curl -sLO https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl
   chmod +x ./kubectl
   mv -f kubectl ${BIN_DIR}/kubectl
else
    emph "Kubectl ${KUBECTL_VERSION} already present."
fi
mkdir -p ${HOME}/.kube
kubectl version

# Get helm binary
if ! helm version 2>/dev/null | grep ${HELM_VERSION} >/dev/null; then
    emph "Installing Helm ${HELM_VERSION}..."
    curl -sLO https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz
    tar xzvf helm-*.tar.gz
    chmod +x linux-amd64/helm
    mv -f linux-amd64/helm ${BIN_DIR}/helm
else
    emph "Helm ${HELM_VERSION} already present."
fi
helm version


# Setup gcloud CI
if [ -z "${FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT:-}" ]
then
    emph "No CI credentials provided. Assuming that gcloud has been authenticated already."
else
    emph "Setting up gcloud with provided CI credentials..."
    # get gcloud credentials
    gcloud auth activate-service-account --key-file <(echo ${FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT} | base64 -d)
    curl -sLO https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v1.4.3/docker-credential-gcr_linux_amd64-1.4.3.tar.gz
    tar xzvf docker-credential-*.tar.gz
    mv docker-credential-gcr ${BIN_DIR}/docker-credential-gcr
    docker-credential-gcr configure-docker
    unset FISSION_WORKFLOWS_CI_SERVICE_ACCOUNT
fi

# get kube config
gcloud container clusters get-credentials fission-workflows-ci-1 --zone us-central1-a --project fission-ci

# Get Fission binary
if ! fission --version 2>/dev/null | grep ${FISSION_VERSION} >/dev/null; then
    emph "Installing Fission ${FISSION_VERSION}..."
#    curl -sLo fission https://github.com/fission/fission/releases/download/${FISSION_VERSION}/fission-cli-linux
#    chmod +x fission
#    mv -f fission ${BIN_DIR}/fission
else
    emph "Fission ${FISSION_VERSION} already present."
fi
#fission --version

# Is Kubernetes setup correctly?
if [ ! -f ${HOME}/.kube/config ]
then
    echo "Missing kubeconfig"
    exit 1
fi
kubectl get node


# Install helm
echo "Setting up helm..."
helm init

printf "Waiting for Helm"
until helm list >/dev/null 2>&1
do
  printf "."
  sleep 3
done
printf "\n"

# Ensure that fission-charts helm repo is added
if ! helm repo list | grep fission-charts >/dev/null 2>&1 ; then
    echo "Setting up fission-charts Helm repo..."
    helm repo add fission-charts https://fission.github.io/fission-charts/
    printf "Updating repo"
    until helm fetch fission-charts/fission-all >/dev/null 2>&1
    do
        printf "."
    done
    printf "\n"
fi
helm repo update

# Check if Fission
if ! helm list | grep fission-all-${FISSION_VERSION} ; then
    # Clean up existing
    emph "Removing existing Fission and Fission Workflow deployments..."
    helm_uninstall_release ${fissionWorkflowsHelmId} &
    helm_uninstall_release ${fissionHelmId} &

    emph "Removing custom resources..."
    clean_tpr_crd_resources || true

    # Trigger deletion of all namespaces before waiting - for concurrency of deletion
    emph "Forcing deletion of namespaces..."
    kubectl delete ns/${NS} > /dev/null 2>&1 & # Sometimes it is not deleted by helm delete
    kubectl delete ns/${NS_BUILDER} > /dev/null 2>&1 & # Sometimes it is not deleted by helm delete
    kubectl delete ns/${NS_FUNCTION} > /dev/null 2>&1 & # Sometimes it is not deleted by helm delete

    # Wait until all namespaces are actually deleted!
    emph "Awaiting deletion of namespaces..."
    retry kubectl delete ns/${NS} 2>&1  | grep -qv "Error from server (Conflict):"
    retry kubectl delete ns/${NS_BUILDER} 2>&1 | grep -qv "Error from server (Conflict):"
    retry kubectl delete ns/${NS_FUNCTION} 2>&1  | grep -qv "Error from server (Conflict):"

    emph "Cleaning up local filesystem..."
    rm -f ./fission-workflows-bundle ./wfcli
    sleep 5

    # Deploy Fission
    # TODO use test specific namespace
    emph "Deploying Fission: helm chart '${fissionHelmId}' in namespace '${NS}'..."
    controllerPort=31234
    routerPort=31235
    helm_install_fission ${fissionHelmId} ${NS} ${FISSION_VERSION} \
        "controllerPort=${controllerPort},routerPort=${routerPort},pullPolicy=Always,analytics=false"

    # Wait for Fission to get ready
    emph "Waiting for fission to be ready..."
    sleep 5
    retry fission fn list
    echo
    emph "Fission deployed!"
else
    emph "Reusing existing Fission ${FISSION_VERSION} deployment"
    emph "Removing custom resources..."
    clean_tpr_crd_resources || true
fi
