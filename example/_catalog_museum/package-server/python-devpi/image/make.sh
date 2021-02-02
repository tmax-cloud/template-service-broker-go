#!/bin/bash

set -ex

TAG=devpi:latest
REGISTRY=192.168.6.110:5000

TAG_REMOTE="$REGISTRY/$TAG"

docker build --no-cache --rm --network host -t $TAG .
docker tag $TAG $TAG_REMOTE
docker push $TAG_REMOTE

