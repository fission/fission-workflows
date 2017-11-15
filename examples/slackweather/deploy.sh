#!/usr/bin/env bash

fission env create --name nodejs --image fission/node-env:0.3.0

kubectl apply -f https://raw.githubusercontent.com/fission/functions/master/slack/function.yaml
fission fn create --name tempconv --env nodejs --deploy ./tempconv.js
fission fn create --name wunderground-conditions --env nodejs --deploy ./wunderground-conditions.js
fission fn create --name formdata2json --env nodejs --deploy ./formdata2json.js

fission fn create --name slackweather --env workflow --src ./slackweather.wf.yaml
fission fn create --name slackslashweather --env workflow --src ./slackslashweather.wf.yaml
