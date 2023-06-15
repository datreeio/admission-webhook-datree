#!/bin/bash

NAMESPACE=datree

# webhook
kubectl delete validatingwebhookconfiguration.admissionregistration.k8s.io/datree-webhook
kubectl delete service/datree-webhook-server -n $NAMESPACE
kubectl delete deployment/datree-webhook-server -n $NAMESPACE
kubectl delete clusterrolebinding/datree-webhook-server-cluster-role
kubectl delete serviceaccount/datree-webhook-server -n $NAMESPACE
kubectl delete clusterrole/datree-webhook-server-cluster-role
kubectl delete secret/webhook-server-tls -n $NAMESPACE

# scanner
kubectl delete deployment/datree-cluster-scanner-server -n $NAMESPACE
kubectl delete clusterrolebinding/cluster-scanner-role-binding
kubectl delete serviceaccount/cluster-scanner-service-account -n $NAMESPACE
kubectl delete clusterrole/cluster-scanner-role

# misc
kubectl delete job/datree-wait-server-ready-hook-post-install -n $NAMESPACE
kubectl label namespace kube-system admission.datree/validate-

# namespace
kubectl delete clusterrole/datree-validationwebhook-delete
kubectl delete clusterrolebinding/datree-validationwebhook-delete
kubectl delete clusterrole/datree-namespaces-update
kubectl delete clusterrolebinding/datree-namespaces-update

kubectl delete namespace/datree
