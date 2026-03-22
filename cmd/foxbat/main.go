package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/fulcrum29/foxbat/internal/handlers"
	"github.com/fulcrum29/foxbat/internal/k8s"
	"github.com/fulcrum29/foxbat/internal/prom"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Prometheus struct {
		URL string `yaml:"url"`
	} `yaml:"prometheus"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

func main() {
	config := Config{}

	// Default values
	config.Prometheus.URL = "http://localhost:9090"
	config.Server.Port = ":8080"

	// Load from yaml if available
	if yamlFile, err := os.ReadFile("config.yaml"); err == nil {
		if err := yaml.Unmarshal(yamlFile, &config); err != nil {
			log.Printf("Warning: error parsing config.yaml: %v", err)
		}
	}

	// Environment variable overrides
	if promURL := os.Getenv("PROMETHEUS_URL"); promURL != "" {
		config.Prometheus.URL = promURL
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}

	// Initialize Prometheus client
	promClient, err := prom.NewClient(config.Prometheus.URL)
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v", err)
	}

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Printf("Warning: error creating Kubernetes client: %v", err)
	}

	// Initialize handlers
	recommenderHandler := handlers.NewRecommenderHandler(promClient, k8sClient, "web/templates")

	// Set up router with middleware
	handler := handlers.NewRouter(recommenderHandler)

	// Start HTTP server
	server := &http.Server{
		Addr:    config.Server.Port,
		Handler: handler,
	}

	slog.Info("Starting server",
		slog.String("port", config.Server.Port),
		slog.String("prometheus_url", config.Prometheus.URL),
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
