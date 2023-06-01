#!/bin/bash

while true; do
  kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=datree-webhook-server" --output=name) 8443 5555
  if [ $? -ne 0 ]; then
    echo "Rollout failed, running port-forward command again"
    kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=datree-webhook-server" --output=name) 8443 5555
  fi
  sleep 1
done