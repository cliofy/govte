# GoVTE Examples

This directory contains example programs demonstrating various features and use cases of GoVTE.

## Available Examples

### 1. Animated Progress Bar (`animated_progress/`)
Demonstrates creating animated progress bars with ANSI escape sequences.

```bash
cd animated_progress
go run .
```

Features:
- Simple progress bar with percentage
- Animated spinner with Unicode characters
- Colored progress bars with gradients
- Multi-style progress indicators

### 2. TUI Program Capture (`capture_tui/`)
Shows how to capture and render output from TUI applications.

```bash
cd capture_tui
go build
./capture_tui htop  # or any other TUI program
```

Features:
- PTY (pseudo-terminal) creation
- Real-time TUI output capture
- ANSI sequence parsing and rendering
- Terminal buffer management

### 3. Parse Log (`parselog/`)
A debugging tool that logs all ANSI escape sequences in detail.

```bash
cd parselog
echo -e "\033[31mRed\033[0m Text" | go run main.go
```

Features:
- Detailed logging of all VTE actions
- CSI, OSC, and DCS sequence debugging
- Control character identification
- State machine visualization

### 4. VTE Animation (`vte_animation/`)
Advanced animation examples using VTE features.

```bash
cd vte_animation
go run .
```

Features:
- Matrix rain effect
- Loading animations
- Color transitions
- Terminal buffer manipulation

## Running All Examples

You can run all examples sequentially:

```bash
for dir in */; do
    echo "Running $dir..."
    (cd "$dir" && go run . 2>/dev/null || go build && ./$(basename "$dir"))
    echo ""
done
```

## Creating Your Own Examples

To create a new example:

1. Create a new directory
2. Initialize a Go module: `go mod init example`
3. Add GoVTE dependency: `go get github.com/cliofy/govte@latest`
4. Import and use GoVTE:

```go
package main

import (
    "fmt"
    "github.com/cliofy/govte"
    "github.com/cliofy/govte/terminal"
)

func main() {
    // Your example code here
    output := terminal.ParseBytesWithColors(
        []byte("\x1b[32mHello GoVTE!\x1b[0m"), 
        80, 24,
    )
    fmt.Println(output)
}
```

## Learning Path

1. Start with `parselog/` to understand ANSI sequences
2. Try `animated_progress/` for basic terminal manipulation
3. Explore `capture_tui/` for advanced PTY handling
4. Study `vte_animation/` for complex animations

## Need Help?

- Check the [main documentation](https://pkg.go.dev/github.com/cliofy/govte)
- Open an issue on [GitHub](https://github.com/cliofy/govte/issues)
- See [CONTRIBUTING.md](../CONTRIBUTING.md) to contribute