#!/bin/bash

NAMESPACE=datree

kubectl delete validatingwebhookconfiguration.admissionregistration.k8s.io/datree-webhook
kubectl delete service/datree-webhook-server -n $NAMESPACE
kubectl delete deployment/datree-webhook-server -n $NAMESPACE
kubectl delete deployment/datree-cluster-scanner-server -n $NAMESPACE
kubectl delete secret/webhook-server-tls -n $NAMESPACE
kubectl delete clusterrolebinding/datree-webhook-server-cluster-role
kubectl delete clusterrolebinding/cluster-scanner-role-binding
kubectl delete serviceaccount/datree-webhook-server -n datree
kubectl delete serviceaccount/cluster-scanner-service-account -n datree
kubectl delete clusterrole/datree-webhook-server-cluster-role
kubectl delete clusterrole/cluster-scanner-role

kubectl label namespace kube-system admission.datree/validate-
kubectl delete namespace/datree
