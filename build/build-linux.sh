#!/bin/sh

rm -f ./fission-workflows-bundle ./wfcli
GOOS=linux GOARCH=386 go build github.com/fission/fission-workflows/cmd/fission-workflows-bundle/
GOOS=linux GOARCH=386 go build github.com/fission/fission-workflows/cmd/wfcli/