#!/bin/bash
set -ex

# required to have access to the repository tags in the github action: https://github.com/actions/checkout/issues/206#issuecomment-607496604
git fetch --prune --unshallow --tags

# code duplication from the datree repository: https://github.com/datreeio/datree/blob/cb6948f1b3cf36abf10201f9a4fd1a2ba7865fea/scripts/deploy_release_candidate.sh#L4
MAJOR_VERSION=0
MINOR_VERSION=1

currentVersion=$(git tag --sort=-version:refname | grep -E "^${MAJOR_VERSION}\.${MINOR_VERSION}\.[0-9]+$" | head -n 1)

if [ "$currentVersion" == "" ]; then
    nextVersion=$MAJOR_VERSION.$MINOR_VERSION.0
else
    nextVersion=$(echo "$currentVersion" | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')
fi

echo "$nextVersion"
