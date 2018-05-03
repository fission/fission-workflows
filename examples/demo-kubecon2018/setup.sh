#!/bin/bash
# Run this in current process (aka source it)
set -xe

BUILD_TAG=latest
BUILD_NS=erwinvaneyk
FISSION_WORKFLOWS_DIR=$(dirname $0)/../..
# Ensure fresh minikube deployment
# NOTE: you need to set your docker to point to minikube's: `eval $(minikube docker-env)`
for i in {1..5}; do kubectl get pods && break || sleep 5; done

# Requirements
helm list || helm init --wait
for i in {1..5}; do helm list && break || sleep 5; done

# Install Fission
fission fn list || helm install --name fission --debug --namespace fission --set "analytics=false,serviceType=NodePort" https://github.com/fission/fission/releases/download/0.6.0/fission-all-0.6.0.tgz 
for i in {1..5}; do fission fn list && break || sleep 5; done

# Install newest version of fission workflows

helm install ${FISSION_WORKFLOWS_DIR}/charts/fission-workflows --namespace fission --set "pullPolicy=Always,tag=${BUILD_TAG},bundleImage=${BUILD_NS}/fission-workflows-bundle,envImage=${BUILD_NS}/workflow-env,buildEnvImage=${BUILD_NS}/workflow-build-env" --debug --wait -n fission-workflows
for i in {1..5}; do wfcli status && break || sleep 5; done

# Misc
fission env create --name binary --image fission/binary-env
