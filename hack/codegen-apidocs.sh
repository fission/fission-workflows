#!/bin/bash

set -ex

swagger-codegen generate -i ./pkg/apiserver/apiserver.swagger.json -l html -o ./Docs/api/
