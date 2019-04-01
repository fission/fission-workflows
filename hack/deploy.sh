#!/usr/bin/env bash

#
# deploy.sh - (Almost) automatic setup of a Fission Workflows deployment
#

# Configs
FISSION_VERSION=${FISSION_VERSION:-1.1.0}
WORKFLOWS_VERSION=${WORKFLOWS_VERSION:-0.6.0}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Fission Workflows Deploy Script v1.4"

source ${DIR}/minikube.sh

# Install Fission
if ! fission fn list >/dev/null 2>&1 ; then
    echo "Installing Fission ${FISSION_VERSION}..."
    if ! helm install --namespace fission --set "serviceType=NodePort,pullPolicy=IfNotPresent,analytics=false"  -n fission-all fission-charts/fission-all --wait --version ${FISSION_VERSION} ; then
        echo "Failed to install fission"
        exit 0
    fi
fi

sleep 10

# Install Fission Workflows
if ! fission env get --name workflow >/dev/null 2>&1 ; then
    echo "Installing Fission Workflows ${WORKFLOWS_VERSION}..."
    if [[ -z "${WORKFLOWS_VERSION// }" ]] ; then
        helm install --namespace fission -n fission-workflows fission-charts/fission-workflows --version ${WORKFLOWS_VERSION} --wait
    else
        helm install --namespace fission -n fission-workflows fission-charts/fission-workflows --wait
    fi
fi

# Output debug logs
echo "---------- Debug ----------"
minikube version
printf "K8S "
kubectl version --client --short
printf "K8S "
kubectl version --short | grep Server
printf "Helm "
helm version --short -c
printf "Helm "
helm version --short -s
echo "Fission: ${FISSION_VERSION}"
echo "Fission Workflows: ${WORKFLOWS_VERSION}"
echo "---------------------------"
