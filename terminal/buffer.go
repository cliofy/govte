//! Terminal buffer implementation
//! Go port of the Rust implementation - production-ready terminal buffer

package terminal

import (
	"strings"

	"github.com/cliofy/govte"
)

// TerminalBuffer implements a complete terminal buffer with VTE integration
type TerminalBuffer struct {
	// Screen dimensions
	width  int
	height int

	// Terminal state
	viewport     []Row
	cursor       Cursor
	savedCursor  *SavedCursor
	title        *string
	scrollRegion *ScrollRegion

	// Current character styles
	currentStyles CharacterStyles
}

// ScrollRegion represents the terminal scroll region
type ScrollRegion struct {
	top    int
	bottom int
}

// NewTerminalBuffer creates a new terminal buffer with specified dimensions
func NewTerminalBuffer(width, height int) *TerminalBuffer {
	viewport := make([]Row, height)
	for i := range viewport {
		viewport[i] = NewRowWithWidth(width)
	}

	return &TerminalBuffer{
		width:         width,
		height:        height,
		viewport:      viewport,
		cursor:        NewCursor(),
		currentStyles: DefaultCharacterStyles(),
	}
}

// GetDisplay returns the rendered display as plain text
func (tb *TerminalBuffer) GetDisplay() string {
	var result strings.Builder

	for i, row := range tb.viewport {
		result.WriteString(row.ToString())
		if i < len(tb.viewport)-1 {
			result.WriteString("\n")
		}
	}

	return strings.TrimRight(result.String(), " \t\n")
}

// GetDisplayWithColors returns the rendered display with ANSI color codes
func (tb *TerminalBuffer) GetDisplayWithColors() string {
	var result strings.Builder
	currentStyles := DefaultCharacterStyles()

	for rowIdx, row := range tb.viewport {
		for _, character := range row.Columns {
			// Only emit style changes when styles actually change
			if character.Styles.DiffersFrom(&currentStyles) {
				// Reset if we had any previous styles
				defaultStyles := DefaultCharacterStyles()
				if !currentStyles.equals(&defaultStyles) {
					result.WriteString("\x1b[0m")
				}

				// Apply new styles
				styleSequence := character.Styles.ToAnsiSequence()
				if styleSequence != "" {
					result.WriteString(styleSequence)
				}

				currentStyles = character.Styles
			}

			result.WriteRune(character.Character)
		}

		if rowIdx < len(tb.viewport)-1 {
			result.WriteString("\n")
		}
	}

	// Reset styles at the end if we had any
	defaultStyles := DefaultCharacterStyles()
	if !currentStyles.equals(&defaultStyles) {
		result.WriteString("\x1b[0m")
	}

	return strings.TrimRight(result.String(), " \t\n")
}

// Dimensions returns the terminal dimensions
func (tb *TerminalBuffer) Dimensions() (int, int) {
	return tb.width, tb.height
}

// CursorPosition returns the current cursor position
func (tb *TerminalBuffer) CursorPosition() (int, int) {
	return tb.cursor.X, tb.cursor.Y
}

// Resize resizes the terminal buffer
func (tb *TerminalBuffer) Resize(width, height int) {
	tb.width = width
	tb.height = height

	// Resize existing rows
	for i := range tb.viewport {
		tb.viewport[i].EnsureWidth(width)
		if tb.viewport[i].Len() > width {
			tb.viewport[i].Truncate(width)
		}
	}

	// Add or remove rows as needed
	if len(tb.viewport) < height {
		// Add new rows
		for len(tb.viewport) < height {
			tb.viewport = append(tb.viewport, NewRowWithWidth(width))
		}
	} else if len(tb.viewport) > height {
		// Remove excess rows
		tb.viewport = tb.viewport[:height]
	}

	// Ensure cursor is within bounds
	if tb.cursor.X >= width {
		tb.cursor.X = width - 1
	}
	if tb.cursor.Y >= height {
		tb.cursor.Y = height - 1
	}
}

// === Performer interface implementation ===

