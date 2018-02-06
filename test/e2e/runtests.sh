#!/usr/bin/env bash

# Run tests against current deployment
# Use `test/e2e/runtests.sh 2>&1 | grep '\(FAILURE\|SUCCESS\).*|'` to get a short report
# TODO differentiate between clusters

source $(dirname $0)/utils.sh
TEST_DIR=$(dirname $0)/tests

run_all_tests ${TEST_DIR}