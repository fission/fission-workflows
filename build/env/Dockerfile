FROM alpine

ADD fission-workflow-bundle /

EXPOSE 8888

ENV FNENV_FISSION_CONTROLLER http://controller.fission
ENV FNENV_FISSION_POOLMGR http://poolmgr.fission
ENV ES_NATS_URL nats://nats-streaming:4222
ENV ES_NATS_CLUSTER fissionMQTrigger

ENTRYPOINT ["/fission-workflow-bundle", "--nats", "--fission", "--controller", "--api-http", "--api-workflow-invocation", "--api-workflow", "--api-admin"]
