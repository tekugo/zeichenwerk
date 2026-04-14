package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	collmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
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
		info := resourceSessionInfo(rm.Resource)
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				recv.handleMetric(m, rm.Resource, info)
			}
		}
	}
	return &collmetricspb.ExportMetricsServiceResponse{}, nil
}

func (recv *metricsReceiver) handleMetric(m *metricspb.Metric, res *resourcepb.Resource, info SessionInfo) {
	points := numberDataPoints(m)
	for _, dp := range points {
		sessionID := attrStr(dp.Attributes, "session.id")
		if sessionID == "" {
			continue
		}

		// Fill in identity fields that may come from data-point attributes.
		dpInfo := info
		if dpInfo.Name == "" {
			dpInfo.Name = sessionName(dp.Attributes, "")
		}
		if dpInfo.OrgID == "" {
			dpInfo.OrgID = attrStr(dp.Attributes, "organization.id")
		}
		if dpInfo.TerminalType == "" {
			dpInfo.TerminalType = attrStr(dp.Attributes, "terminal.type")
		}
		if dpInfo.UserEmail == "" {
			dpInfo.UserEmail = attrStr(dp.Attributes, "user.email")
		}

		ts := time.Unix(0, int64(dp.TimeUnixNano))
		startTS := time.Unix(0, int64(dp.StartTimeUnixNano))
		typ := attrStr(dp.Attributes, "type")

		metric := Metric{
			Timestamp:      ts,
			StartTimestamp: startTS,
			Name:           m.Name,
			Model:          attrStr(dp.Attributes, "model"),
			Decision:       attrStr(dp.Attributes, "decision"),
			Language:       attrStr(dp.Attributes, "language"),
			Source:         attrStr(dp.Attributes, "source"),
			ToolName:       attrStr(dp.Attributes, "tool_name"),
		}

		switch m.Name {
		case "claude_code.token.usage":
			v := dp.GetAsInt()
			switch typ {
			case "input":
				metric.InputTokens = v
			case "output":
				metric.OutputTokens = v
			case "cacheRead":
				metric.CacheReadTokens = v
			case "cacheCreation":
				metric.CacheCreationTokens = v
			}
			recv.store.Log.add(otlpEntry{
				time:      ts,
				session:   dpInfo.Name,
				metric:    m.Name,
				tokenType: typ,
				value:     fmt.Sprintf("%d", v),
			})

		case "claude_code.cost.usage":
			metric.CostUSD = dp.GetAsDouble()
			recv.store.Log.add(otlpEntry{
				time:    ts,
				session: dpInfo.Name,
				metric:  m.Name,
				value:   fmt.Sprintf("$%.6f", metric.CostUSD),
			})

		case "claude_code.active_time.total":
			v := dp.GetAsDouble()
			switch typ {
			case "user":
				metric.ActiveTimeUser = v
			case "cli":
				metric.ActiveTimeCLI = v
			}
			recv.store.Log.add(otlpEntry{
				time:      ts,
				session:   dpInfo.Name,
				metric:    m.Name,
				tokenType: typ,
				value:     fmt.Sprintf("%.2fs", v),
			})

		case "claude_code.lines_of_code.count":
			v := dp.GetAsInt()
			switch typ {
			case "added":
				metric.LinesAdded = v
			case "removed":
				metric.LinesRemoved = v
			}
			recv.store.Log.add(otlpEntry{
				time:      ts,
				session:   dpInfo.Name,
				metric:    m.Name,
				tokenType: typ,
				value:     fmt.Sprintf("%d", v),
			})

		case "claude_code.code_edit_tool.decision":
			v := dp.GetAsInt()
			switch metric.Decision {
			case "accept":
				metric.EditDecisionsAccepted = v
			case "reject":
				metric.EditDecisionsRejected = v
			}
			recv.store.Log.add(otlpEntry{
				time:      ts,
				session:   dpInfo.Name,
				metric:    m.Name,
				tokenType: metric.Decision,
				value:     fmt.Sprintf("%d", v),
			})

		case "claude_code.session.count":
			metric.SessionCount = dp.GetAsInt()
			recv.store.Log.add(otlpEntry{
				time:    ts,
				session: dpInfo.Name,
				metric:  m.Name,
				value:   fmt.Sprintf("%d", metric.SessionCount),
			})

		default:
			continue
		}

		recv.store.AddMetric(sessionID, dpInfo, metric)
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

// resourceSessionInfo extracts session identity fields from resource-level attributes.
func resourceSessionInfo(res *resourcepb.Resource) SessionInfo {
	if res == nil {
		return SessionInfo{}
	}
	attrs := res.GetAttributes()
	name := ""
	for _, key := range []string{"claude_code.session.path", "process.working_directory", "service.instance.id"} {
		if v := attrStr(attrs, key); v != "" {
			name = shortPath(v)
			break
		}
	}
	return SessionInfo{
		Name:         name,
		OrgID:        attrStr(attrs, "organization.id"),
		TerminalType: attrStr(attrs, "terminal.type"),
		UserEmail:    attrStr(attrs, "user.email"),
	}
}

// sessionName returns the best display name from data-point attributes.
func sessionName(attrs []*commonpb.KeyValue, resourceName string) string {
	if resourceName != "" {
		return resourceName
	}
	for _, key := range []string{"process.working_directory", "claude_code.session.path"} {
		if v := attrStr(attrs, key); v != "" {
			return shortPath(v)
		}
	}
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
