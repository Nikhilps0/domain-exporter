package collector

import (
	"context"
	"net"
	"strings"
	"time"
)

// DNSSECCollector checks whether a domain publishes DS records.
// If one or more DS records exist at the parent zone, the domain is
// considered DNSSEC-enabled.
type DNSSECCollector struct {
	resolver *net.Resolver
}

func NewDNSSECCollector() *DNSSECCollector {
	return &DNSSECCollector{
		resolver: net.DefaultResolver,
	}
}

// Collect checks DNSSEC status for a domain.
func (c *DNSSECCollector) Collect(ctx context.Context, domain string) (DomainResult, error) {
	result := DomainResult{
		Domain:      domain,
		CollectedAt: time.Now(),
	}

	domain = strings.TrimSuffix(domain, ".")

	// LookupDS returns the DS records published by the parent zone.
	dsRecords, err := c.resolver.LookupDS(ctx, domain)
	if err != nil {
		// Most domains without DNSSEC return "no such host" or
		// "no records". This isn't considered a collector failure.
		result.DNSSECEnabled = false
		return result, nil
	}

	result.DNSSECEnabled = len(dsRecords) > 0

	return result, nil
}
