#!/bin/bash

# In case $BUILD_ENVIRONMENT env var is not set, will pass default - "staging"
if [ "$BUILD_ENVIRONMENT" = "" ]; then
  buildEnv="staging"
else
  buildEnv=$BUILD_ENVIRONMENT
fi

# In case $IMAGE_NAME env var is not set, will pass default - "webhook-server"
if [ "$IMAGE_NAME" = "" ]; then
  imageName="webhook-server"
else
  imageName=$IMAGE_NAME
fi

if [[ "$GIT_BRANCH" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  #  In production build, will inject the tag "x.x.x" as version
  version=$GIT_BRANCH
else
  #  In staging, will inject the tag "staging" instead
  version="staging"
fi

# "DOCKER_BUILDKIT=1" will make sure buildkit is enabled (out of the box for docker 1.18+ and dockerhub builds, so just in case)
# "--build-arg BUILDKIT_INLINE_CACHE=1" allows for better caching but not mandatory
DOCKER_BUILDKIT=1 docker build --build-arg BUILD_ENVIRONMENT="$buildEnv" --build-arg BUILDKIT_INLINE_CACHE=1 --build-arg WEBHOOK_VERSION="$version" -t "$imageName" .
