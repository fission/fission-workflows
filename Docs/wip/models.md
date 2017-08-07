# Models

## Workflow


## WorkflowInvocation

A WorkflowInvocation consists of a spec and a status.

The spec consists of the data provided to the Workflow Engine in order to start the actual WorkflowInvocation.
After it is supplied to the Workflow Engine it is considered to be immutable.
The only way to manipulate is by canceling the invocation.

WorkflowInvocation
```
ID
Spec
Status
```

WorkflowInvocationSpec
```
InvocationID <String>
WorkflowID <String>
ExecuteAfter <Timestamp/int64>
Constraints{}
```
- `InvocationID` is a unique identifier for the WorkflowInvocation provided by the engine.
- `WorkflowID` is the identifier of the associated Workflow.
- `ExecuteAfter` indicates when the workflow should be executed. It defaults to 'now'.
- `Constraints` TODO: things like deadline, priority, location, etc..

Event types
- `WORKFLOW_CREATED`


WorkflowInvocationStatus
```
InvocationID
Status <Status>
LastUpdate
LastHeartbeatTime
```
- `InvocationID` is a unique identifier for the WorkflowInvocation provided by the engine.
- `Status` is a value indicating the general status of the invocation. Possible values: NOT_STARTED|IN_PROGRESS|SUCCEEDED|FAILED|ABORTED|UNKNOWN.
- `LastUpdate` holds the last WorkflowInvocationEvent (without data).

WorkflowInvocationEvent
```
ID
InvocationID
Time
Type
Metadata{taskId : 1234}
[DATA]
```

Event types:
- `TASK_STARTED` A task has started executing. 
- `TASK_HEARTBEAT` A task has responded to a heartbeat check of the Workflow Engine. 
- `TASK_SUCCEEDED` A task has succeeded. 
- `TASK_FAILED` A task has failed.
- `TASK_ABORTED` A task was aborted by the engine 
- `TASK_SKIPPED` A task was skipped by the engine. E.g. If/else condition
- TODO indicate some decision/branch/etc (e.g. IF/ELSE, MAP)

Valid Task transitions:
- `S -> TASK_STARTED|TASK_SKIPPED`
- `TASK_STARTED -> !(TASK_STARTED|TASK_SKIPPED)`
- `TASK_HEARTBEAT -> !(TASK_STARTED|TASK_SKIPPED)`

References
- https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md
