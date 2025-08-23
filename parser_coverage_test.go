package govte

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// TestParserEscapeIntermediate tests escape intermediate state
func TestParserEscapeIntermediate(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}

	// Enter escape intermediate state
	parser.Advance(performer, []byte{0x1B}) // ESC
	assert.Equal(t, StateEscape, parser.State())

	parser.Advance(performer, []byte{0x20}) // Space (intermediate)
	assert.Equal(t, StateEscapeIntermediate, parser.State())

	// Execute control in intermediate
	parser.Advance(performer, []byte{0x0A}) // LF
	assert.Equal(t, StateEscapeIntermediate, parser.State())
	assert.Contains(t, performer.executed, byte(0x0A))

	// Collect more intermediates
	parser.Advance(performer, []byte{0x21}) // !
	assert.Equal(t, StateEscapeIntermediate, parser.State())

	// Dispatch
	parser.Advance(performer, []byte{0x41}) // A
	assert.Equal(t, StateGround, parser.State())
	assert.Len(t, performer.escDispatched, 1)

	// Test ignore
	parser = NewParser()
	performer = &MockPerformer{}
	parser.Advance(performer, []byte("\x1b ")) // ESC space
	parser.Advance(performer, []byte{0x7F})    // DEL - should be ignored
	assert.Equal(t, StateEscapeIntermediate, parser.State())
}

// TestParserCSIIgnore tests CSI ignore state
func TestParserCSIIgnore(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}

	// Enter CSI ignore state from CSI param with invalid intermediate
	parser.Advance(performer, []byte("\x1b[1"))
	assert.Equal(t, StateCSIParam, parser.State())

	parser.Advance(performer, []byte{0x3F}) // ? causes ignore in CSI param state
	assert.Equal(t, StateCSIIgnore, parser.State())

	// Execute control in ignore
	parser.Advance(performer, []byte{0x0A}) // LF
	assert.Contains(t, performer.executed, byte(0x0A))

	// Ignore characters
	parser.Advance(performer, []byte("123"))
	assert.Equal(t, StateCSIIgnore, parser.State())

	// Exit on dispatch
	parser.Advance(performer, []byte{0x40}) // @
	assert.Equal(t, StateGround, parser.State())

	// Test DEL ignore
	parser = NewParser()
	performer = &MockPerformer{}
	parser.Advance(performer, []byte("\x1b["))
	parser.Advance(performer, []byte{0x3C}) // < is collected as intermediate, stays in CSI param
	assert.Equal(t, StateCSIParam, parser.State())
	parser.Advance(performer, []byte{0x3C}) // Second < causes ignore
	assert.Equal(t, StateCSIIgnore, parser.State())
	parser.Advance(performer, []byte{0x7F}) // DEL
	assert.Equal(t, StateCSIIgnore, parser.State())
}

// TestParserDCSIgnore tests DCS ignore state
func TestParserDCSIgnore(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}

	// Enter DCS ignore from DCS intermediate with invalid char
	parser.Advance(performer, []byte("\x1bP ")) // DCS with space intermediate
	assert.Equal(t, StateDCSIntermediate, parser.State())

	parser.Advance(performer, []byte{0x3F}) // ? (invalid, causes ignore)
	assert.Equal(t, StateDCSIgnore, parser.State())

	// ESC in ignore (might be ST)
	parser.Advance(performer, []byte{0x1B})
	assert.Equal(t, StateDCSIgnore, parser.State())

	// CAN exits to ground
	parser.Advance(performer, []byte{0x18}) // CAN
	assert.Equal(t, StateGround, parser.State())

	// Test SUB exit
	parser = NewParser()
	performer = &MockPerformer{}
	parser.Advance(performer, []byte("\x1bP"))
	parser.Advance(performer, []byte{0x3C}) // < is collected as intermediate in DCS entry
	assert.Equal(t, StateDCSParam, parser.State())
	parser.Advance(performer, []byte{0x3C}) // Second < causes ignore
	assert.Equal(t, StateDCSIgnore, parser.State())
	parser.Advance(performer, []byte{0x1A}) // SUB
	assert.Equal(t, StateGround, parser.State())
}

