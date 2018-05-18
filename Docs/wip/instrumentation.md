# Instrumentation

This document contains description and high-level documentation on the instrumentation of the system.
The instrumentation here is considered to encompass tracing, logging and metrics.

Terminology:
- ctrl: name of the controller if applicable
- wf: id of the workflow if applicable
- wfi: id of the workflow invocation if applicable
- component: name of component (controller, api, apiserver, fnenv)

## Logging

## Metrics

For metrics the Prometheus time series and monitoring system is used.
The following metrics are collected and available under `:8080/metrics` in the Prometheus data format.

Metrics:
- Active invocations (Gauge)
- Cache size (Gauge)

- Completed invocations (Counter)
    - Failed invocations (Counter)
    - Successful invocations (Counter) 
    - Aborted invocations (Counter) 
