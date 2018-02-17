# Input Expressions

Often there are trivial data transformations between subsequent workflow steps.
For example, you might need to select a field within a larger object or normalize the input text.

To help you with these simple data transformations Fission Workflows offers support for _input expressions_.
An _input expression_ is an inline function to perform a trivial data transformation on an input value just before 
the task is invoked. 
In contrast to tasks, these _input expressions_ are not recorded or stored by the workflow engine, and do not
 have options to retry or fallbacks.
In case an input expression errors, the task will automatically fail without being invoked.

## JavaScript
The workflow engine supports input expressions written in javascript.
The underlying implementation uses the Golang-based  
[Otto Javascript interpreter](https://github.com/robertkrimen/otto) to provide a JavaScript support with almost the 
complete standard library (including regex, json, math, and date).
See the [Otto documentation](https://github.com/robertkrimen/otto) for a full reference on all builtin functions.

The workflow engine interprets any string input that is wrapped with `{ ... }` as a JavaScript expression.
An example of a JavaScript Expression:
```javascript
{$.Tasks.MyTask.Output}
``` 

See the [Examples section](#Examples) for more examples.

### Data Model
To avoid tedious and verbose querying of the internal data structures used for workflow invocations, the workflow 
engine provides a compressed data model to query with the JavaScript expression.

The workflow invocation is stored in the `$` variable; commonly referred to as the _scope_.
It consists out of the following fields:
```javascript
$ = {
    Workflow : Object,          // See Workflow
    Invocation : Object,        // See Invocation
    Tasks : {
        String : Object         // See Task
        // ...
    }
}
```

The `Workflow` object provides information about the workflow definition.
```javascript
Workflow = {
    Id : String,                // ID of the workflow
    CreatedAt: Integer,         // Unix timestamp
    UpdatedAt: Integer,         // Unix timestamp
    Status: String,             // Status of the workflow (during input evaluation it is always 'READY')
    Tasks: {
        String : {
            Src : String,       // The user provided function reference
            Runtime : String,   // The runtime responsible for executing the function
            Resolved : String   // The runtime-specific function identifier
        }
        // ...
    }
}
```

The `Invocation` object provides information about the current invocation.
````javascript
Invocation = {
    Id : String,                // ID of the workflow invocation
    CreatedAt: Integer,         // Unix timestamp
    UpdatedAt: Integer,         // Unix timestamp
    Inputs: {
        String : Object         // The input to the invocation. The value of it depends on the value type.
        // ...
    }
}
````

The `Task` object holds information about a specific task execution within the current workflow invocation.
```javascript
Task = {
    Id : String,                // ID of the task (invocation)
    CreatedAt: Integer,         // Unix timestamp
    UpdatedAt: Integer,         // Unix timestamp
    Inputs: {
        String : Object         // The input to the task. The value of it depends on the value type.
        // ...
    },
    Requires: {
        String : Object         // The key is the task ID of the dependency
        // ...
    },
    Output: Object,             // The output of the function (if available)
    Resolved : {
        Src : String,           // The user provided function reference
        Runtime : String,       // The runtime responsible for executing the function
        Resolved : String       // The runtime-specific function identifier
    },
    Status: String             // Status of the task
}
``` 

Source: `https://github.com/fission/fission-workflows/blob/master/pkg/controller/expr/scope.go` 

For convenience, the expression resolver provides the id of the current task in the `taskId` variable.

The variables are case-sensitive, which requires you to reference fields appropriately.
Additionally, the expression is truly plain javascript, so the user is responsible for avoiding NPEs.
Undefined `tasks`, `requires` or `outputs` will resolve to `undefined`.
For example `$.Workflow.Status` is valid, whereas `$.workflow.Status` will error.

Note that in the case of `inputs`, if there is a single input without an explicit key defined, it will be stored 
under the default key: `default`.

### Built-in Functions
Besides the standard library of JavaScript, the expression interpreter provides a couple of additional utility 
functions.
These functions do not have access to any additional functionality not provided to the user; they are generally 
short-hand functions to simplify common operations.
The following functions are provided by default:

name | Usage      | Description
-----|------------|-------------------------------
uid  | `uid()`    | Generates a unique (string) id
input | `input("taskId", "key")` | Gets the input of a task for the given key. If no key is provided, the default key is used.    
output | `output("taskId")` | Gets the output of a task. If no argument is provided the output of the current task is returned.
param | `param("key")` | Gets the invocation param for the given key.  
task | `task("taskId")` | Gets the task for the given taskId.
me

### Adding Custom Function
The JavaScript expression interpreter is fully extensible, allowing you to add your own functions to the existing 
standard library.

To do so, you will need to implement the [expr.Function](https://github.com/fission/fission-workflows/blob/master/pkg/controller/expr/functions.go#L17) interface, add your function to the list of 
builtin functions, and [compile](../compiling.md) the workflow engine again.
In the future functionality will be added to allow functions to be added at runtime in the form of plugins.

### Examples
This section contains various examples of common uses of JavaScript-based input expressions.

Get the timestamp of the last update to this workflow:
```javascript
{ $.Workflow.UpdatedAt }
```

Get the default (body) input of the workflow invocation:
```javascript
{ $.Invocation.Inputs.default }
// Or the function equivalent:
{ param("default") }
``` 

Get the 'Foo' header from the workflow invocation inputs, or default to 'Bar':
```javascript
{ $.Invocation.Inputs.headers.Foo || "bar" }
// Or the function equivalent:
{ param("headers").Foo || "bar" }
``` 

Get the default input of the 'other' task:
```javascript
{ $.Tasks.other.Inputs.default }
// Or the function equivalent:
{ input("other") }
```

Get the output of the 'example' task and convert it to a string:
```javascript
{ String($.Tasks.example.Output) }
// Or the function equivalent:
{ String(output("example")) }
```

Get a unique id:
```javascript
{ uid() }
```