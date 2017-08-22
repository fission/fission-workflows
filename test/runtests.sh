#!/bin/bash

set -e

# Run unit and integration tests, exclude dependencies
go test -v -i $(go list ./... | grep -v '/vendor/' )
go test -v $(go list ./... | grep -v '/vendor/' )
