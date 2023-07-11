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

<table>
	<thead>
		<th>Parameter</th>
		<th>Description</th>
		<th>Default</th>
	</thead>
	<tbody>
		<tr>
			<td>namespace</td>
			<td>The name of the namespace all resources will be created in, if not specified in the release.</td>
			<td><pre lang="json">
""
</pre>
</td>
		</tr>
		<tr>
			<td>replicaCount</td>
			<td>The number of Datree webhook-server replicas to deploy for the webhook.</td>
			<td><pre lang="json">
2
</pre>
</td>
		</tr>
		<tr>
			<td>customLabels</td>
			<td>Additional labels to add to all resources.</td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>customAnnotations</td>
			<td>Additional annotations to add to all resources.</td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>rbac.serviceAccount</td>
			<td>Create service Account for the webhook</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "datree-webhook-server"
}
</pre>
</td>
		</tr>
		<tr>
			<td>rbac.clusterRole</td>
			<td>Create service Role for the webhook</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "datree-webhook-server-cluster-role"
}
</pre>
</td>
		</tr>
		<tr>
			<td>datree.token</td>
			<td>The token used to link Datree to your dashboard. (string, required)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.existingSecret</td>
			<td>The token may also be provided via secret, note if the existingSecret is provided the token field above is ignored.</td>
			<td><pre lang="json">
{
  "key": "",
  "name": ""
}
</pre>
</td>
		</tr>
		<tr>
			<td>datree.verbose</td>
			<td>Display 'How to Fix' link for failed rules in output. (boolean, optional)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.output</td>
			<td>The format output of the policy check results: yaml, json, xml, simple, JUnit. (string, optional)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.noRecord</td>
			<td>Don’t send policy checks metadata to the backend. (boolean, optional)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.enabledWarnings</td>
			<td>Choose which warnings to enable. (string array ,optional)</td>
			<td><pre lang="json">
[
  "failedPolicyCheck",
  "skippedBySkipList",
  "passedPolicyCheck",
  "RBACBypassed"
]
</pre>
</td>
		</tr>
		<tr>
			<td>datree.clusterName</td>
			<td>The name of the cluster link for cluster name in your dashboard (string ,optional)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.scanIntervalHours</td>
			<td>How often should the scan run in hours. (int, optional, default: 1 )</td>
			<td><pre lang="json">
1
</pre>
</td>
		</tr>
		<tr>
			<td>datree.configFromHelm</td>
			<td>If false, the webhook will be configured from the dashboard, otherwise it will be configured from here. Affected configurations: policy, enforce, customSkipList.</td>
			<td><pre lang="json">
false
</pre>
</td>
		</tr>
		<tr>
			<td>datree.policy</td>
			<td>The name of the policy to check, e.g: staging. (string, optional)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.enforce</td>
			<td>Block resources that fail the policy check. (boolean ,optional)</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>datree.customSkipList</td>
			<td>Excluded resources from policy checks. ("namespace;kind;name" ,optional)</td>
			<td><pre lang="json">
[
  "(.*);(.*);(^aws-node.*)"
]
</pre>
</td>
		</tr>
		<tr>
			<td>datree.labelKubeSystem</td>
			<td>set admission.datree/validate=skip label on kube-system resources. (openshift/okd users should set it to false)</td>
			<td><pre lang="json">
true
</pre>
</td>
		</tr>
		<tr>
			<td>image.repository</td>
			<td>Image repository for the webhook</td>
			<td><pre lang="json">
"datree/admission-webhook"
</pre>
</td>
		</tr>
		<tr>
			<td>image.tag</td>
			<td>The image release tag to use for the webhook</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>image.pullPolicy</td>
			<td>Image pull policy for the webhook</td>
			<td><pre lang="json">
"Always"
</pre>
</td>
		</tr>
		<tr>
			<td>imageCredentials</td>
			<td>For private registry which contains all the required images</td>
			<td><pre lang="json">
{
  "email": null,
  "enabled": false,
  "password": null,
  "registry": null,
  "username": null
}
</pre>
</td>
		</tr>
		<tr>
			<td>securityContext</td>
			<td>Security context applied on the containers</td>
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
		</tr>
		<tr>
			<td>resources</td>
			<td>The resource request/limits for the webhook container image</td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>nodeSelector</td>
			<td>Used to select on which node a pod is scheduled to run</td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>affinity</td>
			<td></td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>tolerations</td>
			<td></td>
			<td><pre lang="json">
[]
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.resources</td>
			<td>The resource request/limits for the scanner container image</td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.annotations</td>
			<td></td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.rbac.serviceAccount</td>
			<td>Create service Account for the scanner</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "cluster-scanner-service-account"
}
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.rbac.clusterRole</td>
			<td>Create service Role for the scanner</td>
			<td><pre lang="json">
{
  "create": true,
  "name": "cluster-scanner-role"
}
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.rbac.clusterRoleBinding</td>
			<td>Create service RoleBinding for the scanner</td>
			<td><pre lang="json">
{
  "name": "cluster-scanner-role-binding"
}
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.image.repository</td>
			<td>Image repository for the scanner</td>
			<td><pre lang="json">
"datree/cluster-scanner"
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.image.pullPolicy</td>
			<td>Image pull policy for the scanner</td>
			<td><pre lang="json">
"Always"
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.image.tag</td>
			<td>The image release tag to use for the scanner</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.image.resources</td>
			<td></td>
			<td><pre lang="json">
{}
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.livenessProbe.enabled</td>
			<td></td>
			<td><pre lang="json">
true
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.livenessProbe.scheme</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.livenessProbe.initialDelaySeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.livenessProbe.periodSeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.readinessProbe.enabled</td>
			<td></td>
			<td><pre lang="json">
true
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.readinessProbe.scheme</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.readinessProbe.initialDelaySeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>clusterScanner.readinessProbe.periodSeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>hooks.timeoutTime</td>
			<td>The timeout time the hook will wait for the webhook-server is ready.</td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>hooks.ttlSecondsAfterFinished</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>hooks.image.repository</td>
			<td></td>
			<td><pre lang="json">
"clastix/kubectl"
</pre>
</td>
		</tr>
		<tr>
			<td>hooks.image.tag</td>
			<td></td>
			<td><pre lang="json">
"v1.25"
</pre>
</td>
		</tr>
		<tr>
			<td>hooks.image.pullPolicy</td>
			<td></td>
			<td><pre lang="json">
"IfNotPresent"
</pre>
</td>
		</tr>
		<tr>
			<td>validatingWebhookConfiguration.failurePolicy</td>
			<td></td>
			<td><pre lang="json">
"Ignore"
</pre>
</td>
		</tr>
		<tr>
			<td>livenessProbe.enabled</td>
			<td></td>
			<td><pre lang="json">
true
</pre>
</td>
		</tr>
		<tr>
			<td>livenessProbe.scheme</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>livenessProbe.initialDelaySeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>livenessProbe.periodSeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>readinessProbe.enabled</td>
			<td></td>
			<td><pre lang="json">
true
</pre>
</td>
		</tr>
		<tr>
			<td>readinessProbe.scheme</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>readinessProbe.initialDelaySeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>readinessProbe.periodSeconds</td>
			<td></td>
			<td><pre lang="json">
null
</pre>
</td>
		</tr>
		<tr>
			<td>devMode.enabled</td>
			<td></td>
			<td><pre lang="json">
false
</pre>
</td>
		</tr>
	</tbody>
</table>

