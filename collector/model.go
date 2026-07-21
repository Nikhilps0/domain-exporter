package collector

import "time"

// DomainResult contains all information collected for a domain.
type DomainResult struct {
	Domain string

	CollectedAt time.Time

	// ----------------------------------------------------
	// Lookup
	// ----------------------------------------------------

	// "rdap" or "whois"
	LookupMethod string

	// ----------------------------------------------------
	// RDAP
	// ----------------------------------------------------

	RDAPSuccess  bool
	RDAPDuration time.Duration

	// ----------------------------------------------------
	// WHOIS
	// ----------------------------------------------------

	WHOISSuccess  bool
	WHOISDuration time.Duration

	// ----------------------------------------------------
	// Expiration
	// ----------------------------------------------------

	ExpirationTime time.Time

	// ----------------------------------------------------
	// TLS
	// ----------------------------------------------------

	TLSSuccess bool

	TLSDuration time.Duration

	TLSExpiryTime time.Time

	TLSDaysRemaining int

	// ----------------------------------------------------
	// DNSSEC
	// ----------------------------------------------------

	DNSSECEnabled bool

	// ----------------------------------------------------
	// Error
	// ----------------------------------------------------

	Error string
}

// DaysUntilExpiry returns the remaining days until domain expiry.
func (d DomainResult) DaysUntilExpiry() int {

	if d.ExpirationTime.IsZero() {
		return -1
	}

	return int(time.Until(d.ExpirationTime).Hours() / 24)
}

// Expired returns true if the domain is already expired.
func (d DomainResult) Expired() bool {

	if d.ExpirationTime.IsZero() {
		return false
	}

	return time.Now().After(d.ExpirationTime)
}

// HasTLS indicates whether TLS information was successfully collected.
func (d DomainResult) HasTLS() bool {

	return d.TLSSuccess && !d.TLSExpiryTime.IsZero()
}

// HasExpiration indicates whether an expiration date is available.
func (d DomainResult) HasExpiration() bool {

	return !d.ExpirationTime.IsZero()
}

// HasRDAP indicates whether RDAP lookup succeeded.
func (d DomainResult) HasRDAP() bool {

	return d.RDAPSuccess
}

// HasWHOIS indicates whether WHOIS lookup succeeded.
func (d DomainResult) HasWHOIS() bool {

	return d.WHOISSuccess
}
