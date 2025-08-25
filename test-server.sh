#!/bin/bash

echo "Testing server startup..."

# Make sure we have the latest builds
make wasm
make server

echo "Starting server..."
./bin/server -dev -port 8080 &
SERVER_PID=$!

echo "Server PID: $SERVER_PID"

# Wait a moment for server to start
sleep 2

echo "Testing if server responds..."
if curl -s http://localhost:8080/ > /dev/null; then
    echo "✅ Server is responding!"
    echo "🌐 Visit: http://localhost:8080"
else
    echo "❌ Server is not responding"
fi

echo "Checking what's listening on port 8080:"
lsof -i :8080

echo "To stop the server manually: kill $SERVER_PID"
echo "Or press Ctrl+C"

# Keep script running
wait $SERVER_PID