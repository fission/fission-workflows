# Development and Testing of Fission Workflows 

## Development

To test with local builds of the workflow engine in the Kubernetes cluster in the current context. First [install a 
workflow version to your cluster](../INSTALL.md). After that you can swap the deployed workflow engine with a local one 
using [telepresence](https://telepresence.io):

```bash
telepresence --method=vpn-tcp --namespace fission --swap-deployment workflows:workflows --expose 5555 --expose 8080
```

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