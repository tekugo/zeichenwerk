package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	collmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	z "github.com/tekugo/zeichenwerk"
)

func main() {
	port      := flag.Int("port", 4317, "OTLP gRPC listen port")
	themeName := flag.String("t", "tokyo", "theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	flag.Parse()

	theme := resolveTheme(*themeName)
	store := NewStore()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		slog.Error("failed to listen for OTLP", "port", *port, "err", err)
		return
	}

	srv := grpc.NewServer()
	collmetricspb.RegisterMetricsServiceServer(srv, &metricsReceiver{store: store})
	go func() {
		slog.Info("OTLP receiver started", "addr", lis.Addr())
		if err := srv.Serve(lis); err != nil {
			slog.Error("OTLP receiver stopped", "err", err)
		}
	}()

	ui := buildUI(theme, store)
	ui.Run()
	srv.Stop()
}

func resolveTheme(name string) *z.Theme {
	switch name {
	case "midnight":
		return z.MidnightNeonTheme()
	case "nord":
		return z.NordTheme()
	case "gruvbox-dark":
		return z.GruvboxDarkTheme()
	case "gruvbox-light":
		return z.GruvboxLightTheme()
	case "lipstick":
		return z.LipstickTheme()
	default:
		return z.TokyoNightTheme()
	}
}
