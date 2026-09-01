# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN make build

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/goawd-server .
COPY --from=builder /app/bin/goawd-roundworm .
COPY --from=builder /app/bin/goawd-guardian .

# Copy public directory (frontend)
COPY --from=builder /app/public ./public

# Copy plugins
COPY --from=builder /app/plugins ./plugins

# Expose ports
EXPOSE 1337 8023

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:1337/api/ping || exit 1

# Entry point
ENTRYPOINT ["./goawd-server"]
CMD ["-http", "0.0.0.0:1337", "-tcp", "0.0.0.0:8023", "-storage", "memory"]
