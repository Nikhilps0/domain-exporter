package collector

import (
	"context"
	"net"
	"strings"
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

	domain = strings.TrimSuffix(domain, ".")

	parent := parentZone(domain)

	if parent == "" {
		return result, nil
	}

	m := new(dns.Msg)

	m.SetQuestion(
		dns.Fqdn(domain),
		dns.TypeDS,
	)

	client := &dns.Client{
		Timeout: 5 * time.Second,
	}

	response, _, err := client.ExchangeContext(
		ctx,
		m,
		c.server,
	)

	if err != nil {

		// Try system resolver once
		server, lookupErr := systemDNSServer()

		if lookupErr == nil {

			response, _, err = client.ExchangeContext(
				ctx,
				m,
				server,
			)

			if err != nil {
				return result, nil
			}

		} else {
			return result, nil
		}
	}

	for _, answer := range response.Answer {

		if _, ok := answer.(*dns.DS); ok {

			result.DNSSECEnabled = true
			return result, nil
		}
	}

	result.DNSSECEnabled = false

	return result, nil
}

func parentZone(domain string) string {

	parts := strings.Split(domain, ".")

	if len(parts) < 2 {
		return ""
	}

	return parts[len(parts)-1]
}

func systemDNSServer() (string, error) {

	cfg, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return "", err
	}

	if len(cfg.Servers) == 0 {
		return "", nil
	}

	return net.JoinHostPort(
		cfg.Servers[0],
		cfg.Port,
	), nil
}
