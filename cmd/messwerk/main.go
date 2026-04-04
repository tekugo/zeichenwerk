package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	collmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	. "github.com/tekugo/zeichenwerk"
)

func main() {
	port := flag.Int("port", 4317, "OTLP gRPC listen port")
	timeout := flag.Duration("timeout", 2*time.Minute, "idle session timeout")
	themeName := flag.String("t", "tokyo", "theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	flag.Parse()

	theme := resolveTheme(*themeName)
	store := newStore(*timeout)

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

	// Ticker to keep status dots fresh even when no new data arrives.
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				store.mu.RLock()
				fn := store.onChange
				store.mu.RUnlock()
				if fn != nil {
					fn()
				}
			case <-stop:
				return
			}
		}
	}()

	ui.Run()
	close(stop)
	srv.Stop()
}

func resolveTheme(name string) *Theme {
	switch name {
	case "midnight":
		return MidnightNeonTheme()
	case "nord":
		return NordTheme()
	case "gruvbox-dark":
		return GruvboxDarkTheme()
	case "gruvbox-light":
		return GruvboxLightTheme()
	case "lipstick":
		return LipstickTheme()
	default:
		return TokyoNightTheme()
	}
}
