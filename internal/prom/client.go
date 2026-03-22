package prom

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type QueryParams struct {
	Namespace string
	Pod       string
	Workload  string
}

func (p QueryParams) ToFilters() (string, string) {
	var filters []string
	var ownerFilters []string

	if p.Namespace != "" {
		filters = append(filters, fmt.Sprintf("namespace=%q", p.Namespace))
		ownerFilters = append(ownerFilters, fmt.Sprintf("namespace=%q", p.Namespace))
	}
	if p.Pod != "" {
		filters = append(filters, fmt.Sprintf("pod=%q", p.Pod))
	}
	if p.Workload != "" {
		ownerFilters = append(ownerFilters, fmt.Sprintf("owner_name=%q", p.Workload))
	}

	filterStr := ""
	if len(filters) > 0 {
		filterStr = "," + strings.Join(filters, ",")
	}

	ownerFilterStr := ""
	if len(ownerFilters) > 0 {
		ownerFilterStr = "," + strings.Join(ownerFilters, ",")
	}

	return filterStr, ownerFilterStr
}

type Client struct {
	api promapi.API
}

// NewClient creates a new Prometheus client wrapper.
func NewClient(address string) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}
	return &Client{api: promapi.NewAPI(client)}, nil
}

// GetCPURecommendations fetches CPU resource recommendations from Prometheus.
func (c *Client) GetCPURecommendations(ctx context.Context, params QueryParams) (model.Value, error) {
	f, of := params.ToFilters()
	query := fmt.Sprintf(`quantile by (namespace,owner_name,container)(0.98,(rate(container_cpu_usage_seconds_total{container!="POD",container!=""%s}[5m])) * on(namespace,pod) group_left(owner_name) avg by (namespace,pod,owner_name)(kube_pod_owner{owner_kind=~"DaemonSet|StatefulSet|Deployment"%s}))`, f, of)
	result, _, err := c.api.Query(ctx, query, time.Now(), promapi.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error querying CPU recommendations: %w", err)
	}
	return result, nil
}

// GetMemoryRecommendations fetches Memory resource recommendations from Prometheus.
func (c *Client) GetMemoryRecommendations(ctx context.Context, params QueryParams) (model.Value, error) {
	f, of := params.ToFilters()
	query := fmt.Sprintf(`quantile by (namespace,owner_name,container)(0.98,(container_memory_usage_bytes{container!="POD",container!=""%s}) * on(namespace,pod) group_left(owner_name) avg by (namespace,pod,owner_name)(kube_pod_owner{owner_kind=~"DaemonSet|StatefulSet|Deployment"%s}))`, f, of)
	result, _, err := c.api.Query(ctx, query, time.Now(), promapi.WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error querying Memory recommendations: %w", err)
	}
	return result, nil
}
