#!/bin/bash
selector=$1
ports=$2
while true; do
  kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=$selector" --output=name) $ports
  if [ $? -ne 0 ]; then
    echo "Rollout failed, running port-forward command again"
    kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=$selector" --output=name) $ports
  fi
  sleep 1
done
