#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

# Constants
versionPath="github.com/fission/fission-workflows/pkg/version"

# Handle arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -d| --date)
      date=$2
      shift
      ;;
    -c| --commit)
      gitcommit=$2
      shift
      ;;
    -v| --version)
      version=$2
      shift
      ;;
    --os)
      goos=$2
      shift
      ;;
    --arch)
      goarch=$2
      shift
      ;;
    --output-bundle)
      output_bundle=$2
      shift
      ;;
    --output-cli)
      output_cli=$2
      shift
      ;;
    --output-cli)
      output_proxy=$2
      shift
      ;;
    -h| --help)
      echo "usage: build.sh [-c commit] [-d date] [-v version] [--arch arch] [--os os] [--output-bundle name] [--output-cli name]"
      exit 0;
  esac
  shift
done

# Set defaults
if [ -z ${version:-} ]; then
    version=$(git rev-parse HEAD)
fi

if [ -z ${date:-} ] ; then
  date=$(date -R)
fi

if [ -z ${gitcommit:-} ] ; then
  gitcommit=$(git rev-parse HEAD)
fi
goos=${goos:-linux}
goarch=${goarch:-amd64}
output_cli=${output_cli:-fission-workflows}
output_bundle=${output_bundle:-fission-workflows-bundle}
output_proxy=${output_proxy:-fission-workflows-proxy}

echo "-------- Build config --------"
echo "version: ${version}"
echo "date: ${date}"
echo "commit: ${gitcommit}"
echo "goarch: ${goarch}"
echo "goos: ${goos}"
echo "output-cli: ${output_cli}"
echo "output-bundle: ${output_bundle}"
echo "output-proxy: ${output_proxy}"
echo "------------------------------"

# Build bundle
CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build\
  -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH}\
  -ldflags '-X "${versionPath}.buildDate=${date}"'\
  -o ${output_bundle}\
  github.com/fission/fission-workflows/cmd/fission-workflows-bundle/
echo "$(pwd)/${output_bundle}"

# Build proxy
CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build\
  -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH}\
  -ldflags '-X "${versionPath}.buildDate=${date}"'\
  -o ${output_proxy}\
  github.com/fission/fission-workflows/cmd/fission-workflows-proxy/
echo "$(pwd)/${output_proxy}"

# Build client
CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build \
  -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH}\
  -ldflags '-X "${versionPath}.buildDate=${date}"'\
  -o ${output_cli}\
  github.com/fission/fission-workflows/cmd/fission-workflows/
echo "$(pwd)/${output_cli}"