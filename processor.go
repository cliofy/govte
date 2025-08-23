package govte

import (
	"io"
	"time"
)

// SyncState manages synchronized update state.
type SyncState struct {
	enabled   bool
	buffer    []byte
	startTime time.Time
	timeout   time.Duration
}

// DCSState manages DCS sequence state.
type DCSState struct {
	active bool
	buffer []byte
}

// Processor wraps a Parser and provides high-level terminal operations.
// It translates low-level Performer callbacks into Handler method calls.
type Processor struct {
	parser    *Parser
	handler   Handler
	output    io.Writer
	syncState *SyncState
	dcsState  *DCSState
	modes     map[Mode]bool
}

// NewProcessor creates a new Processor with a handler.
func NewProcessor(handler Handler) *Processor {
	return &Processor{
		parser:  NewParser(),
		handler: handler,
		modes:   make(map[Mode]bool),
		syncState: &SyncState{
			timeout: 150 * time.Millisecond, // Default timeout
		},
		dcsState: &DCSState{
			active: false,
			buffer: make([]byte, 0),
		},
	}
}

// NewProcessorWithBuffer creates a new Processor with a buffer and handler.
func NewProcessorWithBuffer(output io.Writer, handler Handler) *Processor {
	p := NewProcessor(handler)
	p.output = output
	return p
}

// Advance processes bytes and calls appropriate Handler methods.
func (p *Processor) Advance(handler Handler, bytes []byte) {
	// Check for synchronized update mode
	if p.syncState.enabled {
		// In sync mode, buffer the data
		p.syncState.buffer = append(p.syncState.buffer, bytes...)

		// Check for timeout
		if time.Since(p.syncState.startTime) > p.syncState.timeout {
			// Timeout - flush buffer and disable sync
			p.processSyncBuffer(handler)
			p.syncState.enabled = false
		}
		return
	}

	// Normal processing
	performer := &processorPerformer{handler: handler, processor: p}
	p.parser.Advance(performer, bytes)
}

// processSyncBuffer processes buffered data in synchronized mode.
func (p *Processor) processSyncBuffer(handler Handler) {
	if len(p.syncState.buffer) == 0 {
		return
	}

	performer := &processorPerformer{handler: handler, processor: p}
	p.parser.Advance(performer, p.syncState.buffer)
	p.syncState.buffer = p.syncState.buffer[:0]
}

// SetSyncTimeout sets the synchronized update timeout.
func (p *Processor) SetSyncTimeout(timeout time.Duration) {
	p.syncState.timeout = timeout
}

// BeginSynchronizedUpdate starts synchronized update mode.
func (p *Processor) BeginSynchronizedUpdate() {
	p.syncState.enabled = true
	p.syncState.startTime = time.Now()
	p.syncState.buffer = p.syncState.buffer[:0] // Clear buffer
}

// EndSynchronizedUpdate ends synchronized update mode and flushes buffer.
func (p *Processor) EndSynchronizedUpdate() {
	if p.syncState.enabled {
		if p.output != nil && len(p.syncState.buffer) > 0 {
			// Write buffered data to output
			_, _ = p.output.Write(p.syncState.buffer)
		}
		p.syncState.enabled = false
		p.syncState.buffer = p.syncState.buffer[:0]
	}
}

// IsInSynchronizedUpdate returns true if in synchronized update mode.
func (p *Processor) IsInSynchronizedUpdate() bool {
	return p.syncState.enabled
}

// SetMode sets a terminal mode on or off.
func (p *Processor) SetMode(mode Mode, enabled bool) {
	if p.modes == nil {
		p.modes = make(map[Mode]bool)
	}
	p.modes[mode] = enabled
}

// IsMode returns true if the specified mode is enabled.
func (p *Processor) IsMode(mode Mode) bool {
	if p.modes == nil {
		return false
	}
	return p.modes[mode]
}

// Write writes data to the processor (for buffered output).
func (p *Processor) Write(data string) {
	if p.syncState.enabled {
		// Buffer the data during synchronized updates
		p.syncState.buffer = append(p.syncState.buffer, []byte(data)...)
	} else if p.output != nil {
		// Write directly to output
		_, _ = p.output.Write([]byte(data))
	}
}

// Process processes raw bytes through the parser.
func (p *Processor) Process(data []byte) {
	if p.handler != nil {
		performer := &processorPerformer{handler: p.handler, processor: p}
		p.parser.Advance(performer, data)
	}
}

