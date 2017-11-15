# Fission Workflows
[![Build Status](https://travis-ci.org/fission/fission-workflows.svg?branch=master)](https://travis-ci.org/fission/fission-workflows)
[![Go Report Card](https://goreportcard.com/badge/github.com/fission/fission-workflows)](https://goreportcard.com/report/github.com/fission/fission-workflows)
[![Fission Slack](http://slack.fission.io/badge.svg)](http://slack.fission.io)

[fission.io](http://fission.io)  [@fissionio](http://twitter.com/fissionio)

Fission Workflows is a workflow-based serverless function composition framework built on top of the [Fission](https://github.com/fission/fission) Function-as-a-Service (FaaS) platform.

### Highlights
- **Fault-Tolerant**: Fission Workflows engine keeps track of state, re-tries, handling of errors, etc. By internally utilizing event sourcing, it allows the engine to recover from failures and continue exactly were it left off.   
- **Scalable**: Other than a backing data store, the workflow system is stateless and can easily be scaled. The independent nature of workflows allows for relatively straightforward sharding of workloads over multiple workflow engines.
- **High Performance**: In contrast to existing workflow engines targeting other domains, Fission Workflows is designed from the ground up for low overhead, low latency workflow executions.
- **Extensible**: All main aspects of the engine are extensible. For example, you can even define your own control flow constructs.
- **Lightweight**: With just the need for a single data store (NATS Streaming) and a FaaS platform (Fission), the engine consumes minimal resources.

### Philosophy

The Fission Functions-as-a-Service framework provides simplicity and
quick time-to-value for functions on any infrastructure using
Kubernetes.  

Functions tend to do one logically separate task, and they're usually
short-lived.  For many relatively simple applications this is good
enough.  But a more complex application that uses serverless functions
must, in some way, _compose_ functions together.

There are several ways to compose functions.  A function could invoke
another function using the Fission API or HTTP.  But this requires the
calling function to handle serialization, networking, etc.

Functions could also be set up to be invoked via message queue topics.
This requires less boilerplate within each function, but the structure
of the application is not explicit; dependencies are buried inside
mappings of message queue topics to functions.

In addition, both these approaches are operationally difficult, in
terms of error analysis, performance debugging, upgrades, etc.

Workflows have been popular in other domains, such as data processing
and scientific applications, and recently got introduced into the
serverless paradigm by AWS Step Functions and Azure Logic Apps.

Fission Workflows is an open-source alternative to these workflow
systems.  It allows users to compose Fission functions in powerful
ways. Users can define their own control flow constructs, which in
most other workflow engines are part of the internal API.

### Concepts

**Workflows** can generally be represented in as a Directed Acyclic Graph (DAG).
Consider the following example, a common pattern, the diamond-shaped workflow:
```
(A) --> (B) ---> (C) ---
          \              \     
            --> (D) ---->  (E) ---> (F)
```

In this graph there is a single _starting task_ A, a _scatter task_ B
triggering parallel execution of two _branches_ with tasks C and D,
followed by a _synchronizing task_ E collecting the outputs of the
parallel tasks.

Finally the graph concludes once _final task_ F has been completed.

Although Fission Workflow has additional functionality such as
conditional branches and advanced control flow options, it
fundamentally executes a dependency graph.

```yaml
apiVersion: 1
output: WhaleWithFortune
tasks:
  GenerateFortune:
    run: fortune
    inputs: "{$.Invocation.Inputs.default}"

  WhaleWithFortune:
    run: whalesay
    inputs: "{$.Tasks.GenerateFortune.Output}"
    requires:
    - GenerateFortune
```
**Task** (also called a function here) is an atomic task, the 'building block' of a workflows. 

Currently there are two options of executing tasks.  First, Fission is
used as the main function execution runtime, using _fission functions_
as tasks.  Second, for very small tasks, such as flow control
constructs, _internal functions_ are executed within the workflow
engine, in order to minimize the network overhead.

A workflow execution is called a **(Workflow) Invocation**.  An
invocation is assigned a UID and is stored, persistently, in the data
store of Fission Workflow.  This allows users to reference the
invocation during and after the execution, for example to view the
progress so far.

Finally, **selectors** and **data transformations** are _inline
functions_ which can be used to manipulate data without having to
create a task for it.  These _inline functions_ consist of commonly
used transformations, such as getting the length of an array or
string.  Additionally, selectors allow users to only pass through
certain properties of data. In the example workflow, the JSONPath-like
selector selects the `default` input of the invocation:

See the [docs](./Docs) for a more extensive, in-depth overview of the system.

### Usage
```bash
#
# Add binary environment and create two test function on your Fission setup 
#
fission env create --name binary --image fission/binary-env:v0.3.0
fission function create --name whalesay --env binary --deploy examples/whales/whalesay.sh
fission function create --name fortune --env binary --deploy examples/whales/fortune.sh

#
# Create a workflow that uses those two functions; a workflow
# is just a function that uses the special "workflow" environment.
#
fission function create --name fortunewhale --env workflow --src examples/whales/fortunewhale.wf.yaml

#
# Map a HTTP GET to your new workflow function
#
$ fission route create --method GET --url /fortunewhale --function fortunewhale

#
# Invoke the workflow with an HTTP request
#
curl $FISSION_ROUTER/fortunewhale
``` 
See [examples](./examples) for other workflow examples.

### Installation
See the [installation guide](./INSTALL.md).

### Compiling
See the [compilation guide](./compiling.md).

### Status and Roadmap

This is an early release for community participation and user
feedback. It is under active development; none of the interfaces are
stable yet. It should not be used in production.

Contributions are welcome in whatever form, including testing,
examples, use cases, or adding features. For an overview of current
issues, checkout the [roadmap](./Docs/roadmap.md) or browse through
the open issues on Github.

Finally, we're looking for early developer feedback -- if you do use
Fission or Fission Workflow, we'd love to hear how it's working for
you, what parts in particular you'd like to see improved, and so on.

Talk to us on [slack](http://slack.fission.io) or
[twitter](https://twitter.com/fissionio).
