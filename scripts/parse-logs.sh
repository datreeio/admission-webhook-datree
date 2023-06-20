#!/bin/bash

# get all the webhook internal logs from the past 72 hours
rm -r ./datree-webhook-logs
mkdir -p ./datree-webhook-logs

for podId in $(kubectl get pods -n datree --output name); 
do {
 podName=$(echo "$podId" | cut -c 5-)
 kubectl logs -n datree --since=72h "$podId" > ./datree-webhook-logs/pod-"$podName".logs; 
 ts-node ./scripts/parseLogs.ts ./datree-webhook-logs/pod-"$podName".logs ./datree-webhook-logs/parsed-pod-"$podName".json;
} 
done
