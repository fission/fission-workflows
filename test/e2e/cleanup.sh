#!/usr/bin/env bash


set -euo pipefail

. $(dirname $0)/utils.sh

NS=fission
NS_FUNCTION=fission-function
NS_BUILDER=fission-builder
fissionHelmId=fission
fissionWorkflowsHelmId=fission-workflows
TEST_STATUS=0
TEST_LOGFILE_PATH=tests.log
BIN_DIR="${BIN_DIR:-$HOME/testbin}"

cleanup() {
    emph "Removing Fission and Fission Workflow deployments..."
    helm_uninstall_release ${fissionWorkflowsHelmId} &
    helm_uninstall_release ${fissionHelmId} &

    emph "Removing custom resources..."
    clean_tpr_crd_resources || true

    # Trigger deletion of all namespaces before waiting - for concurrency of deletion
    emph "Forcing deletion of namespaces..."
    kubectl delete ns/${NS} > /dev/null 2>&1 & # Sometimes it is not deleted by helm delete
    kubectl delete ns/${NS_BUILDER} > /dev/null 2>&1 & # Sometimes it is not deleted by helm delete
    kubectl delete ns/${NS_FUNCTION} > /dev/null 2>&1 & # Sometimes it is not deleted by helm delete

    # Wait until all namespaces are actually deleted!
    emph "Awaiting deletion of namespaces..."
    retry kubectl delete ns/${NS} 2>&1  | grep -qv "Error from server (Conflict):"
    retry kubectl delete ns/${NS_BUILDER} 2>&1 | grep -qv "Error from server (Conflict):"
    retry kubectl delete ns/${NS_FUNCTION} 2>&1  | grep -qv "Error from server (Conflict):"

    emph "Cleaning up local filesystem..."
    rm -f ./fission-workflows-bundle ./wfcli
    sleep 5
}

print_report() {
    emph "--- Test Report ---"
    if ! cat ${TEST_LOGFILE_PATH} | grep '\(FAILURE\|SUCCESS\).*|' ; then
        echo "No report found."
    fi
    emph "--- End Test Report ---"
}

on_exit() {
    emph "[Buildtest exited]"
    # Dump all the logs
    dump_logs ${NS} ${NS_FUNCTION} ${NS_BUILDER} || true

    # Ensure teardown after tests finish
    # TODO provide option to not cleanup the test setup after tests (e.g. for further tests)
    emph "Cleaning up cluster..."
    retry cleanup

    # Print a short test report
    print_report

    # Ensure correct exist status
    echo "TEST_STATUS: ${TEST_STATUS}"
    if [ ${TEST_STATUS} -ne 0 ]; then
        exit 1
    fi
}

# Ensure printing of report
on_exit
