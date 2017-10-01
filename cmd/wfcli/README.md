# Fission Workflows CLI (wfcli)

This provides a CLI for Fission Workflows, providing additional functionality for workflows over the regular functions 
available in the Fission CLI.

_Note: this is an experimental CLI -- in the near future wfcli will be merged into the Fission CLI._

## Installation
```bash
go install
``` 

## Usage
```bash
wfcli status # View whether the Fission Workflows deployment can be reached.

wfcli workflow get # List all workflows in the workflow engine.

wfcli workflow get <id> # Get the definition of a specifc workflow

wfcli invocation get # List all invocations so-far (both in-progress and finished)

wfcli invocation get <id> # Get all info of a specific invocation

wfcli invocation status <id> # Get a concise overview of the progress of an invocation 
```
