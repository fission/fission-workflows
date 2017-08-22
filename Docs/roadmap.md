# Roadmap
Roughly ordered based on priority.

- **Performance optimizations**: Pre-warm functions, the main optimization that would make the workflow engine faster than calling functions yourself. 
The workflow engine interprets the dependency graph, notices that a certain function will be called ‘soon’, and triggers the specialization of the function by Fission before it has to call it.
- **Workflow language**: Provide simple Python library, that allows users to generate workflows by just writing code and feeding it to a parser.
- **Visualization**: Provide a simple visualization tool, that allows users to see the status of the workflows and See the execution visually of a workflow invocation.
- **Error handling in workflows**: Exception Handling, allowing users to deal with errors in functions. 
For example, users could provide a fallback function that is executed when the first function fails.
- **Observability**: Add initial telemetry support, measuring function runtime, workflow engine overhead.
- **Function Versioning**: provide support for dealing with versions of functions. Users can indicate how to workflow should deal with new versions. 
Examples of strategies for dealing with versioning: never upgrade, canary deployment, blue/green deployment...
- Build a user-friendly GUI.
- Add support for passing around data ‘by reference’. 
Currently, all data passes through. the workflow engine, which can be costly for data-intensive tasks. 
Support can be added to allow functions to pass data to each other directly.
- Create an online ‘catalog’ of functions, such that users can re-use functions created by others.
- Inline data transformation, the ability to perform some basic data transformations without having to specify a task for it. (e.g. transform data to uppercase, remove whitespace).
- Extract components of the engine in order to improve the scalability. 
- Improve structure of scheduler to enable smarter scheduling.
- Create benchmark that measures the performance of Fission Workflow.