// Print handles printable characters
func (tb *TerminalBuffer) Print(c rune) {
	tb.ensureCursorInBounds()

	// Create character with current styles
	char := NewStyledTerminalCharacter(c, tb.currentStyles)

	// Ensure the current row has enough width
	if tb.cursor.Y < len(tb.viewport) {
		tb.viewport[tb.cursor.Y].EnsureWidth(tb.width)

		// Place the character
		if tb.cursor.X < tb.width {
			tb.viewport[tb.cursor.Y].Set(tb.cursor.X, char)
			tb.cursor.MoveRight(char.Width)

			// Handle line wrapping
			if tb.cursor.X >= tb.width {
				tb.cursor.CarriageReturn()
				tb.cursor.LineFeed()
				tb.ensureCursorInBounds()
			}
		}
	}
}

// Execute handles control characters
func (tb *TerminalBuffer) Execute(b byte) {
	switch b {
	case 0x07: // BEL - Bell
		// Terminal bell - could trigger notification
	case 0x08: // BS - Backspace
		tb.cursor.MoveLeft(1)
		tb.ensureCursorInBounds()
	case 0x09: // HT - Horizontal Tab
		// Move to next tab stop (every 8 columns)
		nextTab := ((tb.cursor.X / 8) + 1) * 8
		if nextTab < tb.width {
			tb.cursor.X = nextTab
		} else {
			tb.cursor.X = tb.width - 1
		}
	case 0x0A: // LF - Line Feed
		tb.cursor.LineFeed()
		tb.ensureCursorInBounds()
	case 0x0D: // CR - Carriage Return
		tb.cursor.CarriageReturn()
	case 0x0E: // SO - Shift Out (activate G1 charset)
		// Character set handling - could be implemented
	case 0x0F: // SI - Shift In (activate G0 charset)
		// Character set handling - could be implemented
	}
}

// Hook handles DCS sequence start
func (tb *TerminalBuffer) Hook(params *govte.Params, intermediates []byte, ignore bool, action rune) {
	// Device Control String handling - could be implemented for special features
}

// Put handles DCS data
func (tb *TerminalBuffer) Put(b byte) {
	// DCS data handling
}

// Unhook handles DCS sequence end
func (tb *TerminalBuffer) Unhook() {
	// DCS cleanup
}

// OscDispatch handles Operating System Command sequences
func (tb *TerminalBuffer) OscDispatch(params [][]byte, bellTerminated bool) {
	if len(params) == 0 {
		return
	}

	// Parse OSC command
	if len(params[0]) == 0 {
		return
	}

	cmd := string(params[0])

	// Handle different OSC commands
	switch cmd {
	case "0", "2": // Set window title
		if len(params) > 1 {
			title := string(params[1])
			tb.title = &title
		}
	case "1": // Set icon name (similar to title)
		if len(params) > 1 {
			title := string(params[1])
			tb.title = &title
		}
	}
}

