// Package govte provides high-level terminal control interfaces.
package govte

// Handler defines high-level terminal operations.
// This interface provides semantic methods for terminal control,
// abstracting away the low-level escape sequence details.
type Handler interface {
	// Text and Display
	
	// Input handles a character to be displayed.
	Input(c rune)

	// Bell rings the terminal bell.
	Bell()

	// LineFeed moves cursor down one line.
	LineFeed()

	// CarriageReturn moves cursor to beginning of line.
	CarriageReturn()

	// Backspace moves cursor back one column.
	Backspace()

	// Tab moves cursor to next tab stop.
	Tab()

	// SetTabStop sets a tab stop at the current cursor position.
	SetTabStop()

	// ClearTabStop clears tab stops according to the specified mode.
	ClearTabStop(mode TabulationClearMode)

	// TabForward moves cursor forward by n tab stops.
	TabForward(count int)

	// TabBackward moves cursor backward by n tab stops.
	TabBackward(count int)

	// SetTitle sets the window title.
	SetTitle(title string)

	// Cursor Movement
	
	// Goto moves cursor to absolute position (1-based).
	Goto(line, col int)

	// GotoLine moves cursor to specific line (1-based).
	GotoLine(line int)

	// GotoCol moves cursor to specific column (1-based).
	GotoCol(col int)

	// MoveUp moves cursor up by n lines.
	MoveUp(lines int)

	// MoveDown moves cursor down by n lines.
	MoveDown(lines int)

	// MoveForward moves cursor forward by n columns.
	MoveForward(cols int)

	// MoveBackward moves cursor backward by n columns.
	MoveBackward(cols int)

	// MoveDownAndCR moves cursor down n lines and to column 1.
	MoveDownAndCR(lines int)

	// MoveUpAndCR moves cursor up n lines and to column 1.
	MoveUpAndCR(lines int)

	// SaveCursorPosition saves current cursor position.
	SaveCursorPosition()

	// RestoreCursorPosition restores saved cursor position.
	RestoreCursorPosition()

	// Text Modification
	
	// InsertBlank inserts n blank characters at cursor.
	InsertBlank(count int)

	// DeleteChars deletes n characters at cursor.
	DeleteChars(count int)

	// EraseChars erases n characters at cursor (replace with space).
	EraseChars(count int)

	// InsertLines inserts n blank lines at cursor line.
	InsertLines(count int)

	// DeleteLines deletes n lines at cursor line.
	DeleteLines(count int)

	// Screen Operations
	
	// ClearLine clears line according to mode.
	ClearLine(mode LineClearMode)

	// ClearScreen clears screen according to mode.
	ClearScreen(mode ClearMode)

	// ScrollUp scrolls screen up by n lines.
	ScrollUp(lines int)

	// ScrollDown scrolls screen down by n lines.
	ScrollDown(lines int)

	// SetScrollingRegion sets the scrolling region (1-based).
	SetScrollingRegion(top, bottom int)

	// Text Attributes
	
	// SetAttribute sets text rendering attribute.
	SetAttribute(attr Attr)

	// ResetAttributes resets all text attributes to default.
	ResetAttributes()

	// SetForeground sets foreground color.
	SetForeground(color Color)

	// SetBackground sets background color.
	SetBackground(color Color)

	// ResetColors resets colors to default.
	ResetColors()

	// Cursor Appearance
	
	// SetCursorStyle sets cursor appearance.
	SetCursorStyle(style CursorStyle)

	// SetCursorVisible sets cursor visibility.
	SetCursorVisible(visible bool)

	// Terminal Modes
	
	// SetMode enables a terminal mode.
	SetMode(mode Mode)

	// ResetMode disables a terminal mode.
	ResetMode(mode Mode)

	// Device Operations
	
	// DeviceStatus reports device status.
	DeviceStatus(kind int)

	// IdentifyTerminal identifies the terminal type.
	IdentifyTerminal()

	// Reset performs a soft terminal reset.
	Reset()

	// HardReset performs a hard terminal reset.
	HardReset()

	// Device Control String (DCS) Support

	// Hook is called when a DCS sequence begins.
	// params: parameters parsed from the DCS sequence
	// intermediates: intermediate characters
	// ignore: true if sequence should be ignored due to overflow
	// action: the final character that triggered the DCS
	Hook(params [][]uint16, intermediates []byte, ignore bool, action rune)

	// Put receives data bytes within a DCS sequence.
	// This is called for each data byte after Hook until Unhook.
	Put(data []byte)

	// Unhook is called when a DCS sequence ends.
	// This signals the completion of the DCS sequence.
	Unhook()

	// Character Set Support

	// ConfigureCharset configures a character set for a specific charset index.
	// index: the charset index (G0, G1, G2, G3)
	// charset: the standard charset to assign
	ConfigureCharset(index CharsetIndex, charset StandardCharset)

	// SetActiveCharset sets the active character set.
	// index: the charset index to activate
	SetActiveCharset(index CharsetIndex)
}

