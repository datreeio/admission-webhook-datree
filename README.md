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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| namespace | string | `nil` | The name of the namespace all resources will be created in, if not specified in the release. |
| replicaCount | int | `2` | The number of Datree webhook-server replicas to deploy for the webhook. |
| customLabels | object | `{}` | Additional labels to add to all resources. |
| customAnnotations | object | `{}` | Additional annotations to add to all resources. |
| rbac.serviceAccount | object | `{"create":true,"name":"datree-webhook-server"}` | Create service Account for the webhook |
| rbac.clusterRole | object | `{"create":true,"name":"datree-webhook-server-cluster-role"}` | Create service Account for the webhook |
| datree.token | string | `nil` | The token used to link the CLI to your dashboard. (string, required) |
| datree.existingSecret | object | `{"key":"","name":""}` | The token may also be provided via secret, note if the existingSecret is provided the token field above is ignored. |
| datree.verbose | string | `nil` | Display 'How to Fix' link for failed rules in output. (boolean, optional) |
| datree.output | string | `nil` | The format output of the policy check results: yaml, json, xml, simple, JUnit. (string, optional) |
| datree.noRecord | string | `nil` | Don’t send policy checks metadata to the backend. (boolean, optional) |
| datree.clusterName | string | `nil` | The name of the cluster link for cluster name in your dashboard (string ,optional) |
| datree.scanIntervalHours | int | `1` | How often should the scan run in hours. (int, optional, default: 1 ) |
| datree.configFromHelm | bool | `false` | If false, the webhook will be configured from the dashboard, otherwise it will be configured from here. Affected configurations: policy, enforce, customSkipList. |
| datree.policy | string | `nil` | The name of the policy to check, e.g: staging. (string, optional) |
| datree.enforce | string | `nil` | Block resources that fail the policy check. (boolean ,optional) |
| datree.customSkipList | list | `["(.*);(.*);(^aws-node.*)"]` | Excluded resources from policy checks. ("namespace;kind;name" ,optional) |
| image.repository | string | `"datree/admission-webhook"` | Image repository for the webhook |
| image.tag | string | `nil` | The image release tag to use for the webhook |
| image.pullPolicy | string | `"Always"` | Image pull policy for the webhook |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":25000,"seccompProfile":{"type":"RuntimeDefault"}}` | Security context applied on the containers |
| resources | object | `{}` | The resource request/limits for the container image |
| nodeSelector | object | `{}` | Used to select on which node a pod is scheduled to run |
| affinity | object | `{}` | Offers more expressive syntax for fine-grained control of how Pods are scheduled to specific nodes |
| tolerations | list | `[]` |  |
| clusterScanner.resources | object | `{}` |  |
| clusterScanner.annotations | object | `{}` |  |
| clusterScanner.rbac.serviceAccount | object | `{"create":true,"name":"cluster-scanner-service-account"}` | Create service Account for the scanner |
| clusterScanner.rbac.clusterRole | object | `{"create":true,"name":"cluster-scanner-role"}` | Create service Role for the scanner |
| clusterScanner.rbac.clusterRoleBinding | object | `{"name":"cluster-scanner-role-binding"}` | Create service RoleBinding for the scanner |
| clusterScanner.image.repository | string | `"datree/cluster-scanner"` | Image repository for the scanner |
| clusterScanner.image.pullPolicy | string | `"Always"` | Image pull policy for the scanner |
| clusterScanner.image.tag | string | `nil` | The image release tag to use for the scanner |
| hooks.timeoutTime | string | `nil` | The timeout time the hook will wait for the webhook-server is ready. |
| hooks.ttlSecondsAfterFinished | string | `nil` |  |
| hooks.image.repository | string | `"clastix/kubectl"` |  |
| hooks.image.tag | string | `"v1.25"` |  |
| hooks.image.pullPolicy | string | `"IfNotPresent"` |  |
| validatingWebhookConfiguration.failurePolicy | string | `"Ignore"` |  |
