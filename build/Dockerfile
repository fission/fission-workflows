# To run (from repo root): docker build -t fission -f ./build/Dockerfile .
ARG GOLANG_VERSION=1.15.2
FROM golang:$GOLANG_VERSION AS builder
ARG NOBUILD

WORKDIR /go/src/github.com/fission/fission-workflows

COPY . .

ENV GO111MODULE=on

RUN if ! $NOBUILD ; then \
        go mod download \
        build/build-linux.sh; \
    else \
        echo "NOBUILD provided; assuming binaries exist in context."; \
    fi

# Bundle image
FROM scratch

COPY --from=builder /go/src/github.com/fission/fission-workflows/fission-workflows-bundle /fission-workflows-bundle
COPY --from=builder /go/src/github.com/fission/fission-workflows/fission-workflows-proxy /fission-workflows-proxy
COPY --from=builder /go/src/github.com/fission/fission-workflows/fission-workflows /fission-workflows

# Sensible default: run fission-workflows with in-memory event store and all basic components
# --api - serves the gRPC and HTTP interface at :5000 and :8080 respectively
# --internal - enables the internal function runtime
# --controller - runs the controller and scheduler components
# --metrics - enables Prometheus and Jaeger metric collection.
ENTRYPOINT ["/fission-workflows-bundle"]

CMD ["--api", "--internal", "--controller", "--metrics"]
