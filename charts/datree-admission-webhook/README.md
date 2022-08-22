# Datree Admission Webhook

A Kubernetes validating webhook for policy enforcement within the cluster, on every CREATE, APPLY and UPDATE operation on a resource.

## TL;DR

```bash
  helm repo add datree-webhook https://datreeio.github.io/admission-webhook-datree/
  helm repo update
  helm install datree-webhook datree-webhook/datree-admission-webhook --create-namespace --set datree.token=<DATREE_TOKEN>
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
```

For further information about Datree flags see [CLI arguments](https://hub.datree.io/setup/cli-arguments).

### Parameters

| Parameter                  | Description                                                                               | Default                                                                                                 |     |     |
| -------------------------- | ----------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | --- | --- |
| namespace                  | The name of the namespace all resources will be created in.                               | datree                                                                                                  |     |     |
| replicaCount               | The number of Datree webhook-server replicas to deploy for the webhook.                   | 2                                                                                                       |     |     |
| customLabels               | Additional labels for Datree webhook-server pods.                                         | {}                                                                                                      |     |     |
| rbac.serviceAccount.create | Create a ServiceAccount                                                                   | true                                                                                                    |     |     |
| rbac.serviceAccount.name   | The ServiceAccount name                                                                   | webhook-server-datree                                                                                   |     |     |
| rbac.clusterRole.create    | Create a ClusterRole                                                                      | true                                                                                                    |     |     |
| rbac.clusterRole.name      | The ClusterRole name                                                                      | webhook-server-datree                                                                                   |     |     |
| image.repository           | Image repository.                                                                         | datree/admission-webhook                                                                                |     |     |
| image.tag                  | The image release tag to use.                                                             | Defaults to Chart appVersion                                                                            |     |     |
| image.pullPolicy           | Image pull policy                                                                         | Always                                                                                                  |     |     |
| securityContext            | Security context applied on the container.                                                | {"allowPrivilegeEscalation":false,"readOnlyRootFilesystem":true, "runAsNonRoot":true,"runAsUser":25000} |     |     |
| resources                  | The resource request/limits for the container image.                                      | limits :cpu: 1000m, memory: 512Mi requests: cpu:100m, memory:256Mi                                      |     |     |
| datree.token               | The token used to link the CLI to your dashboard. (required)                              | nil                                                                                                     |     |     |
| datree.policy              | The name of the policy to check, e.g: staging. (optional)                                 | "" (i.e "default")                                                                                      |     |     |
| datree.verbose             | Display 'How to Fix' link for failed rules in output. (optional)                          | false                                                                                                   |     |     |
| datree.output              | The format output of the policy check results: yaml, json, xml, simple, JUnit. (optional) | "" (i.e beautiful😊)                                                                                    |     |     |
| datree.noRecord            | Don’t send policy checks metadata to the backend. (optional)                              | false                                                                                                   |     |     |
