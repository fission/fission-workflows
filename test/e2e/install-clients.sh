#!/usr/bin/env bash

set -euo pipefail

. $(dirname $0)/utils.sh

BIN_DIR=${BIN_DIR:-/tmp/fission-workflow-ci/bin}
HELM_VERSION=${HELM_VERSION:-2.11.0}
FISSION_VERSION=${FISSION_VERSION:-0.10.0}

# Install kubectl
if ! kubectl version ; then
    sudo apt-get install -y apt-transport-https
    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
    sudo touch /etc/apt/sources.list.d/kubernetes.list
    echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list
    sudo apt-get update
    sudo apt-get install -y kubectl
fi
emph "Using kubectl $(kubectl version --client --short) already present."
mkdir -p ${HOME}/.kube
which kubectl

# Install helm client
if ! helm version -c 2>/dev/null | grep ${HELM_VERSION} >/dev/null; then
    emph "Installing Helm ${HELM_VERSION} to ${BIN_DIR}/helm..."
    curl -sLO https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz
    tar xzvf helm-*.tar.gz >/dev/null
    chmod +x linux-amd64/helm
    mv -f linux-amd64/helm ${BIN_DIR}/helm
else
    emph "Helm ${HELM_VERSION} already present."
fi
which helm

# Install Fission client
if ! fission --version 2>/dev/null | grep ${FISSION_VERSION} >/dev/null; then
    emph "Installing Fission ${FISSION_VERSION} to ${BIN_DIR}/fission..."
    curl -sLo fission https://github.com/fission/fission/releases/download/${FISSION_VERSION}/fission-cli-linux
    chmod +x fission
    mv -f fission ${BIN_DIR}/fission
else
    emph "Fission ${FISSION_VERSION} already present."
fi
which fission

emph "Clients installed in ${BIN_DIR}"