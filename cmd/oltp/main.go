package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 4317, "OTLP gRPC listen port")
	logFile := flag.String("log", "otlp.log", "log file path (use - for stdout only)")
	flag.Parse()

	// Build a writer that fans out to both stdout and a file.
	var out io.Writer = os.Stdout
	if *logFile != "-" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			slog.Error("cannot open log file", "path", *logFile, "err", err)
			os.Exit(1)
		}
		defer f.Close()
		out = io.MultiWriter(os.Stdout, f)
	}

	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Use a compact time format.
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Error("failed to listen", "port", *port, "err", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	colmetricspb.RegisterMetricsServiceServer(srv, &metricsReceiver{log: logger})
	coltracepb.RegisterTraceServiceServer(srv, &traceReceiver{log: logger})
	collogspb.RegisterLogsServiceServer(srv, &logsReceiver{log: logger})

	logger.Info("OTLP receiver listening", "addr", lis.Addr(), "log", *logFile)
	if err := srv.Serve(lis); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
