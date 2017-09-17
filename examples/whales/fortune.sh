#!/bin/sh

if ! hash fortune 2>/dev/null; then
    apk add fortune > /dev/null
fi

fortune -s
