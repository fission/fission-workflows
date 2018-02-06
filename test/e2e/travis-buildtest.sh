#!/usr/bin/env bash

# Run e2e tests using minikube

set -euo pipefail

. $(dirname $0)/utils.sh
ROOT=$(dirname $0)/../..

# Install helm on cluster
if ! helm list >/dev/null 2>&1 ; then
    echo "Installing helm..."
    kubectl -n kube-system create sa tiller

    kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller

    helm init --service-account tiller

    printf "Waiting for Helm"
    until helm list >/dev/null 2>&1
    do
      printf "."
      sleep 3
    done
    printf "\n"
fi

# Ensure that fission-charts helm repo is added
if ! helm repo list | grep fission-charts >/dev/null 2>&1 ; then
    echo "Setting up fission-charts Helm repo..."
    helm repo add fission-charts https://fission.github.io/fission-charts/
    printf "Updating repo"
    until helm fetch fission-charts/fission-all >/dev/null 2>&1
    do
        printf "."
        helm repo update
    done
    printf "\n"
fi

# Build and run tests
. ${ROOT}/test/e2e/buildtest.sh

# TODO is cluster cleaned up?