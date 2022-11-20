# Datree Admission Webhook

<p align="center">
<img src="/internal/images/diagram.png" width="80%" />
</p>
  
# Overview
[Datree](https://datree.io) prevents misconfigurations by blocking resources that do not meet your policy, using an admission webhook.

The webhook will catch **create**, **apply** and **edit** operations and initiate a policy check against the resources associated with each operation. If any misconfigurations are found, the webhook will reject the operation, and display a detailed output with instructions on how to resolve each misconfiguration.

For the full documentation, click [here](https://hub.datree.io)👈🏼

The webhook can be configured via the helm values file under the `datree` key. See the [Datree webhook Helm chart](https://github.com/datreeio/admission-webhook-datree/tree/gh-pages) for more details.

