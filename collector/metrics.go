package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	DaysUntilExpiry *prometheus.GaugeVec
	ExpiryTimestamp *prometheus.GaugeVec

	RDAPSuccess  *prometheus.GaugeVec
	RDAPDuration *prometheus.GaugeVec

	TLSSuccess     *prometheus.GaugeVec
	TLSDuration    *prometheus.GaugeVec
	TLSExpiry      *prometheus.GaugeVec
	TLSDaysRemain  *prometheus.GaugeVec

	DNSSECEnabled *prometheus.GaugeVec

	LastRefresh *prometheus.GaugeVec

	mutex sync.RWMutex
}

func NewMetrics() *Metrics {
	m := &Metrics{
		DaysUntilExpiry: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_days_until_expiry",
				Help: "Days remaining until domain expiration.",
			},
			[]string{"domain"},
		),

		ExpiryTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_expiration_timestamp",
				Help: "Unix timestamp of domain expiration.",
			},
			[]string{"domain"},
		),

		RDAPSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_rdap_success",
				Help: "RDAP lookup success (1=success,0=failure).",
			},
			[]string{"domain"},
		),

		RDAPDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_rdap_response_seconds",
				Help: "RDAP lookup duration in seconds.",
			},
			[]string{"domain"},
		),

		TLSSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_tls_lookup_success",
				Help: "TLS lookup success (1=success,0=failure).",
			},
			[]string{"domain"},
		),

		TLSDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_tls_response_seconds",
				Help: "TLS lookup duration in seconds.",
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
				Help: "DNSSEC enabled (1=yes,0=no).",
			},
			[]string{"domain"},
		),

		LastRefresh: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "domain_last_refresh",
				Help: "Unix timestamp of the last successful refresh.",
			},
			[]string{"domain"},
		),
	}

	prometheus.MustRegister(
		m.DaysUntilExpiry,
		m.ExpiryTimestamp,
		m.RDAPSuccess,
		m.RDAPDuration,
		m.TLSSuccess,
		m.TLSDuration,
		m.TLSExpiry,
		m.TLSDaysRemain,
		m.DNSSECEnabled,
		m.LastRefresh,
	)

	return m
}

func (m *Metrics) Update(result DomainResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	domain := result.Domain

	if result.RDAPSuccess {
		m.RDAPSuccess.WithLabelValues(domain).Set(1)
		m.DaysUntilExpiry.WithLabelValues(domain).Set(float64(result.DaysUntilExpiry()))
		m.ExpiryTimestamp.WithLabelValues(domain).Set(float64(result.ExpirationTime.Unix()))
	} else {
		m.RDAPSuccess.WithLabelValues(domain).Set(0)
	}

	m.RDAPDuration.WithLabelValues(domain).Set(result.RDAPDuration.Seconds())

	if result.TLSSuccess {
		m.TLSSuccess.WithLabelValues(domain).Set(1)
		m.TLSExpiry.WithLabelValues(domain).Set(float64(result.TLSExpiryTime.Unix()))
		m.TLSDaysRemain.WithLabelValues(domain).Set(float64(result.TLSDaysRemaining))
	} else {
		m.TLSSuccess.WithLabelValues(domain).Set(0)
	}

	m.TLSDuration.WithLabelValues(domain).Set(result.TLSDuration.Seconds())

	if result.DNSSECEnabled {
		m.DNSSECEnabled.WithLabelValues(domain).Set(1)
	} else {
		m.DNSSECEnabled.WithLabelValues(domain).Set(0)
	}

	m.LastRefresh.WithLabelValues(domain).Set(float64(result.CollectedAt.Unix()))
}
