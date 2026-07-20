package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Nikhilps0/domain-exporter/collector"
	"github.com/Nikhilps0/domain-exporter/config"
)

type Scheduler struct {
	cfg       *config.Config
	collector *collector.Collector
	logger    *slog.Logger
}

func New(
	cfg *config.Config,
	collector *collector.Collector,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{
		cfg:       cfg,
		collector: collector,
		logger:    logger,
	}
}

// Start performs an initial collection and then continues collecting
// at the configured refresh interval.
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("starting scheduler")

	// Initial collection immediately after startup.
	s.run(ctx)

	ticker := time.NewTicker(s.cfg.Scheduler.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return

		case <-ticker.C:
			s.run(ctx)
		}
	}
}

// run executes one refresh cycle using a fixed-size worker pool.
func (s *Scheduler) run(ctx context.Context) {
	start := time.Now()

	s.logger.Info(
		"starting refresh cycle",
		"domains", len(s.cfg.Domains),
		"workers", s.cfg.Scheduler.WorkerCount,
	)

	jobs := make(chan string)
	var wg sync.WaitGroup

	// Start workers.
	for i := 0; i < s.cfg.Scheduler.WorkerCount; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return

				case domain, ok := <-jobs:
					if !ok {
						return
					}

					collectCtx, cancel := context.WithTimeout(
						ctx,
						s.cfg.Scheduler.RequestTimeout,
					)

					s.collector.Collect(collectCtx, domain)

					cancel()
				}
			}
		}(i)
	}

	// Queue all domains.
	for _, domain := range s.cfg.Domains {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return

		case jobs <- domain.Name:
		}
	}

	close(jobs)
	wg.Wait()

	s.logger.Info(
		"refresh cycle completed",
		"duration", time.Since(start),
	)
}
