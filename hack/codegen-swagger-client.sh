#!/usr/bin/env bash


mkdir cmd/wfcli/swagger-client/
swagger generate client -f api/swagger/apiserver.swagger.json -t cmd/wfcli/swagger-client/
