package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	for _, name := range []string{"APP_URL", "GRAFANA_URL", "GRAFANA_ADMIN_USER", "GRAFANA_ADMIN_PASSWORD"} {
		if os.Getenv(name) == "" {
			log.Fatalf("%s is required", name)
		}
	}
	g := generator{
		appURL:     strings.TrimRight(os.Getenv("APP_URL"), "/"),
		grafanaURL: strings.TrimRight(os.Getenv("GRAFANA_URL"), "/"),
		username:   os.Getenv("GRAFANA_ADMIN_USER"), password: os.Getenv("GRAFANA_ADMIN_PASSWORD"),
		client: &http.Client{Timeout: 15 * time.Second},
	}
	routes, err := g.handlers()
	if err != nil {
		log.Fatal(err)
	}
	dashboard := newDashboard(routes)
	if err := g.save(dashboard); err != nil {
		log.Fatal(err)
	}
	log.Printf("saved /d/%s: %d routes, %d panels", dashboardUID, len(routes), len(dashboard["panels"].([]map[string]any)))
}
