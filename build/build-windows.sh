#!/bin/sh

GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o fission-workflows-bundle-windows.exe github.com/fission/fission-workflows/cmd/fission-workflows-bundle/
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o wfcli-windows.exe github.com/fission/fission-workflows/cmd/wfcli/
