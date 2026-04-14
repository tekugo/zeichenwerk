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
	store   *Store
	metrics Metrics // reused per data point: Clear → fill → Add
}

// Export processes an incoming batch of OTLP metrics from Claude Code.
func (recv *metricsReceiver) Export(
	_ context.Context,
	req *collmetricspb.ExportMetricsServiceRequest,
) (*collmetricspb.ExportMetricsServiceResponse, error) {
	for _, rm := range req.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				recv.handleMetric(m, rm.Resource)
			}
		}
	}
	recv.store.Notify()
	return &collmetricspb.ExportMetricsServiceResponse{}, nil
}

func (recv *metricsReceiver) handleMetric(m *metricspb.Metric, res *resourcepb.Resource) {
	for _, dp := range numberDataPoints(m) {
		sessionID := attrStr(dp.Attributes, "session.id")
		if sessionID == "" {
			continue
		}

		ts := time.Unix(0, int64(dp.TimeUnixNano))
		startTS := time.Unix(0, int64(dp.StartTimeUnixNano))
		typ := attrStr(dp.Attributes, "type")

		session, isNew := recv.store.Get(sessionID)
		if isNew {
			session.ID = sessionID
			session.Start = ts
			if res != nil {
				attrs := res.GetAttributes()
				session.HostArch = attrStr(attrs, "host.arch")
				session.OSType = attrStr(attrs, "os.type")
				session.OSVersion = attrStr(attrs, "os.version")
				session.ServiceName = attrStr(attrs, "service.name")
				session.ServiceVersion = attrStr(attrs, "service.version")
			}
			session.OrgID = attrStr(dp.Attributes, "organization.id")
			session.TerminalType = attrStr(dp.Attributes, "terminal.type")
			session.UserAccountID = attrStr(dp.Attributes, "user.account_id")
			session.UserAccountUUID = attrStr(dp.Attributes, "user.account_uuid")
			session.UserEmail = attrStr(dp.Attributes, "user.email")
			session.UserID = attrStr(dp.Attributes, "user.id")
		}

		recv.metrics.Clear()
		recv.metrics.Time = ts
		recv.metrics.Start = startTS
		recv.metrics.Model = attrStr(dp.Attributes, "model")

		var logValue string

		switch m.Name {
		case "claude_code.token.usage":
			v := dpGetInt(dp)
			switch typ {
			case "input":
				recv.metrics.Input = v
			case "output":
				recv.metrics.Output = v
			case "cacheRead":
				recv.metrics.CacheRead = v
			case "cacheCreation":
				recv.metrics.CacheCreation = v
			}
			logValue = fmt.Sprintf("%d", v)

		case "claude_code.cost.usage":
			recv.metrics.Cost = dp.GetAsDouble()
			logValue = fmt.Sprintf("$%.6f", recv.metrics.Cost)

		case "claude_code.active_time.total":
			v := dp.GetAsDouble()
			switch typ {
			case "user":
				recv.metrics.ActiveUser = v
			case "cli":
				recv.metrics.ActiveCLI = v
			}
			logValue = fmt.Sprintf("%.2fs", v)

		case "claude_code.lines_of_code.count":
			v := dpGetInt(dp)
			switch typ {
			case "added":
				recv.metrics.LinesAdded = v
			case "removed":
				recv.metrics.LinesRemoved = v
			}
			logValue = fmt.Sprintf("%d", v)

		case "claude_code.code_edit_tool.decision":
			v := dpGetInt(dp)
			decision := attrStr(dp.Attributes, "decision")
			switch decision {
			case "accept":
				recv.metrics.Accepted = v
			case "reject":
				recv.metrics.Rejected = v
			}
			logValue = fmt.Sprintf("%d", v)

		default:
			continue
		}

		session.Add(&recv.metrics)

		attrs := map[string]string{
			"metric": m.Name,
			"value":  logValue,
		}
		if typ != "" {
			attrs["type"] = typ
		}
		if recv.metrics.Model != "" {
			attrs["model"] = recv.metrics.Model
		}
		session.Log = append(session.Log, Log{
			Time:  ts,
			Body:  m.Name,
			Attrs: attrs,
		})
	}
}

// dpGetInt reads a NumberDataPoint value as int64 regardless of whether the
// underlying proto oneof is AsInt or AsDouble (the Go OTel SDK may use either).
func dpGetInt(dp *metricspb.NumberDataPoint) int64 {
	switch v := dp.Value.(type) {
	case *metricspb.NumberDataPoint_AsInt:
		return v.AsInt
	case *metricspb.NumberDataPoint_AsDouble:
		return int64(v.AsDouble)
	}
	return 0
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

// attrStr returns the string value of the first attribute matching key.
func attrStr(attrs []*commonpb.KeyValue, key string) string {
	for _, kv := range attrs {
		if kv.Key == key {
			return kv.Value.GetStringValue()
		}
	}
	return ""
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
