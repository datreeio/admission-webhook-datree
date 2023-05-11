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
| namespace | string | `"datree"` | The name of the namespace all resources will be created in. |
| replicaCount | int | `2` | The number of Datree webhook-server replicas to deploy for the webhook. |
| customLabels | object | `{}` | Additional labels to add to all resources. |
| customAnnotations | object | `{}` | Additional annotations to add to all resources. |
| rbac.serviceAccount.create | bool | `true` | Create a ServiceAccount |
| rbac.serviceAccount.name | string | `"datree-webhook-server"` | The ServiceAccount name |
| rbac.clusterRole.create | bool | `true` | Create a ClusterRole |
| rbac.clusterRole.name | string | `"datree-webhook-server-cluster-role"` | The ClusterRole name |
| datree.token | string | `nil` | The token used to link the CLI to your dashboard. (string, required) |
| datree.existingSecret | object | `{"key":"","name":""}` | The token may also be provided via secret, note if the existingSecret is provided the token field above is ignored. |
| datree.verbose | string | `nil` | Display 'How to Fix' link for failed rules in output. (boolean ,optional) |
| datree.output | string | `nil` | The format output of the policy check results: yaml, json, xml, simple, JUnit. (string ,optional) |
| datree.noRecord | string | `nil` | Don’t send policy checks metadata to the backend. (boolean ,optional) |
| datree.clusterName | string | `nil` | The name of the cluster link for cluster name in your dashboard. (string ,optional) |
| datree.scanIntervalHours | int | `1` | How often should the scan run in hours. (int, optional, default: 1 ) |
| datree.configFromHelm | bool | `false` | If false, the webhook will be configured from the dashboard, otherwise it will be configured from here. Affected configurations: policy, enforce, customSkipList. |
| datree.policy | string | `nil` | The name of the policy to check, e.g: staging. (string, optional) |
| datree.enforce | string | `nil` | Block resources that fail the policy check. (boolean ,optional) |
| datree.customSkipList | list | `["(.*);(.*);(^aws-node.*)"]` | Excluded resources from policy checks. ("namespace;kind;name" ,optional) |
| image.repository | string | `"datree/admission-webhook"` | Image repository |
| image.tag | string | `nil` | The image release tag to use |
| image.pullPolicy | string | `"Always"` | Image pull policy |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":25000,"seccompProfile":{"type":"RuntimeDefault"}}` | Security context applied on the containers |
| resources | object | `{}` | The resource request/limits for the container image |
| nodeSelector | object | `{}` |  |
| affinity | object | `{}` |  |
| tolerations | list | `[]` |  |
| clusterScanner.resources | object | `{}` |  |
| clusterScanner.annotations | object | `{}` |  |
| clusterScanner.rbac.serviceAccount.create | bool | `true` | Create the ServiceAccount |
| clusterScanner.rbac.serviceAccount.name | string | `"cluster-scanner-service-account"` | The ServiceAccount name |
| clusterScanner.rbac.clusterRole.create | bool | `true` | Create the ClusterRole |
| clusterScanner.rbac.clusterRole.name | string | `"cluster-scanner-role"` | The ClusterRole name |
| clusterScanner.rbac.clusterRoleBinding.name | string | `"cluster-scanner-role-binding"` |  |
| clusterScanner.image.repository | string | `"datree/cluster-scanner"` |  |
| clusterScanner.image.pullPolicy | string | `"Always"` |  |
| clusterScanner.image.tag | string | `nil` |  |
| hooks.timeoutTime | string | `nil` |  |
| hooks.ttlSecondsAfterFinished | string | `nil` |  |
| hooks.image.repository | string | `"clastix/kubectl"` |  |
| hooks.image.tag | string | `"v1.25"` |  |
| hooks.image.pullPolicy | string | `"IfNotPresent"` |  |
| validatingWebhookConfiguration.failurePolicy | string | `"Ignore"` |  |
