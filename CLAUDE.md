# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a comprehensive document generation microservice built in Go that supports multiple output formats: PDF, CSV, and Excel. It provides HTTP APIs for converting data to various document formats, designed for integration with B2B POS systems.

## Development Commands

### Go Commands
```bash
# Build the application
go build -o document-microservice

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
go build -ldflags="-s -w" -o document-microservice
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

# Test CSV generation
curl -X POST -d 'data=[["John","25"],["Jane","30"]]&headers=["Name","Age"]' http://localhost:8080/csv > output.csv

# Test Excel generation
curl -X POST -d 'data=[["John","25"],["Jane","30"]]&headers=["Name","Age"]&filename=report.xlsx' http://localhost:8080/excel > output.xlsx

# Test with API key authentication
curl -X POST -H "X-API-Key: your-secret-key" -d "html=<h1>Hello World</h1>" http://localhost:8080/pdf > output.pdf
```

## Architecture

### Core Components
- **Multi-format document service**: Supports PDF, CSV, and Excel generation
- **Modular architecture**: Separated into main.go, types.go, csv.go, excel.go, utils.go
- **ChromeDP integration**: Uses headless Chrome for HTML to PDF conversion
- **Excel library**: Uses excelize/v2 for Excel file generation
- **Stateless design**: No persistent storage, each request is independent

### API Endpoints
- `POST /pdf`: Accepts HTML content via form data, returns PDF
- `POST /csv`: Accepts JSON data array with headers, returns CSV file
- `POST /excel`: Accepts JSON data with optional styling, returns Excel file
- `GET /health`: Returns JSON health status with service availability

### Key Dependencies
- **chromedp/chromedp**: Headless Chrome automation for PDF generation
- **chromedp/cdproto**: Chrome DevTools Protocol bindings
- **github.com/xuri/excelize/v2**: Excel file generation and styling
- **encoding/csv**: Built-in CSV generation
- Standard Go HTTP server and JSON handling

## Technical Details

### Document Generation Processes

#### PDF Generation
1. Receives HTML content via POST request
2. Creates new Chrome context for isolation
3. Navigates to data URL with HTML content
4. Configures PDF printing parameters (A4, margins)
5. Generates PDF buffer using Chrome's print functionality
6. Returns PDF as HTTP response

#### CSV Generation
1. Receives JSON data array and headers via POST request
2. Validates data structure and headers
3. Creates CSV writer with configurable delimiter
4. Streams data row by row to output buffer
5. Returns CSV file with proper content headers

#### Excel Generation
1. Receives JSON data, headers, and optional styling via POST request
2. Creates new Excel workbook using excelize library
3. Applies header styling (bold, colors, fonts)
4. Populates data rows with type conversion
5. Applies data styling and auto-sizing if requested
6. Returns Excel file (.xlsx format)

### Configuration

Environment variables and command-line options:
- **PORT**: Server port (default: 8080) - can be set via environment variable or `-port` flag
- **API_KEY**: Optional API key for authentication - all endpoints require this header if set
- **DISPLAY**: Chrome display setting for headless operation (default: :99)

Document format settings:
- **PDF Settings**: A4 paper size with 0.4 inch margins, background printing enabled
- **CSV Settings**: Configurable delimiter (default: comma), UTF-8 encoding
- **Excel Settings**: .xlsx format, auto-sizing, custom styling support

Configuration files:
- `.env`: Environment variables (copy from `.env.example`)
- Command-line flags: `-port` to override default port
- Performance options: Buffer sizes, chunk processing, memory limits

### Error Handling

- **Method validation**: Only POST requests accepted for generation endpoints
- **Input validation**: JSON data format, headers structure, parameter validation
- **Authentication**: API key validation when configured
- **File security**: Filename sanitization, path traversal prevention
- **Processing errors**: ChromeDP, CSV writer, and Excel generation error handling
- **Structured responses**: JSON error responses with error codes and details
- **HTTP status codes**: 200, 400, 401, 413, 422, 500, 503

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
scp -r . user@server:/opt/document-microservice/
ssh user@server

# Build the application
cd /opt/document-microservice
go mod download
go build -ldflags="-s -w" -o document-microservice

# Run directly (for testing)
./document-microservice

# Run in background
nohup ./document-microservice > app.log 2>&1 &
```

### Systemd Service Setup

```bash
# Create systemd service file
sudo tee /etc/systemd/system/document-microservice.service > /dev/null <<EOF
[Unit]
Description=Document Generation Microservice
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/document-microservice
ExecStart=/opt/document-microservice/document-microservice
Restart=always
RestartSec=5
Environment=DISPLAY=:99

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable document-microservice
sudo systemctl start document-microservice

# Check service status
sudo systemctl status document-microservice

# View logs
sudo journalctl -u document-microservice -f
```

### Nginx Reverse Proxy

```bash
# Install Nginx
sudo apt install -y nginx

# Create Nginx configuration
sudo tee /etc/nginx/sites-available/document-microservice > /dev/null <<EOF
server {
    listen 80;
    server_name your-domain.com;  # Replace with your domain

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Increase timeout for document generation
        proxy_read_timeout 60s;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        
        # Handle large document responses
        client_max_body_size 50M;
    }
}
EOF

# Enable the site
sudo ln -s /etc/nginx/sites-available/document-microservice /etc/nginx/sites-enabled/
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
docker-compose logs -f document-microservice

# Stop services
docker-compose down
```

### Docker Commands

```bash
# Build manually
docker build -t document-microservice .

# Run with environment variables
docker run -d -p 8080:8080 -e PORT=8080 -e API_KEY=your-key document-microservice

# Run with custom port
docker run -d -p 9000:9000 -e PORT=9000 document-microservice
```

### Monitoring and Maintenance

```bash
# Monitor service logs
sudo journalctl -u document-microservice -f

# Check service performance
htop
netstat -tlnp | grep 8080

# Test deployment
curl http://localhost:8080/health
curl -X POST -d "html=<h1>Test</h1>" http://localhost:8080/pdf > test.pdf
curl -X POST -d 'data=[["Test","Data"]]&headers=["Column1","Column2"]' http://localhost:8080/csv > test.csv
curl -X POST -d 'data=[["Test","Data"]]&headers=["Column1","Column2"]' http://localhost:8080/excel > test.xlsx

# Update deployment
sudo systemctl stop document-microservice
# Replace binary
sudo systemctl start document-microservice
```