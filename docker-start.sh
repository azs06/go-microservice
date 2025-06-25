#!/bin/bash
# Start Chrome in the background
/headless-shell/headless-shell --no-sandbox --use-gl=angle --use-angle=swiftshader --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 &
# Start socat to forward Chrome debugging port
socat TCP4-LISTEN:9223,fork TCP4:127.0.0.1:9222 &
# Wait a moment for Chrome to start
sleep 2
# Start the Go application
exec ./document-microservice