// CsiDispatch handles CSI escape sequences
func (tb *TerminalBuffer) CsiDispatch(params *govte.Params, intermediates []byte, ignore bool, action rune) {
	if ignore {
		return
	}

	// Convert params to [][]uint16 for easier processing
	var paramGroups [][]uint16
	if params != nil {
		paramGroups = params.Iter()
	}

	switch action {
	case 'H', 'f': // CUP - Cursor Position
		row, col := 1, 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			row = int(paramGroups[0][0])
		}
		if len(paramGroups) > 1 && len(paramGroups[1]) > 0 {
			col = int(paramGroups[1][0])
		}

		// Convert to 0-based and clamp to screen bounds
		tb.cursor.X = min(col-1, tb.width-1)
		tb.cursor.Y = min(row-1, tb.height-1)
		tb.ensureCursorInBounds()

	case 'A': // CUU - Cursor Up
		lines := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 && paramGroups[0][0] > 0 {
			lines = int(paramGroups[0][0])
		}
		tb.cursor.MoveUp(lines)
		tb.ensureCursorInBounds()

	case 'B': // CUD - Cursor Down
		lines := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 && paramGroups[0][0] > 0 {
			lines = int(paramGroups[0][0])
		}
		tb.cursor.MoveDown(lines)
		tb.ensureCursorInBounds()

	case 'C': // CUF - Cursor Forward
		cols := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 && paramGroups[0][0] > 0 {
			cols = int(paramGroups[0][0])
		}
		tb.cursor.MoveRight(cols)
		tb.ensureCursorInBounds()

	case 'D': // CUB - Cursor Back
		cols := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 && paramGroups[0][0] > 0 {
			cols = int(paramGroups[0][0])
		}
		tb.cursor.MoveLeft(cols)
		tb.ensureCursorInBounds()

	case 'G': // CHA - Cursor Horizontal Absolute
		col := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			col = int(paramGroups[0][0])
		}
		tb.cursor.X = min(col-1, tb.width-1)
		tb.ensureCursorInBounds()

	case 'd': // VPA - Vertical Position Absolute
		row := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			row = int(paramGroups[0][0])
		}
		tb.cursor.Y = min(row-1, tb.height-1)
		tb.ensureCursorInBounds()

	case 'J': // ED - Erase in Display
		mode := 0
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			mode = int(paramGroups[0][0])
		}
		tb.eraseInDisplay(mode)

	case 'K': // EL - Erase in Line
		mode := 0
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			mode = int(paramGroups[0][0])
		}
		tb.eraseInLine(mode)

	case 'm': // SGR - Select Graphic Rendition
		tb.currentStyles.AddStyleFromAnsiParams(paramGroups)
		tb.cursor.PendingStyles = tb.currentStyles

	case 'r': // DECSTBM - Set Top and Bottom Margins
		top, bottom := 1, tb.height
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			top = int(paramGroups[0][0])
		}
		if len(paramGroups) > 1 && len(paramGroups[1]) > 0 {
			bottom = int(paramGroups[1][0])
		}

		if top < bottom && top >= 1 && bottom <= tb.height {
			tb.scrollRegion = &ScrollRegion{
				top:    top - 1, // Convert to 0-based
				bottom: bottom - 1,
			}
		}

	case 's': // SCOSC - Save Cursor Position
		saved := tb.cursor.SavePosition()
		tb.savedCursor = &saved

	case 'u': // SCORC - Restore Cursor Position
		if tb.savedCursor != nil {
			tb.cursor.RestorePosition(*tb.savedCursor)
			tb.currentStyles = tb.cursor.PendingStyles
		}

	case 'S': // SU - Scroll Up
		lines := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			lines = int(paramGroups[0][0])
		}
		tb.scrollUp(lines)

	case 'T': // SD - Scroll Down
		lines := 1
		if len(paramGroups) > 0 && len(paramGroups[0]) > 0 {
			lines = int(paramGroups[0][0])
		}
		tb.scrollDown(lines)
	}
}

// EscDispatch handles escape sequences
func (tb *TerminalBuffer) EscDispatch(intermediates []byte, ignore bool, b byte) {
	if ignore {
		return
	}

	switch b {
	case 'D': // IND - Index (move cursor down, scroll if needed)
		tb.cursor.LineFeed()
		tb.ensureCursorInBounds()
	case 'M': // RI - Reverse Index (move cursor up, scroll if needed)
		tb.cursor.MoveUp(1)
		tb.ensureCursorInBounds()
	case '7': // DECSC - Save Cursor
		saved := tb.cursor.SavePosition()
		tb.savedCursor = &saved
	case '8': // DECRC - Restore Cursor
		if tb.savedCursor != nil {
			tb.cursor.RestorePosition(*tb.savedCursor)
			tb.currentStyles = tb.cursor.PendingStyles
		}
	case 'c': // RIS - Reset to Initial State
		tb.reset()
	case 'E': // NEL - Next Line
		tb.cursor.NewLine()
		tb.ensureCursorInBounds()
	}
}

// Helper methods

// ensureCursorInBounds ensures cursor position is within screen bounds
func (tb *TerminalBuffer) ensureCursorInBounds() {
	if tb.cursor.X < 0 {
		tb.cursor.X = 0
	}
	if tb.cursor.X >= tb.width {
		tb.cursor.X = tb.width - 1
	}
	if tb.cursor.Y < 0 {
		tb.cursor.Y = 0
	}
	if tb.cursor.Y >= tb.height {
		tb.cursor.Y = tb.height - 1
	}
}

