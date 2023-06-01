#!/bin/bash
echo "starting port-forwarding"
deployment_name="datree-webhook-server"
tilt_app="datree-webhook-server-debuging"

function get_replicas() {
  deployment_name=$1
  available_replicas=$(kubectl get deployments.apps -n datree $deployment_name -o jsonpath='{.status.availableReplicas}')
  replicas=$(kubectl get deployments.apps -n datree $deployment_name -o jsonpath='{.status.replicas}') 
}

function get_pod_name() {
  pod_name=$(kubectl get pods -n datree -l app=$1 -o jsonpath='{.items[0].metadata.name}')
}

function get_pod_status() {
  pod_status=$(kubectl get pods -n datree -l app=$1 -o jsonpath='{.items[0].status.containerStatuses[0].ready}')
}
# kubectl get pods -n datree -l app=datree-webhook-server -o jsonpath='{.items[0].status.phase}
# kubectl get pod -n datree datree-webhook-server-5966d786c9-rwgnf -o jsonpath='{.status.containerStatuses[0].ready}'
while true
do 
if ! get_pod_status $deployment_name; then
  echo " Failed to get pod status for $deployment_name"
  # If the commands fail, run the tilt trigger command
  tilt trigger "$tilt_app"
  echo "-------------------"
  continue
elif [[ $pod_status == "" ]]; then
  echo "Pod status is empty"
  continue
elif [[ $pod_status == "false" ]]; then
    tilt trigger "$tilt_app"
    echo -e "Running $tilt_app port-forward command"
    echo "-------------------"
fi
sleep 1

done