// Reset performs a soft reset.
func (p *Processor) Reset() {
	p.parser = NewParser()
	p.syncState.enabled = false
	p.syncState.buffer = p.syncState.buffer[:0]
	p.dcsState.active = false
	p.dcsState.buffer = p.dcsState.buffer[:0]
}

// processorPerformer implements Performer and translates to Handler calls.
type processorPerformer struct {
	handler   Handler
	processor *Processor
}

// Print implements Performer.
func (pp *processorPerformer) Print(c rune) {
	pp.handler.Input(c)
}

// Execute implements Performer.
func (pp *processorPerformer) Execute(b byte) {
	switch b {
	case C0.BEL:
		pp.handler.Bell()
	case C0.BS:
		pp.handler.Backspace()
	case C0.HT:
		pp.handler.Tab()
	case C0.LF, C0.VT, C0.FF:
		pp.handler.LineFeed()
	case C0.CR:
		pp.handler.CarriageReturn()
	case C0.SO:
		// Shift Out - activate G1 character set
		pp.handler.SetActiveCharset(G1)
	case C0.SI:
		// Shift In - activate G0 character set
		pp.handler.SetActiveCharset(G0)
	}
}

// Hook implements Performer.
func (pp *processorPerformer) Hook(params *Params, intermediates []byte, ignore bool, action rune) {
	// Convert Params to [][]uint16 format for Handler interface
	groups := params.Iter()
	handlerParams := make([][]uint16, len(groups))
	for i, group := range groups {
		handlerParams[i] = make([]uint16, len(group))
		copy(handlerParams[i], group)
	}

	// Mark DCS as active and clear buffer
	pp.processor.dcsState.active = true
	pp.processor.dcsState.buffer = pp.processor.dcsState.buffer[:0]

	// Call handler hook with converted parameters
	pp.handler.Hook(handlerParams, intermediates, ignore, action)
}

// Put implements Performer.
func (pp *processorPerformer) Put(b byte) {
	if pp.processor.dcsState.active {
		// Buffer the data byte
		pp.processor.dcsState.buffer = append(pp.processor.dcsState.buffer, b)
	}
}

// Unhook implements Performer.
func (pp *processorPerformer) Unhook() {
	if pp.processor.dcsState.active {
		// Send buffered data to handler
		if len(pp.processor.dcsState.buffer) > 0 {
			pp.handler.Put(pp.processor.dcsState.buffer)
		}

		// Mark DCS as inactive
		pp.processor.dcsState.active = false

		// Call handler unhook
		pp.handler.Unhook()
	}
}

// OscDispatch implements Performer.
func (pp *processorPerformer) OscDispatch(params [][]byte, bellTerminated bool) {
	if len(params) == 0 {
		return
	}

	// Parse OSC number
	var oscNum int
	if len(params[0]) > 0 {
		for _, b := range params[0] {
			if b >= '0' && b <= '9' {
				oscNum = oscNum*10 + int(b-'0')
			}
		}
	}

	switch oscNum {
	case 0, 2:
		// Set window title
		if len(params) > 1 {
			pp.handler.SetTitle(string(params[1]))
		}
	}
}

