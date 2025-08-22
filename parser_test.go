package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserCreation(t *testing.T) {
	parser := NewParser()
	assert.NotNil(t, parser)
	assert.Equal(t, StateGround, parser.State())
	assert.Empty(t, parser.intermediates)
	assert.False(t, parser.ignoring)
}

func TestParserSimpleText(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// Test simple ASCII text
	input := []byte("Hello")
	parser.Advance(performer, input)
	
	assert.Equal(t, []rune{'H', 'e', 'l', 'l', 'o'}, performer.printed)
	assert.Empty(t, performer.executed)
}

func TestParserControlCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{"Backspace", []byte{0x08}, []byte{0x08}},
		{"Tab", []byte{0x09}, []byte{0x09}},
		{"Line Feed", []byte{0x0A}, []byte{0x0A}},
		{"Carriage Return", []byte{0x0D}, []byte{0x0D}},
		{"Bell", []byte{0x07}, []byte{0x07}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			performer := &MockPerformer{}
			
			parser.Advance(performer, tt.input)
			assert.Equal(t, tt.expected, performer.executed)
			assert.Empty(t, performer.printed)
		})
	}
}

func TestParserMixedTextAndControl(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// Text with embedded control characters
	input := []byte("Hello\nWorld\rX")
	parser.Advance(performer, input)
	
	assert.Equal(t, []rune{'H', 'e', 'l', 'l', 'o', 'W', 'o', 'r', 'l', 'd', 'X'}, performer.printed)
	assert.Equal(t, []byte{0x0A, 0x0D}, performer.executed)
}

func TestParserEscapeSequence(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC should transition to Escape state
	input := []byte{0x1B}
	parser.Advance(performer, input)
	
	assert.Equal(t, StateEscape, parser.State())
	assert.Empty(t, performer.printed)
	assert.Empty(t, performer.executed)
}

func TestParserCSISequence(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC [ should transition to CSI Entry
	input := []byte{0x1B, '['}
	parser.Advance(performer, input)
	
	assert.Equal(t, StateCSIEntry, parser.State())
}

func TestParserSimpleCSIDispatch(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC [ H - Cursor home
	input := []byte{0x1B, '[', 'H'}
	parser.Advance(performer, input)
	
	assert.Len(t, performer.csiDispatched, 1)
	assert.Equal(t, 'H', performer.csiDispatched[0].action)
	assert.Equal(t, StateGround, parser.State())
}

func TestParserCSIWithParams(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC [ 1 ; 2 H - Cursor position with params
	input := []byte{0x1B, '[', '1', ';', '2', 'H'}
	parser.Advance(performer, input)
	
	assert.Len(t, performer.csiDispatched, 1)
	dispatch := performer.csiDispatched[0]
	assert.Equal(t, 'H', dispatch.action)
	assert.NotNil(t, dispatch.params)
	
	// Check parameters
	iter := dispatch.params.Iter()
	assert.Len(t, iter, 2)
	assert.Equal(t, []uint16{1}, iter[0])
	assert.Equal(t, []uint16{2}, iter[1])
}

func TestParserOSCSequence(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC ] 0 ; Title ST
	input := []byte{0x1B, ']', '0', ';', 'T', 'i', 't', 'l', 'e', 0x1B, '\\'}
	parser.Advance(performer, input)
	
	assert.Len(t, performer.oscDispatched, 1)
	assert.Equal(t, [][]byte{[]byte("0"), []byte("Title")}, performer.oscDispatched[0].params)
	assert.False(t, performer.oscDispatched[0].bellTerminated)
	assert.Equal(t, StateGround, parser.State())
}

func TestParserOSCBellTerminated(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC ] 0 ; Title BEL
	input := []byte{0x1B, ']', '0', ';', 'T', 'i', 't', 'l', 'e', 0x07}
	parser.Advance(performer, input)
	
	assert.Len(t, performer.oscDispatched, 1)
	assert.Equal(t, [][]byte{[]byte("0"), []byte("Title")}, performer.oscDispatched[0].params)
	assert.True(t, performer.oscDispatched[0].bellTerminated)
	assert.Equal(t, StateGround, parser.State())
}

func TestParserUTF8Handling(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []rune
	}{
		{"ASCII", []byte("Hello"), []rune{'H', 'e', 'l', 'l', 'o'}},
		{"2-byte UTF-8", []byte("café"), []rune{'c', 'a', 'f', 'é'}},
		{"3-byte UTF-8", []byte("你好"), []rune{'你', '好'}},
		{"4-byte UTF-8", []byte("𝔸𝔹"), []rune{'𝔸', '𝔹'}},
		{"Mixed", []byte("Hi你好!"), []rune{'H', 'i', '你', '好', '!'}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			performer := &MockPerformer{}
			
			parser.Advance(performer, tt.input)
			assert.Equal(t, tt.expected, performer.printed)
		})
	}
}

