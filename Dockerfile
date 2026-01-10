# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /k8s-custom-controller \
    ./main.go

# Final stage - minimal runtime image
FROM scratch

# Copy CA certificates from alpine for TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /k8s-custom-controller /k8s-custom-controller

# Use non-root user (UID 65534 = nobody)
USER 65534:65534

# Expose metrics port (if needed)
EXPOSE 8080

# Run the controller
ENTRYPOINT ["/k8s-custom-controller"]

