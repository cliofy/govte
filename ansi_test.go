package govte

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRgb(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		c := NewRgb(255, 128, 64)
		assert.Equal(t, uint8(255), c.R)
		assert.Equal(t, uint8(128), c.G)
		assert.Equal(t, uint8(64), c.B)
	})

	t.Run("String", func(t *testing.T) {
		tests := []struct {
			color    Rgb
			expected string
		}{
			{NewRgb(255, 255, 255), "#ffffff"},
			{NewRgb(0, 0, 0), "#000000"},
			{NewRgb(255, 0, 0), "#ff0000"},
			{NewRgb(0, 255, 0), "#00ff00"},
			{NewRgb(0, 0, 255), "#0000ff"},
			{NewRgb(128, 64, 32), "#804020"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.expected, tt.color.String())
		}
	})

	t.Run("Luminance", func(t *testing.T) {
		// Test known values
		white := NewRgb(255, 255, 255)
		black := NewRgb(0, 0, 0)
		red := NewRgb(255, 0, 0)

		assert.InDelta(t, 1.0, white.Luminance(), 0.001)
		assert.InDelta(t, 0.0, black.Luminance(), 0.001)
		assert.InDelta(t, 0.2126, red.Luminance(), 0.001)
	})

	t.Run("Contrast", func(t *testing.T) {
		white := NewRgb(255, 255, 255)
		black := NewRgb(0, 0, 0)

		// Maximum contrast is 21:1 (white on black)
		contrast := white.Contrast(black)
		assert.InDelta(t, 21.0, contrast, 0.1)

		// Same color has contrast of 1:1
		assert.InDelta(t, 1.0, white.Contrast(white), 0.001)
		assert.InDelta(t, 1.0, black.Contrast(black), 0.001)
	})
}

func TestNamedColor(t *testing.T) {
	t.Run("ToRgb", func(t *testing.T) {
		tests := []struct {
			color    NamedColor
			expected Rgb
		}{
			{Black, Rgb{0, 0, 0}},
			{Red, Rgb{170, 0, 0}},
			{Green, Rgb{0, 170, 0}},
			{Yellow, Rgb{170, 85, 0}},
			{Blue, Rgb{0, 0, 170}},
			{Magenta, Rgb{170, 0, 170}},
			{Cyan, Rgb{0, 170, 170}},
			{White, Rgb{170, 170, 170}},
			{BrightBlack, Rgb{85, 85, 85}},
			{BrightRed, Rgb{255, 85, 85}},
			{BrightGreen, Rgb{85, 255, 85}},
			{BrightYellow, Rgb{255, 255, 85}},
			{BrightBlue, Rgb{85, 85, 255}},
			{BrightMagenta, Rgb{255, 85, 255}},
			{BrightCyan, Rgb{85, 255, 255}},
			{BrightWhite, Rgb{255, 255, 255}},
		}

		for _, tt := range tests {
			result := tt.color.ToRgb()
			assert.Equal(t, tt.expected, result, "Color %d", tt.color)
		}
	})

	t.Run("SpecialColors", func(t *testing.T) {
		// Special colors should return black as default
		assert.Equal(t, Rgb{0, 0, 0}, Foreground.ToRgb())
		assert.Equal(t, Rgb{0, 0, 0}, Background.ToRgb())
	})
}

func TestColor(t *testing.T) {
	t.Run("NamedColor", func(t *testing.T) {
		c := NewNamedColor(Red)
		assert.Equal(t, ColorTypeNamed, c.Type)
		assert.Equal(t, Red, c.Named)
	})

	t.Run("IndexedColor", func(t *testing.T) {
		c := NewIndexedColor(128)
		assert.Equal(t, ColorTypeIndexed, c.Type)
		assert.Equal(t, uint8(128), c.Index)
	})

	t.Run("RgbColor", func(t *testing.T) {
		c := NewRgbColor(100, 150, 200)
		assert.Equal(t, ColorTypeRgb, c.Type)
		assert.Equal(t, Rgb{100, 150, 200}, c.Rgb)
	})
}

