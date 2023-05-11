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
| namespace | string | `""` |  |
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
