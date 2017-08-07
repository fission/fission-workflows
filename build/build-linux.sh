#!/bin/sh
GOOS=linux GOARCH=386 go build github.com/fission/fission-workflow/cmd/workflow-engine/
