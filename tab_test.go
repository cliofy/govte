package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TabHandler is a test handler that tracks tab operations
type TabHandler struct {
	NoopHandler
	tabStops      map[int]bool
	cursorCol     int
	tabOperations []TabOperation
}

// TabOperation represents a tab-related operation for testing
type TabOperation struct {
	Type   string
	Column int
	Count  int
}

// Tab implements Handler for basic tab movement
func (h *TabHandler) Tab() {
	h.tabOperations = append(h.tabOperations, TabOperation{Type: "Tab", Column: h.cursorCol})
	h.cursorCol = h.nextTabStop(h.cursorCol)
}

// SetTabStop implements Handler for setting tab stops
func (h *TabHandler) SetTabStop() {
	if h.tabStops == nil {
		h.tabStops = make(map[int]bool)
	}
	h.tabStops[h.cursorCol] = true
	h.tabOperations = append(h.tabOperations, TabOperation{Type: "SetTabStop", Column: h.cursorCol})
}

// ClearTabStop implements Handler for clearing tab stops
func (h *TabHandler) ClearTabStop(mode TabulationClearMode) {
	switch mode {
	case TabClearCurrent:
		delete(h.tabStops, h.cursorCol)
		h.tabOperations = append(h.tabOperations, TabOperation{Type: "ClearTabStop", Column: h.cursorCol})
	case TabClearAll:
		h.tabStops = make(map[int]bool)
		h.tabOperations = append(h.tabOperations, TabOperation{Type: "ClearAllTabStops", Column: -1})
	}
}

// TabForward implements Handler for forward tab movement
func (h *TabHandler) TabForward(count int) {
	startCol := h.cursorCol
	for i := 0; i < count; i++ {
		h.cursorCol = h.nextTabStop(h.cursorCol)
	}
	h.tabOperations = append(h.tabOperations, TabOperation{Type: "TabForward", Column: startCol, Count: count})
}

// TabBackward implements Handler for backward tab movement
func (h *TabHandler) TabBackward(count int) {
	startCol := h.cursorCol
	for i := 0; i < count; i++ {
		h.cursorCol = h.prevTabStop(h.cursorCol)
	}
	h.tabOperations = append(h.tabOperations, TabOperation{Type: "TabBackward", Column: startCol, Count: count})
}

// Goto implements Handler to track cursor position
func (h *TabHandler) Goto(line, col int) {
	h.cursorCol = col
}

// GotoCol implements Handler to track cursor column
func (h *TabHandler) GotoCol(col int) {
	h.cursorCol = col
}

// nextTabStop finds the next tab stop after the given column
func (h *TabHandler) nextTabStop(col int) int {
	if h.tabStops == nil {
		// Default tab stops every 8 columns
		return ((col / 8) + 1) * 8
	}
	
	// Find next set tab stop
	for i := col + 1; i <= 120; i++ { // Reasonable terminal width limit
		if h.tabStops[i] {
			return i
		}
	}
	
	// If no custom tab stops found, use default
	return ((col / 8) + 1) * 8
}

// prevTabStop finds the previous tab stop before the given column
func (h *TabHandler) prevTabStop(col int) int {
	if col <= 1 {
		return 1
	}
	
	if h.tabStops == nil {
		// Default tab stops every 8 columns
		prevStop := ((col - 1) / 8) * 8
		if prevStop == 0 {
			return 1
		}
		return prevStop
	}
	
	// Find previous set tab stop
	for i := col - 1; i >= 1; i-- {
		if h.tabStops[i] {
			return i
		}
	}
	
	return 1
}

