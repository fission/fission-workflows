#!/bin/bash

set -e

# Run unit and integration tests, exclude dependencies
go test -race -v $(go list ./... | grep -v '/vendor/')