func TestParserPartialUTF8(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// Split a 3-byte UTF-8 character (你 = E4 BD A0)
	part1 := []byte{0xE4, 0xBD}
	part2 := []byte{0xA0}
	
	parser.Advance(performer, part1)
	assert.Empty(t, performer.printed) // Should not print incomplete UTF-8
	
	parser.Advance(performer, part2)
	assert.Equal(t, []rune{'你'}, performer.printed) // Should print complete character
}

func TestParserStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		finalState  State
		description string
	}{
		{
			name:        "ESC to Escape",
			input:       []byte{0x1B},
			finalState:  StateEscape,
			description: "ESC should transition to Escape state",
		},
		{
			name:        "ESC [ to CSI Entry",
			input:       []byte{0x1B, '['},
			finalState:  StateCSIEntry,
			description: "ESC [ should transition to CSI Entry",
		},
		{
			name:        "ESC ] to OSC String",
			input:       []byte{0x1B, ']'},
			finalState:  StateOSCString,
			description: "ESC ] should transition to OSC String",
		},
		{
			name:        "ESC P to DCS Entry",
			input:       []byte{0x1B, 'P'},
			finalState:  StateDCSEntry,
			description: "ESC P should transition to DCS Entry",
		},
		{
			name:        "Complete CSI returns to Ground",
			input:       []byte{0x1B, '[', 'H'},
			finalState:  StateGround,
			description: "Complete CSI sequence should return to Ground",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			performer := &MockPerformer{}
			
			parser.Advance(performer, tt.input)
			assert.Equal(t, tt.finalState, parser.State(), tt.description)
		})
	}
}

func TestParserIgnoreInvalidSequences(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// Invalid intermediate bytes should set ignore flag
	input := []byte{0x1B, '[', 0x20, 0x21, 0x22, 'H'} // Too many intermediates
	parser.Advance(performer, input)
	
	assert.Len(t, performer.csiDispatched, 1)
	assert.True(t, performer.csiDispatched[0].ignore, "Should set ignore flag for invalid sequence")
}

func TestParserDCSSequence(t *testing.T) {
	parser := NewParser()
	performer := &MockPerformer{}
	
	// ESC P (DCS) followed by data and ST
	input := []byte{0x1B, 'P', '1', '$', 'r', 'D', 'a', 't', 'a', 0x1B, '\\'}
	parser.Advance(performer, input)
	
	assert.True(t, performer.hookCalled)
	assert.Equal(t, []byte{'D', 'a', 't', 'a'}, performer.putBytes)
	assert.True(t, performer.unhookCalled)
	assert.Equal(t, StateGround, parser.State())
}

// Benchmark tests
func BenchmarkParserSimpleText(b *testing.B) {
	parser := NewParser()
	performer := &NoopPerformer{}
	input := []byte("Hello, World! This is a simple text benchmark.")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Advance(performer, input)
	}
}

func BenchmarkParserWithEscapes(b *testing.B) {
	parser := NewParser()
	performer := &NoopPerformer{}
	input := []byte("Normal \x1b[31mRed\x1b[0m Normal \x1b[1;2H")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Advance(performer, input)
	}
}

func BenchmarkParserUTF8(b *testing.B) {
	parser := NewParser()
	performer := &NoopPerformer{}
	input := []byte("Hello 你好 世界 🌍 测试文本")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Advance(performer, input)
	}
}

