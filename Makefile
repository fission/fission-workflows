
PROTO_TARGETS=$(shell find pkg -type f -name "*.pb.go")
PROTO_TARGETS+=$(shell find pkg -type f -name "*.pb.gw.go")
SRC_TARGETS=$(shell find pkg -type f -name "*.go" | grep -v version.gen.go )
CHART_FILES=$(shell find charts/fission-workflows -type f)
VERSION=head

.PHONY: build generate prepush verify test changelog

build fission-workflows fission-workflows-bundle fission-workflows-proxy:
	# TODO toggle between container and local build, support parameters, seperate cli and bundle
	build/build.sh

generate: ${PROTO_TARGETS} examples/workflows-env.yaml pkg/api/events/events.gen.go

prepush: generate verify test

test:
	test/runtests.sh

verify:
	helm lint charts/fission-workflows/ > /dev/null
	hack/verify-workflows.sh
	hack/verify-gofmt.sh
	hack/verify-misc.sh
	hack/verify-govet.sh

clean:
	rm fission-workflows*

version pkg/version/version.gen.go: pkg/version/version.go ${SRC_TARGETS}
	hack/codegen-version.sh -o pkg/version/version.gen.go -v ${VERSION}

changelog:
	test -n "${GITHUB_TOKEN}" # $$GITHUB_TOKEN
	github_changelog_generator -t ${GITHUB_TOKEN} --future-release ${VERSION}

examples/workflows-env.yaml: ${CHART_FILES}
	hack/codegen-helm.sh

%.swagger.json: %.pb.go
	hack/codegen-swagger.sh

%.pb.gw.go %.pb.go: %.proto
	hack/codegen-grpc.sh

pkg/api/events/events.gen.go: pkg/api/events/events.proto
	python3 hack/codegen-events.py

# TODO add: release, docker builds, (quick) deploy, test-e2e