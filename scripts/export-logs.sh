# get all the webhook internal logs from the past 72 hours
mkdir -p ./datree-webhook-logs && for podId in $(kubectl get pods -n datree --output name); do kubectl logs -n datree --since=72h $podId > ./datree-webhook-logs/pod-"$(echo $podId | cut -c 5-)".txt; done
