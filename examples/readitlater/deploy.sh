#!/usr/bin/env bash

# Setup environments
fission env create --name python --version 2 --image fission/python-env:0.4.0rc --builder fission/python-build-env:0.4.0rc
# TODO buildcmd

# Prepare functions
zip -jr notify-pushbullet.zip notify-pushbullet/

# Setup functions
fission fn create --name notify-pushbullet --source notify-pushbullet.zip --entrypoint "notify.main"