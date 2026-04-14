package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// ── Metrics ──────────────────────────────────────────────────────────────────

type metricsReceiver struct {
	colmetricspb.UnimplementedMetricsServiceServer
	log *slog.Logger
}

func (r *metricsReceiver) Export(
	_ context.Context,
	req *colmetricspb.ExportMetricsServiceRequest,
) (*colmetricspb.ExportMetricsServiceResponse, error) {
	for _, rm := range req.ResourceMetrics {
		res := rm.Resource
		for _, sm := range rm.ScopeMetrics {
			scope := sm.Scope
			for _, m := range sm.Metrics {
				r.logMetric(res, scope, m)
			}
		}
	}
	return &colmetricspb.ExportMetricsServiceResponse{}, nil
}

func (r *metricsReceiver) logMetric(
	res *resourcepb.Resource,
	scope *commonpb.InstrumentationScope,
	m *metricspb.Metric,
) {
	base := []any{
		"signal", "metric",
		"name", m.Name,
		"description", m.Description,
		"unit", m.Unit,
	}
	if scope != nil && scope.Name != "" {
		base = append(base, "scope", scope.Name, "scope.version", scope.Version)
	}
	base = append(base, resourceAttrs(res)...)

	switch d := m.Data.(type) {
	case *metricspb.Metric_Gauge:
		for _, dp := range d.Gauge.DataPoints {
			r.log.Info("metric", append(base, numberDPAttrs(dp)...)...)
		}
	case *metricspb.Metric_Sum:
		for _, dp := range d.Sum.DataPoints {
			r.log.Info("metric", append(base,
				append([]any{
					"aggregation_temporality", d.Sum.AggregationTemporality.String(),
					"is_monotonic", d.Sum.IsMonotonic,
				}, numberDPAttrs(dp)...)...)...)
		}
	case *metricspb.Metric_Histogram:
		for _, dp := range d.Histogram.DataPoints {
			r.log.Info("metric", append(base,
				"ts", unixNanoTime(dp.TimeUnixNano),
				"count", dp.Count,
				"sum", dp.Sum,
				"min", dp.Min,
				"max", dp.Max,
				"attrs", attrsToMap(dp.Attributes),
				"exemplars", len(dp.Exemplars),
			)...)
		}
	case *metricspb.Metric_ExponentialHistogram:
		for _, dp := range d.ExponentialHistogram.DataPoints {
			r.log.Info("metric", append(base,
				"ts", unixNanoTime(dp.TimeUnixNano),
				"count", dp.Count,
				"sum", dp.Sum,
				"attrs", attrsToMap(dp.Attributes),
			)...)
		}
	case *metricspb.Metric_Summary:
		for _, dp := range d.Summary.DataPoints {
			r.log.Info("metric", append(base,
				"ts", unixNanoTime(dp.TimeUnixNano),
				"count", dp.Count,
				"sum", dp.Sum,
				"attrs", attrsToMap(dp.Attributes),
			)...)
		}
	default:
		r.log.Info("metric", base...)
	}
}

func numberDPAttrs(dp *metricspb.NumberDataPoint) []any {
	var value any
	switch v := dp.Value.(type) {
	case *metricspb.NumberDataPoint_AsDouble:
		value = v.AsDouble
	case *metricspb.NumberDataPoint_AsInt:
		value = v.AsInt
	}
	return []any{
		"ts", unixNanoTime(dp.TimeUnixNano),
		"start_ts", unixNanoTime(dp.StartTimeUnixNano),
		"value", value,
		"attrs", attrsToMap(dp.Attributes),
		"flags", dp.Flags,
	}
}

// ── Traces ───────────────────────────────────────────────────────────────────

type traceReceiver struct {
	coltracepb.UnimplementedTraceServiceServer
	log *slog.Logger
}

func (r *traceReceiver) Export(
	_ context.Context,
	req *coltracepb.ExportTraceServiceRequest,
) (*coltracepb.ExportTraceServiceResponse, error) {
	for _, rs := range req.ResourceSpans {
		res := rs.Resource
		for _, ss := range rs.ScopeSpans {
			scope := ss.Scope
			for _, span := range ss.Spans {
				r.logSpan(res, scope, span)
			}
		}
	}
	return &coltracepb.ExportTraceServiceResponse{}, nil
}

