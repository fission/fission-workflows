ARG BUNDLE_IMAGE=fission-workflows-bundle
ARG BUNDLE_TAG=latest
FROM $BUNDLE_IMAGE:$BUNDLE_TAG as workflows-bundle

FROM scratch

COPY --from=workflows-bundle /fission-workflows-proxy /fission-workflows-proxy

ENTRYPOINT ["/fission-workflows-proxy"]
CMD ["-h"]