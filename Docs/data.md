# Input and Output

This document describes how the input and output data is handled in Fission Workflows.

## Data Lifecycle

There are two main places where data is converted from internal to and from external formats. 

On incoming request
0. read body
1. attempt to parse body to TypedValue
2. check if task/workflow
    3. add dynamic task to
    
On executing function
1. fetch inputs
2. resolve expressions in function inputs
3. Format inputs to request (infer content-type).

## TypedValue

In contrast to the average workflow engine, Workflows attempts to interpret input and output values.
There are several use cases for this: 
- Check if the output value of a task is a task in order to support dynamic tasks.
- Support and resolve [expressions](../Docs/expressions.md) in inputs (and, in the future, outputs) for manipulation 
and selections of data.
- Ensure that values can be stored and exchanged in a variety of formats (e.g. protobuf and JSON).

The typing is completely optional.
This typing is done automatically, among others, based on the Content-Type headers (for example `application/json`).
If a value could not be parsed into a specific type, it will just be assumed to be a binary 'blob'.
Although it will not be possible to manipulate or inspect as more specific data types, you can still reference and 
pass it around as a whole.
 
In the future, options will be added to prevent type parsing to optimize performance. 

### Implementation

This interpretation is done using an internal construct called `TypedValue`.
This, as the name says, is a representation of the data with some added metadata:

```proto
message TypedValue {
    string type = 1;
    bytes value = 2;
    map<string, string> labels = 3;
}
```

Conceptually, it exists out of an arbitrary `type` (see Supported Types), the corresponding byte-level representation 
of the `value`.
Finally it contains annotations or `labels` that provide metadata about the value.
This metadata could include the original Content-Type header that was associated with the value, the source of 
evaluated expressions, and so on.    

## Supported types

The supported types can be divided into a small set of primitive types, and more complex types. 

### Primitive Types

The primitive types are somewhat analogous to the types in JavaScript. 

As memory consumption of individual values is not very important in the scope of a serverless workflows, no attempt 
is made to model the various kinds of integer and float values.

Type       | GoType                         | Description
-----------|--------------------------------|--------------------------------------------------
bytes      | []byte                         | (aka the 'blob' type) the default type for data that is not recognized. 
nil        | nil                            | Empty, but defined value.
boolean    | bool                           | Boolean value.
number     | float64                        | A number value that represents number values 
string     | string                         | A string value.

### Other Types

Type       | GoType                         | Description
-----------|--------------------------------|--------------------------------------------------
expression | string                         | An expression that resolves at runtime to a TypedValue.
task       | *types.TaskSpec                | A specification of a task (used to implement dynamic tasks)
workflow   | *types.WorkflowSpec            | A specification of a workflow (used to implement dynamic tasks)
map        | map[string]interface{}         | Map of key-value pairs.
list       | []interface{}                  | List of values.

## Values and References
Currently, the workflow engine is very generous with storing all data received from and sent to functions.
Although this helps debuggability, and simplicity, with data-intensive functions - functions that for example output 
video or large images - this becomes a less-than optimal situation wasting storage.

As part of our future work, we are looking into a way to avoid copying and storing all data in the workflow engine's 
storage.
The challenge here is to optimize the usage of storage vs. keeping the simplicity of the current execution model.
One solution that is promising is to have a middleware component that stores and replaces large data sources with 
references to the data instead.