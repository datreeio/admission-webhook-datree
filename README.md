# Datree admission webhook

![image](/internal/images/diagram.png)

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
Simply copy this command and run it in your terminal:
`curl... install.sh`

### Option 2 - manual installation
put in guides folder

## Usage
Once the webhook is installed, every hooked operation will trigger a Datree policy check. If any violations are found, the following output will be displayed:

![image](/internal/images/deny-example.png)

If no violations are found, the resource will be applied/updated normally.

## Behavior
The webhook’s behavior is configured within the `datree-webhook` resource. You can alter this behavior by providing settings as environment variables. The following settings are supported:  
| Setting              | Values                 | Description                  |
| -------------------- | ---------------------- |  --------------------------- |
| DATREE_TOKEN         |                        | Your Datree token, see our [docs](https://hub.datree.io/setup/account-token#1-get-your-account-token-from-the-dashboard) for instructions on how to obtain it |
| DATREE_POLICY        | e.g. "dev", "prod"     | The name of the desired Datree policy to run |
| DATREE_VERBOSE       | true, false            | Display 'How to Fix' link for failed rules in output|
| DATREE_NO_RECORD     | true, false            | Don’t send policy checks metadata to the backend |
| DATREE_OUTPUT        | json, yaml, xml, JUnit | Output the policy check results in the requested format |

To change the behavior, use the `kubectl patch` command to change the `datree-webhook` resource.

## Uninstall
To uninstall the webhook, copy this command and run it in your terminal:
`curl... uninstall.sh`

## Local development
To run the webhook locally (in development), view the DEVELOPER_GUIDE.md file.
