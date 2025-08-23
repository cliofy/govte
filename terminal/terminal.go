//! A complete terminal buffer implementation for GoVTE
//!
//! This package provides a production-ready terminal buffer that implements
//! the VTE Performer interface, handling ANSI escape sequences and maintaining
//! terminal state.
//!
//! Example:
//!
//!	parser := govte.NewParser()
//!	terminal := terminal.NewTerminalBuffer(80, 24)
//!
//!	// Parse some terminal output
//!	bytes := []byte("Hello \x1b[31mRed Text\x1b[0m")
//!	for _, b := range bytes {
//!		parser.Advance(terminal, []byte{b})
//!	}
//!
//!	// Get the rendered output
//!	output := terminal.GetDisplay()

package terminal

import "github.com/cliofy/govte"

// DefaultTerminal creates a default terminal buffer with standard dimensions (80x24)
func DefaultTerminal() *TerminalBuffer {
	return NewTerminalBuffer(80, 24)
}

// ParseBytes parses bytes and returns the rendered display
func ParseBytes(bytes []byte, width, height int) string {
	parser := govte.NewParser()
	terminal := NewTerminalBuffer(width, height)

	for _, b := range bytes {
		parser.Advance(terminal, []byte{b})
	}

	return terminal.GetDisplay()
}

// ParseBytesWithColors parses bytes and returns the rendered display with colors
func ParseBytesWithColors(bytes []byte, width, height int) string {
	parser := govte.NewParser()
	terminal := NewTerminalBuffer(width, height)

	for _, b := range bytes {
		parser.Advance(terminal, []byte{b})
	}

	return terminal.GetDisplayWithColors()
}

// CreateTerminalFromString creates a terminal buffer and parses the given string
func CreateTerminalFromString(input string, width, height int) *TerminalBuffer {
	parser := govte.NewParser()
	terminal := NewTerminalBuffer(width, height)

	bytes := []byte(input)
	for _, b := range bytes {
		parser.Advance(terminal, []byte{b})
	}

	return terminal
}

// RenderString renders a string with VTE parsing and returns plain text
func RenderString(input string, width, height int) string {
	return ParseBytes([]byte(input), width, height)
}

// RenderStringWithColors renders a string with VTE parsing and returns colored output
func RenderStringWithColors(input string, width, height int) string {
	return ParseBytesWithColors([]byte(input), width, height)
}
