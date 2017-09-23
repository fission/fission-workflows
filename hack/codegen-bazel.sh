#!/usr/bin/env bash

set -ex

gazelle -external vendored -go_prefix github.com/fission/fission-workflows ./pkg ./cmd