// CsiDispatch implements Performer.
func (pp *processorPerformer) CsiDispatch(params *Params, intermediates []byte, ignore bool, action rune) {
	if ignore {
		return
	}

	// Get parameter groups
	groups := params.Iter()

	switch action {
	case 'A':
		// CUU - Cursor Up
		n := getParam(groups, 0, 0, 1)
		pp.handler.MoveUp(n)

	case 'B':
		// CUD - Cursor Down
		n := getParam(groups, 0, 0, 1)
		pp.handler.MoveDown(n)

	case 'C':
		// CUF - Cursor Forward
		n := getParam(groups, 0, 0, 1)
		pp.handler.MoveForward(n)

	case 'D':
		// CUB - Cursor Backward
		n := getParam(groups, 0, 0, 1)
		pp.handler.MoveBackward(n)

	case 'E':
		// CNL - Cursor Next Line
		n := getParam(groups, 0, 0, 1)
		pp.handler.MoveDownAndCR(n)

	case 'F':
		// CPL - Cursor Previous Line
		n := getParam(groups, 0, 0, 1)
		pp.handler.MoveUpAndCR(n)

	case 'G':
		// CHA - Cursor Horizontal Absolute
		col := getParam(groups, 0, 0, 1)
		pp.handler.GotoCol(col)

	case 'H', 'f':
		// CUP - Cursor Position
		row := getParam(groups, 0, 0, 1)
		col := getParam(groups, 1, 0, 1)
		pp.handler.Goto(row, col)

	case 'J':
		// ED - Erase Display
		mode := getParam(groups, 0, 0, 0)
		pp.handler.ClearScreen(ClearMode(mode)) //nolint:gosec // mode is validated by getParam

	case 'K':
		// EL - Erase Line
		mode := getParam(groups, 0, 0, 0)
		pp.handler.ClearLine(LineClearMode(mode)) //nolint:gosec // mode is validated by getParam

	case 'L':
		// IL - Insert Lines
		n := getParam(groups, 0, 0, 1)
		pp.handler.InsertLines(n)

	case 'M':
		// DL - Delete Lines
		n := getParam(groups, 0, 0, 1)
		pp.handler.DeleteLines(n)

	case 'P':
		// DCH - Delete Characters
		n := getParam(groups, 0, 0, 1)
		pp.handler.DeleteChars(n)

	case 'S':
		// SU - Scroll Up
		n := getParam(groups, 0, 0, 1)
		pp.handler.ScrollUp(n)

	case 'T':
		// SD - Scroll Down
		n := getParam(groups, 0, 0, 1)
		pp.handler.ScrollDown(n)

	case 'X':
		// ECH - Erase Characters
		n := getParam(groups, 0, 0, 1)
		pp.handler.EraseChars(n)

	case '@':
		// ICH - Insert Characters
		n := getParam(groups, 0, 0, 1)
		pp.handler.InsertBlank(n)

	case 'd':
		// VPA - Vertical Position Absolute
		row := getParam(groups, 0, 0, 1)
		pp.handler.GotoLine(row)

	case 'm':
		// SGR - Select Graphic Rendition
		pp.processSGR(groups)

	case 'r':
		// DECSTBM - Set Scrolling Region
		top := getParam(groups, 0, 0, 1)
		bottom := getParam(groups, 1, 0, 0)
		if bottom == 0 {
			// 0 means default (bottom of screen)
			bottom = 24 // Default terminal height, should be configurable
		}
		pp.handler.SetScrollingRegion(top, bottom)

	case 's':
		// Save cursor position
		pp.handler.SaveCursorPosition()

	case 'u':
		// Restore cursor position
		pp.handler.RestoreCursorPosition()

	case 'h':
		// SM - Set Mode
		if len(intermediates) > 0 && intermediates[0] == '?' {
			// Private mode
			for _, group := range groups {
				if len(group) > 0 {
					pp.handler.SetMode(Mode(0x200 + group[0]))
				}
			}
		} else {
			// Standard mode
			for _, group := range groups {
				if len(group) > 0 {
					pp.handler.SetMode(Mode(group[0]))
				}
			}
		}

	case 'l':
		// RM - Reset Mode
		if len(intermediates) > 0 && intermediates[0] == '?' {
			// Private mode
			for _, group := range groups {
				if len(group) > 0 {
					pp.handler.ResetMode(Mode(0x200 + group[0]))
				}
			}
		} else {
			// Standard mode
			for _, group := range groups {
				if len(group) > 0 {
					pp.handler.ResetMode(Mode(group[0]))
				}
			}
		}

	case 'n':
		// DSR - Device Status Report
		kind := getParam(groups, 0, 0, 0)
		pp.handler.DeviceStatus(kind)

	case 'c':
		// DA - Device Attributes
		pp.handler.IdentifyTerminal()

	case 'g':
		// TBC - Tab Clear
		mode := getParam(groups, 0, 0, 0)
		switch mode {
		case 0:
			pp.handler.ClearTabStop(TabClearCurrent)
		case 3:
			pp.handler.ClearTabStop(TabClearAll)
		}

	case 'I':
		// CHT - Cursor Horizontal Tab (Forward)
		count := getParam(groups, 0, 0, 1)
		pp.handler.TabForward(count)

	case 'Z':
		// CBT - Cursor Backward Tab
		count := getParam(groups, 0, 0, 1)
		pp.handler.TabBackward(count)
	}
}

// EscDispatch implements Performer.
func (pp *processorPerformer) EscDispatch(intermediates []byte, ignore bool, b byte) {
	if ignore {
		return
	}

	switch b {
	case '7':
		// DECSC - Save Cursor
		pp.handler.SaveCursorPosition()

	case '8':
		// DECRC - Restore Cursor
		pp.handler.RestoreCursorPosition()

	case 'c':
		// RIS - Reset to Initial State
		pp.handler.Reset()

	case 'D':
		// IND - Index (move down one line)
		pp.handler.MoveDown(1)

	case 'E':
		// NEL - Next Line
		pp.handler.MoveDownAndCR(1)

	case 'M':
		// RI - Reverse Index (move up one line)
		pp.handler.MoveUp(1)

	case 'B':
		// Configure charset to ASCII
		pp.configureCharset(intermediates, StandardCharsetASCII)

	case '0':
		// Configure charset to special line drawing
		pp.configureCharset(intermediates, StandardCharsetSpecialLineDrawing)

	case 'H':
		// HTS - Horizontal Tab Set
		pp.handler.SetTabStop()
	}
}

