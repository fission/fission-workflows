#!/usr/bin/env bash

gopath=`find pkg -type f -name "*.go" | xargs -n 1 dirname | sort --unique`
while read -r path; do
    (cd ${path} && go generate -x)
done <<< "$gopath"