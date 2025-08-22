//! GoVTE Animated Progress Bar Example
//!
//! Demonstrates how to create real animation effects using GoVTE, including:
//! - Using \r carriage return for in-line updates
//! - Time control for smooth animations
//! - Multiple progress bar styles
//! - ANSI color sequence handling

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cliofy/govte"
)

// simpleProgressBar simple progress bar - directly uses GoVTE to handle ANSI sequences
func simpleProgressBar(durationSecs int) {
	fmt.Printf("Simple progress bar (%d seconds):\n", durationSecs)

	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()

	totalSteps := 100
	delayMs := time.Duration((durationSecs * 1000) / totalSteps)

	for i := 0; i <= totalSteps; i++ {
		// Calculate progress bar width (50 characters wide)
		filled := (i * 50) / totalSteps
		empty := 50 - filled

		// Construct progress bar sequence: \r + progress bar content
		progressText := fmt.Sprintf("\r[%s%s] %d%%",
			repeatString("=", filled),
			repeatString(" ", empty),
			i)

		// Use GoVTE to process the sequence
		processor.Advance(handler, []byte(progressText))
		handler.Flush()

		if i < totalSteps {
			time.Sleep(delayMs * time.Millisecond)
		}
	}

	handler.PrintLineDirect(" Complete!")
}

// animatedProgressBar animated progress bar - with moving indicator
func animatedProgressBar(durationSecs int) {
	fmt.Printf("\nAnimated progress bar (%d seconds):\n", durationSecs)

	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()

	totalSteps := 100
	delayMs := time.Duration((durationSecs * 1000) / totalSteps)

	for i := 0; i <= totalSteps; i++ {
		spinner := GetSpinner(i, "braille")
		bar := handler.RenderUnicodeBar(i, 50)

		// Construct animated progress bar
		progressText := fmt.Sprintf("\r%c %s", spinner, bar)

		processor.Advance(handler, []byte(progressText))
		handler.Flush()

		if i < totalSteps {
			time.Sleep(delayMs * time.Millisecond)
		}
	}

	handler.PrintLineDirect(" ✓")
}

// coloredProgressBar colored progress bar - uses ANSI color codes processed through GoVTE
func coloredProgressBar(durationSecs int) {
	fmt.Printf("\nColored progress bar (%d seconds):\n", durationSecs)

	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()

	totalSteps := 100
	delayMs := time.Duration((durationSecs * 1000) / totalSteps)

	for i := 0; i <= totalSteps; i++ {
		bar := handler.RenderColoredBar(i, 50)

		// Use \r carriage return for in-line updates
		progressText := fmt.Sprintf("\r%s", bar)

		processor.Advance(handler, []byte(progressText))
		handler.Flush()

		if i < totalSteps {
			time.Sleep(delayMs * time.Millisecond)
		}
	}

	handler.PrintLineDirect(" Complete!")
}

// multiProgressBars multi-task progress bars - display multiple progress simultaneously
func multiProgressBars() {
	fmt.Println("\nMulti-task progress bars (simulating downloads):")

	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()

	// Save cursor position - processed through GoVTE
	processor.Advance(handler, []byte("\x1b[s"))

	// Prepare display area
	handler.PrintLineDirect("File 1: [                                                  ] 0%")
	handler.PrintLineDirect("File 2: [                                                  ] 0%")
	handler.PrintLineDirect("File 3: [                                                  ] 0%")
	handler.PrintLineDirect("Total: [                                                  ] 0%")

	progress := [3]int{0, 0, 0}
	speeds := [3]int{3, 5, 2} // Different download speeds

	start := time.Now()

	for hasIncompleteTask(progress[:]) {
		// Update each progress
		for i := 0; i < 3; i++ {
			if progress[i] < 100 {
				progress[i] = min(progress[i]+speeds[i], 100)
			}
		}

		// Calculate total progress
		totalProgress := (progress[0] + progress[1] + progress[2]) / 3

		// Restore cursor position and update display - process ANSI sequences through GoVTE
		processor.Advance(handler, []byte("\x1b[u")) // Restore cursor position

		for i := 0; i < 3; i++ {
			// Move cursor down to corresponding line
			moveSeq := fmt.Sprintf("\x1b[%dB", i+1)
			processor.Advance(handler, []byte(moveSeq))

			// Update progress bar
			bar := renderFileProgressBar(progress[i], 50)
			updateText := fmt.Sprintf("\rFile %d: %s", i+1, bar)
			processor.Advance(handler, []byte(updateText))

			if i < 2 {
				processor.Advance(handler, []byte("\x1b[u")) // Restore to starting position
			}
		}

		// Display total progress
		processor.Advance(handler, []byte("\x1b[u\x1b[4B")) // Restore position and move to total progress line
		totalBar := renderTotalProgressBar(totalProgress, 50)
		totalText := fmt.Sprintf("\rTotal: %s", totalBar)
		processor.Advance(handler, []byte(totalText))

		handler.Flush()
		time.Sleep(100 * time.Millisecond)

		// Prevent running too long
		if time.Since(start).Seconds() > 10 {
			break
		}
	}

	handler.PrintLineDirect("\n\n✅ All downloads complete!")
}

