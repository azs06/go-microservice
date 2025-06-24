#!/bin/bash

# Start script for PDF microservice
# Usage: ./start.sh [port] [api_key]

PORT=${1:-8080}
API_KEY=${2:-""}

export API_KEY="$API_KEY"

echo "Starting PDF microservice..."
echo "Port: $PORT"
if [ -n "$API_KEY" ]; then
    echo "API Key: ********"
else
    echo "API Key: Not configured (open access)"
fi

go run . -port="$PORT"