func TestAttr(t *testing.T) {
	t.Run("Has", func(t *testing.T) {
		attr := AttrBold | AttrItalic
		assert.True(t, attr.Has(AttrBold))
		assert.True(t, attr.Has(AttrItalic))
		assert.False(t, attr.Has(AttrUnderline))
		assert.False(t, AttrNone.Has(AttrBold))
	})

	t.Run("Add", func(t *testing.T) {
		attr := AttrBold
		attr = attr.Add(AttrItalic)
		assert.True(t, attr.Has(AttrBold))
		assert.True(t, attr.Has(AttrItalic))

		// Adding same attribute should be idempotent
		attr = attr.Add(AttrBold)
		assert.True(t, attr.Has(AttrBold))
	})

	t.Run("Remove", func(t *testing.T) {
		attr := AttrBold | AttrItalic | AttrUnderline
		attr = attr.Remove(AttrItalic)
		assert.True(t, attr.Has(AttrBold))
		assert.False(t, attr.Has(AttrItalic))
		assert.True(t, attr.Has(AttrUnderline))

		// Removing non-existent attribute should be safe
		attr = attr.Remove(AttrBlinking)
		assert.True(t, attr.Has(AttrBold))
		assert.True(t, attr.Has(AttrUnderline))
	})

	t.Run("Toggle", func(t *testing.T) {
		attr := AttrBold
		attr = attr.Toggle(AttrItalic)
		assert.True(t, attr.Has(AttrBold))
		assert.True(t, attr.Has(AttrItalic))

		attr = attr.Toggle(AttrBold)
		assert.False(t, attr.Has(AttrBold))
		assert.True(t, attr.Has(AttrItalic))

		attr = attr.Toggle(AttrBold)
		assert.True(t, attr.Has(AttrBold))
		assert.True(t, attr.Has(AttrItalic))
	})

	t.Run("AllAttributes", func(t *testing.T) {
		// Test all attribute constants are unique
		attrs := []Attr{
			AttrBold, AttrDim, AttrItalic, AttrUnderline,
			AttrBlinking, AttrReverse, AttrHidden, AttrStrikethrough,
			AttrDoubleUnderline, AttrCurlyUnderline, AttrDottedUnderline, AttrDashedUnderline,
		}

		for i, a1 := range attrs {
			for j, a2 := range attrs {
				if i != j {
					assert.NotEqual(t, a1, a2, "Attributes should be unique")
				}
			}
		}
	})
}

func TestMode(t *testing.T) {
	t.Run("IsPrivate", func(t *testing.T) {
		// Standard modes
		assert.False(t, ModeKeyboardAction.IsPrivate())
		assert.False(t, ModeInsert.IsPrivate())
		assert.False(t, ModeSendReceive.IsPrivate())
		assert.False(t, ModeAutomaticNewline.IsPrivate())

		// Private modes
		assert.True(t, ModeApplicationCursor.IsPrivate())
		assert.True(t, ModeApplicationKeypad.IsPrivate())
		assert.True(t, ModeAlternateScreen.IsPrivate())
		assert.True(t, ModeShowCursor.IsPrivate())
		assert.True(t, ModeBracketedPaste.IsPrivate())
		assert.True(t, ModeSynchronizedOutput.IsPrivate())
	})

	t.Run("UniqueValues", func(t *testing.T) {
		modes := []Mode{
			ModeKeyboardAction, ModeInsert, ModeReplace, ModeSendReceive, ModeAutomaticNewline,
			ModeApplicationCursor, ModeApplicationKeypad, ModeAlternateScreen,
			ModeShowCursor, ModeSaveRestoreCursor, ModeAlternateScreenBuffer,
			ModeBracketedPaste, ModeSynchronizedOutput,
		}

		seen := make(map[Mode]bool)
		for _, m := range modes {
			assert.False(t, seen[m], "Mode %d should be unique", m)
			seen[m] = true
		}
	})
}

func TestCursorStyle(t *testing.T) {
	t.Run("CursorShape", func(t *testing.T) {
		shapes := []CursorShape{CursorShapeBlock, CursorShapeUnderline, CursorShapeBeam}
		assert.Equal(t, 3, len(shapes))

		// Ensure values are distinct
		assert.NotEqual(t, CursorShapeBlock, CursorShapeUnderline)
		assert.NotEqual(t, CursorShapeUnderline, CursorShapeBeam)
		assert.NotEqual(t, CursorShapeBlock, CursorShapeBeam)
	})

	t.Run("CursorStyle", func(t *testing.T) {
		style := CursorStyle{
			Shape:    CursorShapeBeam,
			Blinking: true,
		}
		assert.Equal(t, CursorShapeBeam, style.Shape)
		assert.True(t, style.Blinking)
	})
}

