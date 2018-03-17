#!/usr/bin/env bash

set -e

mkdir -p cmd/wfcli/swagger-client/
swagger generate client -f api/swagger/apiserver.swagger.json -t cmd/wfcli/swagger-client/