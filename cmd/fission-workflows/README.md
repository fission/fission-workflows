# Fission Workflows CLI (fission-workflows)

This provides a CLI for Fission Workflows, providing additional functionality for workflows over the regular functions 
available in the Fission CLI.

_Note: this is an experimental CLI -- in the near future fission-workflows will be merged into the Fission CLI._

## Installation
```bash
go install
``` 

## Usage
```bash
fission-workflows status # View whether the Fission Workflows deployment can be reached.

fission-workflows workflow get # List all workflows in the workflow engine.

fission-workflows workflow get <id> # Get the definition of a specific workflow

fission-workflows invocation get # List all invocations so-far (both in-progress and finished)

fission-workflows invocation get <id> # Get all info of a specific invocation

fission-workflows invocation status <id> # Get a concise overview of the progress of an invocation 
```
