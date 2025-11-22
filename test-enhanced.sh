#!/bin/bash

# Enhanced jotr Test Runner
# Demonstrates comprehensive Go testing framework usage

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    echo -e "${BLUE}🧹 Cleaning up temporary files...${NC}"
    rm -f coverage.out coverage.xml 2>/dev/null || true
    echo -e "${GREEN}✨ Cleanup completed${NC}"
}

trap cleanup EXIT

echo -e "${BLUE}🧪 Enhanced jotr Test Suite${NC}"
echo "================================="

echo
echo -e "${YELLOW}📦 Running standard tests...${NC}"
go test ./internal/... -v

echo
echo -e "${YELLOW}🏁 Running table-driven and subtest patterns...${NC}"
go test ./internal/utils -v -run "TableDriven|Subtests"

echo
echo -e "${YELLOW}🔧 Testing with race detection...${NC}"
go test ./internal/... -race

echo
echo -e "${YELLOW}🔥 Running benchmarks...${NC}"
go test ./internal/utils -bench=. -benchtime=1s -run=^$

echo
echo -e "${YELLOW}📊 Generating detailed coverage report...${NC}"
go test ./internal/... -coverprofile=coverage.out -coverpkg=./internal/...
go tool cover -html=coverage.out -o coverage.html

# Generate coverage summary
echo
echo -e "${GREEN}✅ Test Results Summary:${NC}"
echo "========================="

# Count total tests
TOTAL_TESTS=$(go test ./internal/... -list=. | grep -E "^(Test|Example)" | wc -l)
echo -e "${BLUE}📋 Total Tests:${NC} $TOTAL_TESTS"

# Show coverage summary  
echo -e "${BLUE}📊 Coverage Summary:${NC}"
go test ./internal/... -cover | grep "coverage:" | while read line; do
    echo -e "  ${GREEN}✓${NC} $line"
done

# Show benchmark summary
echo -e "${BLUE}⚡ Performance Summary:${NC}"
echo "  See benchmark results above for atomic operation performance"

echo
echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo -e "${BLUE}📄 Detailed coverage report:${NC} coverage.html"
echo -e "${BLUE}💡 Open coverage.html in your browser for interactive coverage analysis${NC}"

echo
echo -e "${YELLOW}🚀 Available Test Commands:${NC}"
echo "  go test ./internal/...                    # All tests"
echo "  go test ./internal/... -v                 # Verbose output"  
echo "  go test ./internal/... -race              # Race detection"
echo "  go test ./internal/... -cover             # Coverage analysis"
echo "  go test ./internal/utils -bench=.         # Benchmarks"
echo "  go test ./internal/utils -run TableDriven # Table-driven tests"