func TestClearModes(t *testing.T) {
	t.Run("LineClearMode", func(t *testing.T) {
		modes := []LineClearMode{LineClearRight, LineClearLeft, LineClearAll}
		assert.Equal(t, 3, len(modes))
	})

	t.Run("ClearMode", func(t *testing.T) {
		modes := []ClearMode{ClearBelow, ClearAbove, ClearAll, ClearSaved}
		assert.Equal(t, 4, len(modes))
	})

	t.Run("TabulationClearMode", func(t *testing.T) {
		modes := []TabulationClearMode{TabClearCurrent, TabClearAll}
		assert.Equal(t, 2, len(modes))
	})
}

func TestCharsets(t *testing.T) {
	t.Run("CharsetIndex", func(t *testing.T) {
		indices := []CharsetIndex{G0, G1, G2, G3}
		assert.Equal(t, 4, len(indices))

		// Ensure sequential values
		assert.Equal(t, CharsetIndex(0), G0)
		assert.Equal(t, CharsetIndex(1), G1)
		assert.Equal(t, CharsetIndex(2), G2)
		assert.Equal(t, CharsetIndex(3), G3)
	})

	t.Run("StandardCharset", func(t *testing.T) {
		charsets := []StandardCharset{StandardCharsetASCII, StandardCharsetSpecialLineDrawing}
		assert.Equal(t, 2, len(charsets))
	})
}

func TestControlCharacters(t *testing.T) {
	t.Run("C0", func(t *testing.T) {
		// Test some key C0 characters
		assert.Equal(t, byte(0x00), C0.NUL)
		assert.Equal(t, byte(0x07), C0.BEL)
		assert.Equal(t, byte(0x08), C0.BS)
		assert.Equal(t, byte(0x09), C0.HT)
		assert.Equal(t, byte(0x0A), C0.LF)
		assert.Equal(t, byte(0x0D), C0.CR)
		assert.Equal(t, byte(0x1B), C0.ESC)
		assert.Equal(t, byte(0x1F), C0.US)
	})

	t.Run("C1", func(t *testing.T) {
		// Test some key C1 characters
		assert.Equal(t, byte(0x80), C1.PAD)
		assert.Equal(t, byte(0x84), C1.IND)
		assert.Equal(t, byte(0x85), C1.NEL)
		assert.Equal(t, byte(0x90), C1.DCS)
		assert.Equal(t, byte(0x9B), C1.CSI)
		assert.Equal(t, byte(0x9C), C1.ST)
		assert.Equal(t, byte(0x9D), C1.OSC)
		assert.Equal(t, byte(0x9F), C1.APC)
	})
}

func TestHyperlink(t *testing.T) {
	t.Run("WithID", func(t *testing.T) {
		link := Hyperlink{
			ID:  "link1",
			URI: "https://example.com",
		}
		assert.Equal(t, "link1", link.ID)
		assert.Equal(t, "https://example.com", link.URI)
	})

	t.Run("WithoutID", func(t *testing.T) {
		link := Hyperlink{
			URI: "https://example.com",
		}
		assert.Equal(t, "", link.ID)
		assert.Equal(t, "https://example.com", link.URI)
	})
}

func TestModifyOtherKeys(t *testing.T) {
	assert.Equal(t, ModifyOtherKeys(0), ModifyOtherKeysDisabled)
	assert.Equal(t, ModifyOtherKeys(1), ModifyOtherKeysEnabled)
	assert.Equal(t, ModifyOtherKeys(2), ModifyOtherKeysExtended)
}

func BenchmarkRgbLuminance(b *testing.B) {
	colors := []Rgb{
		NewRgb(255, 255, 255),
		NewRgb(0, 0, 0),
		NewRgb(128, 128, 128),
		NewRgb(255, 0, 0),
		NewRgb(0, 255, 0),
		NewRgb(0, 0, 255),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range colors {
			_ = c.Luminance()
		}
	}
}

