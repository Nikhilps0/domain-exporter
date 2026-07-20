package collector

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"
)

// TLSCollector collects TLS certificate information.
type TLSCollector struct {
	port    int
	timeout time.Duration
}

func NewTLSCollector(port int, timeout time.Duration) *TLSCollector {
	return &TLSCollector{
		port:    port,
		timeout: timeout,
	}
}

// Collect retrieves the TLS certificate for a domain.
func (c *TLSCollector) Collect(ctx context.Context, domain string) (DomainResult, error) {
	result := DomainResult{
		Domain:      domain,
		CollectedAt: time.Now(),
	}

	start := time.Now()

	address := net.JoinHostPort(domain, strconv.Itoa(c.port))

	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		address,
		&tls.Config{
			ServerName:         domain,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		},
	)
	if err != nil {
		result.TLSDuration = time.Since(start)
		return result, fmt.Errorf("tls dial failed: %w", err)
	}
	defer conn.Close()

	// Respect context cancellation after connection.
	select {
	case <-ctx.Done():
		result.TLSDuration = time.Since(start)
		return result, ctx.Err()
	default:
	}

	state := conn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		result.TLSDuration = time.Since(start)
		return result, fmt.Errorf("no peer certificate received")
	}

	cert := state.PeerCertificates[0]

	result.TLSSuccess = true
	result.TLSDuration = time.Since(start)
	result.TLSExpiryTime = cert.NotAfter
	result.TLSDaysRemaining = int(time.Until(cert.NotAfter).Hours() / 24)

	return result, nil
}
