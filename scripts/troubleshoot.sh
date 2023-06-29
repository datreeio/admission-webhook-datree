#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;36m'
PLAIN='\033[0m'
bold=$(tput bold)
normal=$(tput sgr0)



# Run command and show stdout and stderr
run_command() {
  echo -e "${BLUE}Running: ${YELLOW}$@${PLAIN}"
  "$@"
  echo ""
}


run_command kubectl version

run_command kubectl config current-context

run_command kubectl get ns

run_command helm list -n datree

run_command kubectl get all -n datree

run_command kubectl get ns kube-system -o jsonpath='{.metadata.uid}'

run_command kubectl get validatingwebhookconfigurations

run_command kubectl get mutatingwebhookconfigurations

# get image version of cluster-scanner-server
run_command kubectl get deployments/datree-cluster-scanner-server  -n datree -o jsonpath='{.spec.template.spec.containers[0].image}'

# get latest 10 logs from the cluster-scanner-server
run_command kubectl logs deployments/datree-cluster-scanner-server -n datree | head -n 10
run_command kubectl logs deployments/datree-cluster-scanner-server -n datree --tail=10

#get lates 10 logs from the webhook-server
run_command kubectl logs deployments/datree-webhook-server -n datree | head -n 10
run_command kubectl logs deployments/datree-webhook-server -n datree --tail=10

run_command kubectl describe -n datree deployment/datree-cluster-scanner-server

run_command kubectl describe -n datree deployment/datree-webhook-server
