#!/bin/bash

NAMESPACE=datree

kubectl delete validatingwebhookconfiguration.admissionregistration.k8s.io/webhook-datree
kubectl delete service/webhook-server -n $NAMESPACE
kubectl delete deployment/webhook-server -n $NAMESPACE
kubectl delete deployment/datree-cluster-scanner-server -n $NAMESPACE
kubectl delete secret/webhook-server-tls -n $NAMESPACE
kubectl delete clusterrolebinding --all -n $NAMESPACE
kubectl delete serviceaccounts --all -n $NAMESPACE
kubectl delete clusterrole --all -n $NAMESPACE
kubectl label namespace kube-system admission.datree/validate-
kubectl delete namespace/datree