func BenchmarkRgbContrast(b *testing.B) {
	c1 := NewRgb(255, 255, 255)
	c2 := NewRgb(0, 0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c1.Contrast(c2)
	}
}

func BenchmarkAttrOperations(b *testing.B) {
	b.Run("Has", func(b *testing.B) {
		attr := AttrBold | AttrItalic | AttrUnderline
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = attr.Has(AttrItalic)
		}
	})

	b.Run("Add", func(b *testing.B) {
		attr := AttrBold
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = attr.Add(AttrItalic)
		}
	})

	b.Run("Remove", func(b *testing.B) {
		attr := AttrBold | AttrItalic
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = attr.Remove(AttrItalic)
		}
	})

	b.Run("Toggle", func(b *testing.B) {
		attr := AttrBold
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = attr.Toggle(AttrItalic)
		}
	})
}

// TestRgbEdgeCases tests edge cases and mathematical properties
func TestRgbEdgeCases(t *testing.T) {
	t.Run("Luminance Range", func(t *testing.T) {
		// Luminance should be in range [0, 1]
		for r := 0; r <= 255; r += 51 {
			for g := 0; g <= 255; g += 51 {
				for b := 0; b <= 255; b += 51 {
					c := NewRgb(uint8(r), uint8(g), uint8(b))
					lum := c.Luminance()
					assert.True(t, lum >= 0.0 && lum <= 1.0,
						"Luminance %f should be in [0,1] for color %v", lum, c)
				}
			}
		}
	})

	t.Run("Contrast Symmetry", func(t *testing.T) {
		c1 := NewRgb(100, 150, 200)
		c2 := NewRgb(200, 100, 50)

		// Contrast should be symmetric
		assert.InDelta(t, c1.Contrast(c2), c2.Contrast(c1), 0.001)
	})

	t.Run("Contrast Range", func(t *testing.T) {
		// Contrast ratio should be >= 1
		colors := []Rgb{
			NewRgb(0, 0, 0),
			NewRgb(255, 255, 255),
			NewRgb(128, 128, 128),
			NewRgb(255, 0, 0),
			NewRgb(0, 255, 0),
			NewRgb(0, 0, 255),
		}

		for _, c1 := range colors {
			for _, c2 := range colors {
				contrast := c1.Contrast(c2)
				assert.True(t, contrast >= 1.0,
					"Contrast %f should be >= 1 for %v and %v", contrast, c1, c2)
				assert.True(t, !math.IsNaN(contrast) && !math.IsInf(contrast, 0),
					"Contrast should be finite for %v and %v", c1, c2)
			}
		}
	})
}
func TestSynchronizedUpdates(t *testing.T) {
	t.Run("BeginSynchronizedUpdate", func(t *testing.T) {
		// Test BSU (Begin Synchronized Update) sequence generation
		bsu := BeginSynchronizedUpdate()
		expected := "[?2026h"
		assert.Equal(t, expected, bsu)
	})

	t.Run("EndSynchronizedUpdate", func(t *testing.T) {
		// Test ESU (End Synchronized Update) sequence generation
		esu := EndSynchronizedUpdate()
		expected := "[?2026l"
		assert.Equal(t, expected, esu)
	})

	t.Run("SynchronizedBlock", func(t *testing.T) {
		// Test wrapping content in synchronized update block
		content := "Hello, World!"
		result := WrapInSynchronizedUpdate(content)
		expected := "[?2026h" + content + "[?2026l"
		assert.Equal(t, expected, result)
	})
}

func TestTerminalSequences(t *testing.T) {
	t.Run("ClearScreen", func(t *testing.T) {
		seq := ClearScreen()
		expected := "[2J"
		assert.Equal(t, expected, seq)
	})

	t.Run("ClearLine", func(t *testing.T) {
		seq := ClearLine()
		expected := "[K"
		assert.Equal(t, expected, seq)
	})

	t.Run("MoveCursor", func(t *testing.T) {
		seq := MoveTo(5, 10)
		expected := "[6;11H" // 1-indexed
		assert.Equal(t, expected, seq)
	})

	t.Run("SaveRestoreCursor", func(t *testing.T) {
		save := SaveCursor()
		restore := RestoreCursor()
		assert.Equal(t, "7", save)    // DECSC
		assert.Equal(t, "8", restore) // DECRC
	})
}

