# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a PDF generation microservice built in Go that converts HTML to PDF using headless Chrome via ChromeDP. The service provides a simple HTTP API for generating PDFs from HTML content.

## Development Commands

### Go Commands
```bash
# Build the application
go build -o pdf-microservice

# Run the application
go run main.go

# Install dependencies
go mod tidy

# Download dependencies
go mod download

# Test the application (when tests are added)
go test ./...

# Run with verbose output
go run -v main.go

# Build for production
go build -ldflags="-s -w" -o pdf-microservice
```

### Running the Service

```bash
# Start the service (default port 8080)
go run main.go

# Start with custom port
go run main.go -port=9000

# Start with environment variables
PORT=9000 API_KEY=your-secret-key go run main.go

# Test the health endpoint
curl http://localhost:8080/health

# Test with API key (if configured)
curl -H "X-API-Key: your-secret-key" http://localhost:8080/health

# Test PDF generation
curl -X POST -d "html=<h1>Hello World</h1>" http://localhost:8080/pdf > output.pdf

# Test PDF generation with API key
curl -X POST -H "X-API-Key: your-secret-key" -d "html=<h1>Hello World</h1>" http://localhost:8080/pdf > output.pdf
```

## Architecture

### Core Components
- **Single-file microservice**: Entire application contained in `main.go`
- **HTTP handlers**: Two endpoints for PDF generation and health checks
- **ChromeDP integration**: Uses headless Chrome for HTML to PDF conversion
- **Stateless design**: No persistent storage, each request is independent

### API Endpoints
- `POST /pdf`: Accepts HTML content via form data, returns PDF
- `GET /health`: Returns JSON health status

### Key Dependencies
- **chromedp/chromedp**: Headless Chrome automation
- **chromedp/cdproto**: Chrome DevTools Protocol bindings
- Standard Go HTTP server

## Technical Details

### PDF Generation Process
1. Receives HTML content via POST request
2. Creates new Chrome context for isolation
3. Navigates to data URL with HTML content
4. Configures PDF printing parameters (A4, margins)
5. Generates PDF buffer using Chrome's print functionality
6. Returns PDF as HTTP response

### Configuration

Environment variables and command-line options:
- **PORT**: Server port (default: 8080) - can be set via environment variable or `-port` flag
- **API_KEY**: Optional API key for authentication - all endpoints require this header if set
- **PDF Settings**: A4 paper size with 0.4 inch margins
- **Chrome Options**: Print background enabled by default

Configuration files:
- `.env`: Environment variables (copy from `.env.example`)
- Command-line flags: `-port` to override default port

### Error Handling

- Method validation for POST requests
- HTML parameter validation
- ChromeDP execution error handling
- HTTP error responses with appropriate status codes

## Ubuntu Server Deployment

### Prerequisites

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Go (if not already installed)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Chrome dependencies for headless operation
sudo apt install -y wget gnupg ca-certificates
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
sudo apt update
sudo apt install -y google-chrome-stable

# Install additional dependencies for headless Chrome
sudo apt install -y xvfb
```

### Manual Deployment

```bash
# Clone/copy your code to server
scp -r . user@server:/opt/pdf-microservice/
ssh user@server

# Build the application
cd /opt/pdf-microservice
go mod download
go build -ldflags="-s -w" -o pdf-microservice

# Run directly (for testing)
./pdf-microservice

# Run in background
nohup ./pdf-microservice > app.log 2>&1 &
```

### Systemd Service Setup

```bash
# Create systemd service file
sudo tee /etc/systemd/system/pdf-microservice.service > /dev/null <<EOF
[Unit]
Description=PDF Generation Microservice
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/pdf-microservice
ExecStart=/opt/pdf-microservice/pdf-microservice
Restart=always
RestartSec=5
Environment=DISPLAY=:99

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable pdf-microservice
sudo systemctl start pdf-microservice

# Check service status
sudo systemctl status pdf-microservice

# View logs
sudo journalctl -u pdf-microservice -f
```

### Nginx Reverse Proxy

```bash
# Install Nginx
sudo apt install -y nginx

# Create Nginx configuration
sudo tee /etc/nginx/sites-available/pdf-microservice > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;  # Replace with your domain

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Increase timeout for PDF generation
        proxy_read_timeout 60s;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        
        # Handle large PDF responses
        client_max_body_size 50M;
    }
}
EOF

# Enable the site
sudo ln -s /etc/nginx/sites-available/pdf-microservice /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### Docker Deployment

```bash
# Copy environment configuration
cp .env.example .env
# Edit .env with your configuration

# Build and run with Docker Compose
docker-compose up -d

# Build and run with custom port
PORT=9000 API_KEY=your-secret-key docker-compose up -d

# Run with Nginx reverse proxy
docker-compose --profile with-nginx up -d

# View logs
docker-compose logs -f pdf-microservice

# Stop services
docker-compose down
```

### Docker Commands

```bash
# Build manually
docker build -t pdf-microservice .

# Run with environment variables
docker run -d -p 8080:8080 -e PORT=8080 -e API_KEY=your-key pdf-microservice

# Run with custom port
docker run -d -p 9000:9000 -e PORT=9000 pdf-microservice
```

### Monitoring and Maintenance

```bash
# Monitor service logs
sudo journalctl -u pdf-microservice -f

# Check service performance
htop
netstat -tlnp | grep 8080

# Test deployment
curl http://localhost:8080/health
curl -X POST -d "html=<h1>Test</h1>" http://localhost:8080/pdf > test.pdf

# Update deployment
sudo systemctl stop pdf-microservice
# Replace binary
sudo systemctl start pdf-microservice
```