// NoopHandler is a no-op implementation of Handler.
// It can be embedded in custom handlers to avoid implementing all methods.
type NoopHandler struct{}

// Ensure NoopHandler implements Handler
var _ Handler = (*NoopHandler)(nil)

// Input implements Handler.
func (h *NoopHandler) Input(c rune) {}

// Bell implements Handler.
func (h *NoopHandler) Bell() {}

// LineFeed implements Handler.
func (h *NoopHandler) LineFeed() {}

// CarriageReturn implements Handler.
func (h *NoopHandler) CarriageReturn() {}

// Backspace implements Handler.
func (h *NoopHandler) Backspace() {}

// Tab implements Handler.
func (h *NoopHandler) Tab() {}

// SetTabStop implements Handler.
func (h *NoopHandler) SetTabStop() {}

// ClearTabStop implements Handler.
func (h *NoopHandler) ClearTabStop(mode TabulationClearMode) {}

// TabForward implements Handler.
func (h *NoopHandler) TabForward(count int) {}

// TabBackward implements Handler.
func (h *NoopHandler) TabBackward(count int) {}

// SetTitle implements Handler.
func (h *NoopHandler) SetTitle(title string) {}

// Goto implements Handler.
func (h *NoopHandler) Goto(line, col int) {}

// GotoLine implements Handler.
func (h *NoopHandler) GotoLine(line int) {}

// GotoCol implements Handler.
func (h *NoopHandler) GotoCol(col int) {}

// MoveUp implements Handler.
func (h *NoopHandler) MoveUp(lines int) {}

// MoveDown implements Handler.
func (h *NoopHandler) MoveDown(lines int) {}

// MoveForward implements Handler.
func (h *NoopHandler) MoveForward(cols int) {}

// MoveBackward implements Handler.
func (h *NoopHandler) MoveBackward(cols int) {}

// MoveDownAndCR implements Handler.
func (h *NoopHandler) MoveDownAndCR(lines int) {}

// MoveUpAndCR implements Handler.
func (h *NoopHandler) MoveUpAndCR(lines int) {}

// SaveCursorPosition implements Handler.
func (h *NoopHandler) SaveCursorPosition() {}

// RestoreCursorPosition implements Handler.
func (h *NoopHandler) RestoreCursorPosition() {}

// InsertBlank implements Handler.
func (h *NoopHandler) InsertBlank(count int) {}

// DeleteChars implements Handler.
func (h *NoopHandler) DeleteChars(count int) {}

// EraseChars implements Handler.
func (h *NoopHandler) EraseChars(count int) {}

// InsertLines implements Handler.
func (h *NoopHandler) InsertLines(count int) {}

// DeleteLines implements Handler.
func (h *NoopHandler) DeleteLines(count int) {}

// ClearLine implements Handler.
func (h *NoopHandler) ClearLine(mode LineClearMode) {}

// ClearScreen implements Handler.
func (h *NoopHandler) ClearScreen(mode ClearMode) {}

// ScrollUp implements Handler.
func (h *NoopHandler) ScrollUp(lines int) {}

// ScrollDown implements Handler.
func (h *NoopHandler) ScrollDown(lines int) {}

// SetScrollingRegion implements Handler.
func (h *NoopHandler) SetScrollingRegion(top, bottom int) {}

// SetAttribute implements Handler.
func (h *NoopHandler) SetAttribute(attr Attr) {}

// ResetAttributes implements Handler.
func (h *NoopHandler) ResetAttributes() {}

// SetForeground implements Handler.
func (h *NoopHandler) SetForeground(color Color) {}

// SetBackground implements Handler.
func (h *NoopHandler) SetBackground(color Color) {}

// ResetColors implements Handler.
func (h *NoopHandler) ResetColors() {}

// SetCursorStyle implements Handler.
func (h *NoopHandler) SetCursorStyle(style CursorStyle) {}

// SetCursorVisible implements Handler.
func (h *NoopHandler) SetCursorVisible(visible bool) {}

// SetMode implements Handler.
func (h *NoopHandler) SetMode(mode Mode) {}

// ResetMode implements Handler.
func (h *NoopHandler) ResetMode(mode Mode) {}

// DeviceStatus implements Handler.
func (h *NoopHandler) DeviceStatus(kind int) {}

// IdentifyTerminal implements Handler.
func (h *NoopHandler) IdentifyTerminal() {}

// Reset implements Handler.
func (h *NoopHandler) Reset() {}

// HardReset implements Handler.
func (h *NoopHandler) HardReset() {}

// Hook implements Handler.
func (h *NoopHandler) Hook(params [][]uint16, intermediates []byte, ignore bool, action rune) {}

// Put implements Handler.
func (h *NoopHandler) Put(data []byte) {}

// Unhook implements Handler.
func (h *NoopHandler) Unhook() {}

// ConfigureCharset implements Handler.
func (h *NoopHandler) ConfigureCharset(index CharsetIndex, charset StandardCharset) {}

// SetActiveCharset implements Handler.
func (h *NoopHandler) SetActiveCharset(index CharsetIndex) {}