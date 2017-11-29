#!/bin/sh

# fortune - generate wise words
#
# Output: string (plain text)

# TODO move this to the build step once the binary environment is a v2 environment
if ! hash fortune 2>/dev/null; then
    apk add fortune > /dev/null
fi

# Pretty useless, but used to show headers
prefix=${HTTP_PREFIX}

if [ ! -z "${prefix}" ] ; then
    printf "${prefix}"
fi

fortune -s
