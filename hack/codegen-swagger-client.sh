#!/usr/bin/env bash

set -e

mkdir -p cmd/wfcli/swagger-client/
docker run --rm -it -e GOPATH=$HOME/go:/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger:0.11.0 generate client
-f api/swagger/apiserver.swagger.json -t cmd/wfcli/swagger-client/
