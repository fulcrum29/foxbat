## Recommender

Prometheus query for finding the CPU request for 98 quantile so we have some spare space
`quantile by (namespace,owner_name,container)(0.98,(rate(container_cpu_usage_seconds_total{container!="POD",container!=""}[5m])) * on(namespace,pod) group_left(owner_name) avg by (namespace,pod,owner_name)(kube_pod_owner{owner_kind=~"DaemonSet|StatefulSet|Deployment"}))`

Prometheus query for finding the memory request for 98 quantile so we have some spare space
`quantile by (namespace,owner_name,container)(0.98,(container_memory_usage_bytes{container!="POD",container!=""}) * on(namespace,pod) group_left(owner_name) avg by (namespace,pod,owner_name)(kube_pod_owner{owner_kind=~"DaemonSet|StatefulSet|Deployment"}))`
