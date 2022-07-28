# Datree Admission Webhook

<p align="center">
<img src="/internal/images/diagram.png" width="80%" />
</p>
  
# Overview
Datree offers cluster integration that allows you to validate your resources against your configured policy upon pushing them into a cluster, by using an admission webhook.  

The webhook will catch **create**, **apply** and **edit** operations and initiate a policy check against the configs associated with each operation. If any misconfigurations are found, the webhook will reject the operation, and display a detailed output with instructions on how to resolve each misconfiguration.

## Webhook validation triggers

K8s use different abstractions to simplify and automate complex processes. For example, when explicitly applying an object type “Deployment”, under the hood, K8s will “translate” this object into implicit objects of type “Pod.”

When installed on your cluster, other policy enforcement tools will validate both explicit and implicit objects. This approach may create a lot of noise and false positive failures since it will cause the webhook to validate objects that the users don’t manage and, in some cases, are not even accessible.

To avoid such issues, we decided to define the specific operations that the admission webhook should validate:

* Kubectl - validate objects that were created or updated using kubectl `create`, `edit`, and `apply` commands. Objects that were implicitly created (e.g., pods created via deployment) are ignored since the webhook validates the deployment that generated them and is accessible to the user. 
* Gitops CD tools -  validate objects that were explicitly created and distinguish them from other objects (custom resources) that were implicitly created during the installation and are required for the ongoing operation of these tools (e.g., ArgoCD, FluxCD, etc.) 

# Installation

The webhook officially supports Kubernetes version *1.19* and higher, and has been tested with EKS.

**Prerequisites**  
The following applications need to be installed on the machine:
- kubectl
- openssl - *required for creating a certificate authority (CA).*
- curl

**Installation**  
Simply copy the following command and run it in your terminal:  
```
bash <(curl https://get.datree.io/admission-webhook)
```

**[NOTE]** the link above will prompt you to enter your Datree token during installation.  
To install without a prompt, you can provide your token as part of the installation command, by running this in your terminal:  
```
DATREE_TOKEN=[your-token] bash <(curl https://get.datree.io/admission-webhook)
```

## Usage
Once the webhook is installed, every hooked operation will trigger a Datree policy check.  
If any misconfigurations are found, the following output will be displayed:

![image](/internal/images/deny-example.png)

If no misconfigurations are found, the resource will be applied/updated normally.

## Behavior
The webhook’s behavior is configured within the `webhook-server` resource.  
The following settings are supported:  
| Setting              | Values                 | Description                  |
| -------------------- | ---------------------- |  --------------------------- |
| DATREE_TOKEN         |                        | Your Datree token, see our [docs](https://hub.datree.io/setup/account-token#1-get-your-account-token-from-the-dashboard) for instructions on how to obtain it |
| DATREE_POLICY        | e.g. "Argo", "NSA"     | The name of the desired Datree policy to run |
| DATREE_VERBOSE       | true, false            | Display 'How to Fix' link for failed rules in output|
| DATREE_NO_RECORD     | true, false            | Don’t send policy checks metadata to the backend |
| DATREE_OUTPUT        | json, yaml, xml, JUnit | Output the policy check results in the requested format |

**To change the behavior:**  
1. Create a YAML file in your repository with this content:  
```yaml
spec:
  template:
    spec:
      containers:
        - name: server
          env:
            - name: DATREE_POLICY
              value: ""
            - name: DATREE_VERBOSE
              value: ""
            - name: DATREE_OUTPUT
              value: ""
            - name: DATREE_NO_RECORD
              value: ""
```
2. Change the values of your settings as you desire.
3. Run the following command to apply your changes to the webhook resource:  
```
kubectl patch deployment webhook-server -n datree --patch-file /path/to/patch/file.yaml
```

🤫 Since your token is sensitive and you would not want to keep it in your repository, we recommend to set/change it by running a separate `kubectl patch` command:  
```yaml
kubectl patch deployment webhook-server -n datree -p '
spec:
  template:
    spec:
      containers:
        - name: server
          env:
            - name: DATREE_TOKEN
              value: "<your-token>"'
```
Simply replace `<your-token>` with your actual token, then copy the entire command and run it in your terminal. 

## Uninstallation
To uninstall the webhook, copy the following command and run it in your terminal:  
```
bash <(curl https://get.datree.io/admission-webhook-uninstall)
```

## Local development
To run the webhook locally (in development), view our [developer guide](/guides/developer-guide.md).
