# uniget CLI

This repository contains a Go based project for the `uniget` CLI tool, which is used for managing and installing packages. The project includes various commands and functionalities to facilitate package management. Please follow these guidelines when contributing:

## Code Standards

### Required Before Each Commit

Run `go fmt ./...` before committing any changes to ensure proper code formatting

### Development Flow

- Build: `goreleaser build --auto-snapshot --clean --single-target`
- Test: `go test ./...`
- Lint: `golangci-lint run`

## Repository Structure

- `cmd/`: Main service entry points and executables
- `pkg/`: Core Go packages

## Key Guidelines

1. Follow Go best practices and idiomatic patterns
2. Maintain existing code structure and organization
3. Use dependency injection patterns where appropriate
4. Write unit tests for new functionality. Use table-driven unit tests when possible
5. Document public APIs and complex logic
