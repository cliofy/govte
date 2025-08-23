# Changelog

All notable changes to GoVTE will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-08-23

### Added
- Complete VT100/xterm ANSI escape sequence parsing
- Full terminal buffer implementation with cursor management
- Support for 24-bit true color (RGB), 256-color palette, and named colors
- Unicode/UTF-8 character handling
- CSI (Control Sequence Introducer) sequences support
- OSC (Operating System Command) sequences support
- DCS (Device Control String) sequences support
- SGR (Select Graphic Rendition) parameters
- Terminal title and icon sequences
- Comprehensive test coverage
- Example programs for various use cases
  - Animated progress bars
  - TUI program capture
  - VTE animations

### Features
- High-performance state machine implementation
- Modular design with Performer interface for custom implementations
- Production-ready terminal emulation
- Convenience functions for quick parsing
- Full documentation and examples

### Dependencies
- Minimal external dependencies (only testify for testing)
- Go 1.21+ support

## [0.1.0] - 2025-08-22

### Added
- Initial release
- Basic VTE parser implementation
- Core state machine
- Basic terminal buffer

---

[0.2.0]: https://github.com/cliofy/govte/releases/tag/v0.2.0