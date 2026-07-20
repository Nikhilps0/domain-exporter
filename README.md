
# Domain Exporter

A Prometheus exporter that monitors domain health using **RDAP**, **TLS**, and **DNSSEC**.

## Features

- ✅ RDAP domain expiration lookup
- ✅ TLS certificate expiration
- ✅ DNSSEC detection
- ✅ Prometheus metrics
- ✅ Worker pool
- ✅ Graceful shutdown
- ✅ Structured logging (`log/slog`)
- ✅ Configurable refresh interval
- ✅ Docker support
- ✅ GitHub Actions CI

---

## Project Structure

```
domain-exporter/
├── cmd/
│   └── main.go
├── collector/
├── config/
├── scheduler/
├── internal/
├── config.yaml
├── Dockerfile
└── go.mod
```

---

## Build

```bash
go mod tidy
go build ./...
```

Run:

```bash
./domain-exporter
```

Metrics:

```
http://localhost:9118/metrics
```

---

## Docker

Build:

```bash
docker build -t domain-exporter .
```

Run:

```bash
docker run -p 9118:9118 \
    -v $(pwd)/config.yaml:/app/config.yaml \
    domain-exporter
```

---

## Configuration

Example:

```yaml
server:
  listen_address: ":9118"

scheduler:
  refresh_interval: 6h
  worker_count: 5
  request_timeout: 10s

rdap:
  bootstrap_url: https://data.iana.org/rdap/dns.json
  max_retries: 3

tls:
  enabled: true
  port: 443

domains:
  - name: google.com
  - name: github.com
  - name: openai.com
```

---

## Exported Metrics

### RDAP

```
domain_days_until_expiry
domain_expiration_timestamp
domain_rdap_success
domain_rdap_response_seconds
```

### TLS

```
domain_tls_lookup_success
domain_tls_response_seconds
domain_tls_expiration_timestamp
domain_tls_days_remaining
```

### DNSSEC

```
domain_dnssec_enabled
```

### General

```
domain_last_refresh
```

---

## Example Prometheus Configuration

```yaml
scrape_configs:
  - job_name: domain-exporter

    static_configs:
      - targets:
          - localhost:9118
```

---

## Roadmap

- [ ] Retry with exponential backoff
- [ ] Persistent RDAP bootstrap cache
- [ ] Multiple DNS resolvers
- [ ] WHOIS fallback
- [ ] OpenMetrics support
- [ ] Unit tests
- [ ] Integration tests

---
