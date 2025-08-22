package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cliofy/govte"
)

// ProgressHandler is a Handler implementation specifically for progress bar display
// It inherits from NoopHandler and overrides necessary methods to support progress bar rendering
type ProgressHandler struct {
	govte.NoopHandler
	output strings.Builder
	currentLine strings.Builder
}

// NewProgressHandler creates a new ProgressHandler instance
func NewProgressHandler() *ProgressHandler {
	return &ProgressHandler{}
}

// Input processes character output, adds characters to current line
func (h *ProgressHandler) Input(c rune) {
	h.currentLine.WriteRune(c)
}

// CarriageReturn handles carriage return, used for in-line updates (\r)
func (h *ProgressHandler) CarriageReturn() {
	// Carriage return without line feed, just moves cursor to beginning of line
	// Next output will overwrite current line content
	fmt.Print("\r")
	h.currentLine.Reset()
}

// LineFeed handles line feed, outputs current line and creates new line
func (h *ProgressHandler) LineFeed() {
	if h.currentLine.Len() > 0 {
		fmt.Print(h.currentLine.String())
		h.currentLine.Reset()
	}
	fmt.Println()
}

// Bell handles bell character
func (h *ProgressHandler) Bell() {
	fmt.Print("\a") // Send bell character to terminal
}

// SetForeground sets foreground color
func (h *ProgressHandler) SetForeground(color govte.Color) {
	// Output corresponding ANSI sequence based on GoVTE's Color struct
	switch color.Type {
	case govte.ColorTypeNamed:
		switch color.Named {
		case govte.Red:
			fmt.Print("\x1b[31m")
		case govte.Green:
			fmt.Print("\x1b[32m")
		case govte.Yellow:
			fmt.Print("\x1b[33m")
		case govte.Blue:
			fmt.Print("\x1b[34m")
		case govte.Magenta:
			fmt.Print("\x1b[35m")
		case govte.Cyan:
			fmt.Print("\x1b[36m")
		case govte.White:
			fmt.Print("\x1b[37m")
		case govte.Black:
			fmt.Print("\x1b[30m")
		}
	case govte.ColorTypeIndexed:
		// 256-color mode
		fmt.Printf("\x1b[38;5;%dm", color.Index)
	case govte.ColorTypeRgb:
		// RGB true color mode
		fmt.Printf("\x1b[38;2;%d;%d;%dm", color.Rgb.R, color.Rgb.G, color.Rgb.B)
	}
}

// ResetColors resets colors to default values
func (h *ProgressHandler) ResetColors() {
	fmt.Print("\x1b[0m")
}

// ClearLine clears line content (used for progress bar updates)
func (h *ProgressHandler) ClearLine(mode govte.LineClearMode) {
	switch mode {
	case govte.LineClearRight:
		fmt.Print("\x1b[K") // Clear from cursor to end of line
	case govte.LineClearLeft:
		fmt.Print("\x1b[1K") // Clear from beginning of line to cursor
	case govte.LineClearAll:
		fmt.Print("\x1b[2K") // Clear entire line
	}
}

// Flush immediately flushes output buffer, ensures progress bar displays in time
func (h *ProgressHandler) Flush() {
	if h.currentLine.Len() > 0 {
		fmt.Print(h.currentLine.String())
		h.currentLine.Reset()
	}
	os.Stdout.Sync()
}

// PrintDirect outputs text directly, bypassing ANSI processing (for simple output)
func (h *ProgressHandler) PrintDirect(text string) {
	fmt.Print(text)
	os.Stdout.Sync()
}

// PrintLineDirect outputs a line of text directly
func (h *ProgressHandler) PrintLineDirect(text string) {
	fmt.Println(text)
	os.Stdout.Sync()
}

// Progress bar rendering helper functions

// RenderSimpleBar renders simple ASCII progress bar
func (h *ProgressHandler) RenderSimpleBar(progress, width int) string {
	filled := (progress * width) / 100
	empty := width - filled
	
	bar := "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"
	return fmt.Sprintf("%s %d%%", bar, progress)
}

// RenderUnicodeBar renders Unicode-style progress bar
func (h *ProgressHandler) RenderUnicodeBar(progress, width int) string {
	filled := (progress * width) / 100
	empty := width - filled
	
	var bar strings.Builder
	bar.WriteString("[")
	
	// Completed portion
	bar.WriteString(strings.Repeat("█", filled))
	
	// Partially complete character (if there's a remainder)
	remainder := (progress * width) % 100
	if remainder > 0 && filled < width {
		bar.WriteString("▓")
		empty--
	}
	
	// Incomplete portion
	bar.WriteString(strings.Repeat("░", empty))
	bar.WriteString("]")
	
	return fmt.Sprintf("%s %d%%", bar.String(), progress)
}

// RenderColoredBar renders colored progress bar
func (h *ProgressHandler) RenderColoredBar(progress, width int) string {
	filled := (progress * width) / 100
	
	// Choose color based on progress
	var colorCode string
	if progress < 33 {
		colorCode = "\x1b[31m" // Red
	} else if progress < 66 {
		colorCode = "\x1b[33m" // Yellow
	} else {
		colorCode = "\x1b[32m" // Green
	}
	
	var bar strings.Builder
	bar.WriteString("Progress: [")
	
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("=")
		} else if i == filled && progress < 100 {
			bar.WriteString(">")
		} else {
			bar.WriteString(" ")
		}
	}
	
	bar.WriteString("]")
	
	return fmt.Sprintf("%s%s %d%%\x1b[0m", colorCode, bar.String(), progress)
}

// GetSpinner gets spinner character
func GetSpinner(step int, spinnerType string) rune {
	switch spinnerType {
	case "braille":
		spinners := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		return spinners[step%len(spinners)]
	case "blocks":
		spinners := []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}
		return spinners[step%len(spinners)]
	default:
		spinners := []rune{'|', '/', '-', '\\'}
		return spinners[step%len(spinners)]
	}
}