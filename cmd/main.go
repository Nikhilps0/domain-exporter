package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nikhilps0/domain-exporter/collector"
	"github.com/Nikhilps0/domain-exporter/config"
	"github.com/Nikhilps0/domain-exporter/internal"
	"github.com/Nikhilps0/domain-exporter/scheduler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	// -------------------------------------------------------
	// Logger
	// -------------------------------------------------------

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	// -------------------------------------------------------
	// Load Configuration
	// -------------------------------------------------------

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// -------------------------------------------------------
	// HTTP Client
	// -------------------------------------------------------

	httpClient := internal.NewHTTPClient(
		cfg.Scheduler.RequestTimeout,
	)

	// -------------------------------------------------------
	// Collectors
	// -------------------------------------------------------

	rdapClient := collector.NewRDAPClient(
		httpClient,
		cfg.RDAP.BootstrapURL,
	)

	whoisCollector := collector.NewWhoisCollector()

	tlsCollector := collector.NewTLSCollector(
		cfg.TLS.Port,
		cfg.Scheduler.RequestTimeout,
	)

	dnssecCollector := collector.NewDNSSECCollector()

	metrics := collector.NewMetrics()

	domainCollector := collector.New(
		rdapClient,
		whoisCollector,
		tlsCollector,
		dnssecCollector,
		metrics,
		logger,
	)

	// -------------------------------------------------------
	// Context
	// -------------------------------------------------------

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	// -------------------------------------------------------
	// Scheduler
	// -------------------------------------------------------

	s := scheduler.New(
		cfg,
		domainCollector,
		logger,
	)

	go s.Start(ctx)

	// -------------------------------------------------------
	// HTTP Server
	// -------------------------------------------------------

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         cfg.Server.ListenAddress,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {

		logger.Info(
			"starting exporter",
			"listen", cfg.Server.ListenAddress,
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.Error(
				"http server failed",
				"error", err,
			)

			cancel()
		}
	}()

	// -------------------------------------------------------
	// Wait for shutdown
	// -------------------------------------------------------

	<-ctx.Done()

	logger.Info("shutdown requested")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {

		logger.Error(
			"graceful shutdown failed",
			"error", err,
		)
	}

	logger.Info("exporter stopped")
}
