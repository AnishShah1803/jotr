#!/bin/bash

# jotr Test Runner
# Runs all automated tests for the jotr project

set -e

# Cleanup function for coverage files
cleanup() {
    echo "🧹 Cleaning up temporary files..."
    rm -f coverage.out 2>/dev/null || true
    echo "✨ Cleanup completed"
}

# Set trap to ensure cleanup happens even if script exits early
trap cleanup EXIT

echo "🧪 Running jotr Test Suite"
echo "=========================="

echo
echo "📦 Testing internal packages..."
go test ./internal/... -v -cover

echo
echo "🔧 Testing with race detection..."
go test ./internal/... -race

echo
echo "📊 Generating coverage report..."
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

echo
echo "✅ All tests completed!"
echo "📄 Coverage report saved to: coverage.html"
echo "💡 Open coverage.html in your browser to view detailed coverage"
echo
echo "Test Coverage Summary:"
go test ./internal/... -cover | grep "coverage:"