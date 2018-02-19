#!/usr/bin/env bash

#
# minikube.sh - sets up a minikube cluster along with all required components (like Helm)
#
KUBERNETES_VERSION=v1.8.0
MINIKUBE_MEMORY=4096
FISSION_CHARTS_HELM_REPO=https://fission.github.io/fission-charts/

# Check if kubectl is installed
if ! command -v kubectl >/dev/null 2>&1; then
    echo "kubectl is not installed"
    exit 1;
fi

# Check if minikube is installed
if ! command -v minikube >/dev/null 2>&1 ; then
    echo "minikube is not installed."
    exit 1;
fi

# Check if helm is installed
if ! command -v helm >/dev/null 2>&1 ; then
    echo "helm is not installed."
    exit 1;
fi

# Ensure that minikube is running
if ! minikube ip >/dev/null 2>&1 ; then
    echo "Starting minikube-based Kubernetes ${KUBERNETES_VERSION} cluster..."
    if ! minikube start --kubernetes-version ${KUBERNETES_VERSION} --memory ${MINIKUBE_MEMORY} ; then
        echo "Failed to setup minikube cluster."
        exit 1;
    fi
fi

# Use docker registry
eval $(minikube docker-env)

# Install helm on cluster
if ! helm list >/dev/null 2>&1 ; then
    echo "Installing helm..."
    helm init

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
