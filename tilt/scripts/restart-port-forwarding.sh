#!/bin/bash
# This script is used to restart the port-forwarding command when the pod is not ready


# deployment_name is the name of the deployment that we want to check
deployment_name=$1
# tilt_app is the name of the tilt app that we want to restart
tilt_app=$2

echo "starting port-forwarding"

function get_pod_status() {
    label=$1
    pod_status=$(kubectl get pods -n datree -l app=$label -o jsonpath='{.items[0].status.containerStatuses[0].ready}')
}

while true; do
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
