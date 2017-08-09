#!/bin/bash

set -ex

# API server (generate HTTP gateway + swagger)
echo "Generating golang HTTP gateway and swagger definition for proto-file: pkg/apiserver/apiserver.proto"
protoc -I . \
        -I/usr/local/include \
        -I ./pkg/ \
        -I $GOPATH/src \
        -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
        --swagger_out=logtostderr=true:. \
        pkg/apiserver/apiserver.proto

mkdir -p api/swagger/
mv pkg/apiserver/apiserver.swagger.json api/swagger/

swagger-codegen generate -i api/swagger/apiserver.swagger.json -l html -o ./Docs/api/
