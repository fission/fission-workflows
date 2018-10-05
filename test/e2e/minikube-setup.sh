#!/bin/bash

set -eu

. $(dirname $0)/utils.sh

BIN_DIR=/tmp/fission-workflow-ci/bin
FISSION_VERSION=${FISSION_VERSION:-0.11.0}
fissionHelmId=fission
fissionWorkflowsHelmId=fission-workflows
NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder

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
fission version
emph "Fission deployed!"