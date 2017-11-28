# Simple Examples
This directory contains multiple simple example workflows (and whales!), along with the required functions.
The intention of each workflow is to show-off and explain a concept in Fission Workflows

## Setup
- Make sure that you have a running Fission Cluster with the workflow engine, see the [readme](../../README.md) for details.
- Deploy both the functions, using the [deploy.sh](./deploy.sh) script, or run the commands listed there manually.

## Examples


### Echowhale
The `echowhale` workflow is a minimal workflow, consisting of whale echoing the input.
The main feature it shows is that a workflow from the user's view the same as a Fission function.

Run:
```bash
curl -XPOST -d 'Whale you echo?' ${FISSION_ROUTER}/fission-function/echowhale
```

### Fortunewhale
The `fortunewhale` workflow consists of multiple steps, generating a 'fortune' and passing it through the whalesay function that was used in the first workflow.
The features introduced in this example:
- Dependencies between tasks
- Mapping a task's output to another task's input
- The 'noop' internal function, which is a function executed within the engine itself.

Run:
```bash
curl -XPOST ${FISSION_ROUTER}/fission-function/fortunewhale
```

### Maybewhale
The `maybewhale` workflow consists of a simple if-else condition.
Depending on the length of the input, the workflow transforms the input.
If the input is smaller than 20 characters, the workflow will run the `whalesay` task.
If not, it will simply propagate the user input.

Run:
```bash
# This will wrap the input in a whale, because the input is < 20
curl -XPOST -d 'Whale you echo?' ${FISSION_ROUTER}/fission-function/maybewhale

# This will simply echo the input, because the input is >= 20.
curl -XPOST -d 'Whale you echo this way too long message?' ${FISSION_ROUTER}/fission-function/maybewhale
```

### Nestedwhale
The `nestedwhale` workflow is surprisingly... empty.
The principle that 'workflows are fission functions', allows workflows to call other workflows just like you would call regular Fission Functions.
In this workflow, we generate a fortune and pass it to the dedicated whale echoing workflow: `echowhale`.

Run
```bash
curl -XPOST ${FISSION_ROUTER}/fission-function/nestedwhale
```