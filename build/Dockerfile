# To run (from repo root): docker build -t fission -f ./build/Dockerfile .
ARG GOLANG_VERSION=1.10.0
FROM golang:$GOLANG_VERSION AS builder
ARG NOBUILD

WORKDIR /go/src/github.com/fission/fission-workflows

COPY . .

RUN if ! $NOBUILD ; then \
        go get github.com/Masterminds/glide; \
        glide install -v; \
        build/build-linux.sh; \
    else \
        echo "NOBUILD provided; assuming binaries exist in context."; \
    fi

# Bundle image
FROM scratch

COPY --from=builder /go/src/github.com/fission/fission-workflows/fission-workflows-bundle /fission-workflows-bundle
COPY --from=builder /go/src/github.com/fission/fission-workflows/fission-workflows /fission-workflows
