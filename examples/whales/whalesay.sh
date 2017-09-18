#!/bin/sh

if ! hash cowsay 2> /dev/null; then
    apk update > /dev/null
    apk add curl perl > /dev/null
    curl https://raw.githubusercontent.com/docker/whalesay/master/cowsay > /bin/cowsay 2> /dev/null
    chmod +x /bin/cowsay
    mkdir -p /usr/local/share/cows/
    curl https://raw.githubusercontent.com/docker/whalesay/master/docker.cow > /usr/local/share/cows/default.cow 2> /dev/null
fi

cowsay
