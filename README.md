# GoVTE

[![Go Reference](https://pkg.go.dev/badge/github.com/cliofy/govte.svg)](https://pkg.go.dev/github.com/cliofy/govte)
[![Go Report Card](https://goreportcard.com/badge/github.com/cliofy/govte)](https://goreportcard.com/report/github.com/cliofy/govte)

GoVTE is a Go implementation of a VTE (Virtual Terminal Emulator) parser, providing robust ANSI escape sequence parsing and complete terminal emulation capabilities.

## Features

- 🔍 **ANSI Escape Sequence Parsing** - Complete VT100/xterm compatibility
- 🎨 **Full Color Support** - Named colors, RGB, and 256-color palette
- 📺 **Terminal Emulation** - Complete terminal buffer with cursor management
- 🌐 **Unicode Support** - Full UTF-8 character handling
- 🖥️ **TUI Program Capture** - Capture and render real TUI applications
- ⚡ **High Performance** - Optimized state machine implementation

## Installation

```bash
go get github.com/cliofy/govte
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/cliofy/govte"
    "github.com/cliofy/govte/terminal"
)

func main() {
    parser := govte.NewParser()
    terminalBuffer := terminal.NewTerminalBuffer(80, 24)
    
    // Parse ANSI colored text
    input := []byte("Hello \x1b[31mRed\x1b[0m World!")
    for _, b := range input {
        parser.Advance(terminalBuffer, []byte{b})
    }
    
    fmt.Println(terminalBuffer.GetDisplay())
    
    // Or use convenience functions
    output := terminal.ParseBytesWithColors([]byte("\x1b[32mGreen\x1b[0m"), 80, 24)
    fmt.Println(output)
}
```

## Core Components

### Parser

The `Parser` implements a state machine for processing ANSI escape sequences:

```go
parser := govte.NewParser()
// Process bytes through the parser
parser.Advance(performer, inputBytes)
```

### Performer Interface

The `Performer` interface handles parsed actions. Implement it for custom behavior:

```go
type MyPerformer struct {
    govte.NoopPerformer // Embed for default implementations
}

func (p *MyPerformer) Print(c rune) {
    fmt.Printf("Character: %c\n", c)
}

func (p *MyPerformer) CsiDispatch(params *govte.Params, intermediates []byte, ignore bool, action rune) {
    fmt.Printf("CSI sequence: %c with params %v\n", action, params)
}
```

### Terminal Buffer

The `TerminalBuffer` provides complete terminal emulation:

```go
terminal := terminal.NewTerminalBuffer(width, height)

// Get plain text output
text := terminal.GetDisplay()

// Get output with ANSI color codes preserved
colored := terminal.GetDisplayWithColors()

// Access cursor position
x, y := terminal.CursorPosition()
```

## Advanced Usage

### Custom Performer Implementation

```go
type LoggingPerformer struct {
    govte.NoopPerformer
}

func (l *LoggingPerformer) Execute(b byte) {
    switch b {
    case 0x0A: // Line Feed
        fmt.Println("[LF] New line")
    case 0x0D: // Carriage Return
        fmt.Println("[CR] Carriage return")
    }
}

func (l *LoggingPerformer) CsiDispatch(params *govte.Params, intermediates []byte, ignore bool, action rune) {
    fmt.Printf("[CSI] Action: %c, Params: %v\n", action, params)
}
```

### TUI Program Capture

Capture and render real TUI applications:

```go
import (
    "github.com/creack/pty"
    "os/exec"
)

// Start a TUI program in a PTY
cmd := exec.Command("htop")
ptmx, err := pty.Start(cmd)
if err != nil {
    panic(err)
}

// Capture output
var output []byte
buffer := make([]byte, 4096)
n, _ := ptmx.Read(buffer)
output = append(output, buffer[:n]...)

// Parse and render
parser := govte.NewParser()
terminal := terminal.NewTerminalBuffer(120, 40)
for _, b := range output {
    parser.Advance(terminal, []byte{b})
}

fmt.Println(terminal.GetDisplayWithColors())
```

### Color Handling

```go
// Supports named colors (\x1b[31m), RGB (\x1b[38;2;255;0;0m), and 256-color (\x1b[38;5;196m)
input := "\x1b[38;2;255;0;0mRGB Red\x1b[0m \x1b[38;5;21mBlue\x1b[0m"
output := terminal.ParseBytesWithColors([]byte(input), 80, 24)
```

## Examples & Tools

The repository includes several example programs:

- **`examples/parselog/`** - Debug tool that logs all parsed ANSI actions
- **`examples/capture_tui/`** - Complete TUI program capture and rendering
- **`examples/animated_progress/`** - Animated progress bar demonstration
- **`examples/vte_animation/`** - VTE animation examples

```bash
cd examples/parselog && go run main.go
cd examples/capture_tui && go build && ./capture_tui --colors
```

## Supported Features

- ✅ CSI (Control Sequence Introducer) sequences
- ✅ OSC (Operating System Command) sequences
- ✅ SGR (Select Graphic Rendition) parameters
- ✅ Cursor movement and positioning
- ✅ Text styling (bold, italic, underline, etc.)
- ✅ Color support (3/4-bit, 8-bit, 24-bit)
- ✅ Character set handling
- ✅ UTF-8 unicode support
- ✅ Terminal title and icon sequences

## Contributing

Contributions are welcome! Please ensure:

1. Code follows Go conventions and is well-documented
2. All tests pass: `go test ./...`
3. Add tests for new functionality
4. Performance-critical code includes benchmarks

## License

MIT License - see LICENSE file for details.