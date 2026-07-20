package collector

import "time"

// DomainResult contains the complete collection result for a single domain.
// It is passed between the RDAP, TLS, DNSSEC collectors and finally used by
// the metrics collector.
type DomainResult struct {
	Domain string

	CollectedAt time.Time

	// -------------------------
	// RDAP
	// -------------------------

	RDAPSuccess    bool
	RDAPDuration   time.Duration
	ExpirationTime time.Time

	// -------------------------
	// TLS
	// -------------------------

	TLSSuccess       bool
	TLSDuration      time.Duration
	TLSExpiryTime    time.Time
	TLSDaysRemaining int

	// -------------------------
	// DNSSEC
	// -------------------------

	DNSSECEnabled bool

	// -------------------------
	// Errors
	// -------------------------

	Error string
}

// DomainJob represents one domain scheduled for collection.
type DomainJob struct {
	Domain string
}

// DaysUntilExpiry returns the number of whole days remaining until the
// domain expires. If the expiration date is unavailable, -1 is returned.
func (d DomainResult) DaysUntilExpiry() int {
	if d.ExpirationTime.IsZero() {
		return -1
	}

	return int(time.Until(d.ExpirationTime).Hours() / 24)
}

// Expired reports whether the domain has already expired.
func (d DomainResult) Expired() bool {
	if d.ExpirationTime.IsZero() {
		return false
	}

	return time.Now().After(d.ExpirationTime)
}

// HasTLS reports whether a valid TLS certificate was collected.
func (d DomainResult) HasTLS() bool {
	return d.TLSSuccess && !d.TLSExpiryTime.IsZero()
}

// HasRDAP reports whether RDAP lookup succeeded.
func (d DomainResult) HasRDAP() bool {
	return d.RDAPSuccess && !d.ExpirationTime.IsZero()
}
