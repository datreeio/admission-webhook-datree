kubectl delete validatingwebhookconfiguration.admissionregistration.k8s.io/webhook-datree
kubectl delete service/webhook-server -n datree
kubectl delete deployment/webhook-server -n datree
kubectl delete secret/webhook-server-tls -n datree
kubectl delete clusterrolebinding/rolebinding:webhook-server-datree
kubectl delete serviceaccount/webhook-server-datree -n datree
kubectl delete clusterrole/webhook-server-datree
kubectl label namespace kube-system admission.datree/validate-
# kubectl delete namespace/datree
