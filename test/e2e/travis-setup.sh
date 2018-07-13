#!/bin/bash

set -eu

. $(dirname $0)/utils.sh

BIN_DIR=/tmp/fission-workflow-ci/bin
FISSION_VERSION=${FISSION_VERSION:-0.9.1}
fissionHelmId=fission
fissionWorkflowsHelmId=fission-workflows
NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder

if [ ! -d ${BIN_DIR} ]
then
    mkdir -p ${BIN_DIR}
fi

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
if ! cat ${HOME}/.kube/config >/dev/null 2>&1 | grep "current-context: gke_fission-ci_us-central1-a_fission-workflows-ci-1" ;
then
    emph "Connecting to gcloud cluster..."
    gcloud container clusters get-credentials fission-workflows-ci-1 --zone us-central1-a --project fission-ci
fi

# Is Kubernetes setup correctly?
if [ ! -f ${HOME}/.kube/config ]
then
    echo "Missing kubeconfig"
    exit 1
fi
kubectl get node

# Install helm in Kubernetes cluster
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

. $(dirname $0)/cleanup.sh

# Check if Fission
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
fission --version
emph "Fission deployed!"