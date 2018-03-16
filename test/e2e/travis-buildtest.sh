#!/usr/bin/env bash

# Run e2e tests using minikube

set -euo pipefail

. $(dirname $0)/utils.sh
ROOT=$(dirname $0)/../..

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

# Build and run tests
. $(dirname $0)/buildtest.sh