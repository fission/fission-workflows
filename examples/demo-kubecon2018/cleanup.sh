#!/bin/bash

set -x

fission fn delete --name fortune
fission fn delete --name whalesay
fission fn delete --name fortunewhale
kubectl -n fission-function delete $(kubectl -n fission-function get po -o name | grep workflow)
kubectl -n fission delete $(kubectl -n fission get po -o name | grep router)
kubectl -n fission delete $(kubectl -n fission get po -o name | grep executor)
