#!/bin/sh
#
# Build command run by default.
# TODO allow parsing of multiple workflow definitions
#

# TODO loop through directory to support arbitrary filenames
# TODO check if fission builder supports single files
echo "Parsing ${SRC_PKG}/wf.yaml to ${DEPLOY_PKG}/wf.json"

if ! wfcli parse ${SRC_PKG}/wf.yaml ; then
    echo "Failed to parse ${SRC_PKG}/wf.yaml"
    exit 12
fi

mkdir -p ${DEPLOY_PKG}
cp -r ${SRC_PKG}/wf.json ${DEPLOY_PKG}/wf.json

echo "Finished build."