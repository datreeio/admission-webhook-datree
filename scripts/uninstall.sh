#!/bin/bash

NAMESPACE=datree

kubectl delete validatingwebhookconfiguration.admissionregistration.k8s.io/webhook-datree
kubectl delete service/webhook-server -n $NAMESPACE
kubectl delete deployment/webhook-server -n $NAMESPACE
kubectl delete deployment/datree-cluster-scanner-server -n $NAMESPACE
kubectl delete secret/webhook-server-tls -n $NAMESPACE
kubectl delete clusterrolebinding/rolebinding:webhook-server-datree
kubectl delete clusterrolebinding/rolebinding:cluster-scanner-role-binding
kubectl delete serviceaccount/webhook-server-datree -n datree
kubectl delete serviceaccount/cluster-scanner-service-account -n datree
kubectl delete clusterrole/webhook-server-datree
kubectl delete clusterrole/cluster-scanner-role

kubectl label namespace kube-system admission.datree/validate-
kubectl delete namespace/datree
