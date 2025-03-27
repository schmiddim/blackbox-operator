#!/bin/bash

helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm repo update

# istio
helm upgrade --install istio-base istio/base -n istio-system --set defaultRevision=default --create-namespace
helm upgrade --install istiod istio/istiod -n istio-system --wait

# blackbox exporter
helm upgrade --install blackbox-exporter -n prometheus  prometheus-community/prometheus-blackbox-exporter  --create-namespace -f blackbox-values.yaml

# prometheus

helm upgrade --install prometheus-operator prometheus-community/kube-prometheus-stack -n prometheus --create-namespace --set prometheus.serviceAccount.create=true    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false  --set prometheus.rbac.create=true  #  --set prometheus.prometheusSpec.serviceMonitorNamespaceSelector=[]    --set prometheus.prometheusSpec.serviceMonitorSelector=[]

# Apply Example ServiceEntry
kubectl apply -f serviceEntry.yaml