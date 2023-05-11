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
| namespace234 | string | `""` |  |
| replicaCount | int | `2` |  |
| customLabels | object | `{}` |  |
| customAnnotations | object | `{}` |  |
| rbac.serviceAccount.create | bool | `true` |  |
| rbac.serviceAccount.name | string | `"datree-webhook-server"` |  |
| rbac.clusterRole.create | bool | `true` |  |
| rbac.clusterRole.name | string | `"datree-webhook-server-cluster-role"` |  |
| datree.token | string | `nil` |  |
| datree.existingSecret.name | string | `""` |  |
| datree.existingSecret.key | string | `""` |  |
| datree.verbose | string | `nil` |  |
| datree.output | string | `nil` |  |
| datree.noRecord | string | `nil` |  |
| datree.clusterName | string | `nil` |  |
| datree.scanIntervalHours | int | `1` |  |
| datree.configFromHelm | bool | `false` |  |
| datree.policy | string | `nil` |  |
| datree.enforce | string | `nil` |  |
| datree.customSkipList[0] | string | `"(.*);(.*);(^aws-node.*)"` |  |
| image.repository | string | `"datree/admission-webhook"` |  |
| image.tag | string | `nil` |  |
| image.pullPolicy | string | `"Always"` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `25000` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| resources | object | `{}` |  |
| nodeSelector | object | `{}` |  |
| affinity | object | `{}` |  |
| tolerations | list | `[]` |  |
| clusterScanner.resources | object | `{}` |  |
| clusterScanner.annotations | object | `{}` |  |
| clusterScanner.rbac.serviceAccount.create | bool | `true` |  |
| clusterScanner.rbac.serviceAccount.name | string | `"cluster-scanner-service-account"` |  |
| clusterScanner.rbac.clusterRole.create | bool | `true` |  |
| clusterScanner.rbac.clusterRole.name | string | `"cluster-scanner-role"` |  |
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