// TestParserSOSPMApcString tests SOS/PM/APC string state
func TestParserSOSPMApcString(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}

	// Enter SOS state
	parser.Advance(performer, []byte{0x1B, 0x58}) // ESC X
	assert.Equal(t, StateSOSPMApcString, parser.State())

	// Ignore content
	parser.Advance(performer, []byte("ignored text"))
	assert.Equal(t, StateSOSPMApcString, parser.State())

	// ESC might be ST
	parser.Advance(performer, []byte{0x1B})
	assert.Equal(t, StateSOSPMApcString, parser.State())

	// Backslash completes ST
	parser.Advance(performer, []byte{'\\'})
	assert.Equal(t, StateGround, parser.State())

	// Test PM entry
	parser = NewParser()
	parser.Advance(performer, []byte{0x1B, 0x5E}) // ESC ^
	assert.Equal(t, StateSOSPMApcString, parser.State())

	// Test APC entry
	parser = NewParser()
	parser.Advance(performer, []byte{0x1B, 0x5F}) // ESC _
	assert.Equal(t, StateSOSPMApcString, parser.State())
}

// TestParserDCSStates tests various DCS state transitions
func TestParserDCSStates(t *testing.T) {
	t.Run("DCS entry with params", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1bP"))
		assert.Equal(t, StateDCSEntry, parser.State())

		// Collect intermediate in entry
		parser.Advance(performer, []byte{0x20}) // Space
		assert.Equal(t, StateDCSIntermediate, parser.State())

		// Dispatch to passthrough
		parser.Advance(performer, []byte{0x70}) // p
		assert.Equal(t, StateDCSPassthrough, parser.State())
		assert.True(t, performer.hookCalled)
	})

	t.Run("DCS with subparams", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1bP"))
		parser.Advance(performer, []byte(":")) // Subparam in entry
		assert.Equal(t, StateDCSParam, parser.State())

		parser.Advance(performer, []byte("5"))
		parser.Advance(performer, []byte{0x71}) // q
		assert.Equal(t, StateDCSPassthrough, parser.State())
	})

	t.Run("DCS passthrough with pending ESC", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1bP0q")) // Enter passthrough
		assert.Equal(t, StateDCSPassthrough, parser.State())

		// ESC that's not part of ST
		parser.Advance(performer, []byte{0x1B})
		parser.Advance(performer, []byte{0x41}) // A (not \)
		assert.Contains(t, performer.putBytes, byte(0x1B))
		assert.Contains(t, performer.putBytes, byte(0x41))
	})

	t.Run("DCS intermediate ignore transition", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1bP ")) // Space intermediate
		assert.Equal(t, StateDCSIntermediate, parser.State())

		// Collect more intermediates
		parser.Advance(performer, []byte{0x21}) // !
		assert.Equal(t, StateDCSIntermediate, parser.State())

		// Invalid char causes ignore
		parser.Advance(performer, []byte{0x3F}) // ?
		assert.Equal(t, StateDCSIgnore, parser.State())
	})
}

// TestParserCSIIntermediateTransitions tests CSI intermediate state
func TestParserCSIIntermediateTransitions(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}

	// Enter CSI intermediate from CSI entry
	parser.Advance(performer, []byte("\x1b["))
	parser.Advance(performer, []byte{0x20}) // Space
	assert.Equal(t, StateCSIIntermediate, parser.State())

	// Collect more intermediates
	parser.Advance(performer, []byte{0x21}) // !
	assert.Equal(t, StateCSIIntermediate, parser.State())

	// Invalid causes ignore
	parser.Advance(performer, []byte{0x3F}) // ?
	assert.Equal(t, StateCSIIgnore, parser.State())
}

// TestParserGroundC1Controls tests C1 control handling in ground state
func TestParserGroundC1Controls(t *testing.T) {
	t.Run("DCS via C1", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte{0x90}) // DCS
		assert.Equal(t, StateDCSEntry, parser.State())
	})

	t.Run("CSI via C1", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte{0x9B}) // CSI
		assert.Equal(t, StateCSIEntry, parser.State())
	})

	t.Run("OSC via C1", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte{0x9D}) // OSC
		assert.Equal(t, StateOSCString, parser.State())
	})

	t.Run("Invalid continuation byte", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte{0x85}) // NEL (C1 control)
		assert.Equal(t, StateGround, parser.State())
		// Should print replacement character
		assert.Contains(t, performer.printed, utf8.RuneError)
	})
}

