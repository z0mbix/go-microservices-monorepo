set shell := ["bash", "-uc"]

# Show available targets/recipes
default:
	@just --choose

# Build the binary for the current os/arch
build service:
	go build -o bin/{{service}} ./services/{{service}}/

# Configure your host to use this repo
setup:
	mise trust
	mise install
	mise ls -c

# Show git tags
tags:
	@git tag | sort -V

# Run unit tests
test:
	go test ./... -v -coverprofile=/dev/null

test-packages:
    go test ./pkg/... -v -coverprofile=/dev/null

test-services:
    go test ./services/... -v -coverprofile=/dev/null

# Build a local only, snapshot release
snapshot:
	goreleaser --snapshot --skip-publish --rm-dist --debug

# Create and publish a new release
release:
	goreleaser --rm-dist

# Show help menu
help:
	@just --list --list-prefix '  ❯ '
