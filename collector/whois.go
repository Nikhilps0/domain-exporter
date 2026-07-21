package collector

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type WhoisCollector struct{}

func NewWhoisCollector() *WhoisCollector {
	return &WhoisCollector{}
}

func (w *WhoisCollector) Lookup(ctx context.Context, domain string) (DomainResult, error) {

	result := DomainResult{
		Domain:      domain,
		CollectedAt: time.Now(),
	}

	start := time.Now()

	cmd := exec.CommandContext(ctx, "whois", domain)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.WHOISDuration = time.Since(start)
		return result, fmt.Errorf("whois failed: %v (%s)", err, stderr.String())
	}

	result.WHOISDuration = time.Since(start)

	expiry, err := parseWhoisExpiration(stdout.String())
	if err != nil {
		return result, err
	}

	result.ExpirationTime = expiry
	result.RDAPSuccess = false
	result.WHOISSuccess = true
	result.LookupMethod = "whois"

	return result, nil
}

func parseWhoisExpiration(data string) (time.Time, error) {

	patterns := []*regexp.Regexp{

		regexp.MustCompile(`(?im)^Registry Expiry Date:\s*(.+)$`),

		regexp.MustCompile(`(?im)^Registrar Registration Expiration Date:\s*(.+)$`),

		regexp.MustCompile(`(?im)^Expiration Date:\s*(.+)$`),

		regexp.MustCompile(`(?im)^Expiry Date:\s*(.+)$`),

		regexp.MustCompile(`(?im)^Expires On:\s*(.+)$`),

		regexp.MustCompile(`(?im)^expires:\s*(.+)$`),

		regexp.MustCompile(`(?im)^renewal date:\s*(.+)$`),

		regexp.MustCompile(`(?im)^paid-till:\s*(.+)$`),

		regexp.MustCompile(`(?im)^expire:\s*(.+)$`),
	}

	for _, re := range patterns {

		match := re.FindStringSubmatch(data)

		if len(match) != 2 {
			continue
		}

		value := strings.TrimSpace(match[1])

		if t, err := parseDate(value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("expiration date not found")
}

func parseDate(value string) (time.Time, error) {

	layouts := []string{

		time.RFC3339,

		"2006-01-02",

		"2006-01-02 15:04:05",

		"2006-01-02T15:04:05",

		"2006-01-02T15:04:05Z",

		"2006-01-02T15:04:05.000Z",

		"2006.01.02",

		"02-Jan-2006",

		"02-Jan-2006 15:04:05 UTC",

		"2006/01/02",

		"2006/01/02 15:04:05",

		"02/01/2006",
	}

	for _, layout := range layouts {

		t, err := time.Parse(layout, value)

		if err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %s", value)
}
