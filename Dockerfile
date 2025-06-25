FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o document-microservice .

# Final stage with Chrome
FROM chromedp/headless-shell:latest

# Install certificates for HTTPS requests
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy binary and startup script from builder
COPY --from=builder /app/document-microservice .
COPY docker-start.sh .

# Create non-root user and set permissions
RUN useradd -r -s /bin/false appuser && \
    chown appuser:appuser /app/document-microservice && \
    chmod +x /app/docker-start.sh && \
    chown appuser:appuser /app/docker-start.sh && \
    mkdir -p /home/appuser/.cache/fontconfig && \
    mkdir -p /home/appuser/.fontconfig && \
    chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Override the entrypoint and run our script
ENTRYPOINT []
CMD ["/bin/bash", "/app/docker-start.sh"]