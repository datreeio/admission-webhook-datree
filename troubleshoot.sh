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

# get image version of datree scan-job
run_command kubectl get job.batch/scan-job -n datree -o jsonpath='{.spec.template.spec.containers[0].image}'

run_command kubectl logs job.batch/scan-job -n datree

run_command kubectl get validatingwebhookconfigurations

run_command kubectl get mutatingwebhookconfigurations

