# Functions

In serverless workflows, the notion of a function is a central theme.
This document describes the model, terminology, and the implementation of function execution in Workflows.

TODO explain common lifecycles: fnref -> fnid, fn execution model, data model 

## Terminology

- **function environment** (or fnenv) is a platform, or service that provides the capability to reference and execute 
functions.
- **function runtime** is the part of the fnenv that is responsible for executing the functions. 
- **function resolver** is the part of the fnenv that that parses function references to deterministic function 
identifiers.
This is done to ensure that workflows remain robust in the face of frequent changes to functions.
- **function** (or fn) is a black box function that has zero or more inputs and a single output.
A function is said to be 'executed' or 'run', where it runs atomically to completion or fails.
Although side-effects are often unavoidable, the model proposes mitigating side-effects as much as possible - for 
instance by ensuring that these effects are idempotent. 
In the future, the workflow engine will allow labeling functions being pure or impure (from a functional programming 
perspective). 
- **function reference** (or fnref) is a reference to a function.
This reference is allowed to be ambigious in terms of what fnenv to use, and which version to use. 
It could be that multiple versions of the function exist, or that the function is available on multiple fnenv.
The workflows engine, upon loading the workflow, will resolve the function references to function identifiers.
In the future commands will be added to automate upgrading a workflow to use new function versions.
- **function identifier** (or fnid) is a unique, consistent identifier to a specific version of a specific function 
on a specific fnenv. 
Using this identifier to identify the function to execute should result in the exact same function to be executed.
Note that currently, for Fission function this identifier is not yet considering versions of the same function.  

## Function Environments

There are currently two function environments: **Fission** and **Internal**.

