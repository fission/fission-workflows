#!/bin/bash

# Contains useful functions for testing (most copied from Fission's repo)

ROOT=$(dirname $0)/..

dump_system_info() {
    echo "--- System Info ---"
    echo "--- go ---"
    go version
    echo "--- docker ---"
    docker version
    echo "--- kubectl ---"
    kubectl version
    echo "--- Helm ---"
    helm version
    echo "--- fission ---"
    fission
    echo "--- End System Info ---"
}

dump_logs() {
    dump_system_info
}

clean_tpr_crd_resources() {
    # clean tpr & crd resources to avoid testing error (ex. no kind "HttptriggerList" is registered for version "fission.io/v1")
    # thirdpartyresources part should be removed after kubernetes test cluster is upgrade to 1.8+
    kubectl --namespace default get thirdpartyresources| grep -v NAME| grep "fission.io"| awk '{print $1}'|xargs -I@ bash -c "kubectl --namespace default delete thirdpartyresources @" || true
    kubectl --namespace default get crd| grep -v NAME| grep "fission.io"| awk '{print $1}'|xargs -I@ bash -c "kubectl --namespace default delete crd @"  || true
}

helm_setup() {
    helm init
}

gcloud_login() {
    KEY=${HOME}/gcloud-service-key.json
    if [ ! -f $KEY ]
    then
	echo $FISSION_CI_SERVICE_ACCOUNT | base64 -d - > $KEY
    fi

    gcloud auth activate-service-account --key-file $KEY
}

generate_test_id() {
    echo $(date |md5 | head -c8; echo)
}

set_environment() {
    id=$1
    ns=f-$id

    export FISSION_URL=http://$(kubectl -n $ns get svc controller -o jsonpath='{...ip}')
    export FISSION_ROUTER=$(kubectl -n $ns get svc router -o jsonpath='{...ip}')
}

run_all_tests() {
    id=$1

    export FISSION_NAMESPACE=f-$id
    export FUNCTION_NAMESPACE=f-func-$id

    for file in $ROOT/test/tests/test_*.sh
    do
	echo ------- Running $file -------
	if $file
	then
	    echo SUCCESS: $file
	else
	    echo FAILED: $file
	    export FAILURES=$(($FAILURES+1))
	fi
    done
}

dump_all_resources_in_ns() {
    ns=$1

    echo "--- All objects in the namespace $ns ---"
    kubectl -n $ns get all
    echo "--- End objects in the namespace $ns ---"
}

helm_install_fission() {
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
	 https://github.com/fission/fission/releases/download/0.4.0/fission-all-0.4.0.tgz
	 # TODO pick the latest release
}

helm_uninstall_release() {
    if [ ! -z ${FISSION_TEST_SKIP_DELETE:+} ]
    then
	echo "Fission uninstallation skipped"
	return
    fi
    echo "Uninstalling $1"
    helm delete --purge $1
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
    path=$1

    for file in ${path}/test_*.sh
    do
	echo ------- Running $file -------
	if . $file
	then
	    echo SUCCESS: $file
	else
	    echo FAILED: $file
	    export FAILURES=$(($FAILURES+1))
	fi
    done
}