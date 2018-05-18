#!/usr/bin/env bash

$(dirname $0)/build.sh --os windows --arch 386 \
  --output-bundle "fission-workflows-bundle-windows" \
  --output-cli "wfcli-windows"
