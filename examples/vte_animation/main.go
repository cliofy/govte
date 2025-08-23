//! GoVTE Animated Terminal Demo
//!
//! Demonstrates how to combine GoVTE with real-time updates to create various terminal animation effects
//! This is a Go implementation of the Rust version vte_animation.rs

package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// Check terminal support
	if !checkTerminalSupport() {
		fmt.Println("Warning: Terminal may not fully support required features, display may be abnormal")
		time.Sleep(2 * time.Second)
	}

	// Enter alternate screen buffer, hide cursor, and clear screen
	EnterAlternateScreen()

	// Ensure terminal state is restored on exit
	defer func() {
		ExitAlternateScreen()
		fmt.Println("Thank you for using the GoVTE animation demo!")
	}()

	// Display title in alternate screen
	fmt.Print("\x1b[1;1H=== VTE Animated Terminal Demo (GoVTE) ===")
	fmt.Print("\x1b[2;1HImplementing various terminal animation effects using GoVTE parser")
	fmt.Print("\x1b[4;1HThis demo includes the following animations:")
	fmt.Print("\x1b[5;3H• 📊 Animated Progress Bar")
	fmt.Print("\x1b[6;3H• ⌨️  Typewriter Effect")
	fmt.Print("\x1b[7;3H• 💊 Matrix Rain Effect")
	fmt.Print("\x1b[8;3H• 📈 Live Chart")
	fmt.Print("\x1b[9;3H• 🌊 Wave Animation (Enhanced)")
	fmt.Print("\x1b[10;3H• 🌀 Spiral Animation (Enhanced)")
	fmt.Print("\x1b[11;3H• 🎆 Fireworks Animation (Innovative)")
	fmt.Print("\x1b[13;1HPress Ctrl+C to exit at any time")
	fmt.Print("\x1b[15;1HStarting demo...")
	os.Stdout.Sync()

	// Start countdown
	for i := 3; i > 0; i-- {
		fmt.Printf("\x1b[15;15HCountdown: %d", i)
		os.Stdout.Sync()
		time.Sleep(1 * time.Second)
	}

	// Run each demo
	runAllDemos()

	// Display completion message
	fmt.Print("\x1b[H\x1b[2J")
	fmt.Print("\x1b[10;20H✨ All demos completed!")
	fmt.Print("\x1b[12;15HThis demo showcases the powerful features of GoVTE:")
	fmt.Print("\x1b[13;17H• VTE parser's ANSI sequence processing")
	fmt.Print("\x1b[14;17H• Terminal buffer management and rendering")
	fmt.Print("\x1b[15;17H• Real-time animations and visual effects")
	fmt.Print("\x1b[16;17H• Cursor control and screen operations")
	fmt.Print("\x1b[18;20HThanks for watching! Auto-exit in 3 seconds...")
	os.Stdout.Sync()
	time.Sleep(3 * time.Second)
}

// runAllDemos runs all demos
func runAllDemos() {
	demos := []struct {
		name string
		fn   func()
	}{
		{"Animated Progress Bar", DemoProgressBar},
		{"Typewriter Effect", DemoTypewriter},
		{"Matrix Rain Effect", DemoMatrixRain},
		{"Live Chart", DemoLiveChart},
		{"Wave Animation", DemoWaveAnimation},
		{"Spiral Animation", DemoSpiralAnimation},
		{"Fireworks Animation", DemoFireworks},
	}

	for i, demo := range demos {
		// Display current demo information
		showDemoTransition(i+1, len(demos), demo.name)

		// Run the demo
		demo.fn()

		// Interval between demos
		if i < len(demos)-1 {
			showTransitionMessage("Preparing next demo...")
			time.Sleep(1 * time.Second)
		}
	}
}

// showDemoTransition displays demo transition information
func showDemoTransition(current, total int, name string) {
	fmt.Print("\x1b[H\x1b[2J") // Clear screen
	fmt.Printf("\x1b[8;20HDemo Progress: %d/%d", current, total)
	fmt.Printf("\x1b[10;20HCurrent Demo: %s%s%s", ColorBrightYellow, name, ColorReset)
	fmt.Print("\x1b[12;25HPreparing...")
	os.Stdout.Sync()
	time.Sleep(800 * time.Millisecond)
}

// showTransitionMessage displays transition message
func showTransitionMessage(message string) {
	fmt.Print("\x1b[H\x1b[2J")
	fmt.Printf("\x1b[10;%dH%s", (80-len(message))/2, message)
	os.Stdout.Sync()
}

// checkTerminalSupport checks terminal support
func checkTerminalSupport() bool {
	// Check basic environment variables
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// Check if colors are supported
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "" && term != "xterm-256color" && term != "screen-256color" {
		return false
	}

	return true
}

// Demo functionality description
func init() {
	// Set random seed
	SetSeed(uint64(time.Now().UnixNano()))
}
