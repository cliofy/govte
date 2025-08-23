# Release Checklist for GoVTE v0.2.0

## Pre-release Steps

### Code Quality
- [x] All tests pass (`go test ./...`)
- [ ] No linting errors (`golangci-lint run`)
- [ ] Code coverage is acceptable (run `make coverage`)
- [ ] All examples work correctly

### Documentation
- [x] README.md is up to date
- [x] CHANGELOG.md includes all changes for this version
- [x] Package documentation is complete
- [x] All public APIs have godoc comments
- [x] Examples demonstrate key features

### Repository Setup
- [x] .gitignore is properly configured
- [x] GitHub Actions workflows are set up
- [x] Issue and PR templates are created
- [x] Contributing guidelines are clear
- [x] License file is present

## Release Steps

1. **Final Testing**
   ```bash
   make test
   make lint
   make bench
   ```

2. **Commit All Changes**
   ```bash
   git add .
   git commit -m "Prepare for v0.2.0 release"
   ```

3. **Create and Push Tag**
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0 - Stable release with complete VTE implementation"
   git push origin main
   git push origin v0.2.0
   ```

4. **Verify GitHub Actions**
   - Check that CI workflow runs successfully
   - Ensure release workflow creates GitHub release

5. **Update pkg.go.dev**
   ```bash
   # Force pkg.go.dev to update (optional)
   curl https://proxy.golang.org/github.com/cliofy/govte/@v/v0.2.0.info
   ```

6. **Post-release Verification**
   - [ ] GitHub release is created with changelog
   - [ ] pkg.go.dev shows the new version
   - [ ] Installation works: `go get github.com/cliofy/govte@v0.2.0`
   - [ ] CI badges show passing status

## Announcement Template

```markdown
🎉 GoVTE v0.2.0 Released!

We're excited to announce the stable release of GoVTE v0.2.0, a high-performance VTE parser for Go.

## Highlights
✅ Complete ANSI escape sequence support
✅ Full terminal emulation with buffer management
✅ 24-bit color support
✅ Unicode/UTF-8 handling
✅ Production-ready with comprehensive tests

## Installation
go get github.com/cliofy/govte@v0.2.0

## Links
📦 Package: https://pkg.go.dev/github.com/cliofy/govte
📝 Release: https://github.com/cliofy/govte/releases/tag/v0.2.0
📚 Docs: https://github.com/cliofy/govte#readme

Thanks to all contributors! 🙏
```

## Troubleshooting

If the release doesn't appear on pkg.go.dev:
1. Wait 5-10 minutes for automatic indexing
2. Visit https://pkg.go.dev/github.com/cliofy/govte@v0.2.0
3. Use the "Request" button if needed

If GitHub Actions fail:
1. Check workflow syntax in `.github/workflows/`
2. Ensure secrets are configured if needed
3. Review action logs for specific errors