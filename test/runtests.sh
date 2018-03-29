#!/bin/bash

set -e

# Install test dependencies if requested
go test -v -i $(go list ./... | grep -v '/vendor/')

# Run unit and integration tests, exclude dependencies
go test -race -v $(go list ./... | grep -v '/vendor/') "$@"