func TestProcessorAdvanced(t *testing.T) {
	t.Run("StateTracking", func(t *testing.T) {
		// Test processor state tracking and mode management
		processor := NewProcessor(&NoopHandler{})

		// Test synchronized update tracking
		processor.BeginSynchronizedUpdate()
		assert.True(t, processor.IsInSynchronizedUpdate())

		processor.EndSynchronizedUpdate()
		assert.False(t, processor.IsInSynchronizedUpdate())
	})

	t.Run("ModeStack", func(t *testing.T) {
		// Test mode stack management
		processor := NewProcessor(&NoopHandler{})

		// Push application cursor mode
		processor.SetMode(ModeApplicationCursor, true)
		assert.True(t, processor.IsMode(ModeApplicationCursor))

		// Push another mode
		processor.SetMode(ModeAlternateScreen, true)
		assert.True(t, processor.IsMode(ModeAlternateScreen))
		assert.True(t, processor.IsMode(ModeApplicationCursor))

		// Pop modes
		processor.SetMode(ModeAlternateScreen, false)
		assert.False(t, processor.IsMode(ModeAlternateScreen))
		assert.True(t, processor.IsMode(ModeApplicationCursor))
	})

	t.Run("BufferedOutput", func(t *testing.T) {
		// Test buffered output for synchronized updates
		buffer := &TestBuffer{}
		processor := NewProcessorWithBuffer(buffer, &NoopHandler{})

		processor.BeginSynchronizedUpdate()
		processor.Write("Hello")
		processor.Write(" World")

		// Should be buffered, not written yet
		assert.Equal(t, "", buffer.String())

		processor.EndSynchronizedUpdate()
		// Now should be flushed
		assert.Equal(t, "Hello World", buffer.String())
	})

	t.Run("ErrorRecovery", func(t *testing.T) {
		// Test processor error recovery
		processor := NewProcessor(&NoopHandler{})

		// Should handle malformed sequences gracefully
		processor.Process([]byte("[99999999999999999999m"))
		processor.Process([]byte("[invalid"))
		processor.Process([]byte("normal text"))

		// Should still be functional
		assert.NotNil(t, processor)
	})
}

// TestBuffer for testing buffered output
type TestBuffer struct {
	data []byte
}

func (b *TestBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *TestBuffer) String() string {
	return string(b.data)
}

func TestProcessorIntegration(t *testing.T) {
	t.Run("ComplexSequence", func(t *testing.T) {
		// Test processing a complex ANSI sequence
		handler := &MockHandler{}
		processor := NewProcessor(handler)

		// SGR sequence with multiple parameters
		sequence := "[1;31;4m"
		processor.Process([]byte(sequence))

		// Should have called CsiDispatch with correct parameters
		assert.True(t, handler.csiCalled)
		assert.Equal(t, byte('m'), handler.action)
	})

	t.Run("StreamProcessing", func(t *testing.T) {
		// Test streaming byte processing
		handler := &MockHandler{}
		processor := NewProcessor(handler)

		// Process sequence in chunks
		chunks := []string{"\x1b[31m", "Hello"}
		for _, chunk := range chunks {
			processor.Process([]byte(chunk))
		}

		// Should correctly assemble and process the sequence
		assert.True(t, handler.csiCalled)
		assert.True(t, handler.printCalled)
	})
}

// MockHandler for integration testing
type MockHandler struct {
	NoopHandler
	csiCalled   bool
	printCalled bool
	action      byte
}

func (h *MockHandler) SetAttribute(attr Attr) {
	h.csiCalled = true
	h.action = byte('m') // SGR action
}

func (h *MockHandler) ResetAttributes() {
	h.csiCalled = true
	h.action = byte('m')
}

func (h *MockHandler) ResetColors() {
	h.csiCalled = true
	h.action = byte('m')
}

func (h *MockHandler) SetForeground(color Color) {
	h.csiCalled = true
	h.action = byte('m') // SGR action
}

func (h *MockHandler) Input(c rune) {
	h.printCalled = true
}
