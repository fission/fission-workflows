# Workflows and Waiting

## Problem

Users have points in their workflows where a long-running but
asynchronous function needs to be invoked.

For example, a workflow might involve sending an email to the user and
waiting for them to open it and click on a link in it. Or it may
involve sending something to a third-party async API and waiting for a
result.

Today the recommended approach with workflows is to split the workflow
into two at the async task point, and invoke the latter part of the
workflow when the async task completes, either using an HTTP or
Message Queue trigger.

This is not ideal, because logically the workflow should be a single
workflow. We want some solution that does not involve splitting the
workflow at every async task.

## Approaches 


### Sleep-and-Poll

A static task (either a builtin or a fission function) is invoked and
poll for completion of an asynchronously invoked thing. The task runs
as long as the thing it’s waiting for doesn’t finish.

While polling is generally simple to implement and gets the job done,
there are some problems with polling:

1. Long-running tasks occupy memory while waiting, contradicting our
   goal of having resource usage only when doing “actual work”. In a
   highly async workload, you’ll end up with a lot of waiting tasks, even
   nested ones.
1. Short running tasks make scheduling easier, especially on
   spot/preemptible instances, which are much cheaper.
1. Polling is unsuitable for workloads that need low latency. The
   polling interval is a trade-off: the polling interval can be made very
   small (to poll very often), but that tends to cost CPU, memory,
   network and hit API rate limits.


### Timer-based Poll

The idea here is that the workflow engine (or fission core perhaps)
would have a primitive that allows you to run a short-running function
repeatedly using a timer.

This is a little better than sleep-and-poll, since tasks are short
running. It still has the other problem of poll, which is CPU usage,
API rate limits, and unsuitability for low-latency workloads.


### Wait-and-Notify

In this design the workflow engine would add one API and one builtin
function.

The builtin function would be `data = wait(ID)` and the API would be
`notify(ID, data)`.

The implementation of the builtin would cause the workflow engine to
subscribe to a topic named `ID`. The notify API implementation would
send a message containing `data` onto the topic named `ID`.

The async task implementation would need to take care of calling back
the notify API when it’s done. (We would pass the async task a unique
ID and an authentication cookie, which it would send back into the
notify API when it’s done.)

### Callbacks
*(This one is not really different from the last one, basically the
async task invokes a special fission function instead of a special
notify API.)*

Let’s try this:
`startAsyncTask(arguments, callbackFunctionName)`

Essentially you’d pass a callback function name to a task. That task
would need to figure out how to call the callback function when its
work is done. After that function is called, it restarts the workflow.
We can’t actually implement this as-is, because we need to turn the
callback into a workflow task. So we need a dynamic task:

`async(startAsyncTask, arguments, callbackFunctionName)`

Here, `async` is a new dynamic task. It calls
`startAsyncTask(arguments, wrapper)`. Here, `wrapper` wraps the
callback. It’s a function that re-invokes the workflow, perhaps by
passing a message on the workflow event queue.


### Channels (Communicating Sequential Processes)

In this case we expose an API and builtins. This is similar to
wait-and-notify but more generic.

We expose an API and workflow builtins to create, destroy and close
channels. Channels can be written to or read from, and they can be
created as uni-directional or bi-directional.

This would essentially be an abstraction over the message queue we’re
using, plus bindings into re-starting the workflow on reception of a
message on a channel. Workflow tasks would have a special input type
to say “read from so-and-so channel”.

This is powerful but also a bit complex.  Are there buffers? Retries?
What are the reliability guarantees? Needs more thought on whether
it’s a good idea; perhaps we need to design it down to details to
decide.


### Futures/Promises/Deferred variables

In a programming language, a future is a variable that will eventually
have a value, but you can use it before it actually has a value. The
language runtime automatically blocks code that uses a future until
the future is “ready”.

We’d expose APIs and builtins for creating, writing to, and reading
from futures.

So you’d have to create a task, write its output to a “future” (maybe
we need another dynamic task wrapper to do this). Then you create
another task that takes this future as input. You would explicitly
define a dependency from a task to this future. There would be a
workflow engine API for assigning to a future. On assignment, the
task’s dependency would be satisfied and it would be ready to run.

(So, tasks would now be “ready to run” when both their control flow and
future-variable dependencies are satisfied.)

### Async/Await, Generators/Coroutines, and Yield

This doesn’t address quite the same use case as the rest. But it’s
useful when you want to stream items from a collection into some other
task. For example if you want to chunk a large file into pieces and
start processing each piece before you finish the splitting of the
file, you’d want to use some sort of generator pattern.

[details tbd]


