# Datree admission webhook

<p align="center">
<img src="/internal/images/diagram.png" width="700px" />
</p>
  
## Overview
Datree offers cluster integration that allows you to validate your resources against your configured policy upon pushing them into a cluster, by using an admission webhook.  
The webhook will catch kubectl **create**, **apply** and **edit** operations and initiate a policy check against the resource/s associated with each operation.  
If any violations are found, the webhook will reject the operation, and display a detailed output with instructions on how to resolve each violation.

## Specifications
The webhook officially supports Kubernetes version *1.19* and higher, and has been tested with EKS.

## Installation
To install and start using the webhook, choose one of the following options:

### Option 1 - via an installation script

**Prerequisites**  
The following applications need to be installed on the machine:
- kubectl
- openssl - *required for creating a certificate authority (CA).*
- curl

**Installation**  
Simply copy the following command and run it in your terminal:  
`bash <(curl https://get.datree.io/webhook)`

**NOTE:** the link above will prompt you to enter your Datree token during installation.  
To install without a prompt, you can provide your token as part of the installation command, by running this in your terminal:  
`DATREE_TOKEN=<your-token> bash <(curl https://get.datree.io/webhook)`

### Option 2 - manual installation
See the [manual installation guide](/guides/manual-installation.md)

## Usage
Once the webhook is installed, every hooked operation will trigger a Datree policy check. If any violations are found, the following output will be displayed:

![image](/internal/images/deny-example.png)

If no violations are found, the resource will be applied/updated normally.

## Behavior
The webhook’s behavior is configured within the `datree-webhook` resource.  
The following settings are supported:  
| Setting              | Values                 | Description                  |
| -------------------- | ---------------------- |  --------------------------- |
| DATREE_TOKEN         |                        | Your Datree token, see our [docs](https://hub.datree.io/setup/account-token#1-get-your-account-token-from-the-dashboard) for instructions on how to obtain it |
| DATREE_POLICY        | e.g. "dev", "prod"     | The name of the desired Datree policy to run |
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
`kubectl patch deployment webhook-server -n datree --patch-file /path/to/patch/file.yaml`

⚠️ Since your token is sensitive and you would not want to keep it in your repository, we recommend to set/change it by running a separate `patch` command:  
```
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

## Uninstall
To uninstall the webhook, copy this command and run it in your terminal:  
`bash <(curl https://get.datree.io/webhook-uninstall)`

## Local development
To run the webhook locally (in development), view the DEVELOPER_GUIDE.md file.
