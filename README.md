# GoVTE

A Go implementation of the VTE (Virtual Terminal Emulator) parser, based on Paul Williams' ANSI parser state machine.

## Overview

GoVTE is a port of the Rust VTE library to Go, providing a robust parser for implementing terminal emulators. The library follows Test-Driven Development (TDD) practices to ensure code quality and correctness.

## Features

- ✅ Complete ANSI escape sequence parser
- ✅ UTF-8 support
- ✅ Minimal allocations for performance
- ✅ Clean interface design with the `Performer` interface
- ✅ Comprehensive test coverage

## Installation

```bash
go get github.com/cliofy/govte
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/cliofy/govte"
)

// Implement the Performer interface
type MyPerformer struct {
    govte.NoopPerformer
}

func (p *MyPerformer) Print(c rune) {
    fmt.Printf("Print: %c\n", c)
}

func (p *MyPerformer) Execute(b byte) {
    fmt.Printf("Execute: 0x%02x\n", b)
}

func main() {
    parser := govte.NewParser()
    performer := &MyPerformer{}
    
    // Parse some input
    input := []byte("Hello\x1b[31mRed Text\x1b[0m")
    parser.Advance(performer, input)
}
```

## Project Structure

```
govte/
├── state.go          # State machine states
├── performer.go      # Performer interface
├── params.go         # Parameter handling
├── parser.go         # Core parser (in progress)
├── ansi.go          # ANSI definitions (planned)
└── doc/
    └── go-impl.md   # Implementation plan
```

## Development Status

### Completed ✅
- Basic type definitions (State, Performer, Params)
- TDD test suite for core components
- Project structure and documentation

### In Progress 🚧
- Parser implementation
- State machine logic
- UTF-8 handling

### Planned 📋
- ANSI feature support
- Examples and benchmarks
- Performance optimizations

## Testing

Run tests with:
```bash
go test ./...
```

Check coverage:
```bash
go test -cover ./...
```

## Contributing

This project follows TDD principles. Please:
1. Write tests first
2. Implement minimal code to pass tests
3. Refactor while keeping tests green
4. Maintain high test coverage

## License

Apache-2.0 OR MIT (same as the original Rust VTE)

## References

- [Original Rust VTE](https://github.com/alacritty/vte)
- [VT100.net Parser](https://vt100.net/emu/dec_ansi_parser)
- [ANSI Escape Codes](https://en.wikipedia.org/wiki/ANSI_escape_code)