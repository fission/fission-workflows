# Development and Testing of Fission Workflows 

## Development

To test with local builds of the workflow engine in the Kubernetes cluster in the current context. First [install a 
workflow version to your cluster](../INSTALL.md). After that you can swap the deployed workflow engine with a local one 
using [telepresence](https://telepresence.io):

```bash
telepresence --method=vpn-tcp --namespace fission --swap-deployment workflows:workflows --expose 5555 --expose 8080
```

### Local OpenTracing

The locally running instance does not have access to the in-cluster Jaeger deployment. To view the invocations, the 
easiest option is to run a development all-in-one Jaeger deployment locally:  

```bash
docker run -d --rm --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.6
``` 

You can then navigate to `http://localhost:16686` to access the Jaeger UI.

## Testing

To run local unit and integration tests:

```bash
bash tests/runtests.sh
```

If you have an instance of Workflows deployed in the Kubernetes cluster in the current context, you can also run 
end-to-end tests (the same that are used by the CI):

```bash
bash tests/e2e/runtests.sh
```

Since each e2e test is standalone, you can also run a single e2e test:

```bash
bash tests/e2e/tests/test_inputs.sh
```