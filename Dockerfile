FROM golang:1.23-alpine AS builder

# Install gcc for SQLite (CGO)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY *.go ./

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 go build -o apartment-notifier .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and tzdata for timezone
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/apartment-notifier /apartment-notifier

# Create data directory for SQLite
RUN mkdir -p /data

CMD ["/apartment-notifier"]
