#!/bin/bash

# Contains useful functions for testing (most copied from Fission's repo)

ROOT=$(dirname $0)/..

clean_tpr_crd_resources() {
    # TODO fix once resources are no longer pinned to default namespace
    # clean tpr & crd resources to avoid testing error (ex. no kind "HttptriggerList" is registered for version "fission.io/v1")
    # thirdpartyresources part should be removed after kubernetes test cluster is upgrade to 1.8+
    kubectl --namespace default get thirdpartyresources| grep -v NAME| grep "fission.io"| awk '{print $1}'|xargs -I@ bash -c "kubectl --namespace default delete thirdpartyresources @" || true
    kubectl --namespace default get crd| grep -v NAME| grep "fission.io"| awk '{print $1}'|xargs -I@ bash -c "kubectl --namespace default delete crd @"  || true
}

reset_fission_crd_resources() {
    kubectl --namespace default get crd| grep -v NAME| grep "fission.io"| awk '{print $1}'|xargs -I@ bash -c "kubectl --namespace default delete @ --all"  || true
}

helm_setup() {
    helm init
}

gcloud_login() {
    KEY=${HOME}/gcloud-service-key.json
    if [ ! -f ${KEY} ]
    then
	echo ${FISSION_CI_SERVICE_ACCOUNT} | base64 -d - > ${KEY}
    fi

    gcloud auth activate-service-account --key-file ${KEY}
}

generate_test_id() {
    if command -v md5sum 2>/dev/null >/dev/null; then
        # Linux
        date|md5sum|cut -c1-6
    else
        # Mac OS X
        date|md5|cut -c1-6
    fi
}

dump_all_resources_in_ns() {
    ns=$1

    echo "--- All objects in the namespace $ns ---"
    kubectl -n ${ns} get all
    echo "--- End objects in the namespace $ns ---"
}

helm_install_fission() {
    helmReleaseId=$1
    ns=$2
    version=$3
    helmVars=$4

    helm install		    \
	 --wait			        \
	 --timeout 600	        \
	 --name ${helmReleaseId}\
	 --set ${helmVars}	    \
	 --namespace ${ns}      \
	 --debug                \
	 --version ${version}   \
	 fission-charts/fission-all
}

helm_uninstall_release() {
    helm delete --debug --purge $1
}

helm_install_fission_workflows() {
    helmReleaseId=$1
    ns=$2
    helmVars=$3

    helm install		    \
	 --wait			        \
	 --timeout 600	        \
	 --name ${helmReleaseId}\
	 --set ${helmVars}	    \
	 --namespace ${ns}      \
	 --debug                \
	 ${ROOT}/charts/fission-workflows
}

run_all_tests() {
    local path=$1
    local failures=0

    for file in ${path}/test_*
    do
    TEST_UID=$(generate_test_id)
	echo "------- Running '${file}' -------"
	if ${file}
	then
        print_test_result ${file} 0
    else
        failures=$(($failures+1))
        print_test_result ${file} 1
	fi
    done
    echo "------- Completed '${file}' -------"
    return ${failures}
}

print_test_result() {
    TEST_NAME=$1
    STATUS=$2

    echo
    if (( STATUS == 0 )); then
        print_success "SUCCESS"
    else
        print_error "FAILURE"
    fi
    printf " | "
    printf "${TEST_NAME}\n"
}

# retry function adapted from:
# https://unix.stackexchange.com/questions/82598/how-do-i-write-a-retry-logic-in-script-to-keep-retrying-to-run-it-upto-5-times/82610
function retry {
  local n=1
  RETRY_LIMIT=${RETRY_LIMIT:-5}
  RETRY_DELAY=${RETRY_DELAY:-5}
  while true; do
    "$@" && break || {
      if [[ ${n} -lt ${RETRY_LIMIT} ]]; then
        ((n++))
        echo "Command '$@' failed. Attempt $n/${RETRY_LIMIT}:"
        sleep ${RETRY_DELAY};
      else
        >&2 echo "The command has failed after $n attempts."
        return 1;
      fi
    }
  done
}

print_error() {
    local NC='\033[0m' # No Color
    local RED='\033[0;31m'
    printf "$RED$1$NC"
}

print_success() {
    local NC='\033[0m' # No Color
    local GREEN='\033[0;32m'
    printf "$GREEN$1$NC"
}

emph() {
    local BLUE='\033[0;34m'
    local NC='\033[0m' # No Color
    printf "$BLUE$1$NC\n"
}

dump_fission_logs() {
    ns=$1
    fns=$2
    component=$3

    echo --- $component logs ---
    kubectl -n $ns get pod -o name  | grep $component | xargs kubectl -n $ns logs
    echo --- end $component logs ---
}


dump_fission_crd() {
    type=$1
    echo --- All objects of type $type ---
    kubectl --all-namespaces=true get $type -o yaml
    echo --- End objects of type $type ---
}

dump_fission_crds() {
    dump_fission_crd environments.fission.io
    dump_fission_crd functions.fission.io
    dump_fission_crd httptriggers.fission.io
    dump_fission_crd kuberneteswatchtriggers.fission.io
    dump_fission_crd messagequeuetriggers.fission.io
    dump_fission_crd packages.fission.io
    dump_fission_crd timetriggers.fission.io
}

dump_env_pods() {
    fns=$1

    echo --- All environment pods ---
    kubectl -n ${fns} get pod -o yaml
    echo --- End environment pods ---
}

dump_all_fission_resources() {
    ns=$1

    echo "--- All objects in the fission namespace $ns ---"
    kubectl -n ${ns} get all
    echo "--- End objects in the fission namespace $ns ---"
}


dump_system_info() {
    echo "--- System Info ---"
    echo "--- go ---"
    go version
    echo
    echo "--- python ---"
    python --version
    python3 --version
    echo
    echo "--- docker ---"
    docker version
    echo
    echo "--- kubectl ---"
    kubectl version
    echo
    echo "--- Helm ---"
    helm version
    echo
    echo "--- fission ---"
    fission -v
    echo
    echo "--- fission-workflows ---"
    fission-workflows version
    echo
    echo "--- End System Info ---"
}

dump_function_pod_logs() {
    ns=$1
    fns=$2

    functionPods=$(kubectl -n $fns get pod -o name)
    for p in $functionPods
    do
	echo "--- function pod logs $p ---"
	containers=$(kubectl -n $fns get $p -o jsonpath={.spec.containers[*].name})
	for c in $containers
	do
	    echo "--- function pod logs $p: container $c ---"
	    kubectl -n $fns logs $p $c
	    echo "--- end function pod logs $p: container $c ---"
	done
	echo "--- end function pod logs $p ---"
    done
}

dump_logs() {
    ns=$1
    fns=$2
    buildns=$3

    dump_all_fission_resources $ns
    dump_env_pods $fns
    dump_env_pods $buildns
    dump_fission_logs $ns $fns controller
    dump_fission_logs $ns $fns router
    dump_fission_logs $ns $fns buildermgr
    dump_fission_logs $ns $fns executor
    dump_function_pod_logs $ns $fns
    dump_fission_crds
    dump_system_info
}

cleanup_fission_workflows() {
    helmID=$1
    emph "Removing Fission Workflow deployment..."
    helm_uninstall_release ${helmID}
    # TODO remove workflows?
}
