package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const dashboardsPath = "/apis/dashboard.grafana.app/v1/namespaces/default/dashboards"

type generator struct {
	appURL, grafanaURL string
	username, password string
	client             *http.Client
}

type Metadata struct {
	Name            string `json:"name"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

type Resource struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   Metadata  `json:"metadata"`
	Spec       Dashboard `json:"spec"`
}

func (g *generator) handlers() ([]Route, error) {
	resp, err := g.client.Get(g.appURL + "/handlers")
	if err != nil {
		return nil, fmt.Errorf("get handlers: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get handlers: %s", resp.Status)
	}
	var routes []Route
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&routes); err != nil {
		return nil, fmt.Errorf("decode handlers: %w", err)
	}
	if len(routes) == 0 {
		return nil, fmt.Errorf("application returned no handlers")
	}
	seen := make(map[Route]bool)
	for _, route := range routes {
		if err := route.validate(); err != nil {
			return nil, err
		}
		if seen[route] {
			return nil, fmt.Errorf("duplicate route: %s %s", route.Method, route.Path)
		}
		seen[route] = true
	}
	return routes, nil
}

func (g *generator) save(d Dashboard) error {
	resource := Resource{
		APIVersion: "dashboard.grafana.app/v1", Kind: "Dashboard",
		Metadata: Metadata{Name: dashboardUID}, Spec: d,
	}
	path, method := dashboardsPath, http.MethodPost
	resp, err := g.grafanaRequest(http.MethodGet, path+"/"+dashboardUID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var existing Resource
		if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&existing); err != nil {
			return fmt.Errorf("decode dashboard: %w", err)
		}
		resource.Metadata.ResourceVersion = existing.Metadata.ResourceVersion
		resource.Spec["version"] = existing.Spec["version"]
		method, path = http.MethodPut, path+"/"+dashboardUID
	case http.StatusNotFound:
	default:
		return fmt.Errorf("get dashboard: %s", resp.Status)
	}
	body, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	saved, err := g.grafanaRequest(method, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer saved.Body.Close()
	if saved.StatusCode != http.StatusOK && saved.StatusCode != http.StatusCreated {
		return fmt.Errorf("save dashboard: %s", saved.Status)
	}
	return nil
}

func (g *generator) grafanaRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, g.grafanaURL+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(g.username, g.password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Grafana %s: %w", method, err)
	}
	return resp, nil
}
