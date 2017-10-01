# Internal Functions

Internal Functions are functions that are executed within the workflow engines.
This resu

## Built-in

### Sleep
Description: A function that simply sleeps for the provided duration. 
Output: nil

| Input   | Type   | Description | Example |
|---------|--------|-------------|---------|
| default | string | Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". | "300ms", "-1.5h" or "2h45m" |
|         | int    | In miliseconds | 1000 |

