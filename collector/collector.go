package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Collector struct {
	rdap    *RDAPClient
	tls     *TLSCollector
	dnssec  *DNSSECCollector
	metrics *Metrics
	logger  *slog.Logger
}

func New(
	rdap *RDAPClient,
	tls *TLSCollector,
	dnssec *DNSSECCollector,
	metrics *Metrics,
	logger *slog.Logger,
) *Collector {
	return &Collector{
		rdap:    rdap,
		tls:     tls,
		dnssec:  dnssec,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *Collector) Collect(ctx context.Context, domain string) {
	result := DomainResult{
		Domain:      domain,
		CollectedAt: time.Now(),
	}

	var wg sync.WaitGroup

	var rdapResult DomainResult
	var tlsResult DomainResult
	var dnssecResult DomainResult

	wg.Add(1)
	go func() {
		defer wg.Done()

		r, err := c.rdap.Lookup(ctx, domain)
		if err != nil {
			c.logger.Error(
				"rdap lookup failed",
				"domain", domain,
				"error", err,
			)
			return
		}

		rdapResult = r
	}()

	if c.tls != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			r, err := c.tls.Collect(ctx, domain)
			if err != nil {
				c.logger.Warn(
					"tls lookup failed",
					"domain", domain,
					"error", err,
				)
				return
			}

			tlsResult = r
		}()
	}

	if c.dnssec != nil {
		wg.Add(1)

		go func() {
			defer wg.Done()

			r, err := c.dnssec.Collect(ctx, domain)
			if err != nil {
				c.logger.Warn(
					"dnssec lookup failed",
					"domain", domain,
					"error", err,
				)
				return
			}

			dnssecResult = r
		}()
	}

	wg.Wait()

	// Merge RDAP
	result.RDAPSuccess = rdapResult.RDAPSuccess
	result.RDAPDuration = rdapResult.RDAPDuration
	result.ExpirationTime = rdapResult.ExpirationTime

	// Merge TLS
	result.TLSSuccess = tlsResult.TLSSuccess
	result.TLSDuration = tlsResult.TLSDuration
	result.TLSExpiryTime = tlsResult.TLSExpiryTime
	result.TLSDaysRemaining = tlsResult.TLSDaysRemaining

	// Merge DNSSEC
	result.DNSSECEnabled = dnssecResult.DNSSECEnabled

	c.metrics.Update(result)

	c.logger.Info(
		"collection completed",
		"domain", domain,
		"rdap", result.RDAPSuccess,
		"tls", result.TLSSuccess,
		"dnssec", result.DNSSECEnabled,
	)
}
