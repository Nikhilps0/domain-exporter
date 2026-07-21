package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	mutex sync.RWMutex

	// -----------------------------
	// Domain Expiration
	// -----------------------------

	DaysUntilExpiry *prometheus.GaugeVec
	ExpiryTimestamp *prometheus.GaugeVec
	LastRefresh     *prometheus.GaugeVec

	// -----------------------------
	// Lookup
	// -----------------------------

	RDAPSuccess      *prometheus.GaugeVec
	RDAPDuration     *prometheus.GaugeVec
	WHOISSuccess     *prometheus.GaugeVec
	WHOISDuration    *prometheus.GaugeVec
	LookupMethod     *prometheus.GaugeVec

	// -----------------------------
	// TLS
	// -----------------------------

	TLSSuccess      *prometheus.GaugeVec
	TLSDuration     *prometheus.GaugeVec
	TLSExpiry       *prometheus.GaugeVec
	TLSDaysRemain   *prometheus.GaugeVec

	// -----------------------------
	// DNSSEC
	// -----------------------------

	DNSSECEnabled *prometheus.GaugeVec
}

func NewMetrics() *Metrics {

	m := &Metrics{

		DaysUntilExpiry: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_days_until_expiry",
				Help: "Days until domain expiration.",
			},
			[]string{"domain"},
		),

		ExpiryTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_expiration_timestamp",
				Help: "Domain expiration timestamp.",
			},
			[]string{"domain"},
		),

		LastRefresh: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_last_refresh",
				Help: "Last refresh unix timestamp.",
			},
			[]string{"domain"},
		),

		RDAPSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_rdap_success",
				Help: "RDAP lookup success.",
			},
			[]string{"domain"},
		),

		RDAPDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_rdap_response_seconds",
				Help: "RDAP lookup duration.",
			},
			[]string{"domain"},
		),

		WHOISSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_whois_success",
				Help: "WHOIS lookup success.",
			},
			[]string{"domain"},
		),

		WHOISDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_whois_response_seconds",
				Help: "WHOIS lookup duration.",
			},
			[]string{"domain"},
		),

		LookupMethod: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_lookup_method",
				Help: "0=RDAP 1=WHOIS",
			},
			[]string{"domain"},
		),

		TLSSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_tls_lookup_success",
				Help: "TLS lookup success.",
			},
			[]string{"domain"},
		),

		TLSDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_tls_response_seconds",
				Help: "TLS lookup duration.",
			},
			[]string{"domain"},
		),

		TLSExpiry: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_tls_expiration_timestamp",
				Help: "TLS certificate expiration timestamp.",
			},
			[]string{"domain"},
		),

		TLSDaysRemain: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_tls_days_remaining",
				Help: "Days remaining before TLS certificate expires.",
			},
			[]string{"domain"},
		),

		DNSSECEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_dnssec_enabled",
				Help: "DNSSEC enabled.",
			},
			[]string{"domain"},
		),
	}

	prometheus.MustRegister(
		m.DaysUntilExpiry,
		m.ExpiryTimestamp,
		m.LastRefresh,

		m.RDAPSuccess,
		m.RDAPDuration,

		m.WHOISSuccess,
		m.WHOISDuration,
		m.LookupMethod,

		m.TLSSuccess,
		m.TLSDuration,
		m.TLSExpiry,
		m.TLSDaysRemain,

		m.DNSSECEnabled,
	)

	return m
}

func (m *Metrics) Update(r DomainResult) {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	domain := r.Domain

	// -----------------------------
	// Expiration
	// -----------------------------

	if !r.ExpirationTime.IsZero() {

		m.ExpiryTimestamp.
			WithLabelValues(domain).
			Set(float64(r.ExpirationTime.Unix()))

		m.DaysUntilExpiry.
			WithLabelValues(domain).
			Set(float64(r.DaysUntilExpiry()))
	}

	m.LastRefresh.
		WithLabelValues(domain).
		Set(float64(r.CollectedAt.Unix()))

	// -----------------------------
	// RDAP
	// -----------------------------

	if r.RDAPSuccess {
		m.RDAPSuccess.WithLabelValues(domain).Set(1)
	} else {
		m.RDAPSuccess.WithLabelValues(domain).Set(0)
	}

	m.RDAPDuration.
		WithLabelValues(domain).
		Set(r.RDAPDuration.Seconds())

	// -----------------------------
	// WHOIS
	// -----------------------------

	if r.WHOISSuccess {
		m.WHOISSuccess.WithLabelValues(domain).Set(1)
	} else {
		m.WHOISSuccess.WithLabelValues(domain).Set(0)
	}

	m.WHOISDuration.
		WithLabelValues(domain).
		Set(r.WHOISDuration.Seconds())

	// -----------------------------
	// Lookup Method
	// -----------------------------

	switch r.LookupMethod {

	case "rdap":
		m.LookupMethod.WithLabelValues(domain).Set(0)

	case "whois":
		m.LookupMethod.WithLabelValues(domain).Set(1)

	default:
		m.LookupMethod.WithLabelValues(domain).Set(-1)
	}

	// -----------------------------
	// TLS
	// -----------------------------

	if r.TLSSuccess {

		m.TLSSuccess.
			WithLabelValues(domain).
			Set(1)

		m.TLSExpiry.
			WithLabelValues(domain).
			Set(float64(r.TLSExpiryTime.Unix()))

		m.TLSDaysRemain.
			WithLabelValues(domain).
			Set(float64(r.TLSDaysRemaining))

	} else {

		m.TLSSuccess.
			WithLabelValues(domain).
			Set(0)
	}

	m.TLSDuration.
		WithLabelValues(domain).
		Set(r.TLSDuration.Seconds())

	// -----------------------------
	// DNSSEC
	// -----------------------------

	if r.DNSSECEnabled {
		m.DNSSECEnabled.WithLabelValues(domain).Set(1)
	} else {
		m.DNSSECEnabled.WithLabelValues(domain).Set(0)
	}
}
