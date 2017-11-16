#!/bin/sh

# TODO move to buildstep once binary has builder-support
if ! hash curl 2>/dev/null; then
    apk add --no-cache curl > /dev/null
fi

# TODO add options (e.g. not just HTTP GET)
cat - | xargs curl -L