// TestParserMaxLimits tests buffer limits
func TestParserMaxLimits(t *testing.T) {
	t.Run("Max intermediates", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte{0x1B})
		// Try to add more than MaxIntermediates
		for i := 0; i < MaxIntermediates+2; i++ {
			parser.Advance(performer, []byte{byte(0x20 + i)})
		}
		// Should mark as ignoring after MaxIntermediates
		parser.Advance(performer, []byte{0x41}) // A
		assert.True(t, performer.escDispatched[0].ignore)
	})

	t.Run("Max OSC size", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1b]"))
		// Try to add more than MaxOSCRaw bytes
		longData := make([]byte, MaxOSCRaw+100)
		for i := range longData {
			longData[i] = 'A'
		}
		parser.Advance(performer, longData)
		parser.Advance(performer, []byte{0x07}) // BEL

		// Should truncate to MaxOSCRaw
		assert.LessOrEqual(t, len(performer.oscDispatched[0].params[0]), MaxOSCRaw)
	})
}

// TestParserEdgeCases tests various edge cases
func TestParserEdgeCases(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte{})
		assert.Equal(t, StateGround, parser.State())
	})

	t.Run("DEL in various states", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		// DEL in ground - should be ignored
		parser.Advance(performer, []byte{0x7F})
		assert.Equal(t, StateGround, parser.State())
		assert.Empty(t, performer.executed)

		// DEL in escape
		parser.Advance(performer, []byte{0x1B, 0x7F})
		assert.Equal(t, StateEscape, parser.State())

		// DEL in CSI param
		parser.Advance(performer, []byte{'[', '1', 0x7F})
		assert.Equal(t, StateCSIParam, parser.State())
	})

	t.Run("Control chars in OSC", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1b]"))
		// Control chars < 0x20
		parser.Advance(performer, []byte{0x01, 0x02, 0x03})
		assert.Equal(t, StateOSCString, parser.State())

		// High bytes >= 0x80
		parser.Advance(performer, []byte{0x80, 0x81})
		assert.Equal(t, StateOSCString, parser.State())

		parser.Advance(performer, []byte{0x07}) // BEL
		assert.Equal(t, StateGround, parser.State())
	})

	t.Run("OSC with ST terminator", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		parser.Advance(performer, []byte("\x1b]0;Title"))
		// ST is ESC \
		parser.Advance(performer, []byte{0x1B})
		parser.Advance(performer, []byte{'\\'})
		assert.Equal(t, StateGround, parser.State())
		assert.Len(t, performer.oscDispatched, 1)
		assert.False(t, performer.oscDispatched[0].bellTerminated)
	})

	t.Run("Parameter separator with no current param", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}

		// Empty parameters
		parser.Advance(performer, []byte("\x1b[;;H"))
		assert.Len(t, performer.csiDispatched, 1)
		// Should have two zero parameters
		params := performer.csiDispatched[0].params
		iter := params.Iter()
		// First group
		group1 := iter[0]
		assert.Equal(t, []uint16{0}, group1)
		// Second group
		if len(iter) > 1 {
			group2 := iter[1]
			assert.Equal(t, []uint16{0}, group2)
		}
	})
}

// TestParserDCSPassthroughExecute tests execute in DCS passthrough
func TestParserDCSPassthroughExecute(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}

	// Enter DCS passthrough
	parser.Advance(performer, []byte("\x1bP0q"))
	assert.Equal(t, StateDCSPassthrough, parser.State())

	// CAN exits to ground with execute
	parser.Advance(performer, []byte{0x18}) // CAN
	assert.Equal(t, StateGround, parser.State())
	assert.True(t, performer.unhookCalled)
	assert.Contains(t, performer.executed, byte(0x18))

	// Test SUB
	parser = NewParser()
	performer = &MockPerformer{}
	parser.Advance(performer, []byte("\x1bP0q"))
	parser.Advance(performer, []byte{0x1A}) // SUB
	assert.Equal(t, StateGround, parser.State())
	assert.Contains(t, performer.executed, byte(0x1A))
}
