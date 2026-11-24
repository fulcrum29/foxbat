## The goal of this thesis is to design and implement an application that recommends optimal computing resource values (CPU and RAM) for container environments running on the Kubernetes platform. A key element of this approach is the use of historical resource usage metrics collected and stored in the Prometheus monitoring system. The application analyzes resource usage patterns and, based on this, provides precise, contextual recommendations that will help balance performance and costs in a cloud environment. The application will present recommendations in the form of a periodic report available in a browser or sent to one of the popular messaging applications used in enterprises, such as MS Teams or Slack. Based on these recommendations, the person responsible for managing a given application will be able to better balance the performance and costs of running applications.

## Technologies used
- go
- HTMX
- alpine.js

## Project Structure
- `cmd/foxbat/main.go`: Application entry point, handles configuration and initialization.
- `internal/prom/`: Contains the Prometheus client and metrics fetching logic.
- `internal/recommender/`: Implements the core logic for resource recommendation (CPU/RAM).
- `internal/reporter/`: Formats and sends reports (periodic reports, Slack/Teams notifications).
- `internal/handlers/`: HTTP server handlers for the web interface.
- `web/templates/`: HTML templates (HTMX/Alpine.js).
- `web/static/`: Static assets like CSS and JS files.
- `kindconfig.yaml`: Configuration for local Kubernetes development with KinD.

## Implemented Features
- **Prometheus Client (`internal/prom/client.go`)**: A Go module for executing queries against a Prometheus server. It implements methods to retrieve the 98th percentile of CPU and Memory usage for given Kubernetes workloads using PromQL queries.
- **Kubernetes Client (`internal/k8s/client.go`)**: A module that uses `client-go` to interact with the Kubernetes API (via `~/.kube/config`). It provides methods to list namespaces and workloads (Deployments, StatefulSets, DaemonSets).
- **HTTP Handlers (`internal/handlers/recommender.go`)**: An HTTP handler that processes requests for resource recommendations, fetches the CPU and memory metrics via the Prometheus client, and responds with the raw recommendations in JSON format. It also serves the HTML UI and provides endpoints for dynamic data loading.
- **HTTP Router (`internal/handlers/router.go`)**: A central point to register application routes, using the standard Go `net/http.ServeMux`.
- **HTTP Server & Configuration (`cmd/foxbat/main.go`)**: The application entry point initializes the Prometheus and Kubernetes clients, sets up the router, and starts the HTTP server. It supports configuration via a `config.yaml` file or environment variables (`PROMETHEUS_URL`, `SERVER_PORT`).
- **Structured Logging (`internal/handlers/middleware.go`)**: A custom middleware that logs every HTTP request using Go's standard `log/slog` library. It captures the method, path, status code, duration, and IP address in a structured format.
- **Optimized Prometheus Integration (`internal/prom/client.go`)**:
    - **Recommendations Logic**: Uses `pod=~'workload-.*'` regex matching to query metrics for all pods associated with a specific workload.
    - **CPU Recommendations**: Uses `quantile_over_time(0.95, ...)` on `container_cpu_usage_seconds_total` over a 7-day window. Results are converted to millicores (m) directly in the query.
    - **Memory Recommendations**: Uses `max_over_time(...)` on `container_memory_working_set_bytes` over a 7-day window. Results are converted to MiB directly in the query.
- **Web UI (`web/templates/`)**: A user-friendly interface built with **HTMX** and **Franken-ui**. It features dynamic cascading dropdowns for Namespaces and Workloads, allowing users to easily select targets for resource recommendations. Results tables now correctly display the pod associated with each recommendation.
- **Template Architecture (`internal/templates/data.go`)**: Moved UI data structures to a dedicated module to improve organization and type safety.
- **Code Maintenance**: Refactored handler constructors for better readability and added helper functions for safer template operations.

- **Required namespace**: Recommendation requests without a namespace return HTTP 400, so a single request never scans the whole cluster.
- **Time Range Support**: The analysis time range (1d, 7d, 14d, 30d; default 7d) can be selected in the UI and passed to the API as the `window` query parameter. Invalid values return HTTP 400.

## Deployment
- **Container image**: `Dockerfile` builds a static binary in a multi-stage build and runs it in a distroless image as a non-root user.
- **Helm chart** (`clusters/foxbat/services/foxbat`): Deployment, Service (NodePort 30080), ServiceAccount and a read-only ClusterRole for listing namespaces and workloads.
- **Argo CD** (`clusters/foxbat/argocd-apps/templates/application-foxbat.yaml`): the application is synchronised from this repository like Prometheus.
- **Prometheus retention**: kube-prometheus-stack keeps 30 days of metrics on a 10Gi volume, so 30-day recommendations are possible.

Local run on kind:
```
kind create cluster --config kindconfig.yaml
docker build -t foxbat:dev .
kind load docker-image foxbat:dev --name foxbat-cluster
```
After Argo CD syncs, the UI is available at http://localhost:9080.

## Experiment
`experiments/` contains test workloads with known resource usage and the PromQL queries
used to evaluate the quality of recommendations (backtesting). See `experiments/README.md`.

## TODO
- **Add kube-state-metrics**: Add kube-state-metrics and join kube_pod_info metric on to prom queries to show in which workload does it belongs

Updated query that returns `workload` label based on `ReplicaSet`, `DaemonSet` or `StatefulSet` name
```
max by (namespace, container, workload_name) (
  label_replace(
    (
      quantile_over_time(0.95,
        rate(container_cpu_usage_seconds_total{container!="POD", container!="", namespace="kube-system"}[5m])
      [7d:5m])
    )
    * on (pod, namespace) group_left (created_by_name)
    (
      kube_pod_info{namespace="kube-system"}
    ),
    "workload_name", "$1", "created_by_name", "^(.*?)(-[0-9a-z]{8,10})?$"
  )
)
```
