#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

protopaths=`find pkg -type f -name "*.proto"`
while read -r path; do
   echo "$path"
   protoc -I . \
        -I ${GOPATH}/src \
        -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
         --proto_path=.\
         --go_out=plugins=grpc:.\
         --grpc-gateway_out=logtostderr=true:. \
         ${path}
done <<< "$protopaths"