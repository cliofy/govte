package govte

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessorCreation(t *testing.T) {
	p := NewProcessor(&NoopHandler{})
	assert.NotNil(t, p)
	assert.NotNil(t, p.parser)
	assert.NotNil(t, p.syncState)
	assert.Equal(t, 150*time.Millisecond, p.syncState.timeout)
}

func TestProcessorBasicText(t *testing.T) {
	p := NewProcessor(&NoopHandler{})
	h := NewTestHandler()

	p.Advance(h, []byte("Hello"))

	assert.Equal(t, []rune{'H', 'e', 'l', 'l', 'o'}, h.inputChars)
}

func TestProcessorControlCharacters(t *testing.T) {
	p := NewProcessor(&NoopHandler{})
	h := NewTestHandler()

	// Test various control characters
	p.Advance(h, []byte("\x07")) // BEL
	assert.Equal(t, 1, h.bellCount)

	p.Advance(h, []byte("\x08")) // BS
	// Backspace doesn't have a test handler method, but it shouldn't panic

	p.Advance(h, []byte("\x0A")) // LF
	assert.Equal(t, 1, h.lineFeedCount)

	p.Advance(h, []byte("\x0D")) // CR
	assert.Equal(t, 1, h.carriageReturns)
}

func TestProcessorCursorMovement(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
		checkFn  func(*testing.T, *TestHandler)
	}{
		{
			name:     "Cursor up",
			sequence: "\x1b[5A",
			checkFn: func(t *testing.T, h *TestHandler) {
				// MoveUp should be called with 5
			},
		},
		{
			name:     "Cursor position",
			sequence: "\x1b[10;20H",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Equal(t, 10, h.cursorPos.line)
				assert.Equal(t, 20, h.cursorPos.col)
			},
		},
		{
			name:     "Cursor position with defaults",
			sequence: "\x1b[H",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Equal(t, 1, h.cursorPos.line)
				assert.Equal(t, 1, h.cursorPos.col)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(&NoopHandler{})
			h := NewTestHandler()

			p.Advance(h, []byte(tt.sequence))
			tt.checkFn(t, h)
		})
	}
}

func TestProcessorColors(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
		checkFn  func(*testing.T, *TestHandler)
	}{
		{
			name:     "Simple foreground color",
			sequence: "\x1b[31m",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Len(t, h.foregroundColors, 1)
				assert.Equal(t, ColorTypeNamed, h.foregroundColors[0].Type)
				assert.Equal(t, Red, h.foregroundColors[0].Named)
			},
		},
		{
			name:     "Simple background color",
			sequence: "\x1b[44m",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Len(t, h.backgroundColors, 1)
				assert.Equal(t, ColorTypeNamed, h.backgroundColors[0].Type)
				assert.Equal(t, Blue, h.backgroundColors[0].Named)
			},
		},
		{
			name:     "RGB foreground color",
			sequence: "\x1b[38:2:255:128:64m",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Len(t, h.foregroundColors, 1)
				assert.Equal(t, ColorTypeRgb, h.foregroundColors[0].Type)
				assert.Equal(t, uint8(255), h.foregroundColors[0].Rgb.R)
				assert.Equal(t, uint8(128), h.foregroundColors[0].Rgb.G)
				assert.Equal(t, uint8(64), h.foregroundColors[0].Rgb.B)
			},
		},
		{
			name:     "256-color palette",
			sequence: "\x1b[38:5:128m",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Len(t, h.foregroundColors, 1)
				assert.Equal(t, ColorTypeIndexed, h.foregroundColors[0].Type)
				assert.Equal(t, uint8(128), h.foregroundColors[0].Index)
			},
		},
		{
			name:     "Bright colors",
			sequence: "\x1b[91m",
			checkFn: func(t *testing.T, h *TestHandler) {
				assert.Len(t, h.foregroundColors, 1)
				assert.Equal(t, ColorTypeNamed, h.foregroundColors[0].Type)
				assert.Equal(t, BrightRed, h.foregroundColors[0].Named)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(&NoopHandler{})
			h := NewTestHandler()

			p.Advance(h, []byte(tt.sequence))
			tt.checkFn(t, h)
		})
	}
}

func TestProcessorAttributes(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
		expected []Attr
	}{
		{"Bold", "\x1b[1m", []Attr{AttrBold}},
		{"Italic", "\x1b[3m", []Attr{AttrItalic}},
		{"Underline", "\x1b[4m", []Attr{AttrUnderline}},
		{"Multiple", "\x1b[1;3;4m", []Attr{AttrBold, AttrItalic, AttrUnderline}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(&NoopHandler{})
			h := NewTestHandler()

			p.Advance(h, []byte(tt.sequence))
			assert.Equal(t, tt.expected, h.attributes)
		})
	}
}

