#!/bin/bash

# Quick test script for cloud database connection
echo "🧪 Testing Smart Payment Infrastructure with Cloud Database"
echo "==========================================================="

# Test 1: Database connection
echo ""
echo "1️⃣ Testing database connection..."
if go run cmd/db-migrate/main.go -action=version >/dev/null 2>&1; then
    echo "✅ Database connection: PASS"
else
    echo "❌ Database connection: FAIL"
    echo "Please run ./scripts/setup-cloud-db.sh to configure your database"
    exit 1
fi

# Test 2: Run unit tests
echo ""
echo "2️⃣ Running unit tests..."
if go test ./internal/services ./internal/models ./pkg/auth ./pkg/xrpl -v >/dev/null 2>&1; then
    echo "✅ Unit tests: PASS"
else
    echo "❌ Unit tests: FAIL"
    echo "Running unit tests with verbose output:"
    go test ./internal/services ./internal/models ./pkg/auth ./pkg/xrpl -v
    exit 1
fi

# Test 3: Run integration tests
echo ""
echo "3️⃣ Running integration tests..."
if go test ./test/integration -v; then
    echo "✅ Integration tests: PASS"
else
    echo "❌ Integration tests: FAIL"
    echo "Check the output above for details"
    exit 1
fi

# Test 4: Test API server startup
echo ""
echo "4️⃣ Testing API server startup..."
timeout 10s go run cmd/identity-service/main.go >/dev/null 2>&1 &
SERVER_PID=$!
sleep 3

if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ API server startup: PASS"
    kill $SERVER_PID 2>/dev/null
else
    echo "❌ API server startup: FAIL"
fi

echo ""
echo "🎉 All tests completed!"
echo ""
echo "Your Smart Payment Infrastructure is ready to use!"
echo ""
echo "To start the services:"
echo "• Identity Service: go run cmd/identity-service/main.go"
echo "• XRPL Service: go run cmd/xrpl-service/main.go"
echo ""
echo "API will be available at: http://localhost:8080"