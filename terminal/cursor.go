//! Terminal cursor management
//! Go port of the Rust implementation - simplified from Zellij's implementation

package terminal

// Cursor represents cursor position and state
type Cursor struct {
	X             int
	Y             int
	PendingStyles CharacterStyles
	Shape         CursorShape
	IsHidden      bool
}

// NewCursor creates a new cursor at the origin
func NewCursor() Cursor {
	return Cursor{
		X:             0,
		Y:             0,
		PendingStyles: DefaultCharacterStyles(),
		Shape:         CursorShapeBlock,
		IsHidden:      false,
	}
}

// Goto moves cursor to a specific position
func (c *Cursor) Goto(x, y int) {
	c.X = x
	c.Y = y
}

// MoveUp moves cursor up by n lines
func (c *Cursor) MoveUp(n int) {
	c.Y = max(0, c.Y-n)
}

// MoveDown moves cursor down by n lines
func (c *Cursor) MoveDown(n int) {
	c.Y += n
}

// MoveLeft moves cursor left by n columns
func (c *Cursor) MoveLeft(n int) {
	c.X = max(0, c.X-n)
}

// MoveRight moves cursor right by n columns
func (c *Cursor) MoveRight(n int) {
	c.X += n
}

// CarriageReturn moves cursor to beginning of line
func (c *Cursor) CarriageReturn() {
	c.X = 0
}

// LineFeed moves cursor to next line
func (c *Cursor) LineFeed() {
	c.Y++
}

// NewLine moves cursor to next line and beginning of line
func (c *Cursor) NewLine() {
	c.LineFeed()
	c.CarriageReturn()
}

// SavePosition saves current cursor position
func (c *Cursor) SavePosition() SavedCursor {
	return SavedCursor{
		X:      c.X,
		Y:      c.Y,
		Styles: c.PendingStyles,
	}
}

// RestorePosition restores cursor position from saved state
func (c *Cursor) RestorePosition(saved SavedCursor) {
	c.X = saved.X
	c.Y = saved.Y
	c.PendingStyles = saved.Styles
}

// ChangeShape changes cursor shape
func (c *Cursor) ChangeShape(shape CursorShape) {
	c.Shape = shape
}

// Show shows cursor
func (c *Cursor) Show() {
	c.IsHidden = false
}

// Hide hides cursor
func (c *Cursor) Hide() {
	c.IsHidden = true
}

// SavedCursor represents saved cursor state
type SavedCursor struct {
	X      int
	Y      int
	Styles CharacterStyles
}

// CursorShape represents cursor shape
type CursorShape int

const (
	CursorShapeBlock CursorShape = iota
	CursorShapeBeam
	CursorShapeUnderline
)

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
