package main

import (
	"github.com/cliofy/govte"
)

// TerminalBuffer implements terminal buffer, similar to Rust version TerminalBuffer
// It implements the govte.Performer interface to handle VTE parser callbacks
type TerminalBuffer struct {
	// Screen buffer - 2D character array
	buffer [][]rune
	// Cursor position
	cursorRow int
	cursorCol int
	// Terminal dimensions
	width  int
	height int
}

// NewTerminalBuffer creates a new terminal buffer
func NewTerminalBuffer(width, height int) *TerminalBuffer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
		// Initialize with space characters
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}
	
	return &TerminalBuffer{
		buffer:    buffer,
		cursorRow: 0,
		cursorCol: 0,
		width:     width,
		height:    height,
	}
}

// Clear clears buffer and resets cursor
func (t *TerminalBuffer) Clear() {
	for i := range t.buffer {
		for j := range t.buffer[i] {
			t.buffer[i][j] = ' '
		}
	}
	t.cursorRow = 0
	t.cursorCol = 0
}

// GetBuffer gets buffer content (for rendering)
func (t *TerminalBuffer) GetBuffer() [][]rune {
	return t.buffer
}

// GetCursor gets current cursor position
func (t *TerminalBuffer) GetCursor() (int, int) {
	return t.cursorRow, t.cursorCol
}

// GetDimensions gets terminal dimensions
func (t *TerminalBuffer) GetDimensions() (int, int) {
	return t.width, t.height
}

// === Implement govte.Performer interface ===

// Print handles printable characters
func (t *TerminalBuffer) Print(c rune) {
	if t.cursorRow < t.height && t.cursorCol < t.width {
		t.buffer[t.cursorRow][t.cursorCol] = c
		t.cursorCol++
		
		// Auto line wrap
		if t.cursorCol >= t.width {
			t.cursorCol = 0
			if t.cursorRow < t.height-1 {
				t.cursorRow++
			}
		}
	}
}

// Execute handles control characters
func (t *TerminalBuffer) Execute(b byte) {
	switch b {
	case 0x08: // BS - Backspace
		if t.cursorCol > 0 {
			t.cursorCol--
		}
	case 0x0A: // LF - Line Feed
		if t.cursorRow < t.height-1 {
			t.cursorRow++
		}
	case 0x0D: // CR - Carriage Return
		t.cursorCol = 0
	}
}

// Hook DCS sequence start (not implemented yet)
func (t *TerminalBuffer) Hook(params *govte.Params, intermediates []byte, ignore bool, action rune) {
}

// Put DCS data (not implemented yet)
func (t *TerminalBuffer) Put(b byte) {
}

// Unhook DCS sequence end (not implemented yet)
func (t *TerminalBuffer) Unhook() {
}

// OscDispatch handles OSC sequences (not implemented yet)
func (t *TerminalBuffer) OscDispatch(params [][]byte, bellTerminated bool) {
}

// CsiDispatch handles CSI sequences (core terminal control)
func (t *TerminalBuffer) CsiDispatch(params *govte.Params, intermediates []byte, ignore bool, action rune) {
	if ignore {
		return
	}
	
	// Convert Params to []uint16 slice for processing
	var paramsVec []uint16
	if params != nil {
		groups := params.Iter()
		for _, group := range groups {
			if len(group) > 0 {
				paramsVec = append(paramsVec, group[0])
			}
		}
	}
	
	switch action {
	case 'H', 'f': // CUP - Cursor Position
		row := 1
		col := 1
		
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			row = int(paramsVec[0])
		}
		if len(paramsVec) > 1 && paramsVec[1] > 0 {
			col = int(paramsVec[1])
		}
		
		// Convert to 0-based index and limit to valid range
		t.cursorRow = min(row-1, t.height-1)
		t.cursorCol = min(col-1, t.width-1)
		if t.cursorRow < 0 {
			t.cursorRow = 0
		}
		if t.cursorCol < 0 {
			t.cursorCol = 0
		}
		
	case 'J': // ED - Erase Display
		if len(paramsVec) == 0 || paramsVec[0] == 0 {
			// Clear from cursor to end of screen
			for row := t.cursorRow; row < t.height; row++ {
				startCol := 0
				if row == t.cursorRow {
					startCol = t.cursorCol
				}
				for col := startCol; col < t.width; col++ {
					t.buffer[row][col] = ' '
				}
			}
		} else if paramsVec[0] == 1 {
			// Clear from beginning of screen to cursor
			for row := 0; row <= t.cursorRow; row++ {
				endCol := t.width
				if row == t.cursorRow {
					endCol = t.cursorCol + 1
				}
				for col := 0; col < endCol; col++ {
					t.buffer[row][col] = ' '
				}
			}
		} else if paramsVec[0] == 2 {
			// Clear entire screen
			for row := range t.buffer {
				for col := range t.buffer[row] {
					t.buffer[row][col] = ' '
				}
			}
			t.cursorRow = 0
			t.cursorCol = 0
		}
		
	case 'K': // EL - Erase Line
		if t.cursorRow < t.height {
			if len(paramsVec) == 0 || paramsVec[0] == 0 {
				// Clear to end of line
				for col := t.cursorCol; col < t.width; col++ {
					t.buffer[t.cursorRow][col] = ' '
				}
			} else if paramsVec[0] == 1 {
				// Clear from beginning of line to cursor
				for col := 0; col <= t.cursorCol && col < t.width; col++ {
					t.buffer[t.cursorRow][col] = ' '
				}
			} else if paramsVec[0] == 2 {
				// Clear entire line
				for col := 0; col < t.width; col++ {
					t.buffer[t.cursorRow][col] = ' '
				}
			}
		}
		
	case 'A': // CUU - Cursor Up
		lines := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			lines = int(paramsVec[0])
		}
		t.cursorRow = max(0, t.cursorRow-lines)
		
	case 'B': // CUD - Cursor Down  
		lines := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			lines = int(paramsVec[0])
		}
		t.cursorRow = min(t.height-1, t.cursorRow+lines)
		
	case 'C': // CUF - Cursor Forward
		cols := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			cols = int(paramsVec[0])
		}
		t.cursorCol = min(t.width-1, t.cursorCol+cols)
		
	case 'D': // CUB - Cursor Back
		cols := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			cols = int(paramsVec[0])
		}
		t.cursorCol = max(0, t.cursorCol-cols)
	}
}

// EscDispatch handles ESC sequences (not implemented yet)
func (t *TerminalBuffer) EscDispatch(intermediates []byte, ignore bool, b byte) {
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}