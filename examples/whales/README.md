# Simple Examples
This directory contains two simple example workflows (and two whales!), along with the required functions.

## Setup
- Make sure that you have a running Fission Cluster with the workflow engine, see the [readme](../../README.md) for details.
- Deploy both the functions, using the [deploy.sh](./deploy.sh) script, or run the commands listed there manually.

### Example: Echowhale
The 'echowhale' workflow is a minimal workflow, consisting of whale echoing the input.
The main feature it shows is that a workflow from the user's view the same as a Fission function.

To invoke:
```bash
curl -XPOST -d 'Whale you echo?' $FISSION_ROUTER/fission-function/echowhale
```

### Example: Fortunewhale
The fortunewhale workflow consists of multiple steps, generating a 'fortune' and passing it through the whalesay function that was used in the first workflow.
The features introduced in this example:
- Dependencies between tasks
- Mapping a task's output to another task's input
- The 'noop' internal function, which is a function executed within the engine itself.

To invoke:
```bash
curl -XPOST $FISSION_ROUTER/fission-function/fortunewhale
```

### Example: 
