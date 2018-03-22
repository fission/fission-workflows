# Builders
ARG BUNDLE_IMAGE=fission-workflows-bundle
ARG BUNDLE_TAG=latest
ARG FISSION_BUILDER_IMAGE=fission/builder
ARG FISSION_TAG=0.6.0

FROM $BUNDLE_IMAGE:$BUNDLE_TAG as workflows-bundle
FROM $FISSION_BUILDER_IMAGE:$FISSION_TAG

COPY --from=workflows-bundle /wfcli /usr/local/bin/wfcli
ADD defaultBuild.sh /usr/local/bin/defaultBuild

EXPOSE 8001