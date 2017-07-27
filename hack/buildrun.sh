#!/usr/bin/env bash

set -ex

hack/codegen-grpc.sh
hack/codegen-bazel.sh
hack/format.sh

bazel build //cmd/workflow-engine

bazel-bin/cmd/workflow-engine/workflow-engine