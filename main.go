package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
)

const DemoPrometheusURL = "http://localhost:9090"

const RecCPURequests = "quantile_over_time (0.99,max(rate(container_cpu_usage_seconds_total{container!=\"POD\",container!=\"\"}[5m])) by (container,namespace) [1h:1m])"

func main() {
	client, err := api.NewClient(api.Config{
		Address: DemoPrometheusURL,
	})
	if err != nil {
		log.Fatal("error creating client", err)
	}

	queryClient := promapi.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := queryClient.Query(ctx, RecCPURequests, time.Now(), promapi.WithTimeout(5*time.Second))
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		log.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)
}