func TestProcessorClearOperations(t *testing.T) {
	tests := []struct {
		name           string
		sequence       string
		expectedLines  []LineClearMode
		expectedScreen []ClearMode
	}{
		{
			name:          "Clear line right",
			sequence:      "\x1b[K",
			expectedLines: []LineClearMode{LineClearRight},
		},
		{
			name:          "Clear line left",
			sequence:      "\x1b[1K",
			expectedLines: []LineClearMode{LineClearLeft},
		},
		{
			name:          "Clear entire line",
			sequence:      "\x1b[2K",
			expectedLines: []LineClearMode{LineClearAll},
		},
		{
			name:           "Clear screen below",
			sequence:       "\x1b[J",
			expectedScreen: []ClearMode{ClearBelow},
		},
		{
			name:           "Clear screen above",
			sequence:       "\x1b[1J",
			expectedScreen: []ClearMode{ClearAbove},
		},
		{
			name:           "Clear entire screen",
			sequence:       "\x1b[2J",
			expectedScreen: []ClearMode{ClearAll},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(&NoopHandler{})
			h := NewTestHandler()

			p.Advance(h, []byte(tt.sequence))

			if tt.expectedLines != nil {
				assert.Equal(t, tt.expectedLines, h.clearedLines)
			}
			if tt.expectedScreen != nil {
				assert.Equal(t, tt.expectedScreen, h.clearedScreens)
			}
		})
	}
}

func TestProcessorModes(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
		mode     Mode
		enabled  bool
	}{
		{
			name:     "Set private mode",
			sequence: "\x1b[?25h",
			mode:     ModeShowCursor,
			enabled:  true,
		},
		{
			name:     "Reset private mode",
			sequence: "\x1b[?25l",
			mode:     ModeShowCursor,
			enabled:  false,
		},
		{
			name:     "Set standard mode",
			sequence: "\x1b[4h",
			mode:     ModeInsert,
			enabled:  true,
		},
		{
			name:     "Reset standard mode",
			sequence: "\x1b[4l",
			mode:     ModeInsert,
			enabled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(&NoopHandler{})
			h := NewTestHandler()

			p.Advance(h, []byte(tt.sequence))

			val, exists := h.modes[tt.mode]
			assert.True(t, exists)
			assert.Equal(t, tt.enabled, val)
		})
	}
}

func TestProcessorOSC(t *testing.T) {
	tests := []struct {
		name          string
		sequence      string
		expectedTitle string
	}{
		{
			name:          "Set window title with BEL",
			sequence:      "\x1b]0;Test Title\x07",
			expectedTitle: "Test Title",
		},
		{
			name:          "Set window title with ST",
			sequence:      "\x1b]2;Another Title\x1b\\",
			expectedTitle: "Another Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcessor(&NoopHandler{})
			h := NewTestHandler()

			p.Advance(h, []byte(tt.sequence))
			assert.Equal(t, tt.expectedTitle, h.title)
		})
	}
}

func TestProcessorReset(t *testing.T) {
	p := NewProcessor(&NoopHandler{})

	// Modify some state
	p.Advance(&NoopHandler{}, []byte("Test"))

	// Reset
	p.Reset()

	assert.NotNil(t, p.parser)
	assert.False(t, p.syncState.enabled)
	assert.Empty(t, p.syncState.buffer)
}

func TestProcessorSyncTimeout(t *testing.T) {
	p := NewProcessor(&NoopHandler{})

	// Set custom timeout
	p.SetSyncTimeout(200 * time.Millisecond)
	assert.Equal(t, 200*time.Millisecond, p.syncState.timeout)
}

func TestGetParam(t *testing.T) {
	groups := [][]uint16{
		{1, 2, 3},
		{4},
		{5, 6},
	}

	tests := []struct {
		groupIdx     int
		paramIdx     int
		defaultValue int
		expected     int
	}{
		{0, 0, 10, 1},  // First param of first group
		{0, 1, 10, 2},  // Second param of first group
		{0, 2, 10, 3},  // Third param of first group
		{1, 0, 10, 4},  // First param of second group
		{2, 1, 10, 6},  // Second param of third group
		{3, 0, 10, 10}, // Out of bounds group - use default
		{0, 5, 10, 10}, // Out of bounds param - use default
		{0, 0, 0, 1},   // Default is 0, value is non-zero
		{1, 1, 20, 20}, // Param doesn't exist - use default
	}

	for _, tt := range tests {
		result := getParam(groups, tt.groupIdx, tt.paramIdx, tt.defaultValue)
		assert.Equal(t, tt.expected, result)
	}
}

func TestMinUint16(t *testing.T) {
	assert.Equal(t, uint16(5), minUint16(5, 10))
	assert.Equal(t, uint16(3), minUint16(10, 3))
	assert.Equal(t, uint16(7), minUint16(7, 7))
	assert.Equal(t, uint16(0), minUint16(0, 100))
	assert.Equal(t, uint16(255), minUint16(1000, 255))
}
