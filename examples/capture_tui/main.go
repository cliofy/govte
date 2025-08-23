package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/cliofy/govte"
	"github.com/cliofy/govte/terminal"
	"github.com/creack/pty"
)

func main() {
	// Start htop
	cmd := exec.Command("htop")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer ptmx.Close()
	defer cmd.Process.Kill()

	// Set terminal size
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(24),
		Cols: uint16(80),
	})

	// Create parser and terminal buffer
	parser := govte.NewParser()
	term := terminal.NewTerminalBuffer(80, 24)

	// Channel to signal completion
	done := make(chan bool)

	// Goroutine to continuously read and parse output
	go func() {
		buf := make([]byte, 4096)
		for {
			select {
			case <-done:
				return
			default:
				ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, _ := ptmx.Read(buf)
				if n > 0 {
					// Parse bytes into terminal buffer
					for i := 0; i < n; i++ {
						parser.Advance(term, []byte{buf[i]})
					}
				}
			}
		}
	}()

	// Timer to print terminal content every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Timer to exit after 5 seconds
	timeout := time.After(5 * time.Second)

	// Initial wait for htop to start
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Starting htop monitoring (5 seconds)...")

	for {
		select {
		case <-ticker.C:
			// Clear screen (ANSI escape code)
			fmt.Print("\033[H\033[2J")
			// Print current terminal buffer
			fmt.Println("=== Terminal Output (Updated every 1s) ===")
			fmt.Println(term.GetDisplayWithColors())
			fmt.Println("\n[Press Ctrl+C to exit early]")

		case <-timeout:
			// Signal goroutine to stop
			close(done)
			fmt.Println("\n5 seconds elapsed. Exiting...")
			return
		}
	}
}
