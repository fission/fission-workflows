#!/bin/bash

set -e

# Install test dependencies if requested
if [ -z "$1" ]; then
    go test -v -i $(go list ./... | grep -v '/vendor/')
fi
# Run unit and integration tests, exclude dependencies
go test -race -v $(go list ./... | grep -v '/vendor/')
