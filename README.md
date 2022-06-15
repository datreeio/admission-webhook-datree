# Datree admission webhook

![image](/internal/images/diagram.png)

## Overview

Datree offers cluster integration that allows you to validate your resources against your desired policy upon pushing them into a cluster, by using an admission webhook.  
The webhook will catch kubectl **create**, **apply** and **edit** operations and initiate a policy check against the resource/s associated with each operation.  
If any violations are found, the webhook will reject the operation, and display a detailed output with instructions on how to resolve each violation.

## Installation
To install and start using the webhook, choose one of the following options:

### Option 1 - via an installation script

**Prerequisites**
The following applications need to be installed on the machine:
- kubectl
- openssl - *required for creating a certificate authority (CA).*

**Installation**
1. Download the latest release from this repository.
2. Unzip and run `./install.sh`

### Option 2 - manual installation
aaaaaa

## Usage
Once the webhook is installed, every hooked operation will trigger a Datree policy check. If any violations are found, the following output will be displayed:

![image](/internal/images/deny-example.png)

## Behavior
The webhook’s behavior is configured within the `datree-webhook` resource. You can alter this behavior by providing settings as environment variables. The following settings are supported:
DATREE_POLICY
DATREE_VERBOSE
DATREE_NO_RECORD
DATREE_OUTPUT


For running the webhook locally (in development), view the DEVELOPER_GUIDE.md file.
