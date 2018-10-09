#!/usr/bin/env bash

$(dirname $0)/build.sh --os darwin --arch 386 \
  --output-bundle "fission-workflows-bundle-osx" \
  --output-cli "fission-workflows-osx" \
  --output-proxy "fission-workflows-proxy-osx"
