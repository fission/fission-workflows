#!/bin/sh

GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -o wfcli-osx github.com/fission/fission-workflows/cmd/wfcli/
GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -o fission-workflows-bundle-osx github.com/fission/fission-workflows/cmd/fission-workflows-bundle/