func TestTabClearModeEnum(t *testing.T) {
	tests := []struct {
		name     string
		mode     TabulationClearMode
		expected string
	}{
		{"Clear current tab stop", TabClearCurrent, "TabClearCurrent"},
		{"Clear all tab stops", TabClearAll, "TabClearAll"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

func TestBasicTabMovement(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{cursorCol: 1}

	// Test basic tab character (HT)
	processor.Advance(handler, []byte("\t"))

	assert.Len(t, handler.tabOperations, 1)
	assert.Equal(t, "Tab", handler.tabOperations[0].Type)
	assert.Equal(t, 1, handler.tabOperations[0].Column)
	assert.Equal(t, 8, handler.cursorCol) // Default tab stop at column 8
}

func TestTabStopSetting(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{cursorCol: 10}

	// Test HTS (Horizontal Tab Set) - ESC H
	processor.Advance(handler, []byte("\x1bH"))

	assert.Len(t, handler.tabOperations, 1)
	assert.Equal(t, "SetTabStop", handler.tabOperations[0].Type)
	assert.Equal(t, 10, handler.tabOperations[0].Column)
	assert.True(t, handler.tabStops[10])
}

func TestTabStopClearing(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{
		cursorCol: 10,
		tabStops:  map[int]bool{5: true, 10: true, 15: true},
	}

	tests := []struct {
		name            string
		sequence        string
		expectedType    string
		expectedColumn  int
		remainingStops  map[int]bool
	}{
		{
			name:           "Clear current tab stop",
			sequence:       "\x1b[0g",
			expectedType:   "ClearTabStop",
			expectedColumn: 10,
			remainingStops: map[int]bool{5: true, 15: true},
		},
		{
			name:           "Clear all tab stops",
			sequence:       "\x1b[3g",
			expectedType:   "ClearAllTabStops",
			expectedColumn: -1,
			remainingStops: map[int]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset handler state
			handler.tabOperations = nil
			handler.tabStops = map[int]bool{5: true, 10: true, 15: true}
			
			processor.Advance(handler, []byte(tt.sequence))

			assert.Len(t, handler.tabOperations, 1)
			assert.Equal(t, tt.expectedType, handler.tabOperations[0].Type)
			assert.Equal(t, tt.expectedColumn, handler.tabOperations[0].Column)
			assert.Equal(t, tt.remainingStops, handler.tabStops)
		})
	}
}

func TestCursorHorizontalTab(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{cursorCol: 1}

	tests := []struct {
		name           string
		sequence       string
		expectedType   string
		expectedCount  int
		expectedCol    int
	}{
		{
			name:          "CHT with default parameter (1)",
			sequence:      "\x1b[I",
			expectedType:  "TabForward",
			expectedCount: 1,
			expectedCol:   8,
		},
		{
			name:          "CHT with parameter 3",
			sequence:      "\x1b[3I",
			expectedType:  "TabForward",
			expectedCount: 3,
			expectedCol:   24, // 1 -> 8 -> 16 -> 24
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.tabOperations = nil
			handler.cursorCol = 1
			
			processor.Advance(handler, []byte(tt.sequence))

			assert.Len(t, handler.tabOperations, 1)
			assert.Equal(t, tt.expectedType, handler.tabOperations[0].Type)
			assert.Equal(t, tt.expectedCount, handler.tabOperations[0].Count)
			assert.Equal(t, tt.expectedCol, handler.cursorCol)
		})
	}
}

func TestCursorBackwardTab(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{cursorCol: 25}

	tests := []struct {
		name           string
		sequence       string
		expectedType   string
		expectedCount  int
		expectedCol    int
	}{
		{
			name:          "CBT with default parameter (1)",
			sequence:      "\x1b[Z",
			expectedType:  "TabBackward",
			expectedCount: 1,
			expectedCol:   24, // 25 -> 24 (previous 8-column boundary)
		},
		{
			name:          "CBT with parameter 2",
			sequence:      "\x1b[2Z",
			expectedType:  "TabBackward",
			expectedCount: 2,
			expectedCol:   16, // 25 -> 24 -> 16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.tabOperations = nil
			handler.cursorCol = 25
			
			processor.Advance(handler, []byte(tt.sequence))

			assert.Len(t, handler.tabOperations, 1)
			assert.Equal(t, tt.expectedType, handler.tabOperations[0].Type)
			assert.Equal(t, tt.expectedCount, handler.tabOperations[0].Count)
			assert.Equal(t, tt.expectedCol, handler.cursorCol)
		})
	}
}

func TestCustomTabStops(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{cursorCol: 1}

	// Set custom tab stops at columns 5, 12, 20
	handler.cursorCol = 5
	processor.Advance(handler, []byte("\x1bH"))
	
	handler.cursorCol = 12
	processor.Advance(handler, []byte("\x1bH"))
	
	handler.cursorCol = 20
	processor.Advance(handler, []byte("\x1bH"))

	// Now test tab movement with custom stops
	handler.cursorCol = 1
	handler.tabOperations = nil

	// Tab forward should go to column 5 (first custom stop)
	processor.Advance(handler, []byte("\t"))
	assert.Equal(t, 5, handler.cursorCol)

	// Tab forward again should go to column 12
	processor.Advance(handler, []byte("\t"))
	assert.Equal(t, 12, handler.cursorCol)

	// Tab forward again should go to column 20
	processor.Advance(handler, []byte("\t"))
	assert.Equal(t, 20, handler.cursorCol)
}

func TestTabIntegration(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &TabHandler{cursorCol: 1}

	// Complete scenario: set custom tab stops, move, and clear
	
	// Set tab stops at columns 10 and 20
	handler.cursorCol = 10
	processor.Advance(handler, []byte("\x1bH"))
	
	handler.cursorCol = 20
	processor.Advance(handler, []byte("\x1bH"))

	// Move to start and tab forward
	handler.cursorCol = 1
	handler.tabOperations = nil
	
	// Use CHT to move forward 2 tab stops (should go to column 20)
	processor.Advance(handler, []byte("\x1b[2I"))
	assert.Equal(t, 20, handler.cursorCol)

	// Use CBT to move backward 1 tab stop (should go to column 10)
	processor.Advance(handler, []byte("\x1b[Z"))
	assert.Equal(t, 10, handler.cursorCol)

	// Clear current tab stop (column 10)
	processor.Advance(handler, []byte("\x1b[0g"))
	assert.False(t, handler.tabStops[10])
	assert.True(t, handler.tabStops[20])

	// Clear all tab stops
	processor.Advance(handler, []byte("\x1b[3g"))
	assert.Empty(t, handler.tabStops)
}