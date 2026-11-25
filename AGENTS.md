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

<!-- CODEGRAPH_START -->
## CodeGraph

In repositories indexed by CodeGraph (a `.codegraph/` directory exists at the repo root), reach for it BEFORE grep/find or reading files when you need to understand or locate code:

- **MCP tool** (when available): `codegraph_codegraph_explore` answers most code questions in one call — the relevant symbols' verbatim source plus the call paths between them, including dynamic-dispatch hops grep can't follow. Name a file or symbol in the query to read its current line-numbered source. If it's listed but deferred, load it by name via tool search.
- **Shell** (always works): `codegraph explore "<symbol names or question>"` prints the same output.

If there is no `.codegraph/` directory, skip CodeGraph entirely — indexing is the user's decision.

OpenCode puts the MCP server's key in front of each tool's name, so it lists this tool as `codegraph_codegraph_explore` where other agents list `codegraph_explore`; if the server is registered under another key, that key replaces the first `codegraph_`.
<!-- CODEGRAPH_END -->
