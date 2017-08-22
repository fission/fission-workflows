# Fission Workflow
Fission Workflow is a workflow-based serverless function composition framework built on top of the [Fission](https://github.com/fission/fission) Function-as-a-Service (FaaS) platform.

### Highlights
- **Lightweight**: With just the need for a single data store (NATS Streaming) and a FaaS platform (Fission), the engine consumes minimal resources.
- **Fault-Tolerant**: Fission Workflow engine keeps track of state, re-tries, handling of errors, etc. By internally utilizing event sourcing, it allows the engine to recover from failures and continue exactly were it left off.   
- **Scalable**: Other than a backing data store, the workflow system is stateless and can easily be scaled. The independent nature of workflows allows for relatively straightforward sharding of workloads over multiple workflow engines.
- **High Performance**: In contrast to existing workflow engines targeting other domains, Fission Workflow is designed from the ground up for low overhead, low latency workflow executions.
- **Extensible**: All main aspects of the engine are extensible. For example, you can even define your own control flow constructs.

### Philosophy
The FaaS paradigm is a relatively new development that provides multiple advantages over traditional approaches, such as pay-per-use and delegating operational logic, such as managing servers and scaling, to an experienced cloud provider.
In order to allow for minimum setup and fast execution, FaaS providers advocate using small, specific functions. 

For functions that 'glue' together a couple of APIs this constraint is good enough, but as FaaS is increasingly being used for more complex applications, developers run into issues.
Function orchestration using functions that internally call other functions, or by publishing and subscribing events on a shared event bus, quickly results into a mess of implicitly coupled functions.
Additionally, FaaS platforms lack a clear view of the dependencies between functions, making it harder to optimize performance, upgrade versions, etc.

Workflows have been popular in other domains, such as data processing and scientific applications, and recently got introduced into the serverless paradigm by AWS Step Functions and Azure Logic Apps.
Fission Workflow is an open-source alternative to these workflow systems, which both constrains the functionality in some areas, while extending it in others.
In contrast to others Fission Workflow is purely focused on function composition, excluding events and other integrations from the workflow.
On the other hand the system is more flexible, by allowing users to define their own control flow constructs, which in most other workflow engines are part of the internal API.

### Concepts
**Workflows** can generally be represented in as a Directed Acyclic Graph (DAG).
Consider the following example, a common pattern, the diamond-shaped workflow:
```
(A) --> (B) ---> (C) ---
          \              \     
            --> (D) ---->  (E) ---> (F)
```
In this graph there is a single _starting task_ A, a _scatter task_ B triggering parallel execution of two _branches_ with tasks C and D, followed by a _synchronizing task_ E collecting the outputs of the parallel tasks.
Finally the graph concludes once _final task_ F has been completed.
Although workflow engines, such as Fission Workflow, have various additional functionality, such as conditional branches, advanced data/control flow options, exception handling, all these workflow engines fundamentally consist of a dependency graph.

```javascript
{
  "apiVersion": "alpha",
  "outputTask": "echo",
  "tasks" : {
    "echo" : {
      "name" : "noop",
      "inputs" : {
        "default" : {
          "type" : "ref",
          "value" : "$.invocation.inputs.default"
        }
      }
    }
  }
}
```
**Task** (also called a function here) is an atomic task, the 'building block' of a workflows. 
Currently there are two options of executing tasks. 
First, Fission is used as the main function execution runtime, using _fission functions_ as tasks. 
Second, for very small tasks, such as flow control constructs, _internal functions_ are executed within the workflow engine, in order to minimize the network overhead.

A workflow execution is called a **(Workflow) Invocation**.
An invocation is assigned a UID and is stored, persistently, in the data store of Fission Workflow.
This allows users to reference the invocation during and after the execution, for example to view the progress so far.

Finally, **selectors** and **data transformations** are _inline functions_ which can be used to manipulate data without having to create a task for it.
These _inline functions_ consist of commonly used transformations, such as getting the length of an array or string.
Additionally, selectors allow users to only pass through certain properties of data. In the example workflow, the JSONPath-like selector selects the `default` input of the invocation:

See the [docs](./Docs) for a more extensive, in-depth overview of the system.

### Usage
```bash
# Setup functions
$ fission env create --name nodejs --image fission/node-env
$ fission fn create --name oauthenticator --env nodejs --code scraper/oauthenticator.js
$ fission fn create --name scrape --env nodejs --code scraper/scrape.js
$ fission fn create --name extractLinks --env nodejs --code scraper/extractLinks.js

# Setup workflow
$ fission fn create --name scrapeForLinks --env workflow --code scraper/scrapeLinks.wf.json

# Map POST /hello to your new function
$ fission route create --method POST --url /scrapeLinks --function scrapeForLink

$ curl -XPOST -d 'http://authenticationneeded.com' $FISSION_ROUTER/scrapeLinks
```
Currently, this example is solely explanatory and not functional. 
See [examples](./examples) for actual functional workflow examples.

### Installation
See the [installation guide](./INSTALL.md).

### Compiling
See the [compilation guide](./compiling.md).

### Status and Roadmap
This is an early release for community participation and user feedback. It is under active development; none of the interfaces are stable yet. It should not be used in production.

Contributions are welcome in whatever form, including testing, examples, use cases, or adding features.
For an overview of current issues, checkout the [roadmap](./Docs/roadmap.md) or browse through the open issues on Github.

Finally, we're looking for early developer feedback -- if you do use Fission or Fission Workflow, we'd love to hear how it's working for you, what parts in particular you'd like to see improved, and so on.  
Talk to us on [slack](http://slack.fission.io) or [twitter](https://twitter.com/fissionio).
