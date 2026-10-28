package main

import (
	"fmt"
	"strconv"
)

const dashboardUID = "itmo-high-load-system-1lab"
const dashboardLabels = `job="itmo-high-load-system-1lab",datacenter=~"${datacenter:regex}",pod_id=~"${pod_id:regex}"`

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
	d.addOverview()
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
	labels := dashboardLabels + `,handler=` + strconv.Quote(name)
	d.placePanel(map[string]any{
		"title": name, "type": "row", "collapsed": false,
		"panels": []map[string]any{},
	}, 24, 1)
	buckets := `sum by (le) (rate(http_request_duration_seconds_bucket{` + labels + `}[$__rate_interval]))`
	d.addPanel("timeseries", name+" — latency", "s", []map[string]any{
		{"refId": "A", "expr": "histogram_quantile(0.50, " + buckets + ")", "legendFormat": "p50"},
		{"refId": "B", "expr": "histogram_quantile(0.95, " + buckets + ")", "legendFormat": "p95"},
		{"refId": "C", "expr": "histogram_quantile(0.99, " + buckets + ")", "legendFormat": "p99"},
	})
	d.addPanel("timeseries", name+" — RPS", "reqps", []map[string]any{
		{"refId": "A", "expr": `sum(rate(http_requests_total{` + labels + `}[$__rate_interval]))`, "legendFormat": "requests/s"},
	})
	d.addPanel("timeseries", name+" — errors", "reqps", []map[string]any{
		{"refId": "A", "expr": `sum by (code) (rate(http_requests_total{` + labels + `,code=~"4..|5.."}[$__rate_interval]))`, "legendFormat": "{{code}}"},
	})
}

func (d Dashboard) addOverview() {
	buckets := `sum by (le) (rate(http_request_duration_seconds_bucket{` + dashboardLabels + `}[$__rate_interval]))`
	rps := `sum(rate(http_requests_total{` + dashboardLabels + `}[$__rate_interval]))`
	errors := `(sum(rate(http_requests_total{` + dashboardLabels + `,code=~"4..|5.."}[$__rate_interval])) or (0 * ` + rps + `))`
	for _, stat := range []struct{ title, unit, expr string }{
		{"p50", "s", "histogram_quantile(0.50, " + buckets + ")"},
		{"p95", "s", "histogram_quantile(0.95, " + buckets + ")"},
		{"p99", "s", "histogram_quantile(0.99, " + buckets + ")"},
		{"RPS", "reqps", rps},
		{"Errors", "reqps", errors},
		{"Error ratio", "percent", "100 * " + errors + " / " + rps},
	} {
		d.addPanel("stat", stat.title, stat.unit, []map[string]any{
			{"refId": "A", "expr": stat.expr, "instant": true},
		})
	}
}

func (d Dashboard) addPanel(kind, title, unit string, targets []map[string]any) {
	panel := map[string]any{
		"title": title, "type": kind, "targets": targets,
		"datasource":  map[string]any{"type": "prometheus", "uid": "prometheus"},
		"fieldConfig": map[string]any{"defaults": map[string]any{"unit": unit, "min": 0}},
	}
	width, height := 8, 8
	if kind == "stat" {
		width, height = 4, 4
		panel["description"] = "All application routes across selected pods. Current rate window, not the whole dashboard time range. Percentiles are estimates from histogram buckets."
		panel["options"] = map[string]any{
			"colorMode": "value", "graphMode": "none", "textMode": "auto",
			"reduceOptions": map[string]any{"calcs": []string{"last"}, "values": false},
		}
	} else {
		panel["options"] = map[string]any{
			"tooltip": map[string]any{"mode": "multi"},
			"legend":  map[string]any{"displayMode": "list", "placement": "bottom"},
		}
	}
	d.placePanel(panel, width, height)
}

func (d Dashboard) placePanel(panel map[string]any, width, height int) {
	panels := d["panels"].([]map[string]any)
	x, y, bottom := 0, 0, 0
	for _, previous := range panels {
		grid := previous["gridPos"].(map[string]any)
		x, y = grid["x"].(int)+grid["w"].(int), grid["y"].(int)
		bottom = max(bottom, grid["y"].(int)+grid["h"].(int))
	}
	if x+width > 24 {
		x, y = 0, bottom
	}
	panel["id"] = len(panels) + 1
	panel["gridPos"] = map[string]any{"x": x, "y": y, "w": width, "h": height}
	d["panels"] = append(panels, panel)
}

func (r Route) validate() error {
	if r.Method == "" || len(r.Path) == 0 || r.Path[0] != '/' {
		return fmt.Errorf("invalid route: %q %q", r.Method, r.Path)
	}
	return nil
}
