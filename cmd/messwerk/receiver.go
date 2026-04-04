package main

import (
	"context"
	"path/filepath"
	"time"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	collmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

// metricsReceiver implements the OTLP MetricsService gRPC endpoint.
type metricsReceiver struct {
	collmetricspb.UnimplementedMetricsServiceServer
	store *Store
}

// Export processes an incoming batch of OTLP metrics from Claude Code.
func (recv *metricsReceiver) Export(
	_ context.Context,
	req *collmetricspb.ExportMetricsServiceRequest,
) (*collmetricspb.ExportMetricsServiceResponse, error) {
	for _, rm := range req.ResourceMetrics {
		sessionName := resourceSessionName(rm.Resource)
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				recv.handleMetric(m, sessionName)
			}
		}
	}
	return &collmetricspb.ExportMetricsServiceResponse{}, nil
}

func (recv *metricsReceiver) handleMetric(m *metricspb.Metric, resourceName string) {
	points := numberDataPoints(m)
	for _, dp := range points {
		sessionID := attrStr(dp.Attributes, "session.id")
		if sessionID == "" {
			continue
		}
		name := sessionName(dp.Attributes, resourceName)
		ts := time.Unix(0, int64(dp.TimeUnixNano))

		switch m.Name {
		case "claude_code.token.usage":
			tokenType := attrStr(dp.Attributes, "type")
			recv.store.UpdateTokens(sessionID, name, tokenType, dp.GetAsInt(), ts)
		case "claude_code.cost.usage":
			recv.store.UpdateCost(sessionID, dp.GetAsDouble())
		}
	}
}

// numberDataPoints extracts the data-point slice regardless of aggregation type.
func numberDataPoints(m *metricspb.Metric) []*metricspb.NumberDataPoint {
	if s := m.GetSum(); s != nil {
		return s.DataPoints
	}
	if g := m.GetGauge(); g != nil {
		return g.DataPoints
	}
	return nil
}

// resourceSessionName attempts to derive a human-readable session name from
// resource-level attributes. Falls back to an empty string.
func resourceSessionName(res interface{ GetAttributes() []*commonpb.KeyValue }) string {
	if res == nil {
		return ""
	}
	for _, key := range []string{
		"claude_code.session.path",
		"process.working_directory",
		"service.instance.id",
	} {
		if v := attrStr(res.GetAttributes(), key); v != "" {
			return v
		}
	}
	return ""
}

// sessionName returns the best display name for a session: prefer resource
// name, then data-point attribute, then a short path of the session ID.
func sessionName(attrs []*commonpb.KeyValue, resourceName string) string {
	if resourceName != "" {
		return shortPath(resourceName)
	}
	for _, key := range []string{"process.working_directory", "claude_code.session.path"} {
		if v := attrStr(attrs, key); v != "" {
			return shortPath(v)
		}
	}
	// Fall back to the session ID — caller will truncate if needed.
	return attrStr(attrs, "session.id")
}

// shortPath returns the last two path components of p for compact display.
func shortPath(p string) string {
	parent := filepath.Base(filepath.Dir(p))
	base := filepath.Base(p)
	if parent == "." || parent == "/" || parent == "" {
		return base
	}
	return parent + "/" + base
}

// attrStr returns the string value of the first attribute matching key.
func attrStr(attrs []*commonpb.KeyValue, key string) string {
	for _, kv := range attrs {
		if kv.Key == key {
			return kv.Value.GetStringValue()
		}
	}
	return ""
}
