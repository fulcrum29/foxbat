package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/fulcrum29/foxbat/internal/k8s"
	"github.com/fulcrum29/foxbat/internal/prom"
	"github.com/fulcrum29/foxbat/internal/templates"
	"github.com/prometheus/common/model"
)

type RecommenderHandler struct {
	promClient *prom.Client
	k8sClient  *k8s.Client
	templates  *template.Template
}

// safeDiv handles type assertion for Prometheus metric values and performs division.
func safeDiv(a, b interface{}) float64 {
	toFloat := func(v interface{}) float64 {
		switch val := v.(type) {
		case model.SampleValue:
			return float64(val)
		case float64:
			return val
		default:
			return 0
		}
	}

	af, bf := toFloat(a), toFloat(b)
	if bf == 0 {
		return 0
	}
	return af / bf
}

// NewRecommenderHandler initializes a new handler with Prometheus and Kubernetes clients and parses templates.
func NewRecommenderHandler(pc *prom.Client, kc *k8s.Client, templatePath string) *RecommenderHandler {
	funcMap := template.FuncMap{
		"label": func(s string) model.LabelName { return model.LabelName(s) },
		"div":   safeDiv,
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templatePath, "*.html")))

	return &RecommenderHandler{
		promClient: pc,
		k8sClient:  kc,
		templates:  tmpl,
	}
}

// HandleIndex serves the main UI page with initial lists.
func (h *RecommenderHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespaces, _ := h.k8sClient.ListNamespaces(ctx)

	data := templates.IndexData{
		Namespaces: namespaces,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("Failed to render index template", "error", err)
		http.Error(w, "Failed to render index template", http.StatusInternalServerError)
	}
}

// HandleGetWorkloads returns an HTML fragment of workload options.
func (h *RecommenderHandler) HandleGetWorkloads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := r.URL.Query().Get("namespace")
	workloads, _ := h.k8sClient.ListWorkloads(ctx, namespace)

	data := templates.SelectOptionsData{
		Items:  workloads,
		Label:  "Workload",
		Name:   "workload",
		Target: "/ui/pods",
	}

	if err := h.templates.ExecuteTemplate(w, "select-options.html", data); err != nil {
		slog.Error("Failed to render select-options template", "error", err)
		http.Error(w, "Failed to render select-options template", http.StatusInternalServerError)
	}
}

// HandleGetPods returns an HTML fragment of pod options.
func (h *RecommenderHandler) HandleGetPods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := r.URL.Query().Get("namespace")
	workload := r.URL.Query().Get("workload")
	pods, _ := h.k8sClient.ListPods(ctx, namespace, workload)

	data := struct {
		Items  []string
		Label  string
		Name   string
		Target string
	}{
		Items: pods,
		Label: "Pod",
		Name:  "pod",
	}

	if err := h.templates.ExecuteTemplate(w, "select-options.html", data); err != nil {
		slog.Error("Failed to render select-options template", "error", err)
		http.Error(w, "Failed to render select-options template", http.StatusInternalServerError)
	}
}

// HandleGetRecommendationsUI processes HTMX requests and returns HTML fragments.
func (h *RecommenderHandler) HandleGetRecommendationsUI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queryParams := prom.QueryParams{
		Namespace: r.URL.Query().Get("namespace"),
		Workload:  r.URL.Query().Get("workload"),
	}

	cpuRecs, err := h.promClient.GetCPURecommendations(ctx, queryParams)
	if err != nil {
		slog.Error("Failed to get CPU recommendations", "error", err, "params", queryParams)
		http.Error(w, fmt.Sprintf("Failed to get CPU recommendations: %v", err), http.StatusInternalServerError)
		return
	}

	memRecs, err := h.promClient.GetMemoryRecommendations(ctx, queryParams)
	if err != nil {
		slog.Error("Failed to get Memory recommendations", "error", err, "params", queryParams)
		http.Error(w, fmt.Sprintf("Failed to get Memory recommendations: %v", err), http.StatusInternalServerError)
		return
	}

	cpuVector, ok := cpuRecs.(model.Vector)
	if !ok {
		slog.Error("CPU recommendations result is not a vector", "type", fmt.Sprintf("%T", cpuRecs))
		http.Error(w, "Internal error: CPU result type mismatch", http.StatusInternalServerError)
		return
	}

	memVector, ok := memRecs.(model.Vector)
	if !ok {
		slog.Error("Memory recommendations result is not a vector", "type", fmt.Sprintf("%T", memRecs))
		http.Error(w, "Internal error: Memory result type mismatch", http.StatusInternalServerError)
		return
	}

	data := templates.RecommenderData{
		CPU:       cpuVector,
		Memory:    memVector,
		Namespace: queryParams.Namespace,
		Workload:  queryParams.Workload,
	}

	if err := h.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		slog.Error("Failed to render results template", "error", err)
		http.Error(w, "Failed to render results template", http.StatusInternalServerError)
	}
}

// HandleGetRecommendations processes HTTP requests to fetch and return recommendations.
func (h *RecommenderHandler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queryParams := prom.QueryParams{
		Namespace: r.URL.Query().Get("namespace"),
		Workload:  r.URL.Query().Get("workload"),
	}

	cpuRecs, err := h.promClient.GetCPURecommendations(ctx, queryParams)
	if err != nil {
		slog.Error("API: Failed to get CPU recommendations", "error", err, "params", queryParams)
		http.Error(w, "Failed to get CPU recommendations", http.StatusInternalServerError)
		return
	}

	memRecs, err := h.promClient.GetMemoryRecommendations(ctx, queryParams)
	if err != nil {
		slog.Error("API: Failed to get Memory recommendations", "error", err, "params", queryParams)
		http.Error(w, "Failed to get Memory recommendations", http.StatusInternalServerError)
		return
	}

	// For now, combining results into a single JSON response
	response := map[string]interface{}{
		"cpu_recommendations":    cpuRecs,
		"memory_recommendations": memRecs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
