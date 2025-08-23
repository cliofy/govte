//! Terminal row implementation
//! Go port of the Rust implementation - simplified from Zellij's implementation

package terminal

import "strings"

// Row represents a single row in the terminal buffer
type Row struct {
	Columns     []TerminalCharacter
	IsCanonical bool
}

// NewRow creates a new empty row
func NewRow() Row {
	return Row{
		Columns:     []TerminalCharacter{},
		IsCanonical: false,
	}
}

// NewRowWithWidth creates a row with a specific width filled with spaces
func NewRowWithWidth(width int) Row {
	columns := make([]TerminalCharacter, width)
	emptyChar := EmptyTerminalCharacter()
	for i := range columns {
		columns[i] = emptyChar
	}
	return Row{
		Columns:     columns,
		IsCanonical: true,
	}
}

// Canonical marks this row as canonical (complete line)
func (r Row) Canonical() Row {
	r.IsCanonical = true
	return r
}

// Width gets the display width of this row
func (r *Row) Width() int {
	width := 0
	for _, c := range r.Columns {
		width += c.Width
	}
	return width
}

// Push adds a character to the end of the row
func (r *Row) Push(character TerminalCharacter) {
	r.Columns = append(r.Columns, character)
}

// Get gets a character at a specific column
func (r *Row) Get(index int) *TerminalCharacter {
	if index < 0 || index >= len(r.Columns) {
		return nil
	}
	return &r.Columns[index]
}

// GetMut gets a mutable reference to a character at a specific column
func (r *Row) GetMut(index int) *TerminalCharacter {
	if index < 0 || index >= len(r.Columns) {
		return nil
	}
	return &r.Columns[index]
}

// Set sets a character at a specific column
func (r *Row) Set(index int, character TerminalCharacter) {
	if index >= 0 && index < len(r.Columns) {
		r.Columns[index] = character
	}
}

// Clear clears the row (fill with spaces)
func (r *Row) Clear() {
	emptyChar := EmptyTerminalCharacter()
	for i := range r.Columns {
		r.Columns[i] = emptyChar
	}
}

// Truncate truncates the row to a specific length
func (r *Row) Truncate(length int) {
	if length < len(r.Columns) {
		r.Columns = r.Columns[:length]
	}
}

// EnsureWidth ensures the row has at least the specified width
func (r *Row) EnsureWidth(width int) {
	emptyChar := EmptyTerminalCharacter()
	for len(r.Columns) < width {
		r.Columns = append(r.Columns, emptyChar)
	}
}

// ToString converts the row to a string
func (r *Row) ToString() string {
	var result strings.Builder
	for _, c := range r.Columns {
		result.WriteRune(c.Character)
	}
	return result.String()
}

// VisibleWidth gets the visible width of the row (excluding trailing spaces)
func (r *Row) VisibleWidth() int {
	lastNonSpace := -1

	// Find the last non-space character
	for i, character := range r.Columns {
		if character.Character != ' ' {
			lastNonSpace = i
		}
	}

	if lastNonSpace == -1 {
		return 0
	}

	// Calculate width up to the last non-space character
	width := 0
	for i := 0; i <= lastNonSpace; i++ {
		width += r.Columns[i].Width
	}

	return width
}

// ReplaceRange replaces a range of characters with a single character
func (r *Row) ReplaceRange(start, end int, character TerminalCharacter) {
	if start < 0 {
		start = 0
	}
	if end > len(r.Columns) {
		end = len(r.Columns)
	}

	for i := start; i < end; i++ {
		r.Columns[i] = character
	}
}

// Len returns the number of columns in the row
func (r *Row) Len() int {
	return len(r.Columns)
}

// IsEmpty checks if the row is empty (no columns)
func (r *Row) IsEmpty() bool {
	return len(r.Columns) == 0
}

// Clone creates a deep copy of the row
func (r *Row) Clone() Row {
	columns := make([]TerminalCharacter, len(r.Columns))
	copy(columns, r.Columns)
	return Row{
		Columns:     columns,
		IsCanonical: r.IsCanonical,
	}
}
