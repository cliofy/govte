.PHONY: help test bench coverage lint fmt clean install release

# Default target
help:
	@echo "Available targets:"
	@echo "  test      - Run all tests"
	@echo "  bench     - Run benchmarks"
	@echo "  coverage  - Generate test coverage report"
	@echo "  lint      - Run golangci-lint"
	@echo "  fmt       - Format code with gofmt"
	@echo "  clean     - Remove generated files"
	@echo "  install   - Install the package"
	@echo "  release   - Create a new release (requires VERSION)"

# Run tests
test:
	go test -v -race ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Clean generated files
clean:
	rm -f coverage.out coverage.html
	rm -f examples/*/capture_tui examples/*/animated_progress examples/*/vte_animation
	find . -type f -name "*.test" -delete

# Install package
install:
	go install ./...

# Create a release (usage: make release VERSION=0.2.0)
release:
ifndef VERSION
	$(error VERSION is not set. Usage: make release VERSION=0.2.0)
endif
	@echo "Creating release v$(VERSION)..."
	@echo "1. Updating CHANGELOG.md for version $(VERSION)"
	@echo "2. Committing changes..."
	git add -A
	git commit -m "Release v$(VERSION)"
	@echo "3. Creating tag v$(VERSION)..."
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo "4. Pushing to origin..."
	git push origin main
	git push origin v$(VERSION)
	@echo "Release v$(VERSION) created successfully!"
	@echo "GitHub Actions will automatically create the release."