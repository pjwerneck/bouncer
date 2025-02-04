FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY . .

# Build optimized binary
RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s \
    -X main.version=$(git describe --tags --always) \
    -X main.commit=$(git rev-parse HEAD) \
    -X main.date=$(date -u +%Y-%m-%d)" \
    -trimpath \
    -a \
    -tags netgo,osusergo \
    -o /app/bouncer

# Create minimal production image
FROM scratch

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bouncer /bouncer

# Add metadata
LABEL org.opencontainers.image.source="https://github.com/pjwerneck/bouncer" \
      org.opencontainers.image.description="Rate limiting and synchronization service" \
      org.opencontainers.image.licenses="MIT"

# Use non-root user
USER 65532:65532

EXPOSE 5505

ENTRYPOINT ["/bouncer"]
