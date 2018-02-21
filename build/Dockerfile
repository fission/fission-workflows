# To run (from repo root): docker build -t fission -f ./build/Dockerfile .
ARG NOBUILD
ARG GOLANG_VERSION=1.9.4
FROM golang:$GOLANG_VERSION AS builder

WORKDIR /go/src/github.com/fission/fission-workflows

COPY . .

RUN if [ "_$NOBUILD" = "_" ] ; then \
        go get github.com/Masterminds/glide; \
        glide install -v; \
        build/build-linux.sh; \
    else \
        echo "NOBUILD provided; assuming binaries exist in context."; \
    fi

# Bundle image
FROM scratch

COPY --from=builder /go/src/github.com/fission/fission-workflows/fission-workflows-bundle /fission-workflows-bundle
COPY --from=builder /go/src/github.com/fission/fission-workflows/wfcli /wfcli