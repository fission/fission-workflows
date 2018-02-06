FROM alpine:3.5

ADD fission-workflows-bundle /

EXPOSE 8888

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
            "--api-admin"]
