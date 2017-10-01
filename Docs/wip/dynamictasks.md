# Control Flow Constructs

Control Flow constructs are tasks or computations that change the structure of a workflow. 
This document tracks the design and implementation notes for implementing control constructs in Fission Workflow.

## Use Cases
- Conditional branches; execute branches only if particular conditions hold.
- Data iteration; run branches on items in a collection
- While conditions
- Wrapping a function in a common pattern; e.g. Fallthrough tasks.

## Design
- A control flow construct is just another function, only it returns a task (in other words: a structure) instead of data.
- Should be pluggable, allowing users to write their own control flow constructs.
- In order to keep reasoning about flows simple, tasks can only be executed once.

## Implementation 
- To differentiate between data or workflow, a control flow is a specific type: `internal/flow` 

## Issues
- How to allow functions to access workflow/invocation metadata. 
- How to model a 'flow'? And how to work with them, cancel flows? Alternatively, is it needed?