// eraseInDisplay handles ED command
func (tb *TerminalBuffer) eraseInDisplay(mode int) {
	emptyChar := EmptyTerminalCharacter()

	switch mode {
	case 0: // Clear from cursor to end of display
		// Clear from cursor to end of current line
		if tb.cursor.Y < len(tb.viewport) {
			for x := tb.cursor.X; x < tb.width; x++ {
				tb.viewport[tb.cursor.Y].Set(x, emptyChar)
			}
		}
		// Clear all lines below current line
		for y := tb.cursor.Y + 1; y < len(tb.viewport); y++ {
			tb.viewport[y].Clear()
		}

	case 1: // Clear from beginning of display to cursor
		// Clear all lines above current line
		for y := 0; y < tb.cursor.Y && y < len(tb.viewport); y++ {
			tb.viewport[y].Clear()
		}
		// Clear from beginning of current line to cursor
		if tb.cursor.Y < len(tb.viewport) {
			for x := 0; x <= tb.cursor.X && x < tb.width; x++ {
				tb.viewport[tb.cursor.Y].Set(x, emptyChar)
			}
		}

	case 2, 3: // Clear entire display
		for y := range tb.viewport {
			tb.viewport[y].Clear()
		}
	}
}

// eraseInLine handles EL command
func (tb *TerminalBuffer) eraseInLine(mode int) {
	if tb.cursor.Y >= len(tb.viewport) {
		return
	}

	emptyChar := EmptyTerminalCharacter()
	row := &tb.viewport[tb.cursor.Y]

	switch mode {
	case 0: // Clear from cursor to end of line
		for x := tb.cursor.X; x < tb.width; x++ {
			row.Set(x, emptyChar)
		}

	case 1: // Clear from beginning of line to cursor
		for x := 0; x <= tb.cursor.X && x < tb.width; x++ {
			row.Set(x, emptyChar)
		}

	case 2: // Clear entire line
		row.Clear()
	}
}

// scrollUp scrolls the display up by n lines
func (tb *TerminalBuffer) scrollUp(lines int) {
	if lines <= 0 {
		return
	}

	// Determine scroll region
	top := 0
	bottom := tb.height - 1
	if tb.scrollRegion != nil {
		top = tb.scrollRegion.top
		bottom = tb.scrollRegion.bottom
	}

	// Shift lines up within scroll region
	for i := 0; i < lines; i++ {
		if top < bottom {
			// Remove the top line and add a blank line at the bottom
			for y := top; y < bottom; y++ {
				if y+1 < len(tb.viewport) {
					tb.viewport[y] = tb.viewport[y+1]
				}
			}
			// Add blank line at bottom of scroll region
			if bottom < len(tb.viewport) {
				tb.viewport[bottom] = NewRowWithWidth(tb.width)
			}
		}
	}
}

// scrollDown scrolls the display down by n lines
func (tb *TerminalBuffer) scrollDown(lines int) {
	if lines <= 0 {
		return
	}

	// Determine scroll region
	top := 0
	bottom := tb.height - 1
	if tb.scrollRegion != nil {
		top = tb.scrollRegion.top
		bottom = tb.scrollRegion.bottom
	}

	// Shift lines down within scroll region
	for i := 0; i < lines; i++ {
		if top < bottom {
			// Shift lines down
			for y := bottom; y > top; y-- {
				if y-1 >= 0 && y < len(tb.viewport) {
					tb.viewport[y] = tb.viewport[y-1]
				}
			}
			// Add blank line at top of scroll region
			if top < len(tb.viewport) {
				tb.viewport[top] = NewRowWithWidth(tb.width)
			}
		}
	}
}

// reset resets the terminal to initial state
func (tb *TerminalBuffer) reset() {
	tb.cursor = NewCursor()
	tb.currentStyles = DefaultCharacterStyles()
	tb.savedCursor = nil
	tb.scrollRegion = nil
	tb.title = nil

	// Clear all content
	for i := range tb.viewport {
		tb.viewport[i] = NewRowWithWidth(tb.width)
	}
}
