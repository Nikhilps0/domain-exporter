package collector

import (
	"context"
	"time"

	"github.com/miekg/dns"
)

type DNSSECCollector struct {
	server string
}

func NewDNSSECCollector() *DNSSECCollector {
	return &DNSSECCollector{
		server: "8.8.8.8:53",
	}
}

func (c *DNSSECCollector) Collect(ctx context.Context, domain string) (DomainResult, error) {
	result := DomainResult{
		Domain:      domain,
		CollectedAt: time.Now(),
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeDS)

	client := &dns.Client{}

	resp, _, err := client.ExchangeContext(ctx, m, c.server)
	if err != nil {
		return result, err
	}

	for _, ans := range resp.Answer {
		if _, ok := ans.(*dns.DS); ok {
			result.DNSSECEnabled = true
			return result, nil
		}
	}

	result.DNSSECEnabled = false

	return result, nil
}
