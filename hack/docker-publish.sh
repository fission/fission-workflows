#!/usr/bin/env bash

#
# Usage: docker-publish.sh <docker-ns> <tag>
#

set -e

local_image_exists() { # image_exists <image> <tag>
    local image="${1}:${2}"
    if [[ "$(docker images -q ${image} 2> /dev/null)" == "" ]]; then
      # do something
      echo "Image not found in local docker repo: ${image}"
      exit 1;
    fi
}

# Params
if [[  $# -lt 2 ]]; then
    echo "Usage: docker-publish.sh <docker-ns> <tag>"
    exit 1;
fi
TAG=$2
NS=$1
echo "Using tag: ${TAG} and org ${NS}"
BUNDLE_IMAGE=${NS}/fission-workflows-bundle
ENV_IMAGE=${NS}/workflow-env

# Check if images with tags exist
local_image_exists ${BUNDLE_IMAGE} ${TAG}
local_image_exists ${ENV_IMAGE} ${TAG}

# Publish
read -p "Publish images with '${TAG}' to Dockerhub namespace '${NS}'? " -n 1 -r
echo    # (optional) move to a new line
if [[ $REPLY =~ ^[Yy]$ ]]
then
    # do dangerous stuff
    echo "Publishing ${BUNDLE_IMAGE}:${TAG}"
    docker push ${BUNDLE_IMAGE}:${TAG}
    echo "Publishing ${ENV_IMAGE}:${TAG}"
    docker push ${ENV_IMAGE}:${TAG}
fi
