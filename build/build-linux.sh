#!/usr/bin/env bash

$(dirname $0)/build.sh --os linux --arch amd64 \
  --output-bundle "fission-workflows-bundle" \
  --output-cli "fission-workflows"