kubectl delete validatingwebhookconfiguration.admissionregistration.k8s.io/webhook-datree
kubectl delete service/webhook-server -n datree
kubectl delete deployment/webhook-server -n datree
kubectl delete secret/webhook-server-tls -n datree
kubectl label namespace kube-system admission.datree/validate-
kubectl delete namespace/datree
