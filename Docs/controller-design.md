# Controller

The controller is responsible for managing invocations to completion.
It makes use of the scheduler to determine the strategy towards the completion of these invocations.
It makes use of the InvocationApi, FunctionApi and WorkflowApi to perform actions.

## Control Flow
The controller can best be described as an interruptible control loop.
It contains the following ways of managing the lifecycle of invocations.
- Long control loop
- Short control loop
- Notification interruption

### Long Control Loop
The long control loop has a frequency measured in seconds.
In this control loop, the controller will refresh the cache and check the event store for any active invocation.
This is meant to avoid missing any invocations that for whatever reason did not make it into the controller's active invocation cache.

### Short Control Loop
The short control loop has a frequency as high as possible performance-wise (most likely somewhere between 100-1000 ms)
The controller checks for updates on invocations in its active invocation cache.
This is meant to capture any invocations of which the notification (see notification section) did not occur.
 
### Notification
The notification resembles more or less an interrupt. 
Whenever a relevant event occurs for an invocation, the controller receives a notification of this event.
Note that notifications are entirely optional, and only serve to reduce the latency of a workflow invocation.
In case of errors or an overload of notifications, the short or long control loop will pick up the invocation


## Procedure
The controller can be represented in the following steps.

1. Controller evaluation is triggered for a invocation.
2. Controller gathers required data needed for the evaluation (workflow definition, invocation).
3. Controller submits the invocation to the scheduler for evaluation.
4. Scheduler evaluates the current state of the invocation.
5. Scheduler responds with a set of actions for the controller to undertake.
6. Controller interprets the received actions, and executes them concurrently.
7. The actions (invoking/pausing tasks) result in new events being added 
8. The new events of a specific invocation result trigger step 1. 
