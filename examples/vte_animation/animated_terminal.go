package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cliofy/govte"
)

// AnimatedTerminal animated terminal - contains VTE parser and buffer
// Similar to the Rust version AnimatedTerminal struct
type AnimatedTerminal struct {
	// GoVTE parser
	parser *govte.Parser
	// Terminal buffer
	buffer *TerminalBuffer
}

// NewAnimatedTerminal creates a new animated terminal
func NewAnimatedTerminal(width, height int) *AnimatedTerminal {
	return &AnimatedTerminal{
		parser: govte.NewParser(),
		buffer: NewTerminalBuffer(width, height),
	}
}

// Process processes input and updates buffer
// Equivalent to the Rust version process method
func (a *AnimatedTerminal) Process(input []byte) {
	a.parser.Advance(a.buffer, input)
}

// ProcessString convenience method for processing string input
func (a *AnimatedTerminal) ProcessString(input string) {
	a.Process([]byte(input))
}

// Render renders current buffer to terminal
// Implements bordered terminal display, similar to Rust version render method
func (a *AnimatedTerminal) Render() {
	width, height := a.buffer.GetDimensions()
	buffer := a.buffer.GetBuffer()

	// Hide cursor to avoid flickering
	fmt.Print("\x1b[?25l")

	// Use absolute positioning to draw top border (line 1)
	fmt.Printf("\x1b[1;1H┌%s┐\x1b[K", strings.Repeat("─", width))

	// Use absolute positioning to draw each line content
	for i, line := range buffer {
		// Position to line i+2 (because line 1 is the top border)
		fmt.Printf("\x1b[%d;1H│", i+2)
		for _, ch := range line {
			fmt.Printf("%c", ch)
		}
		fmt.Print("│\x1b[K") // Draw right border and clear to end of line
	}

	// Use absolute positioning to draw bottom border
	bottomRow := height + 2
	fmt.Printf("\x1b[%d;1H└%s┘\x1b[K", bottomRow, strings.Repeat("─", width))

	// Restore cursor display
	fmt.Print("\x1b[?25h")

	// Flush output
	os.Stdout.Sync()
}

// Clear clears terminal buffer
func (a *AnimatedTerminal) Clear() {
	a.buffer.Clear()
}

// GetBuffer gets underlying buffer (for direct access)
func (a *AnimatedTerminal) GetBuffer() *TerminalBuffer {
	return a.buffer
}

// GetDimensions gets terminal dimensions
func (a *AnimatedTerminal) GetDimensions() (int, int) {
	return a.buffer.GetDimensions()
}

// GetCursor gets current cursor position
func (a *AnimatedTerminal) GetCursor() (int, int) {
	return a.buffer.GetCursor()
}

// MoveCursor moves cursor to specified position (convenience method)
func (a *AnimatedTerminal) MoveCursor(row, col int) {
	cmd := fmt.Sprintf("\x1b[%d;%dH", row+1, col+1) // Convert to 1-based index
	a.ProcessString(cmd)
}

// WriteAt writes text at specified position (convenience method)
func (a *AnimatedTerminal) WriteAt(row, col int, text string) {
	a.MoveCursor(row, col)
	a.ProcessString(text)
}

// WriteAtColored writes colored text at specified position
func (a *AnimatedTerminal) WriteAtColored(row, col int, text string, colorCode string) {
	a.MoveCursor(row, col)
	coloredText := fmt.Sprintf("%s%s\x1b[0m", colorCode, text)
	a.ProcessString(coloredText)
}

// ClearScreen clears screen (sends CSI sequence)
func (a *AnimatedTerminal) ClearScreen() {
	a.ProcessString("\x1b[2J")
}

// ClearLine clears current line
func (a *AnimatedTerminal) ClearLine() {
	a.ProcessString("\x1b[K")
}

// SetTitle sets window title (convenience method)
func (a *AnimatedTerminal) SetTitle(title string) {
	titleSeq := fmt.Sprintf("\x1b]0;%s\x07", title)
	fmt.Print(titleSeq)
}

// Color constants for convenience
const (
	ColorReset   = "\x1b[0m"
	ColorRed     = "\x1b[31m"
	ColorGreen   = "\x1b[32m"
	ColorYellow  = "\x1b[33m"
	ColorBlue    = "\x1b[34m"
	ColorMagenta = "\x1b[35m"
	ColorCyan    = "\x1b[36m"
	ColorWhite   = "\x1b[37m"

	// Bright colors
	ColorBrightRed     = "\x1b[91m"
	ColorBrightGreen   = "\x1b[92m"
	ColorBrightYellow  = "\x1b[93m"
	ColorBrightBlue    = "\x1b[94m"
	ColorBrightMagenta = "\x1b[95m"
	ColorBrightCyan    = "\x1b[96m"
	ColorBrightWhite   = "\x1b[97m"
)

// PrintTitle prints title at top of terminal (for transitions between demos)
func PrintTitle(title string) {
	fmt.Print("\x1b[H\x1b[J") // Clear screen and return to top-left corner
	fmt.Printf("%s%s%s\n", ColorBrightCyan, title, ColorReset)
	os.Stdout.Sync()
}

// EnterAlternateScreen enters alternate screen buffer
func EnterAlternateScreen() {
	fmt.Print("\x1b[s\x1b[?1049h\x1b[?25l\x1b[H\x1b[2J")
	os.Stdout.Sync()
}

// ExitAlternateScreen exits alternate screen buffer
func ExitAlternateScreen() {
	fmt.Print("\x1b[?25h\x1b[?1049l\x1b[u")
	os.Stdout.Sync()
}
