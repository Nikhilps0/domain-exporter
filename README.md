
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

## Grafana Dashboard

Example:

<img width="1887" height="706" alt="image" src="https://github.com/user-attachments/assets/dbb4f00c-b9ca-4e71-ac22-a98c85059868" />

<img width="938" height="281" alt="image" src="https://github.com/user-attachments/assets/71e786e7-e026-4577-8772-45455318147c" />

<img width="1877" height="725" alt="image" src="https://github.com/user-attachments/assets/a1157fa4-fda2-4f99-9d96-7cfbf3ae79e2" />

<img width="1871" height="757" alt="image" src="https://github.com/user-attachments/assets/f397a226-2a6d-4301-9418-442abc6725af" />

You can upload  dashboard.json file to Grafana
