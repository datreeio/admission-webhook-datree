# -*- mode: Python -*-



docker_build('datree/admission-webhook', './', dockerfile = './Dockerfile-tilt', build_args={
    "BUILD_ENVIRONMENT":"staging",
    "WEBHOOK_VERSION":"0.0.1",
})
docker_build('datree/cluster-scanner-staging', '../cluster-scanner', dockerfile = '../cluster-scanner/Dockerfile-tilt', build_args={
    "BUILD_ENVIRONMENT":"staging",
    "WEBHOOK_VERSION":"0.0.1",
})
load('ext://namespace', 'namespace_create')
namespace_create('datree')

k8s_yaml(helm('./charts/datree-admission-webhook/', name='admission-webhook', values='./values-tilt.yaml', namespace='datree'))


local_resource(
    name='datree-webhook-server-debuging',
    serve_cmd='bash ./webhook-debugging.sh',
)

local_resource(
    name='webhook - restart port-forwarding',
    serve_cmd='bash ./webhook-restart-port-forwarding.sh',
)

local_resource(
    name='cluster-scanner debugging',
    serve_cmd='bash ./cluster-scanner-debugging.sh',
)

local_resource(
    name='cluster-scanner - restart port-forwarding',
    serve_cmd='bash ./scanner-restart-port-forwarding.sh',
)


# local_resource(
#     name='wait for datree-cluster-scanner-server to be ready',
#     serve_cmd='kubectl port-forward --namespace datree $(kubectl get pods --namespace datree --selector "app=datree-cluster-scanner-server" --output=name) 8080 5556',
#     readiness_probe = probe(
#         # kubectl rollout status -n datree deployment/datree-cluster-scanner-server'
#         exec = exec_action(
#             command=['kubectl', 'rollout', 'status', '-n', 'datree', 'deployment/datree-cluster-scanner-server'],
#             ),
#         period_secs = 1,
#         ),        
#     )
# # Define the path to your Helm chart
# chart_path = '/Users/nivweiss/Documents/datree/admission-webhook-datree/charts/datree-admission-webhook/'
# values_file = './values-tilt.yaml'

# Define the Tilt configuration
# Use the `helm_chart` function to create a Helm chart deployment in Tilt
# helm_chart(chart_path, values=[values_file])

# Optionally, you can specify additional Tilt configurations for your project, such as sync rules, port forwarding, etc.

# Example sync rule to watch for changes in your Helm chart
# sync.sync(chart_path)

# Example port forwarding rule to access services within the cluster
# k8s_port_forward('service-name', local_port=8080, remote_port=80)

# You can add more Tilt configurations based on your project requirements

# Save this Tiltfile and run 'tilt up' to start Tilt and deploy your custom Helm chart.
