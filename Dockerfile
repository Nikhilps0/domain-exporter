# -----------------------------
# Build Stage
# -----------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o domain-exporter ./cmd

# -----------------------------
# Runtime Stage
# -----------------------------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN apk add --no-cache whois

RUN addgroup -S exporter && \
    adduser -S exporter -G exporter

WORKDIR /app

COPY --from=builder /app/domain-exporter .
COPY config.yaml .

USER exporter

EXPOSE 9118

ENTRYPOINT ["./domain-exporter"]