// configureCharset configures a character set based on intermediate bytes.
func (pp *processorPerformer) configureCharset(intermediates []byte, charset StandardCharset) {
	if len(intermediates) != 1 {
		return
	}

	var index CharsetIndex
	switch intermediates[0] {
	case '(':
		index = G0
	case ')':
		index = G1
	case '*':
		index = G2
	case '+':
		index = G3
	default:
		return
	}

	pp.handler.ConfigureCharset(index, charset)
}

// processSGR processes SGR (Select Graphic Rendition) sequences.
func (pp *processorPerformer) processSGR(groups [][]uint16) {
	if len(groups) == 0 {
		// No parameters means reset
		pp.handler.ResetAttributes()
		pp.handler.ResetColors()
		return
	}

	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		switch group[0] {
		case 0:
			// Reset all
			pp.handler.ResetAttributes()
			pp.handler.ResetColors()

		case 1:
			pp.handler.SetAttribute(AttrBold)
		case 2:
			pp.handler.SetAttribute(AttrDim)
		case 3:
			pp.handler.SetAttribute(AttrItalic)
		case 4:
			pp.handler.SetAttribute(AttrUnderline)
		case 5:
			pp.handler.SetAttribute(AttrBlinking)
		case 7:
			pp.handler.SetAttribute(AttrReverse)
		case 8:
			pp.handler.SetAttribute(AttrHidden)
		case 9:
			pp.handler.SetAttribute(AttrStrikethrough)

		case 21:
			pp.handler.SetAttribute(AttrDoubleUnderline)

		case 30, 31, 32, 33, 34, 35, 36, 37:
			// Standard foreground colors
			pp.handler.SetForeground(NewNamedColor(NamedColor(group[0] - 30))) //nolint:gosec // value is validated

		case 38:
			// Extended foreground color
			if len(group) > 1 {
				pp.processExtendedColor(group, true)
			}

		case 39:
			// Default foreground
			pp.handler.SetForeground(NewNamedColor(Foreground))

		case 40, 41, 42, 43, 44, 45, 46, 47:
			// Standard background colors
			pp.handler.SetBackground(NewNamedColor(NamedColor(group[0] - 40))) //nolint:gosec // value is validated

		case 48:
			// Extended background color
			if len(group) > 1 {
				pp.processExtendedColor(group, false)
			}

		case 49:
			// Default background
			pp.handler.SetBackground(NewNamedColor(Background))

		case 90, 91, 92, 93, 94, 95, 96, 97:
			// Bright foreground colors
			pp.handler.SetForeground(NewNamedColor(NamedColor(group[0] - 90 + 8))) //nolint:gosec // value is validated

		case 100, 101, 102, 103, 104, 105, 106, 107:
			// Bright background colors
			pp.handler.SetBackground(NewNamedColor(NamedColor(group[0] - 100 + 8)))
		}
	}
}

// processExtendedColor processes extended color sequences (38/48).
func (pp *processorPerformer) processExtendedColor(group []uint16, isForeground bool) {
	if len(group) < 2 {
		return
	}

	var color Color

	switch group[1] {
	case 2:
		// RGB color
		if len(group) >= 5 {
			r := uint8(minUint16(group[2], 255))
			g := uint8(minUint16(group[3], 255))
			b := uint8(minUint16(group[4], 255))
			color = NewRgbColor(r, g, b)
		}

	case 5:
		// 256-color palette
		if len(group) >= 3 {
			idx := uint8(minUint16(group[2], 255))
			color = NewIndexedColor(idx)
		}
	}

	if isForeground {
		pp.handler.SetForeground(color)
	} else {
		pp.handler.SetBackground(color)
	}
}

// getParam gets a parameter value with defaults.
func getParam(groups [][]uint16, groupIdx, paramIdx int, defaultValue int) int {
	if groupIdx >= len(groups) {
		return defaultValue
	}

	group := groups[groupIdx]
	if paramIdx >= len(group) {
		return defaultValue
	}

	value := int(group[paramIdx])
	if value == 0 && defaultValue != 0 {
		return defaultValue
	}

	return value
}

// minUint16 returns the minimum of two uint16 values.
func minUint16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
