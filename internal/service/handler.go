package service

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Route struct {
	Method  string           `json:"method"`
	Path    string           `json:"path"`
	Handler http.HandlerFunc `json:"-"`
}

func NewHandler() http.Handler {
	req := &itemReq{items: map[string]Item{"1": {Name: "Example", ID: "1"}}}

	metrics := newHTTPMetrics(os.Getenv("POD_ID"), os.Getenv("DATACENTER"))
	mux := http.NewServeMux()
	routes := []Route{
		{Method: http.MethodGet, Path: "/items/{id}", Handler: req.getItem},
		{Method: http.MethodPost, Path: "/items", Handler: req.postItem},
		{Method: http.MethodPut, Path: "/items/{id}", Handler: req.putItem},
	}
	for _, route := range routes {
		metrics.register(mux, route.Method+" "+route.Path, route.Handler)
	}
	mux.HandleFunc("GET /handlers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(routes); err != nil {
			log.Printf("write handlers response: %v", err)
		}
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("GET /metrics", metrics.scrapeHandler())
	return mux
}
