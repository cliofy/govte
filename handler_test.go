package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandler implementation for testing
type TestHandler struct {
	NoopHandler
	
	// Track method calls
	inputChars       []rune
	bellCount        int
	lineFeedCount    int
	carriageReturns  int
	title            string
	cursorPos        struct{ line, col int }
	clearedLines     []LineClearMode
	clearedScreens   []ClearMode
	foregroundColors []Color
	backgroundColors []Color
	attributes       []Attr
	modes            map[Mode]bool
}

func NewTestHandler() *TestHandler {
	return &TestHandler{
		modes: make(map[Mode]bool),
	}
}

func (h *TestHandler) Input(c rune) {
	h.inputChars = append(h.inputChars, c)
}

func (h *TestHandler) Bell() {
	h.bellCount++
}

func (h *TestHandler) LineFeed() {
	h.lineFeedCount++
}

func (h *TestHandler) CarriageReturn() {
	h.carriageReturns++
}

func (h *TestHandler) SetTitle(title string) {
	h.title = title
}

func (h *TestHandler) Goto(line, col int) {
	h.cursorPos.line = line
	h.cursorPos.col = col
}

func (h *TestHandler) ClearLine(mode LineClearMode) {
	h.clearedLines = append(h.clearedLines, mode)
}

func (h *TestHandler) ClearScreen(mode ClearMode) {
	h.clearedScreens = append(h.clearedScreens, mode)
}

func (h *TestHandler) SetForeground(color Color) {
	h.foregroundColors = append(h.foregroundColors, color)
}

func (h *TestHandler) SetBackground(color Color) {
	h.backgroundColors = append(h.backgroundColors, color)
}

func (h *TestHandler) SetAttribute(attr Attr) {
	h.attributes = append(h.attributes, attr)
}

func (h *TestHandler) SetMode(mode Mode) {
	h.modes[mode] = true
}

func (h *TestHandler) ResetMode(mode Mode) {
	h.modes[mode] = false
}

// Tests

func TestNoopHandler(t *testing.T) {
	h := &NoopHandler{}
	
	// Test that all methods can be called without panicking
	h.Input('a')
	h.Bell()
	h.LineFeed()
	h.CarriageReturn()
	h.Backspace()
	h.Tab()
	h.SetTitle("test")
	h.Goto(1, 1)
	h.GotoLine(1)
	h.GotoCol(1)
	h.MoveUp(1)
	h.MoveDown(1)
	h.MoveForward(1)
	h.MoveBackward(1)
	h.MoveDownAndCR(1)
	h.MoveUpAndCR(1)
	h.SaveCursorPosition()
	h.RestoreCursorPosition()
	h.InsertBlank(1)
	h.DeleteChars(1)
	h.EraseChars(1)
	h.InsertLines(1)
	h.DeleteLines(1)
	h.ClearLine(LineClearRight)
	h.ClearScreen(ClearBelow)
	h.ScrollUp(1)
	h.ScrollDown(1)
	h.SetScrollingRegion(1, 24)
	h.SetAttribute(AttrBold)
	h.ResetAttributes()
	h.SetForeground(NewNamedColor(Red))
	h.SetBackground(NewNamedColor(Blue))
	h.ResetColors()
	h.SetCursorStyle(CursorStyle{Shape: CursorShapeBlock})
	h.SetCursorVisible(true)
	h.SetMode(ModeInsert)
	h.ResetMode(ModeInsert)
	h.DeviceStatus(5)
	h.IdentifyTerminal()
	h.Reset()
	h.HardReset()
	
	// If we got here without panicking, test passes
	assert.True(t, true)
}

func TestHandlerInterface(t *testing.T) {
	// Ensure NoopHandler implements Handler
	var _ Handler = (*NoopHandler)(nil)
	
	// Ensure TestHandler implements Handler
	var _ Handler = (*TestHandler)(nil)
}