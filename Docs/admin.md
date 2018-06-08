# Administrating Fission Workflows

Currently, Fission Workflows is in active development.
This also means that it is still rough on edges.
This section provides you with some approaches to debugging, and maintaining a 
Fission Workflows deployment, as well as how to diagnose and fix (common) issues.

## Inspect workflow invocations
Use the `fission-workflows` tool, which allows you to query and inspect workflow invocations.

## View workflow engine logs
To view the logging of the workflow engine:
```bash
# Find and tail the workflows pod for logs
kubectl -n fission-function get all
kubectl -n fission-function logs <workflow-pod> workflow -f
```

Note if nothing seems to happen when you are invoking workflows, you should inspect the 
Fission executor and router logs

## Unresponsive functions/workflows (Fission < 0.7.0)
The workflow engine maintains a lookup table to match workflow invocations to workflows.
In fission < 0.7.0, there can be situations (e.g. after a crash) that the workflow engine 
loses this lookup table, while Fission assumes that the workflow engine still is able to map 
the Fission function to the workflow.
This will result in requests to the functions/workflows to time-out.

To fix simply delete and re-create the functions:
```bash
fission fn delete --name <workflow-name>
fission fn create --name <workflow-name> --env workflow --src <workflow-file>
```    

## Soft reset 
If you suspect that the engine is not functioning correctly, you can try restarting the engine.
By restarting the pod, the engine will restart, replay the events to return to the current state.

```bash
kubectl -n fission-function get po
kubectl -n fission-function delete po <workflow-pod>
```

## Complete reset
There might be cases that you want to completely clear a Fission Workflows deployment.
```bash
# Find and delete the workflows deployment 
helm list
helm delete --purge {fission-workflows-name}

# State is maintained in the NATS deployment; to clear restart the nats-streaming pod
kubectl -n fission get po
kubectl -n fission delete po <nats-streaming-pod>
```

