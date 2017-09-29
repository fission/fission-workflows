#!/usr/bin/env bash

#
# release.sh - Generate all artifacts for a release
#

# wfcli
echo "Building wfcli..."
GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -o wfcli-osx github.com/fission/fission-workflows/cmd/wfcli/
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o wfcli-windows.exe github.com/fission/fission-workflows/cmd/wfcli/
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o wfcli-linux github.com/fission/fission-workflows/cmd/wfcli/

# Deployments
echo "Packaging chart..."
helm package ${GOPATH}/src/github.com/fission/fission-workflows/charts/fission-workflows

echo "Done!"
