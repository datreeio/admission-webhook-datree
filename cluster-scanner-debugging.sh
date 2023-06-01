#!/bin/bash

while true; do
  kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=datree-cluster-scanner-server" --output=name) 8080 5556
  if [ $? -ne 0 ]; then
    echo "Rollout failed, running port-forward command again"
    kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=datree-cluster-scanner-server" --output=name) 8080 5556
  fi
  sleep 1
done