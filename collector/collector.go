package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Collector struct {
	rdap    *RDAPClient
	whois   *WhoisCollector
	tls     *TLSCollector
	dnssec  *DNSSECCollector
	metrics *Metrics
	logger  *slog.Logger
}

func New(
	rdap *RDAPClient,
	whois *WhoisCollector,
	tls *TLSCollector,
	dnssec *DNSSECCollector,
	metrics *Metrics,
	logger *slog.Logger,
) *Collector {

	return &Collector{
		rdap:    rdap,
		whois:   whois,
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

	// -----------------------------------------------------
	// Domain Expiration (RDAP -> WHOIS Fallback)
	// -----------------------------------------------------

	rdapResult, err := c.rdap.Lookup(ctx, domain)

	if err == nil &&
		rdapResult.RDAPSuccess &&
		!rdapResult.ExpirationTime.IsZero() {

		result.LookupMethod = "rdap"
		result.RDAPSuccess = true
		result.RDAPDuration = rdapResult.RDAPDuration
		result.ExpirationTime = rdapResult.ExpirationTime

	} else {

		if err != nil {
			c.logger.Warn(
				"rdap lookup failed",
				"domain", domain,
				"error", err,
			)
		}

		if c.whois != nil {

			whoisResult, whoisErr := c.whois.Lookup(ctx, domain)

			if whoisErr != nil {

				c.logger.Error(
					"whois lookup failed",
					"domain", domain,
					"error", whoisErr,
				)

			} else {

				result.LookupMethod = "whois"

				result.WHOISSuccess = true
				result.WHOISDuration = whoisResult.WHOISDuration

				result.ExpirationTime = whoisResult.ExpirationTime
			}
		}
	}

	// -----------------------------------------------------
	// TLS + DNSSEC in parallel
	// -----------------------------------------------------

	var wg sync.WaitGroup

	if c.tls != nil {

		wg.Add(1)

		go func() {

			defer wg.Done()

			tlsResult, err := c.tls.Collect(ctx, domain)

			if err != nil {

				c.logger.Warn(
					"tls lookup failed",
					"domain", domain,
					"error", err,
				)

				return
			}

			result.TLSSuccess = tlsResult.TLSSuccess
			result.TLSDuration = tlsResult.TLSDuration
			result.TLSExpiryTime = tlsResult.TLSExpiryTime
			result.TLSDaysRemaining = tlsResult.TLSDaysRemaining

		}()
	}

	if c.dnssec != nil {

		wg.Add(1)

		go func() {

			defer wg.Done()

			dnsResult, err := c.dnssec.Collect(ctx, domain)

			if err != nil {

				c.logger.Warn(
					"dnssec lookup failed",
					"domain", domain,
					"error", err,
				)

				return
			}

			result.DNSSECEnabled = dnsResult.DNSSECEnabled

		}()
	}

	wg.Wait()

	c.metrics.Update(result)

	c.logger.Info(
		"collection finished",
		"domain", domain,
		"lookup", result.LookupMethod,
		"expires", result.ExpirationTime,
		"tls", result.TLSSuccess,
		"dnssec", result.DNSSECEnabled,
	)
}
