#!/usr/bin/env bash

# Miscellaneous static code checks

# check <msg> <cmd...> - runs the command and adds a header with the status (OK or FAIL) to the output
check() {
    msg=$1
    shift
    cmd=$@
    printf "[check] ${msg}..."
    output=""
    if output=$(bash -c "${cmd}") ; then
        printf "OK\n"
    else
        printf "FAIL\n"
        if [ ! -z "${output}" ] ; then
            echo ${output}
        fi
        return 1
    fi
}

# Check if we don't accidentally use the gogo protobuf implementation, instead of the golang protobuf implementation.
check "no use of gogo-protobuf" ! grep -R 'github.com/gogo/protobuf' pkg/ cmd/