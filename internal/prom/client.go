package prom

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

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

type QueryParams struct {
	Namespace string
	Workload  string
}

// GetCPURecommendations fetches CPU resource recommendations from Prometheus.
func (c *Client) GetCPURecommendations(ctx context.Context, params QueryParams) (model.Value, error) {
	var filters []string
	if params.Namespace != "" {
		filters = append(filters, fmt.Sprintf("namespace=%q", params.Namespace))
	}
	if params.Workload != "" {
		filters = append(filters, fmt.Sprintf("pod=~%q", params.Workload+"-.*"))
	}

	filterStr := ""
	if len(filters) > 0 {
		filterStr = "," + strings.Join(filters, ",")
	}

	query := fmt.Sprintf(`max by (container,namespace) (quantile_over_time(0.95, rate(container_cpu_usage_seconds_total{container!="POD", container!=""%s}[5m]) [7d:5m]))`, filterStr)

	log.Printf("Querying CPU: %s", query)
	result, _, err := c.api.Query(ctx, query, time.Now(), promapi.WithTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error querying CPU recommendations: %w", err)
	}
	return result, nil
}

// GetMemoryRecommendations fetches Memory resource recommendations from Prometheus.
func (c *Client) GetMemoryRecommendations(ctx context.Context, params QueryParams) (model.Value, error) {
	var filters []string
	if params.Namespace != "" {
		filters = append(filters, fmt.Sprintf("namespace=%q", params.Namespace))
	}
	if params.Workload != "" {
		filters = append(filters, fmt.Sprintf("pod=~%q", params.Workload+"-.*"))
	}

	filterStr := ""
	if len(filters) > 0 {
		filterStr = "," + strings.Join(filters, ",")
	}

	query := fmt.Sprintf(`max_over_time(max by (container, namespace) (container_memory_working_set_bytes{container!="POD",container!=""%s})[7d:5m]) / 1048576`, filterStr)

	log.Printf("Querying Memory: %s", query)
	result, _, err := c.api.Query(ctx, query, time.Now(), promapi.WithTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error querying Memory recommendations: %w", err)
	}
	return result, nil
}
