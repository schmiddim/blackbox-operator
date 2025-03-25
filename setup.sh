#!/bin/bash

helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm repo update

# istio
helm upgrade --install istio-base istio/base -n istio-system --set defaultRevision=default --create-namespace
helm upgrade --install istiod istio/istiod -n istio-system --wait

# blackbox exporter
helm upgrade --install blackbox-exporter -n blackbox-exporter  prometheus-community/prometheus-blackbox-exporter  --create-namespace

# prometheus
helm upgrade --install prometheus -n prometheus prometheus-community/prometheus --create-namespace
# operator for CRDs
helm upgrade --install prometheus-operator prometheus-community/kube-prometheus-stack -n prometheus --create-namespace
