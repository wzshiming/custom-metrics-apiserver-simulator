# custom-metrics-apiserver-simulator

## Create Cluster

``` bash
kwokctl create cluster --runtime docker
```

``` bash
kwokctl scale node --replicas 2
```

## Set up apiservice and service

``` bash
kubectl apply -k ./kustomize/apiservice
```

``` bash
kubectl create service externalname custom-metrics-apiserver -n kube-system --external-name host.docker.internal
```

## Set up CRD and Start Custom Metrics Simulator

``` bash
kubectl apply -k ./kustomize/crd
```

``` bash
go run ./cmd/custom-metrics-apiserver-simulator
```

## Test

### Custom Metrics With HPA

``` bash
kubectl apply -f ./test/custom_metrics.yaml
```

``` console
$ kubectl get --raw '/apis/custom.metrics.k8s.io/v1beta2/namespaces/default/pods/*/http_requests_total' | jq .
{
  "kind": "MetricValueList",
  "apiVersion": "custom.metrics.k8s.io/v1beta2",
  "metadata": {},
  "items": [
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "default",
        "name": "custom-metric-54d77dbf95-fgq8d",
        "apiVersion": "/v1"
      },
      "metric": {
        "name": "http_requests_total",
        "selector": null
      },
      "timestamp": "2023-09-14T05:30:39Z",
      "value": "900m"
    },
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "default",
        "name": "custom-metric-54d77dbf95-kq45w",
        "apiVersion": "/v1"
      },
      "metric": {
        "name": "http_requests_total",
        "selector": null
      },
      "timestamp": "2023-09-14T05:30:39Z",
      "value": "900m"
    }
  ]
}
```

### External Metrics With HPA

``` bash
kubectl apply -f ./test/external_metrics.yaml
```

``` console
$ kubectl get --raw '/apis/external.metrics.k8s.io/v1beta1/namespaces/default/my_external_metric' | jq .
{
  "kind": "ExternalMetricValueList",
  "apiVersion": "external.metrics.k8s.io/v1beta1",
  "metadata": {},
  "items": [
    {
      "metricName": "my_external_metric",
      "metricLabels": null,
      "timestamp": "2023-09-14T05:31:51Z",
      "value": "100m"
    }
  ]
}
```