// TestParserSubparameters tests CSI sequences with subparameters
func TestParserSubparameters(t *testing.T) {
	t.Run("RGB foreground color with subparameters", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// SGR with RGB foreground: ESC[38:2:255:128:64m
		parser.Advance(performer, []byte("\x1b[38:2:255:128:64m"))
		
		assert.Len(t, performer.csiDispatched, 1)
		csi := performer.csiDispatched[0]
		assert.Equal(t, 'm', csi.action)
		
		// Verify params structure
		groups := csi.params.Iter()
		assert.Len(t, groups, 1)
		assert.Equal(t, []uint16{38, 2, 255, 128, 64}, groups[0])
	})

	t.Run("Multiple parameters with subparameters", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// SGR with RGB foreground and indexed background
		parser.Advance(performer, []byte("\x1b[38:2:255:0:0;48:5:16m"))
		
		assert.Len(t, performer.csiDispatched, 1)
		csi := performer.csiDispatched[0]
		
		groups := csi.params.Iter()
		assert.Len(t, groups, 2)
		assert.Equal(t, []uint16{38, 2, 255, 0, 0}, groups[0]) // RGB red
		assert.Equal(t, []uint16{48, 5, 16}, groups[1]) // Indexed color 16
	})

	t.Run("Mixed regular and subparameters", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Bold + RGB color + underline
		parser.Advance(performer, []byte("\x1b[1;38:5:128;4m"))
		
		assert.Len(t, performer.csiDispatched, 1)
		csi := performer.csiDispatched[0]
		
		groups := csi.params.Iter()
		assert.Len(t, groups, 3)
		assert.Equal(t, []uint16{1}, groups[0]) // Bold
		assert.Equal(t, []uint16{38, 5, 128}, groups[1]) // Indexed color
		assert.Equal(t, []uint16{4}, groups[2]) // Underline
	})

	t.Run("Empty subparameters", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Subparameter with missing values
		parser.Advance(performer, []byte("\x1b[38::128m"))
		
		assert.Len(t, performer.csiDispatched, 1)
		csi := performer.csiDispatched[0]
		
		groups := csi.params.Iter()
		assert.Len(t, groups, 1)
		// Empty subparam should be 0, then 128
		assert.Equal(t, []uint16{38, 0, 128}, groups[0])
	})

	t.Run("Subparameter only sequence", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Just a colon without main param
		parser.Advance(performer, []byte("\x1b[:5m"))
		
		assert.Len(t, performer.csiDispatched, 1)
		csi := performer.csiDispatched[0]
		
		groups := csi.params.Iter()
		assert.Len(t, groups, 1)
		// Should have a 0 main param with subparam 5
		assert.Equal(t, []uint16{0, 5}, groups[0])
	})
}