[Fission](https://github.com/fission/fission) is a complete Function-as-a-Service platform - which includes extensive
Kubernetes integration, autoscaling, and offers fine-grained resource management controls.

The **internal** function environment is an extremely light-weight function runtime built into the 
workflow engine itself.
It is useful for lightweight functions - typically control flow functions - where the network overhead of running 
them on an external FaaS platform would be overkill.
However, it has no sandboxing or autoscaling options; intensive functions might slow down the workflow engine.

Whether it is best to execute your functions on either of the function environments is based on the kind of function 
and the type of workload.

If a function happens to be declared on both platforms for now the workflow engine will always choose the internal 
function to use.
You can force the workflow engine to use one fnenv over the other by specifying it in the function reference.
For example, to explicitly use a Fission function called `sleep`, the function reference should: `fission://sleep`.   

In the future we want to improve the scheduler to allow it to make this decision to based on profiling and other data 
sources. 

### Fission

[Fission](https://github.com/fission/fission) is the Function-as-a-Service (FaaS) platform underlying the workflow 
engine.
All Fission functions are compatible with the workflow engine.
Moreover, workflows follow the same API - allowing you to call Fission functions without even needing to know whether 
it is a workflow or an actual Fission function.

#### Specification

Calling a Fission function is no different from calling other (internal) functions.
Like other functions, it has a number of optional input parameters to augment the Fission function execution.

**Input**        | required | types                | description
-----------------|----------|----------------------|---------------------------------
body             | no       | *                    | The body of the Fission function call.
headers          | no       | map[string]string    | The headers that need to be added to the request.
query            | no       | map[string]string    | The key-value values that need to be added to the URL.
method           | no       | string               | HTTP Method to use (Default: GET)
content-type     | no       | string               | Force a specific content-type for the request (default: application/octet-stream).

**Output** (*) the body of the Fission function response.

The body is interpreted based on the content-type.
In case it is interpretable (e.g. `application/json`), the workflow engine will parse it, and make it available to 
the [expressions](./expressions.md) to manipulate/access.
In case it is not interpretable, the output will be considered a binary blob.
You can still pass it as a whole to other tasks, but you can not access or modify the contents with expressions. 

Note: currently you cannot access the metadata of the response. Though, You can access this metadata for the workflow 
invocation.

**Fission's Perspective**

The workflow engine invokes Fission functions with the same API as other types of event sources.
So conceptually there is no difference between Fission functions used or not used in workflows.

**Example**

To invoke the Fission function `example-funciton`, we can define the following task in the workflow:

Note: although you can override the content-type, if nothing is specified the content-type will be inferred from the 
body value. So in this case the content-type will be `text/plain` because the body value is a string.

```yaml
# ...
RunExampleFissionFunction:
  run: example-function
  inputs: 
    body: "Some body input"
    headers:
      foo: bar
    query:
      abc: def
    method: DELETE
# ...
```

This task is basically equivalent to calling the fission function directly:

```bash
curl ${FISSION_ROUTER}/fission-function/example-function?abd=def -XDELETE -H "foo: bar" -H "Content-Type: text/plain" -d "Some body input"
```

#### Notes
- The content-type is important if you want to utilize the full functionality of Workflows; ensure that the functions 
have the correct MIME/content type in their responses.
- There is no need to force content-types on requests in many cases. 
The workflow engine remembers the content-type of how it received the data, and will use it when using the data as 
input for another task (unless the content type is overriden).
- You can return a specification for a task or workflow (to implement dynamic tasks) by using the appropriate 
content-type: `application/vnd.fission.workflows.task` or `application/vnd.fission.workflows.workflow` using the 
protobuf encoding.

### Internal

The internal function environment is a lightweight and limited function runtime inside the workflow engine itself.
It used to execute small, mostly control flow-related, functions, that would otherwise suffer a lot from the unnecessary 
overhead of a FaaS platform. 

It consists out of a number of built-in functions, which aim to cover the common functionality needed in workflows.
Additionally, it has options to extend with your own functions.

#### Built-in

The internal fnenv ships with a number of built-in functions. 
These functions are simply commonly used, utility functions. 
They do not have any additional or special API; you could implement these functions just as well in Fission.

Each function has a listed `available`: in which versions is the function available.
The `status` indicates how reliable and well tested a function is: `stable` is well-tested, and unlikely to change in 
API, `experimental` is working (partially) but is subject to change in future versions, and `concept` is a 
(potentially) unimplemented, non-functioning function that is under active or future development.

---

##### compose

Property  | description
----------|--------
command   | `compose`
available | `^0.1.1`
status    | stable

**Description**

Compose provides a way to merge, modify and create complex values from multiple inputs.
Other than outputting the composed inputs, compose does not perform any other operation.
This is useful when you want to merge the outputs from different tasks (for example in a MapReduce or scatter-gather 
scenario).

**Specification**

**Input**   | required | types  | description
------------|----------|--------|---------------------------------
default     | no       | *      | the inputs to be merged into a single map or outputted if none other.
*           | no       | *      | the inputs to be merged into a single map. 

**Note: custom message does not yet propagate back to the user**

**Output** (*) The composed map, single default input, or nothing.

**Example**

Compose with a single input, similar to `noop`:
```yaml
# ...
foo:
  run: compose
  inputs: "all has failed"
# ...
```

Composing a map inputs:
```yaml
# ...
foo:
  run: compose
  inputs: 
    foo: bar
    fission: workflows
# ...
```

---

##### fail

Property  | description
----------|--------
command   | `fail`
available | `^0.3.0`
status    | experimental

**Description**

Fail is a function that always fails. This can be used to short-circuit workflows in
specific branches. Optionally you can provide a custom message to the failure.

**Specification**

**Input**   | required | types  | description
------------|----------|--------|---------------------------------
default     | no       | string | custom message to show on error 

**Note: custom message does not yet propagate back to the user**

**Output** None 

**Example**

```yaml
# ...
foo:
  run: fail
  inputs: "all has failed"
# ...
```

A complete example of this function can be found in the [failwhale](../examples/whales/failwhale.wf.yaml) example.

---

##### foreach

Property  | description
----------|--------
command   | `foreach`
available | `^0.3.0`
status    | experimental

**Description**

Foreach is a control flow construct to execute a certain task for each element in the provided input.
By default, the tasks are executed in parallel. 
Currently, foreach does not gather or store the outputs of the tasks in any way.

**Specification**

**Input**       | required | types         | description
----------------|----------|---------------|--------------------------------------------------------
foreach/default | yes      | list          | The list of elements that foreach should be looped over.
do              | yes      | task/workflow | The action to perform for every element.
sequential      | no       | bool          | Execute the actions sequentially (default: false).   

The element is made available to the action using the field `element`.

**Output** None 

**Example**

```yaml
# ...
foo:
  run: foreach
  inputs:
    for:
    - a
    - b
    - c
    do:
      run: noop
      inputs: "{ task().element }"
# ...
```

A complete example of this function can be found in the [foreachwhale](../examples/whales/foreachwhale.wf.yaml) example.

---

##### http

Property  | description
----------|--------
command   | `http`
available | `^0.3.0`
status    | experimental

**Description**

Http is a general utility function to perform simple HTTP requests.
It is useful for prototyping and managing low overhead HTTP requests.
To this end it offers basic functionality, such as setting headers, query, method, url, and body inputs.

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
url/default     | yes      | string            | URL of the request.
headers         | no       | map[string|string | The action to perform for every element.
content-type    | no       | string            | Force a specific content-type for the request.
method          | no       | string            | HTTP Method of the request. (default: GET)
body            | no       | *                 | The body of the request. (default: application/octet-stream)

Unless the content type is specified explicitly, the workflow engine will infer the content-type based on the body.

**Output** (*) the body of the response.

Note: currently you cannot access the metadata of the response.

**Example**

```yaml
# ...
httpExample:
  run: http
  inputs:
    url: http://fission.io
    method: post
    body: "foo"
# ...
```

A complete example of this function can be found in the [httpwhale](../examples/whales/httpwhale.wf.yaml) example.

---

##### if

Property  | description
----------|--------
command   | `if`
available | `^0.1.1`
status    | stable

**Description**

If is the simplest ways of altering the control flow of a workflow.
It allows you to implement an if-else construct; executing a branch or returning a specific output based on the 
result of an execution.

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
if              | yes      | bool              | The condition to evaluate.
then            | no       | *                 | Value or action to return if the condition is true.
else            | no       | *                 | Value or action to return if the condition is false.

**Output** (*) Either the input of `then` or `else` (or none if not set).

**Example**

The following example shows the dynamic nature of this control flow.
If the if-expression evaluates to true, a static value is outputted. 
Otherwise, a function is outputted (which in turn is executed).

```yaml
# ...
ifExample:
  run: if
  inputs:
    if: { param() > 42  }
    then: "foo"
    else: 
      run: noop
# ...
```

A complete example of this function can be found in the [maybewhale](../examples/whales/maybewhale.wf.yaml) example. 

---

##### javascript

Property  | description
----------|--------
command   | `javascript`
available | `^0.3.0`
status    | experimental

**Description**

Javascript allows you to create a task that evaluates an arbitrary JavaScript expression.
The implementation is similar to the inline evaluation of JavaScript in [expressions](./expressions.md) in inputs.
In that sense this implementations does not offer more functionality than inline expressions.
However, as it allows you to implement the entire task in JavaScript, this function is useful for prototyping and 
stubbing particular functions. 

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
expr            | yes      | string            | The JavaScript expression
args            | no       | *                 | The arguments that need to be present in the expression.

Note: the `expr` is of type `string` - not a `expression` - to prevent the workflow engine from evaluating the 
expression prematurely. 

**Output** (*) The output of the expression.

**Example**

```yaml
# ...
JsExample:
  run: javascript
  inputs:
    expr: "a ^ b"
    args:
      a: 42
      b: 10
# ...
```

A complete example of this function can be found in the [fibonacci](../examples/misc/fibonacci.wf.yaml) example. 

---

##### noop

Property  | description
----------|--------
command   | `noop` or `nop`
available | `^0.1.0`
status    | stable

**Description**

Noop represents a "no operation" task; it does not do anything.
The input it receives in its default key, will be outputted in the output

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
default         | no       | *                 | The input to pass to the output.

**Output** (*) The output of the default input if provided.

**Example**

```yaml
# ...
NoopExample:
  run: noop
  inputs: foobar
# ...
```

A complete example of this function can be found in the [fortunewhale](../examples/whales/fortunewhale.wf.yaml) example. 

---

#### repeat

Property  | description
----------|--------
command   | `repeat`
available | `^0.3.0`
status    | experimental

**Description**

Repeat, as the name suggests, repeatedly executes a specific function.
The repeating is based on a static number, and is done sequentially.
The subsequent tasks can access the output of the previous task with `prev`.

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
times           | yes      | number            | Number of times to repeat the task.
do              | yes      | task              | The task to execute.

Note: the task `do` gets the output of the previous task injected into `prev`.

**Output** (*) The output of the last task.

**Example**

```yaml
# ...
RepeatExample:
  run: repeat
  inputs:
    times: 5
    do:
      run: noop
      inputs: { task().prev + 1 }}
# ...
```

A complete example of this function can be found in the [repeatwhale](../examples/whales/repeatwhale.wf.yaml) example. 

---

##### sleep

Property  | description
----------|--------
command   | `sleep`
available | `^0.1.1`
status    | stable

**Description**

Sleep is similarly to `noop` a "no operation" function.
However, the `sleep` function will wait for a specific amount of time before "completing".
This can be useful to mock or stub out functions during development, while simulating the realistic execution time. 

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
default         | no       | string            | A string-based representation of the duration of the sleep. (default: 1 second)

Note: the sleep input is parsed based on the [Golang Duration string notation](https://golang.org/pkg/time/#ParseDuration).
Examples: 1 hour and 10 minutes: `1h10m`, 2 minutes and 300 milliseconds: `2m300ms`.

**Output** None

**Example**

```yaml
# ...
NoopExample:
  run: sleep
  inputs: 1h
# ...
```

A complete example of this function can be found in the [sleepalot](../examples/misc/sleepalot.wf.yaml) example. 

---

##### switch

Property  | description
----------|--------
command   | `switch`
available | `^0.3.0`
status    | experimental

**Description**

Switch is very similar to how switch-constructs are implemented in most languages.
In this case the switch is limited to evaluating string keys.
The string-switch is matched to one of the cases, or - if none of those match - the default case.  

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
switch          | yes      | string            | The string to match to one of the cases.
cases           | no       | list              | List of cases to match to.
default         | no       | *                 | The default value if there is no matching case.

**Output** (*) Either the value of the matching case, the default, or nothing (in case the default is not specified).

**Example**

```yaml
# ...
SwitchExample:
  run: switch
  inputs: 
    switch: "{ param() }"
    cases:
    - case: foo
      action: bar
    - case: ac
      action: me
    default: 42
# ...
```

A complete example of this function can be found in the [switchwhale](../examples/whales/switchwhale.wf.yaml) example.

---

##### while
 
Property  | description
----------|--------
command   | `while`
available | `^0.3.0`
status    | concept

**Description**

While consists of a control flow construct that will execute a specific task as long as the condition has
not been met.
The results of the executed action can be accessed using the task ID "action".

**Specification**

**Input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
expr            | yes      | bool              | The condition which determines whether to continue or halt the loop.
do              | yes      | task/workflow     | The action to execute on each iteration.
limit           | no       | number            | The max number of iterations of the loop. (default: unlimited)

Notes: 
- we currently cannot reevaluate the expr. 
There needs to be support for looking up the source of an expression.
Maybe we can add the original expression to the labels.
- we might want to have a `prev` value here to reference the output of the previous iteration.
 

**Output** (*) Either the value of the matching case, the default, or nothing (in case the default is not specified).

**Example**

```yaml
# ...
SwitchExample:
  run: while
  inputs: 
    expr: "{ 42 > 0 }"
    limit: 10
    do:
      run: noop
# ...
```

A complete example of this function can be found in the [whilewhale](../examples/whales/whilewhale.wf.yaml) example.

--- 

#### Extending the internal function environment
As the naming suggests, the internal function environment is not limited to this small set of built-in functions.
The internal function environment can be extended with functions just like Fission.

To add an additional function, you need to implement the `native.Function` Go interface, and add the new function to
the list of functions.

Currently, you will still need to recompile the engine if you want add, change or remove internal functions.
In the near future you will be able to add these functions (in Go or Javascript) to the workflow engine at runtime.
