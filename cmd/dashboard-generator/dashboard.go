package main

import (
	"fmt"
	"strconv"
)

const dashboardUID = "itmo-high-load-system-1lab"

type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Dashboard map[string]any

func newDashboard(routes []Route) Dashboard {
	d := Dashboard{
		"uid": dashboardUID, "title": "itmo-high-load-system-1lab", "schemaVersion": 41,
		"editable": false, "version": 0, "timezone": "browser", "refresh": "15s",
		"time":   map[string]any{"from": "now-15m", "to": "now"},
		"panels": []map[string]any{},
		"templating": map[string]any{"list": []map[string]any{
			newVariable("datacenter", "Datacenter", `label_values(http_requests_total{job="itmo-high-load-system-1lab"}, datacenter)`),
			newVariable("pod_id", "Pod", `label_values(http_requests_total{job="itmo-high-load-system-1lab",datacenter=~"${datacenter:regex}"}, pod_id)`),
		}},
	}
	for _, route := range routes {
		d.addRoute(route)
	}
	return d
}

func newVariable(name, label, query string) map[string]any {
	return map[string]any{
		"name": name, "label": label, "type": "query", "query": query,
		"datasource": map[string]any{"type": "prometheus", "uid": "prometheus"},
		"refresh":    2, "multi": true, "includeAll": true, "allValue": ".*",
		"current": map[string]any{"text": "All", "value": []string{"$__all"}},
	}
}

func (d Dashboard) addRoute(route Route) {
	name := route.Method + " " + route.Path
	labels := `job="itmo-high-load-system-1lab",datacenter=~"${datacenter:regex}",pod_id=~"${pod_id:regex}",handler=` + strconv.Quote(name)
	buckets := `sum by (le) (rate(http_request_duration_seconds_bucket{` + labels + `}[$__rate_interval]))`
	d.addPanel(name+" — latency", "s", []map[string]any{
		{"refId": "A", "expr": "histogram_quantile(0.50, " + buckets + ")", "legendFormat": "p50"},
		{"refId": "B", "expr": "histogram_quantile(0.95, " + buckets + ")", "legendFormat": "p95"},
		{"refId": "C", "expr": "histogram_quantile(0.99, " + buckets + ")", "legendFormat": "p99"},
	})
	d.addPanel(name+" — RPS", "reqps", []map[string]any{
		{"refId": "A", "expr": `sum(rate(http_requests_total{` + labels + `}[$__rate_interval]))`, "legendFormat": "requests/s"},
	})
	d.addPanel(name+" — errors", "reqps", []map[string]any{
		{"refId": "A", "expr": `sum by (code) (rate(http_requests_total{` + labels + `,code=~"4..|5.."}[$__rate_interval]))`, "legendFormat": "{{code}}"},
	})
}

func (d Dashboard) addPanel(title, unit string, targets []map[string]any) {
	panels := d["panels"].([]map[string]any)
	i := len(panels)
	d["panels"] = append(panels, map[string]any{
		"id": i + 1, "title": title, "type": "timeseries", "targets": targets,
		"datasource":  map[string]any{"type": "prometheus", "uid": "prometheus"},
		"gridPos":     map[string]any{"x": (i % 3) * 8, "y": (i / 3) * 8, "w": 8, "h": 8},
		"fieldConfig": map[string]any{"defaults": map[string]any{"unit": unit, "min": 0}},
	})
}

func (r Route) validate() error {
	if r.Method == "" || len(r.Path) == 0 || r.Path[0] != '/' {
		return fmt.Errorf("invalid route: %q %q", r.Method, r.Path)
	}
	return nil
}
