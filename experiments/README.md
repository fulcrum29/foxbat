# Recommendation experiment

This directory contains the test workloads and PromQL queries used to evaluate
foxbat recommendations (thesis section 3.7 and appendix 1).

## Expected values

| Workload | Current request (CPU / memory) | Expected recommendation (CPU / memory) | Expected request with 15% margin |
|---|---|---|---|
| load-constant | 500m / 256Mi | ~200m / ~100 MiB | ~230m / ~115 MiB |
| load-daily | 1000m / 256Mi | ~400m / ~64 MiB | ~460m / ~74 MiB |

`load-daily` uses about 400m for 12 hours a day and close to 0 at night, so its
mean CPU usage is about 200m while the 95th percentile stays at about 400m.

## Procedure

1. Create the cluster and install Argo CD, Prometheus and foxbat (see the main README).
2. `kubectl apply -f experiments/workloads.yaml`
3. Wait at least 7 days (minimal variant: 1 day and the `1d` window).
4. Time t0: read the recommendations for namespace `foxbat-test` and window `7d`
   in the UI (take a screenshot) or with
   `curl "http://localhost:9080/api/recommendations?namespace=foxbat-test&window=7d"`.
5. Set the requests in `workloads.yaml` to recommendation x 1.15, set the memory
   limit equal to the memory request, remove the CPU limit, apply the file again.
6. Wait another 7 days and collect the indicators below for the period [t0, t0 + 7d].

## Queries

Replace `REC_CPU` with the CPU recommendation in millicores for the container.

```promql
# Current requests (CPU in cores, memory in bytes)
max by (pod, container, resource) (kube_pod_container_resource_requests{namespace="foxbat-test"})

# Mean and maximum CPU usage in millicores over the window
avg_over_time((max by (pod, container) (rate(container_cpu_usage_seconds_total{namespace="foxbat-test", container!="", container!="POD"}[5m])))[7d:5m]) * 1000
max_over_time((max by (pod, container) (rate(container_cpu_usage_seconds_total{namespace="foxbat-test", container!="", container!="POD"}[5m])))[7d:5m]) * 1000

# Share of 5-minute samples in which CPU usage exceeded the recommendation
avg_over_time(((max by (pod, container) (rate(container_cpu_usage_seconds_total{namespace="foxbat-test", container!="", container!="POD"}[5m])) * 1000) > bool REC_CPU)[7d:5m])

# Share of CFS periods in which the container was throttled by its CPU limit
sum by (pod, container) (increase(container_cpu_cfs_throttled_periods_total{namespace="foxbat-test", container!=""}[7d]))
  / sum by (pod, container) (increase(container_cpu_cfs_periods_total{namespace="foxbat-test", container!=""}[7d]))

# Restarts and containers whose last termination was an out-of-memory kill
increase(kube_pod_container_status_restarts_total{namespace="foxbat-test"}[7d])
kube_pod_container_status_last_terminated_reason{namespace="foxbat-test", reason="OOMKilled"}
```
