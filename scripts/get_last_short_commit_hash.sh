#!/bin/bash

# will fail in the following scenarios:
# 1. Release is performed and then immediately a new commit is pushed to main - will release a production tag with commit hash (instead of semantic version)
# 2. Release is performed right after merging to main - will release a staging tag with semantic version (instead of commit hash)
# 3. Release is performed twice in a row (without pushing a new commit) - will try to release the same docker tag again (and fail)

LATEST_COMMIT_TAG=$(git tag --points-at HEAD)
TAG=""
if ([ -z "$LATEST_COMMIT_TAG" ]); then
  # No commit tag found - staging deployment
  SHORTHASH="$(git rev-parse --short HEAD)"
  TAG=$DOCKER_REPO$SHORTHASH
else
  # Commit tag found - production deployment
  TAG=$DOCKER_REPO$LATEST_COMMIT_TAG
fi