// TestParserUTF8Boundaries tests UTF-8 parsing edge cases
func TestParserUTF8Boundaries(t *testing.T) {
	t.Run("Split 2-byte UTF-8", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// UTF-8 for "é" (U+00E9) is 0xC3 0xA9
		parser.Advance(performer, []byte{0xC3}) // First byte only
		assert.Empty(t, performer.printed) // Should not print yet
		
		parser.Advance(performer, []byte{0xA9}) // Second byte
		assert.Equal(t, []rune{'é'}, performer.printed)
	})

	t.Run("Split 3-byte UTF-8", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// UTF-8 for "你" (U+4F60) is 0xE4 0xBD 0xA0
		parser.Advance(performer, []byte{0xE4}) // First byte
		assert.Empty(t, performer.printed)
		
		parser.Advance(performer, []byte{0xBD}) // Second byte
		assert.Empty(t, performer.printed)
		
		parser.Advance(performer, []byte{0xA0}) // Third byte
		assert.Equal(t, []rune{'你'}, performer.printed)
	})

	t.Run("Split 4-byte UTF-8", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// UTF-8 for "🌍" (U+1F30D) is 0xF0 0x9F 0x8C 0x8D
		parser.Advance(performer, []byte{0xF0}) // First byte
		assert.Empty(t, performer.printed)
		
		parser.Advance(performer, []byte{0x9F, 0x8C}) // Middle bytes
		assert.Empty(t, performer.printed)
		
		parser.Advance(performer, []byte{0x8D}) // Last byte
		assert.Equal(t, []rune{'🌍'}, performer.printed)
	})

	t.Run("Invalid UTF-8 sequences", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Invalid continuation byte without start
		parser.Advance(performer, []byte{0x80})
		// Should handle gracefully - likely print replacement character
		assert.Len(t, performer.printed, 1)
		performer.printed = nil
		
		// Invalid start byte followed by non-continuation
		parser.Advance(performer, []byte{0xC3, 0x41}) // 0x41 is 'A', not continuation
		// Should handle the invalid sequence and then print 'A'
		assert.Contains(t, performer.printed, 'A')
	})

	t.Run("UTF-8 interrupted by control sequence", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Start UTF-8, then ESC sequence
		parser.Advance(performer, []byte{0xE4}) // Start of "你"
		assert.Empty(t, performer.printed)
		
		// ESC sequence should reset UTF-8 state
		parser.Advance(performer, []byte("\x1b[0m"))
		assert.Len(t, performer.csiDispatched, 1)
		
		// Continue with new UTF-8
		parser.Advance(performer, []byte("Hello"))
		assert.Contains(t, performer.printed, 'H')
	})

	t.Run("Mixed ASCII and UTF-8", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		input := []byte("Hello 世界!")
		parser.Advance(performer, input)
		
		expected := []rune{'H', 'e', 'l', 'l', 'o', ' ', '世', '界', '!'}
		assert.Equal(t, expected, performer.printed)
	})

	t.Run("UTF-8 across multiple advances", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Split "Hello 你好 World" across multiple calls
		parser.Advance(performer, []byte("Hello "))
		parser.Advance(performer, []byte{0xE4, 0xBD}) // Part of "你"
		parser.Advance(performer, []byte{0xA0, 0xE5}) // Rest of "你" and part of "好"
		parser.Advance(performer, []byte{0xA5, 0xBD}) // Rest of "好"
		parser.Advance(performer, []byte(" World"))
		
		expected := []rune{'H', 'e', 'l', 'l', 'o', ' ', '你', '好', ' ', 'W', 'o', 'r', 'l', 'd'}
		assert.Equal(t, expected, performer.printed)
	})

	t.Run("Zero-width characters", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Test with combining diacritical marks
		// "e" + combining acute accent (U+0301)
		input := []byte("e\xCC\x81") // Results in "é"
		parser.Advance(performer, input)
		
		assert.Equal(t, []rune{'e', '\u0301'}, performer.printed)
	})
}

// TestParserAdditionalStateTransitions tests more state transitions
func TestParserAdditionalStateTransitions(t *testing.T) {
	t.Run("Ground to Escape and back", func(t *testing.T) {
		parser := NewParser()
		assert.Equal(t, StateGround, parser.State())
		
		performer := &MockPerformer{}
		parser.Advance(performer, []byte{0x1B}) // ESC
		assert.Equal(t, StateEscape, parser.State())
		
		parser.Advance(performer, []byte{'M'}) // Reverse Index
		assert.Equal(t, StateGround, parser.State())
	})

	t.Run("CSI parameter collection", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		// Test parameter collection state
		parser.Advance(performer, []byte("\x1b["))
		assert.Equal(t, StateCSIEntry, parser.State())
		
		parser.Advance(performer, []byte("1"))
		assert.Equal(t, StateCSIParam, parser.State())
		
		parser.Advance(performer, []byte(";"))
		assert.Equal(t, StateCSIParam, parser.State())
		
		parser.Advance(performer, []byte("2"))
		assert.Equal(t, StateCSIParam, parser.State())
		
		parser.Advance(performer, []byte("H"))
		assert.Equal(t, StateGround, parser.State())
	})

	t.Run("OSC string collection", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		parser.Advance(performer, []byte("\x1b]"))
		assert.Equal(t, StateOSCString, parser.State())
		
		parser.Advance(performer, []byte("0;Title"))
		assert.Equal(t, StateOSCString, parser.State())
		
		parser.Advance(performer, []byte("\x07")) // BEL
		assert.Equal(t, StateGround, parser.State())
	})

	t.Run("DCS passthrough", func(t *testing.T) {
		parser := NewParser()
		performer := &MockPerformer{}
		
		parser.Advance(performer, []byte("\x1bP"))
		assert.Equal(t, StateDCSEntry, parser.State())
		
		parser.Advance(performer, []byte("1"))
		assert.Equal(t, StateDCSParam, parser.State())
		
		parser.Advance(performer, []byte("q"))
		assert.Equal(t, StateDCSPassthrough, parser.State())
		
		parser.Advance(performer, []byte("data"))
		assert.Equal(t, StateDCSPassthrough, parser.State())
		
		parser.Advance(performer, []byte("\x1b\\"))
		assert.Equal(t, StateGround, parser.State())
	})
}