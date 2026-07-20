package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Nikhilps0/domain-exporter/collector"
	"github.com/Nikhilps0/domain-exporter/config"
	"github.com/Nikhilps0/domain-exporter/internal"
	"github.com/Nikhilps0/domain-exporter/scheduler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	httpClient := internal.NewHTTPClient(cfg.Scheduler.RequestTimeout)

	rdapClient := collector.NewRDAPClient(
		httpClient,
		cfg.RDAP.BootstrapURL,
	)

	tlsCollector := collector.NewTLSCollector(
		cfg.TLS.Port,
		cfg.Scheduler.RequestTimeout,
	)

	dnssecCollector := collector.NewDNSSECCollector()

	metrics := collector.NewMetrics()

	domainCollector := collector.New(
		rdapClient,
		tlsCollector,
		dnssecCollector,
		metrics,
		logger,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	s := scheduler.New(
		cfg,
		domainCollector,
		logger,
	)

	go s.Start(ctx)

	http.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    cfg.Server.ListenAddress,
		Handler: http.DefaultServeMux,
	}

	go func() {
		logger.Info(
			"starting prometheus exporter",
			"address", cfg.Server.ListenAddress,
		)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down exporter")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10e9) // 10 seconds
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("exporter stopped")
}
