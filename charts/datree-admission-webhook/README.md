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

<table>
	<thead>
		<th>Key</th>
		<th>Type</th>
		<th>Default</th>
		<th>Description</th>
	</thead>
	<tbody>
		<tr>
			<td>namespace</td>
			<td>string</td>
			<td><pre lang="json">
""
</pre>
</td>
			<td>The name of the namespace all resources will be created in, if not specified in the release.</td>
		</tr>
		<tr>
			<td>replicaCount</td>
			<td>int</td>
			<td><pre lang="json">
2
</pre>
</td>
			<td>The number of Datree webhook-server replicas to deploy for the webhook.</td>
		</tr>
		<tr>
			<td>customLabels</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td>Additional labels to add to all resources.</td>
		</tr>
		<tr>
			<td>customAnnotations</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td>Additional annotations to add to all resources.</td>
		</tr>
		<tr>
			<td>rbac.serviceAccount</td>
			<td>object</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "datree-webhook-server"
}
</pre>
</td>
			<td>Create service Account for the webhook</td>
		</tr>
		<tr>
			<td>rbac.clusterRole</td>
			<td>object</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "datree-webhook-server-cluster-role"
}
</pre>
</td>
			<td>Create service Account for the webhook</td>
		</tr>
		<tr>
			<td>datree.token</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The token used to link the CLI to your dashboard. (string, required)</td>
		</tr>
		<tr>
			<td>datree.existingSecret</td>
			<td>object</td>
			<td><pre lang="json">
{
  "key": "",
  "name": ""
}
</pre>
</td>
			<td>The token may also be provided via secret, note if the existingSecret is provided the token field above is ignored.</td>
		</tr>
		<tr>
			<td>datree.verbose</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>Display 'How to Fix' link for failed rules in output. (boolean, optional)</td>
		</tr>
		<tr>
			<td>datree.output</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The format output of the policy check results: yaml, json, xml, simple, JUnit. (string, optional)</td>
		</tr>
		<tr>
			<td>datree.noRecord</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>Don’t send policy checks metadata to the backend. (boolean, optional)</td>
		</tr>
		<tr>
			<td>datree.clusterName</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The name of the cluster link for cluster name in your dashboard (string ,optional)</td>
		</tr>
		<tr>
			<td>datree.scanIntervalHours</td>
			<td>int</td>
			<td><pre lang="json">
1
</pre>
</td>
			<td>How often should the scan run in hours. (int, optional, default: 1 )</td>
		</tr>
		<tr>
			<td>datree.configFromHelm</td>
			<td>bool</td>
			<td><pre lang="json">
false
</pre>
</td>
			<td>If false, the webhook will be configured from the dashboard, otherwise it will be configured from here. Affected configurations: policy, enforce, customSkipList.</td>
		</tr>
		<tr>
			<td>datree.policy</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The name of the policy to check, e.g: staging. (string, optional)</td>
		</tr>
		<tr>
			<td>datree.enforce</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>Block resources that fail the policy check. (boolean ,optional)</td>
		</tr>
		<tr>
			<td>datree.customSkipList</td>
			<td>list</td>
			<td><pre lang="json">
[
  "(.*);(.*);(^aws-node.*)"
]
</pre>
</td>
			<td>Excluded resources from policy checks. ("namespace;kind;name" ,optional)</td>
		</tr>
		<tr>
			<td>image.repository</td>
			<td>string</td>
			<td><pre lang="json">
"datree/admission-webhook"
</pre>
</td>
			<td>Image repository for the webhook</td>
		</tr>
		<tr>
			<td>image.tag</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The image release tag to use for the webhook</td>
		</tr>
		<tr>
			<td>image.pullPolicy</td>
			<td>string</td>
			<td><pre lang="json">
"Always"
</pre>
</td>
			<td>Image pull policy for the webhook</td>
		</tr>
		<tr>
			<td>securityContext</td>
			<td>object</td>
			<td><pre lang="json">
{
  "allowPrivilegeEscalation": false,
  "capabilities": {
    "drop": [
      "ALL"
    ]
  },
  "readOnlyRootFilesystem": true,
  "runAsNonRoot": true,
  "runAsUser": 25000,
  "seccompProfile": {
    "type": "RuntimeDefault"
  }
}
</pre>
</td>
			<td>Security context applied on the containers</td>
		</tr>
		<tr>
			<td>resources</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td>The resource request/limits for the webhook container image</td>
		</tr>
		<tr>
			<td>nodeSelector</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td>Used to select on which node a pod is scheduled to run</td>
		</tr>
		<tr>
			<td>affinity</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td>Offers more expressive syntax for fine-grained control of how Pods are scheduled to specific nodes</td>
		</tr>
		<tr>
			<td>tolerations</td>
			<td>list</td>
			<td><pre lang="json">
[]
</pre>
</td>
			<td></td>
		</tr>
		<tr>
			<td>clusterScanner.resources</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td>The resource request/limits for the scanner container image</td>
		</tr>
		<tr>
			<td>clusterScanner.annotations</td>
			<td>object</td>
			<td><pre lang="json">
{}
</pre>
</td>
			<td></td>
		</tr>
		<tr>
			<td>clusterScanner.rbac.serviceAccount</td>
			<td>object</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "cluster-scanner-service-account"
}
</pre>
</td>
			<td>Create service Account for the scanner</td>
		</tr>
		<tr>
			<td>clusterScanner.rbac.clusterRole</td>
			<td>object</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "cluster-scanner-role"
}
</pre>
</td>
			<td>Create service Role for the scanner</td>
		</tr>
		<tr>
			<td>clusterScanner.rbac.clusterRoleBinding</td>
			<td>object</td>
			<td><pre lang="json">
{
  "name": "cluster-scanner-role-binding"
}
</pre>
</td>
			<td>Create service RoleBinding for the scanner</td>
		</tr>
		<tr>
			<td>clusterScanner.image.repository</td>
			<td>string</td>
			<td><pre lang="json">
"datree/cluster-scanner"
</pre>
</td>
			<td>Image repository for the scanner</td>
		</tr>
		<tr>
			<td>clusterScanner.image.pullPolicy</td>
			<td>string</td>
			<td><pre lang="json">
"Always"
</pre>
</td>
			<td>Image pull policy for the scanner</td>
		</tr>
		<tr>
			<td>clusterScanner.image.tag</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The image release tag to use for the scanner</td>
		</tr>
		<tr>
			<td>hooks.timeoutTime</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td>The timeout time the hook will wait for the webhook-server is ready.</td>
		</tr>
		<tr>
			<td>hooks.ttlSecondsAfterFinished</td>
			<td>string</td>
			<td><pre lang="json">
null
</pre>
</td>
			<td></td>
		</tr>
		<tr>
			<td>hooks.image.repository</td>
			<td>string</td>
			<td><pre lang="json">
"clastix/kubectl"
</pre>
</td>
			<td></td>
		</tr>
		<tr>
			<td>hooks.image.tag</td>
			<td>string</td>
			<td><pre lang="json">
"v1.25"
</pre>
</td>
			<td></td>
		</tr>
		<tr>
			<td>hooks.image.pullPolicy</td>
			<td>string</td>
			<td><pre lang="json">
"IfNotPresent"
</pre>
</td>
			<td></td>
		</tr>
		<tr>
			<td>validatingWebhookConfiguration.failurePolicy</td>
			<td>string</td>
			<td><pre lang="json">
"Ignore"
</pre>
</td>
			<td></td>
		</tr>
	</tbody>
</table>