func (r *traceReceiver) logSpan(
	res *resourcepb.Resource,
	scope *commonpb.InstrumentationScope,
	span *tracepb.Span,
) {
	attrs := []any{
		"signal", "trace",
		"trace_id", hexID(span.TraceId),
		"span_id", hexID(span.SpanId),
		"parent_span_id", hexID(span.ParentSpanId),
		"name", span.Name,
		"kind", span.Kind.String(),
		"start_ts", unixNanoTime(span.StartTimeUnixNano),
		"end_ts", unixNanoTime(span.EndTimeUnixNano),
		"attrs", attrsToMap(span.Attributes),
		"events", len(span.Events),
		"links", len(span.Links),
		"flags", span.Flags,
	}
	if span.Status != nil {
		attrs = append(attrs,
			"status.code", span.Status.Code.String(),
			"status.message", span.Status.Message,
		)
	}
	if scope != nil && scope.Name != "" {
		attrs = append(attrs, "scope", scope.Name, "scope.version", scope.Version)
	}
	attrs = append(attrs, resourceAttrs(res)...)

	r.log.Info("span", attrs...)

	for _, ev := range span.Events {
		r.log.Info("span.event",
			"trace_id", hexID(span.TraceId),
			"span_id", hexID(span.SpanId),
			"name", ev.Name,
			"ts", unixNanoTime(ev.TimeUnixNano),
			"attrs", attrsToMap(ev.Attributes),
		)
	}
}

// ── Logs ─────────────────────────────────────────────────────────────────────

type logsReceiver struct {
	collogspb.UnimplementedLogsServiceServer
	log *slog.Logger
}

func (r *logsReceiver) Export(
	_ context.Context,
	req *collogspb.ExportLogsServiceRequest,
) (*collogspb.ExportLogsServiceResponse, error) {
	for _, rl := range req.ResourceLogs {
		res := rl.Resource
		for _, sl := range rl.ScopeLogs {
			scope := sl.Scope
			for _, lr := range sl.LogRecords {
				r.logRecord(res, scope, lr)
			}
		}
	}
	return &collogspb.ExportLogsServiceResponse{}, nil
}

func (r *logsReceiver) logRecord(
	res *resourcepb.Resource,
	scope *commonpb.InstrumentationScope,
	lr *logspb.LogRecord,
) {
	body := ""
	if lr.Body != nil {
		body = anyValueString(lr.Body)
	}
	attrs := []any{
		"signal", "log",
		"ts", unixNanoTime(lr.TimeUnixNano),
		"observed_ts", unixNanoTime(lr.ObservedTimeUnixNano),
		"severity", lr.SeverityText,
		"severity_number", lr.SeverityNumber.String(),
		"body", body,
		"trace_id", hexID(lr.TraceId),
		"span_id", hexID(lr.SpanId),
		"attrs", attrsToMap(lr.Attributes),
		"flags", lr.Flags,
	}
	if scope != nil && scope.Name != "" {
		attrs = append(attrs, "scope", scope.Name, "scope.version", scope.Version)
	}
	attrs = append(attrs, resourceAttrs(res)...)

	r.log.Info("log", attrs...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func resourceAttrs(res *resourcepb.Resource) []any {
	if res == nil {
		return nil
	}
	return []any{"resource", attrsToMap(res.Attributes)}
}

func attrsToMap(kvs []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = anyValueString(kv.Value)
	}
	return m
}

func anyValueString(v *commonpb.AnyValue) string {
	if v == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return val.StringValue
	case *commonpb.AnyValue_IntValue:
		return fmt.Sprintf("%d", val.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		return fmt.Sprintf("%g", val.DoubleValue)
	case *commonpb.AnyValue_BoolValue:
		if val.BoolValue {
			return "true"
		}
		return "false"
	case *commonpb.AnyValue_BytesValue:
		return fmt.Sprintf("%x", val.BytesValue)
	case *commonpb.AnyValue_ArrayValue:
		if val.ArrayValue == nil {
			return "[]"
		}
		parts := make([]string, len(val.ArrayValue.Values))
		for i, av := range val.ArrayValue.Values {
			parts[i] = anyValueString(av)
		}
		return "[" + joinStrings(parts) + "]"
	case *commonpb.AnyValue_KvlistValue:
		if val.KvlistValue == nil {
			return "{}"
		}
		parts := make([]string, len(val.KvlistValue.Values))
		for i, kv := range val.KvlistValue.Values {
			parts[i] = kv.Key + "=" + anyValueString(kv.Value)
		}
		return "{" + joinStrings(parts) + "}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func hexID(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

func unixNanoTime(ns uint64) time.Time {
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(ns))
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}
