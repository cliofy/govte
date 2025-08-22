//! Example of capturing and rendering TUI program output (using the new terminal module)
//!
//! This example demonstrates how to use the new TerminalBuffer implementation:
//! 1. Start TUI programs (like htop) in a pseudo terminal (PTY)
//! 2. Capture the program's output stream
//! 3. Parse and render output using GoVTE's terminal module

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cliofy/govte"
	"github.com/cliofy/govte/terminal"
	"github.com/creack/pty"
	"golang.org/x/term"
)

// getTerminalSize gets the current terminal size, returns default values if failed
func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return 120, 40 // default size
	}
	return width, height
}

// captureTUIOutput captures TUI program output
func captureTUIOutput(program string, args []string, duration time.Duration) ([]byte, int, int, error) {
	fmt.Printf("Starting %s ...\n", program)

	// Get current terminal size
	width, height := getTerminalSize()
	fmt.Printf("Detected terminal size: %dx%d\n", width, height)

	// Create command
	cmd := exec.Command(program, args...)

	// Set environment variables
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Create PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("unable to create PTY: %w", err)
	}
	defer ptmx.Close()

	// Set PTY size
	err = pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
	if err != nil {
		log.Printf("Warning: unable to set PTY size: %v", err)
	}

	fmt.Printf("Program started, PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Starting output capture (%.0f seconds)...\n", duration.Seconds())

	// Collect output
	var output []byte
	buffer := make([]byte, 4096)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Use goroutine to read data
	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Set read timeout
				ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, err := ptmx.Read(buffer)
				if err != nil {
					if err != io.EOF && !os.IsTimeout(err) {
						log.Printf("Read error: %v", err)
					}
					continue
				}
				if n > 0 {
					output = append(output, buffer[:n]...)
					// Show capture progress
					fmt.Printf("\rCaptured %d bytes", len(output))
				}
			}
		}
	}()

	// Wait for timeout or completion
	<-ctx.Done()

	// Give read goroutine some time to complete
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	}

	fmt.Println("\nCapture complete, shutting down program...")

	// Try to gracefully terminate the program
	if cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Wait()
	}

	// Give the program some time to clean up
	time.Sleep(100 * time.Millisecond)

	return output, width, height, nil
}

// renderOutput renders output using the new TerminalBuffer
func renderOutput(data []byte, width, height int, withColors bool) string {
	parser := govte.NewParser()
	terminalBuffer := terminal.NewTerminalBuffer(width, height)

	// Parse all data
	for _, b := range data {
		parser.Advance(terminalBuffer, []byte{b})
	}

	fmt.Println("\n=== Render Statistics ===")
	fmt.Printf("Captured bytes: %d\n", len(data))
	fmt.Printf("Terminal size: %dx%d\n", width, height)

	cursorX, cursorY := terminalBuffer.CursorPosition()
	fmt.Printf("Cursor position: (%d, %d)\n", cursorX+1, cursorY+1)

	fmt.Printf("Color output: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[withColors])

	if withColors {
		return terminalBuffer.GetDisplayWithColors()
	}
	return terminalBuffer.GetDisplay()
}

func main() {
	fmt.Println("=== GoVTE TUI Program Capture Example ===")

	// Check if color output is enabled
	enableColors := false
	for _, arg := range os.Args[1:] {
		if arg == "--colors" || arg == "-c" {
			enableColors = true
			break
		}
	}

	if enableColors {
		fmt.Println("🎨 Color output mode enabled")
	} else {
		fmt.Println("💡 Tip: Use --colors or -c flag to enable color output")
	}
	fmt.Println()

	// Try different TUI programs
	programs := []struct {
		name string
		args []string
	}{
		{"htop", []string{}},
		{"btm", []string{}},
		{"top", []string{}},
		{"ps", []string{"aux"}},
	}

	var capturedData []byte
	var usedProgram string
	var terminalWidth, terminalHeight int

	// Try to find an available program
	for _, prog := range programs {
		data, width, height, err := captureTUIOutput(prog.name, prog.args, 3*time.Second)
		if err != nil {
			fmt.Printf("Unable to run %s: %v\n", prog.name, err)
			if prog.name == "htop" {
				fmt.Println("Tip: Please install htop (e.g., apt install htop or brew install htop)")
			}
			continue
		}

		capturedData = data
		terminalWidth = width
		terminalHeight = height
		usedProgram = prog.name
		break
	}

	// Render captured output
	if capturedData != nil {
		fmt.Printf("\nSuccessfully captured %s output\n", usedProgram)
		fmt.Println("\n=== Final Rendered Frame ===")

		rendered := renderOutput(capturedData, terminalWidth, terminalHeight, enableColors)

		// Output rendered result directly (avoid Unicode character truncation issues)
		lines := strings.Split(rendered, "\n")
		for _, line := range lines {
			fmt.Println(line)
		}

		// Optional: save raw data to file
		for _, arg := range os.Args[1:] {
			if arg == "--save" {
				filename := fmt.Sprintf("%s_capture.dat", usedProgram)
				err := os.WriteFile(filename, capturedData, 0644)
				if err != nil {
					log.Printf("Failed to save file: %v", err)
				} else {
					fmt.Printf("\nRaw data saved to: %s\n", filename)
				}
				break
			}
		}
	} else {
		fmt.Println("\nError: Unable to capture any TUI program output")
		fmt.Println("Please ensure at least one of htop, top, or ps is installed")
		os.Exit(1)
	}
}
