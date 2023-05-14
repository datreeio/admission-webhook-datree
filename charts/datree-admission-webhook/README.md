# Datree Admission Webhook

A Kubernetes validating webhook for policy enforcement within the cluster, on every CREATE, APPLY and UPDATE operation
on a resource.

## TL;DR

```bash
  # Install and create namespace with Helm
  helm repo add datree-webhook https://datreeio.github.io/admission-webhook-datree/
  helm repo update

  # Already existing `datree` namespace
  kubectl create ns datree
  helm install -n datree datree-webhook datree-webhook/datree-admission-webhook --set datree.token=<DATREE_TOKEN>
```

### Prerequisites

Helm v3.0.0+

## Configuration Options

Datree admission webhook can be configured via the helm values file under `datree` key:

### Datree Configuration options

```
datree:
  token: <DATREE_TOKEN>     # The token used to link the CLI to your dashboard.
  policy: ""                # The name of the policy to check, e.g: staging. (string, optional)
  verbose: ""               # Display 'How to Fix' link for failed rules in output. (boolean ,optional)
  output: ""                # The format output of the policy check results: yaml, json, xml, simple, JUnit. (string ,optional)
  noRecord: ""              # Don’t send policy checks metadata to the backend. (boolean ,optional)
  enforce: ""               # Block resources that fail the policy check. (boolean ,optional)
  clusterName: ""           # The name of the cluster link for cluster name in your dashboard. (string ,optional)
```

For further information about Datree flags see [CLI arguments](https://hub.datree.io/setup/cli-arguments).

### Parameters

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| namespace | string | `""` | The name of the namespace all resources will be created in, if not specified in the release. |
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
| resources | object | `{}` | The resource request/limits for the webhook container image |
| nodeSelector | object | `{}` | Used to select on which node a pod is scheduled to run |
| affinity | object | `{}` | Offers more expressive syntax for fine-grained control of how Pods are scheduled to specific nodes |
| tolerations | list | `[]` |  |
| clusterScanner.resources | object | `{}` | The resource request/limits for the scanner container image |
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
