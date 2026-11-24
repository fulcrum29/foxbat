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
	// Window is the analysis time range in Prometheus duration format, e.g. "7d".
	Window string
}

// DefaultWindow is used when no time range is given.
const DefaultWindow = "7d"

// AllowedWindows lists the time ranges that can be selected in the UI and API.
var AllowedWindows = []string{"1d", "7d", "14d", "30d"}

// ValidWindow reports whether w is one of the allowed time ranges.
func ValidWindow(w string) bool {
	for _, a := range AllowedWindows {
		if a == w {
			return true
		}
	}
	return false
}

func (p QueryParams) window() string {
	if p.Window == "" {
		return DefaultWindow
	}
	return p.Window
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

	query := fmt.Sprintf(`max by (container,namespace) (quantile_over_time(0.95, rate(container_cpu_usage_seconds_total{container!="POD", container!=""%s}[5m]) [%s:5m]))`, filterStr, params.window())

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

	query := fmt.Sprintf(`max_over_time(max by (container, namespace) (container_memory_working_set_bytes{container!="POD",container!=""%s})[%s:5m]) / 1048576`, filterStr, params.window())

	log.Printf("Querying Memory: %s", query)
	result, _, err := c.api.Query(ctx, query, time.Now(), promapi.WithTimeout(10*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error querying Memory recommendations: %w", err)
	}
	return result, nil
}
