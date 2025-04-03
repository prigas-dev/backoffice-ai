# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands
- Build: `go build`
- Run: `go run main.go`
- Test: `go test ./...`
- Test single file: `go test ./path/to/package -run TestName`
- Lint: `golangci-lint run`
- Format: `gofmt -w .`

## Package Management
- IMPORTANT: First check go.mod before suggesting package installation
- Only suggest installing packages not already in go.mod
- Install packages: `go get package-name` (e.g., `go get github.com/new/package`)
- Always use `go get` commands to install new dependencies instead of directly editing go.mod
- Update dependencies: `go get -u package-name`
- Tidy up dependencies: `go mod tidy`

## Code Style Guidelines
- Follow standard Go naming conventions (CamelCase for exported, camelCase for non-exported)
- Use package anthropic-sdk-go for Claude API integration
- Handle errors explicitly with if err != nil pattern
- Use godotenv for environment variables
- Group imports: standard library first, then third-party
- Use context for API calls
- Indent with tabs, not spaces
- Follow Go's error handling pattern instead of using exceptions
- Document exported functions and types using godoc style comments
- Keep functions concise and focused on a single responsibility