# Contributing to GoVTE

Thank you for your interest in contributing to GoVTE! We welcome contributions from the community and are grateful for any help you can provide.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct. Please be respectful and considerate in all interactions.

## How to Contribute

### Reporting Issues

- Check if the issue has already been reported
- Provide a clear description of the problem
- Include steps to reproduce the issue
- Share relevant code snippets or error messages
- Mention your Go version and operating system

### Submitting Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Write clear commit messages** following conventional commits format
3. **Add tests** for any new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass** by running `go test ./...`
6. **Run benchmarks** if you've made performance changes
7. **Submit a pull request** with a clear description of changes

### Development Setup

```bash
# Clone your fork
git clone https://github.com/your-username/govte.git
cd govte

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# Check test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to catch common issues
- Consider using `golangci-lint` for comprehensive linting
- Write clear, self-documenting code with minimal comments
- Add godoc comments for all exported types and functions

### Testing Guidelines

- Write unit tests for new functionality
- Maintain or improve test coverage
- Include both positive and negative test cases
- Use table-driven tests where appropriate
- Add benchmarks for performance-critical code

### Documentation

- Update README.md if you change user-facing functionality
- Add godoc comments for exported APIs
- Include examples in documentation where helpful
- Update CHANGELOG.md for significant changes

## Areas for Contribution

### High Priority

- Additional terminal emulation features
- Performance optimizations
- Cross-platform compatibility improvements
- More comprehensive examples

### Good First Issues

Look for issues labeled `good first issue` for beginner-friendly tasks.

### Feature Requests

We welcome feature requests! Please open an issue to discuss before implementing major changes.

## Questions?

Feel free to open an issue for any questions about contributing. We're here to help!

## License

By contributing to GoVTE, you agree that your contributions will be licensed under the MIT License.