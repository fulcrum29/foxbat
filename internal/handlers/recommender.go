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
	"github.com/prometheus/common/model"
)
...
// HandleGetRecommendationsUI processes HTMX requests and returns HTML fragments.
func (h *RecommenderHandler) HandleGetRecommendationsUI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queryParams := prom.QueryParams{
		Namespace: r.URL.Query().Get("namespace"),
		Pod:       r.URL.Query().Get("pod"),
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

	data := struct {
		CPU    model.Vector
		Memory model.Vector
	}{
		CPU:    cpuVector,
		Memory: memVector,
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
		Pod:       r.URL.Query().Get("pod"),
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
