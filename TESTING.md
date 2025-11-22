# Testing

This document describes the testing policies and infrastructure for jotr.

## Overview

jotr follows testing patterns established by major Go CLI projects like docker/cli, kubernetes, and gh. All code changes should include appropriate test coverage.

## Unit Tests

All code changes should have unit test coverage. Tests should be:

- **Focused**: Test one thing at a time
- **Fast**: Run quickly to encourage frequent execution  
- **Isolated**: No dependencies on external state
- **Deterministic**: Same input produces same output

### Test Organization

Tests are located alongside source files using the `*_test.go` naming convention:

```
cmd/
├── capture.go
├── capture_test.go
├── daily.go
└── daily_test.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./cmd

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```

## Integration Tests

Integration tests verify end-to-end functionality using the CLI testing helpers in `internal/testhelpers/`.

### CLI Testing

Use the provided helpers for testing CLI commands:

```go
func TestCaptureCommand(t *testing.T) {
    fs := testhelpers.NewTestFS(t)
    defer fs.Cleanup()
    
    configHelper := testhelpers.NewConfigHelper(fs)
    configHelper.CreateBasicConfig(t)
    
    rootCmd := createTestRootCommand()
    result := testhelpers.ExecuteCommand(rootCmd, "capture", "test text")
    
    result.AssertSuccess(t)
    result.AssertStdoutContains(t, "Captured")
}
```

### Table-Driven Tests

Use structured test data for comprehensive coverage:

```go
tests := []testhelpers.NamedTest[TestData]{
    {Name: "success_case", Data: TestData{...}},
    {Name: "error_case", Data: TestData{...}},
}

testhelpers.RunNamedTableTests(t, tests, func(t *testing.T, tt TestData) {
    // Test implementation
})
```

## Testing Utilities

The `internal/testhelpers/` package provides utilities for testing:

### CLI Testing (`testhelpers/cli.go`)
- `ExecuteCommand()` - Execute commands with captured output
- `ExecuteCommandWithInput()` - Test interactive commands  
- `CLIResult` with assertion methods
- Environment variable isolation

### Filesystem Testing (`testhelpers/filesystem.go`)
- `TestFS` - Isolated temporary filesystem
- `ConfigHelper` - Config setup for tests
- `MockTime` - Time-dependent testing
- File assertion methods

### Example Usage

```go
func TestFileOperations(t *testing.T) {
    fs := testhelpers.NewTestFS(t)
    defer fs.Cleanup()
    
    fs.WriteFile(t, "test.md", "# Test")
    fs.AssertFileExists(t, "test.md")
    fs.AssertFileContains(t, "test.md", "# Test")
}
```

## Test Categories

### Unit Tests
- Test individual functions and methods
- Use mocks for external dependencies
- Fast execution, no I/O when possible

### Integration Tests  
- Test command execution end-to-end
- Use temporary filesystems
- Verify file system changes

### Error Tests
- Test failure scenarios
- Validate error messages
- Ensure graceful degradation

## Guidelines

### Do
- Write tests for new features
- Test error conditions
- Use table-driven tests for multiple scenarios
- Keep tests isolated and deterministic
- Use descriptive test names

### Don't  
- Test external dependencies directly
- Use global state in tests
- Write flaky tests
- Skip testing error paths
- Use real user data in tests

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output  
go test -v ./...

# Run specific package
go test ./cmd

# Run parallel tests only
go test -run TestParallel ./...

# Run benchmarks
go test -bench=. ./...

# Run with race detection
go test -race ./...

# Run short tests only
go test -short ./...

# Generate coverage report
go test -cover ./...
```

## Contributing

When adding new features:

1. **Write tests first** - Follow TDD when possible
2. **Test error cases** - Don't just test the happy path  
3. **Use table-driven tests** - For multiple similar scenarios
4. **Keep tests fast** - Use mocks to avoid slow I/O
5. **Test in isolation** - Each test should be independent

See [CONTRIBUTING.md](CONTRIBUTING.md) for general contribution guidelines.