# Slackweather Example
This directory contains an example workflow that combines the Wunderground weather API and the Slack API.

## Setup
- Make sure that you have a running Fission Cluster with the workflow engine, see the [readme](../../README.md) for details.
- Replace `API_KEY` with your Wunderground API key and `hook_url` with the hook to your target Slack channel.
- Deploy workflows, using the [deploy.sh](./deploy.sh) script, or run the commands listed there manually.

## Usage

To invoke:
```bash
curl -XPOST -d 'Sunnyvale, CA' $FISSION_ROUTER/fission-function/slackweather
```

Or, setup a [slash command](https://api.slack.com/slash-commands) and point it to the `slackslashweather` workflow
