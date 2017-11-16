#!/usr/bin/env bash

# Setup external services
kubectl apply -f redis/redis.yaml

# Setup environments
fission env create --name python3 --version 2 --image fission/python-env:0.4.0rc --builder fission/python-build-env:0.4.0rc
fission env create --name binary --image fission/binary-env:0.4.0rc

# Prepare functions
zip -jr notify-pushbullet.zip notify-pushbullet/
zip -jr extract-ogp.zip extract-ogp/
zip -jr parse-article.zip parse-article/
zip -jr redis.zip redis/

# Setup functions
fission fn create --env python3 --name notify-pushbullet --src notify-pushbullet.zip --entrypoint "notify.main" --buildcmd "./build.sh"
fission fn create --env python3 --name extract-ogp --src extract-ogp.zip --entrypoint "ogp.main" --buildcmd "./build.sh"
fission fn create --env python3 --name parse-article --src parse-article.zip --entrypoint "article.main" --buildcmd "./build.sh"
fission fn create --env binary  --name http --deploy http/http.sh

fission fn create --env python3 --name redis-list --src redis.zip --entrypoint "user.list" --buildcmd "./build.sh"
fission fn create --env python3 --name redis-append  --src redis.zip --entrypoint "user.append" --buildcmd "./build.sh"
fission fn create --env python3 --name redis-get --src redis.zip --entrypoint "user.get" --buildcmd "./build.sh"
fission fn create --env python3 --name redis-set  --src redis.zip --entrypoint "user.set" --buildcmd "./build.sh"