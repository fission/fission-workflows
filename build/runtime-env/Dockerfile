# Builders
ARG BUNDLE_IMAGE=fission-workflows-bundle
ARG BUNDLE_TAG=latest
ARG FISSION_BUILDER_IMAGE=fission/builder
ARG FISSION_TAG=latest

FROM $BUNDLE_IMAGE:$BUNDLE_TAG as workflows-bundle
FROM scratch

COPY --from=workflows-bundle /fission-workflows-bundle /fission-workflows-bundle

EXPOSE 8888
EXPOSE 8080

ENV FNENV_FISSION_CONTROLLER http://controller.fission
ENV FNENV_FISSION_EXECUTOR http://executor.fission
ENV ES_NATS_URL nats://defaultFissionAuthToken@nats-streaming.fission:4222
ENV ES_NATS_CLUSTER fissionMQTrigger

# Remove APIs when components stabilize
ENTRYPOINT ["/fission-workflows-bundle", \
            "--nats", \
            "--fission", \
            "--internal", \
            "--controller", \
            "--api-http", \
            "--api-workflow-invocation", \
            "--api-workflow", \
            "--api-admin", \
            "--metrics"]