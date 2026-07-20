package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nikhilps0/domain-exporter/internal"
)

// -----------------------------------------------------------------------------
// IANA Bootstrap Models
// -----------------------------------------------------------------------------

type bootstrapResponse struct {
	Services [][]json.RawMessage `json:"services"`
}

type rdapResponse struct {
	Events []rdapEvent `json:"events"`
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

// -----------------------------------------------------------------------------
// RDAP Client
// -----------------------------------------------------------------------------

type RDAPClient struct {
	http          *internal.Client
	bootstrapURL  string
	bootstrapCache map[string]string
}

func NewRDAPClient(httpClient *internal.Client, bootstrapURL string) *RDAPClient {
	return &RDAPClient{
		http:           httpClient,
		bootstrapURL:   bootstrapURL,
		bootstrapCache: make(map[string]string),
	}
}

// Lookup performs a complete RDAP lookup and returns the expiration time.
func (c *RDAPClient) Lookup(ctx context.Context, domain string) (DomainResult, error) {
	result := DomainResult{
		Domain:      domain,
		CollectedAt: time.Now(),
	}

	start := time.Now()

	server, err := c.getRDAPServer(ctx, domain)
	if err != nil {
		result.RDAPDuration = time.Since(start)
		return result, err
	}

	url := strings.TrimRight(server, "/") + "/domain/" + domain

	resp, err := c.http.Get(ctx, url)
	if err != nil {
		result.RDAPDuration = time.Since(start)
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.RDAPDuration = time.Since(start)
		return result, fmt.Errorf("rdap returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.RDAPDuration = time.Since(start)
		return result, err
	}

	var rdap rdapResponse
	if err := json.Unmarshal(body, &rdap); err != nil {
		result.RDAPDuration = time.Since(start)
		return result, err
	}

	for _, event := range rdap.Events {
		if strings.EqualFold(event.EventAction, "expiration") {
			t, err := time.Parse(time.RFC3339, event.EventDate)
			if err == nil {
				result.ExpirationTime = t
				break
			}
		}
	}

	result.RDAPSuccess = true
	result.RDAPDuration = time.Since(start)

	return result, nil
}

// -----------------------------------------------------------------------------
// Bootstrap
// -----------------------------------------------------------------------------

func (c *RDAPClient) getRDAPServer(ctx context.Context, domain string) (string, error) {

	tld := getTLD(domain)

	// cache hit
	if server, ok := c.bootstrapCache[tld]; ok {
		return server, nil
	}

	resp, err := c.http.Get(ctx, c.bootstrapURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var bootstrap bootstrapResponse

	if err := json.NewDecoder(resp.Body).Decode(&bootstrap); err != nil {
		return "", err
	}

	for _, service := range bootstrap.Services {

		if len(service) != 2 {
			continue
		}

		var tlds []string
		var urls []string

		if err := json.Unmarshal(service[0], &tlds); err != nil {
			continue
		}

		if err := json.Unmarshal(service[1], &urls); err != nil {
			continue
		}

		for _, item := range tlds {
			if strings.EqualFold(item, tld) {

				if len(urls) == 0 {
					break
				}

				server := urls[0]

				c.bootstrapCache[tld] = server

				return server, nil
			}
		}
	}

	return "", fmt.Errorf("no RDAP server found for TLD %s", tld)
}

func getTLD(domain string) string {

	parts := strings.Split(domain, ".")

	if len(parts) < 2 {
		return domain
	}

	return parts[len(parts)-1]
}
