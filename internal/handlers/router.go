package handlers

import (
	"net/http"
)

// NewRouter sets up the routing for the application and wraps it with middleware.
func NewRouter(rh *RecommenderHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", rh.HandleIndex)
	mux.HandleFunc("GET /ui/workloads", rh.HandleGetWorkloads)
	mux.HandleFunc("GET /ui/pods", rh.HandleGetPods)
	mux.HandleFunc("GET /ui/recommendations", rh.HandleGetRecommendationsUI)
	mux.HandleFunc("GET /api/recommendations", rh.HandleGetRecommendations)

	return Logging(mux)
}

// RegisterRoutes sets up the routing for the application.
func RegisterRoutes(mux *http.ServeMux, rh *RecommenderHandler) {
	mux.HandleFunc("GET /", rh.HandleIndex)
	mux.HandleFunc("GET /ui/workloads", rh.HandleGetWorkloads)
	mux.HandleFunc("GET /ui/pods", rh.HandleGetPods)
	mux.HandleFunc("GET /ui/recommendations", rh.HandleGetRecommendationsUI)
	mux.HandleFunc("GET /api/recommendations", rh.HandleGetRecommendations)
}