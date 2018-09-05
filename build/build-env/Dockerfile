# Builders
ARG BUNDLE_IMAGE=fission-workflows-bundle
ARG BUNDLE_TAG=latest

FROM $BUNDLE_IMAGE:$BUNDLE_TAG as workflows-bundle
FROM fission/builder:latest

COPY --from=workflows-bundle /fission-workflows /usr/local/bin/fission-workflows
ADD defaultBuild.sh /usr/local/bin/defaultBuild

EXPOSE 8001