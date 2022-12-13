# Datree Admission Webhook

<p align="center">
<img src="https://github.com/datreeio/admission-webhook-datree/blob/main/internal/images/diagram.png" width="80%" />
</p>
  
# Overview
Datree offers cluster integration that allows you to validate your resources against your configured policy upon pushing them into a cluster, by using an admission webhook.

The webhook will catch **create**, **apply** and **edit** operations and initiate a policy check against the configs associated with each operation. If any misconfigurations are found, the webhook will reject the operation, and display a detailed output with instructions on how to resolve each misconfiguration.


👉🏻 For the full documentation click [here](https://hub.datree.io).

# Values
The following table lists the configurable parameters of the Datree chart and their default values.

| Parameter                             | Description               | Default       |
| ------------------------------------- | -------------------- | ------------------------------- |
| replicaCount                          | The number of Datree webhook-server replicas to deploy for the webhook.                   | 2                    |
| customLabels                          | Additional labels for Datree webhook-server pods.                                         | {}      |
| customAnnotations                     | Additional annotations to add to all resources.                                           | {}              |
| rbac.serviceAccount.create            | Create a ServiceAccount                                                                   | true      |
| rbac.serviceAccount.name              | The ServiceAccount name                                                                   | webhook-server-datree |
| rbac.clusterRole.create               | Create a ClusterRole                                                                      | true |
| rbac.clusterRole.name                 | The ClusterRole name                                                                      | webhook-server-datree |
| image.repository                      | Image repository.                                                                         | datree/admission-webhook |
| image.tag                             | The image release tag to use.                                                             | Defaults to Chart appVersion |
| image.pullPolicy                      | Image pull policy                                                                         | Always |
| securityContext                       | Security context applied on the container.                                                | {"allowPrivilegeEscalation":false,"readOnlyRootFilesystem":true, "runAsNonRoot":true} |
| resources                             | The resource request/limits for the container image.                                      | limits :cpu: 1000m, memory: 512Mi requests: cpu:100m, memory:256Mi|
| datree.token                          | The token used to link the CLI to your dashboard. (required)                              | nil|
| datree.clusterName                    | The name of the cluster link for cluster name in your dashboard                           | nil |
| datree.policy                         | The name of the policy to check, e.g: staging. (optional)                                 | "" (i.e "default")|
| datree.verbose                        | Display 'How to Fix' link for failed rules in output. (optional)                          | false|
| datree.output                         | The format output of the policy check results: yaml, json, xml, simple, JUnit. (optional) | "" (i.e beautiful😊)|
| datree.noRecord                       | Don’t send policy checks metadata to the backend. (optional)                              | false|
| datree.enforce                        | Block resources that fail the policy check. (optional)                                    | false|
| hooks.waitForServerRollout.sleepyTime | The waiting time before the webhook-server is ready to receive requests.                  | nil|
| hooks.waitForServerRollout.image      | An image for running sleep command                                                        | {"repository": "alpine", "sha":"sha256:1304f174557314a7ed9eddb4eab12fed12cb0cd9809e4c28f29af86979a3c870", "pullPolicy":"Always"} |
| hooks.labelNamespace.image.           | An image for running kubectl label command                                                | {"repository": "bitnami/kubectl", "sha":"sha256:d3c17f1dc6e665dcc78e8c14a83ae630bc3d65b07ea11c5f1a012c2c6786d039", "pullPolicy":"Always"} |
| nodeSelector                          | Used to select on which node a pod is scheduled to run | nil |
| affinity                              | Offers more expressive syntax for fine-grained control of how Pods are scheduled to specific nodes | nil |

