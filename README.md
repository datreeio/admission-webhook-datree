# Admission Webhook Datree Helm Chart

```sh
helm repo add datree-webhook https://datreeio.github.io/admission-webhook-datree/
helm repo update
helm install datree-webhook --namespace datree datree-webhook/datree-admission-webhook --create-namespace
```

**Full documentation at:https://github.com/datreeio/admission-webhook-datree**