// streamingProgress real-time data stream progress (simulating log output)
func streamingProgress() {
	fmt.Println("\nReal-time data stream (simulating log processing):")
	fmt.Println(repeatString("─", 60))

	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()

	messages := []string{
		"Initializing system...",
		"Loading configuration files...",
		"Connecting to database...",
		"Verifying permissions...",
		"Loading modules...",
		"Starting services...",
		"Listening on port 8080...",
		"System ready",
	}

	for i, msg := range messages {
		// Display processing animation
		for j := 0; j < 8; j++ {
			spinner := GetSpinner(j, "blocks")
			statusText := fmt.Sprintf("\r%c %s", spinner, msg)

			processor.Advance(handler, []byte(statusText))
			handler.Flush()
			time.Sleep(125 * time.Millisecond)
		}

		// Complete current step
		progress := ((i + 1) * 100) / len(messages)
		completeText := fmt.Sprintf("\r✓ %s [%d%%]\n", msg, progress)
		processor.Advance(handler, []byte(completeText))
		handler.Flush()

		time.Sleep(200 * time.Millisecond)
	}

	handler.PrintLineDirect(repeatString("─", 60))
	handler.PrintLineDirect("🚀 System startup complete!")
}

// Helper functions

// repeatString repeats a string n times (alternative to strings.Repeat for Go versions before 1.21)
func repeatString(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// hasIncompleteTask checks if there are incomplete tasks
func hasIncompleteTask(progress []int) bool {
	for _, p := range progress {
		if p < 100 {
			return true
		}
	}
	return false
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderFileProgressBar renders file download progress bar
func renderFileProgressBar(progress, width int) string {
	filled := (progress * width) / 100

	bar := "["
	for j := 0; j < width; j++ {
		if j < filled {
			if progress == 100 {
				bar += "\x1b[32m=\x1b[0m" // Show green when complete
			} else {
				bar += "="
			}
		} else if j == filled && progress < 100 {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %d%%", progress)

	return bar
}

// renderTotalProgressBar renders total progress bar
func renderTotalProgressBar(progress, width int) string {
	filled := (progress * width) / 100

	bar := "["
	for j := 0; j < width; j++ {
		if j < filled {
			bar += "\x1b[36m▓\x1b[0m" // Cyan
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("] %d%%", progress)

	return bar
}

func main() {
	fmt.Println("=== Animated Progress Bar Example (GoVTE) ===")

	// Check if terminal supports UTF-8
	if os.Getenv("LANG") == "" {
		fmt.Println("Note: If display is abnormal, please ensure your terminal supports UTF-8 encoding")
		fmt.Println()
	}

	// Example 1: Simple progress bar
	simpleProgressBar(3)

	// Example 2: Animated progress bar
	animatedProgressBar(3)

	// Example 3: Colored progress bar
	coloredProgressBar(3)

	// Example 4: Multi-task progress bars
	multiProgressBars()

	// Example 5: Real-time data stream
	streamingProgress()

	fmt.Println("\nAll examples completed!")
	fmt.Println("\nThis example demonstrates the following GoVTE library features:")
	fmt.Println("• ANSI escape sequence parsing and processing")
	fmt.Println("• Terminal cursor control and positioning")
	fmt.Println("• Color sequence processing and rendering")
	fmt.Println("• Real-time output and buffer management")
	fmt.Println("• Unicode character support")
}
