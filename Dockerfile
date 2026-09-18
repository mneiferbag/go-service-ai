# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency specifications
COPY go.mod ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# Final stage
FROM alpine:3.21

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Run as non-root user for security
RUN adduser -D -u 10001 appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]
