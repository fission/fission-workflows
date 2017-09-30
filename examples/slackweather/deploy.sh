#!/usr/bin/env bash

fission env create --name nodejs --image fission/node-env:0.3.0

kubectl apply -f https://raw.githubusercontent.com/fission/functions/master/slack/function.yaml
fission fn create --name tempconv --env nodejs --code ./tempconv.js
fission fn create --name wunderground-conditions --env nodejs --code ./wunderground-conditions.js
fission fn create --name formdata2json --env nodejs --code ./formdata2json.js

fission fn create --name slackweather --env workflow --code ./slackweather.wf.json
fission fn create --name slackslashweather --env workflow --code ./slackslashweather.wf.json
