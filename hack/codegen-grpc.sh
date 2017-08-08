#!/bin/bash

set -e

protopaths=`find pkg -type f -name "*.proto"`
while read -r path; do
   echo "Generating golang implementations for proto-file: $path"
   protoc -I . \
        -I /usr/local/include \
        -I ./pkg/ \
        -I $GOPATH/src \
        -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
         --proto_path=.\
         --go_out=plugins=grpc:.\
         ${path}
done <<< "$protopaths"

echo "Generating golang HTTP gateway for proto-file: pkg/apiserver/apiserver.proto"
protoc -I . \
        -I/usr/local/include \
        -I ./pkg/ \
        -I $GOPATH/src \
        -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
        --grpc-gateway_out=logtostderr=true:. \
        --go_out=plugins=grpc:. \
        pkg/apiserver/apiserver.proto

