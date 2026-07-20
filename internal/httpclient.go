package internal

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Client wraps an HTTP client with sane defaults for exporters.
type Client struct {
	http *http.Client
}

// NewHTTPClient returns a reusable HTTP client.
func NewHTTPClient(timeout time.Duration) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,

		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	return &Client{
		http: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// Do executes an HTTP request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.http.Do(req)
}

// Get performs an HTTP GET request with context.
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "domain-exporter/1.0")
	req.Header.Set("Accept", "application/rdap+json, application/json")

	return c.http.Do(req)
}
