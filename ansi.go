// Package govte provides ANSI terminal control sequence definitions and utilities.
package govte

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Rgb represents an RGB color value.
type Rgb struct {
	R uint8
	G uint8
	B uint8
}

// NewRgb creates a new RGB color.
func NewRgb(r, g, b uint8) Rgb {
	return Rgb{R: r, G: g, B: b}
}

// Luminance calculates the luminance of the color using W3C's algorithm.
// https://www.w3.org/TR/WCAG20/#relativeluminancedef
func (c Rgb) Luminance() float64 {
	channelLuminance := func(channel uint8) float64 {
		ch := float64(channel) / 255.0
		if ch <= 0.03928 {
			return ch / 12.92
		}
		return math.Pow((ch+0.055)/1.055, 2.4)
	}

	rLum := channelLuminance(c.R)
	gLum := channelLuminance(c.G)
	bLum := channelLuminance(c.B)

	return 0.2126*rLum + 0.7152*gLum + 0.0722*bLum
}

// Contrast calculates the contrast ratio between two colors using W3C's algorithm.
// https://www.w3.org/TR/WCAG20/#contrast-ratiodef
func (c Rgb) Contrast(other Rgb) float64 {
	selfLum := c.Luminance()
	otherLum := other.Luminance()

	var lighter, darker float64
	if selfLum > otherLum {
		lighter = selfLum
		darker = otherLum
	} else {
		lighter = otherLum
		darker = selfLum
	}

	return (lighter + 0.05) / (darker + 0.05)
}

// Add returns the result of adding two RGB colors with saturation.
func (c Rgb) Add(other Rgb) Rgb {
	return Rgb{
		R: saturateAdd(c.R, other.R),
		G: saturateAdd(c.G, other.G),
		B: saturateAdd(c.B, other.B),
	}
}

// Sub returns the result of subtracting two RGB colors with saturation.
func (c Rgb) Sub(other Rgb) Rgb {
	return Rgb{
		R: saturateSub(c.R, other.R),
		G: saturateSub(c.G, other.G),
		B: saturateSub(c.B, other.B),
	}
}

// Mul returns the result of multiplying an RGB color by a scalar with clamping.
func (c Rgb) Mul(factor float64) Rgb {
	return Rgb{
		R: clamp(float64(c.R) * factor),
		G: clamp(float64(c.G) * factor),
		B: clamp(float64(c.B) * factor),
	}
}

// Helper functions for arithmetic operations
func saturateAdd(a, b uint8) uint8 {
	result := uint16(a) + uint16(b)
	if result > 255 {
		return 255
	}
	return uint8(result)
}

func saturateSub(a, b uint8) uint8 {
	if a < b {
		return 0
	}
	return a - b
}

func clamp(value float64) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}

// String returns the color as a hex string.
func (c Rgb) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// RgbFromString parses a hex color string into an RGB color.
// Supports formats: "#rrggbb", "0xrrggbb"
// Returns the parsed color and true if successful, or zero color and false if invalid.
func RgbFromString(s string) (Rgb, bool) {
	// Handle empty string
	if len(s) == 0 {
		return Rgb{}, false
	}
	
	// Remove prefix and validate length
	var hexStr string
	if strings.HasPrefix(s, "#") {
		hexStr = s[1:]
	} else if strings.HasPrefix(strings.ToLower(s), "0x") {
		hexStr = s[2:]
	} else {
		return Rgb{}, false
	}
	
	// Must be exactly 6 hex characters
	if len(hexStr) != 6 {
		return Rgb{}, false
	}
	
	// Parse hex string
	val, err := strconv.ParseUint(hexStr, 16, 32)
	if err != nil {
		return Rgb{}, false
	}
	
	// Extract RGB components
	r := uint8((val >> 16) & 0xFF)
	g := uint8((val >> 8) & 0xFF)
	b := uint8(val & 0xFF)
	
	return Rgb{R: r, G: g, B: b}, true
}

// Blend blends this color with another using alpha blending.
// alpha=0.0 returns this color, alpha=1.0 returns other color.
func (c Rgb) Blend(other Rgb, alpha float64) Rgb {
	if alpha <= 0.0 {
		return c
	}
	if alpha >= 1.0 {
		return other
	}
	
	invAlpha := 1.0 - alpha
	return Rgb{
		R: uint8(float64(c.R)*invAlpha + float64(other.R)*alpha),
		G: uint8(float64(c.G)*invAlpha + float64(other.G)*alpha),
		B: uint8(float64(c.B)*invAlpha + float64(other.B)*alpha),
	}
}

// Lerp performs linear interpolation between this color and another.
// t=0.0 returns this color, t=1.0 returns other color.
func (c Rgb) Lerp(other Rgb, t float64) Rgb {
	return c.Blend(other, t)
}

// Distance calculates the Euclidean distance between two colors in RGB space.
func (c Rgb) Distance(other Rgb) float64 {
	dr := float64(c.R) - float64(other.R)
	dg := float64(c.G) - float64(other.G)
	db := float64(c.B) - float64(other.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// PerceptualDistance calculates perceptual color distance weighted by human vision.
// Uses redmean approximation for better perceptual accuracy than Euclidean.
func (c Rgb) PerceptualDistance(other Rgb) float64 {
	rMean := (float64(c.R) + float64(other.R)) / 2.0
	dr := float64(c.R) - float64(other.R)
	dg := float64(c.G) - float64(other.G)
	db := float64(c.B) - float64(other.B)
	
	// Redmean color difference formula
	weightR := 2.0 + rMean/256.0
	weightG := 4.0
	weightB := 2.0 + (255.0-rMean)/256.0
	
	return math.Sqrt(weightR*dr*dr + weightG*dg*dg + weightB*db*db)
}

// NamedColor represents the 16 standard terminal colors plus bright variants.
type NamedColor uint8

const (
	// Standard colors (0-7)
	Black NamedColor = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	// Bright colors (8-15)
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
	// Special colors
	Foreground NamedColor = 16
	Background NamedColor = 17
)

// ToRgb converts a named color to its default RGB value.
func (c NamedColor) ToRgb() Rgb {
	switch c {
	case Black:
		return Rgb{0, 0, 0}
	case Red:
		return Rgb{170, 0, 0}
	case Green:
		return Rgb{0, 170, 0}
	case Yellow:
		return Rgb{170, 85, 0}
	case Blue:
		return Rgb{0, 0, 170}
	case Magenta:
		return Rgb{170, 0, 170}
	case Cyan:
		return Rgb{0, 170, 170}
	case White:
		return Rgb{170, 170, 170}
	case BrightBlack:
		return Rgb{85, 85, 85}
	case BrightRed:
		return Rgb{255, 85, 85}
	case BrightGreen:
		return Rgb{85, 255, 85}
	case BrightYellow:
		return Rgb{255, 255, 85}
	case BrightBlue:
		return Rgb{85, 85, 255}
	case BrightMagenta:
		return Rgb{255, 85, 255}
	case BrightCyan:
		return Rgb{85, 255, 255}
	case BrightWhite:
		return Rgb{255, 255, 255}
	default:
		return Rgb{0, 0, 0}
	}
}

// Color represents a terminal color which can be named, indexed, or RGB.
type Color struct {
	Type  ColorType
	Named NamedColor
	Index uint8
	Rgb   Rgb
}

// ColorType indicates the type of color.
type ColorType uint8

const (
	ColorTypeNamed ColorType = iota
	ColorTypeIndexed
	ColorTypeRgb
)

// NewNamedColor creates a color from a named color.
func NewNamedColor(c NamedColor) Color {
	return Color{Type: ColorTypeNamed, Named: c}
}

// NewIndexedColor creates a color from a palette index (0-255).
func NewIndexedColor(index uint8) Color {
	return Color{Type: ColorTypeIndexed, Index: index}
}

// NewRgbColor creates a color from RGB values.
func NewRgbColor(r, g, b uint8) Color {
	return Color{Type: ColorTypeRgb, Rgb: Rgb{r, g, b}}
}

// ToRgb converts any Color type to its RGB representation.
func (c Color) ToRgb() Rgb {
	switch c.Type {
	case ColorTypeNamed:
		return c.Named.ToRgb()
	case ColorTypeIndexed:
		return indexedColorToRgb(c.Index)
	case ColorTypeRgb:
		return c.Rgb
	default:
		return Rgb{0, 0, 0} // Default to black
	}
}

// indexedColorToRgb converts a palette index (0-255) to RGB
func indexedColorToRgb(index uint8) Rgb {
	switch {
	case index < 16:
		// Standard 16 colors (0-15)
		return NamedColor(index).ToRgb()
	case index < 232:
		// 216-color cube (16-231): 6x6x6 RGB values
		// Formula: index = 16 + 36*r + 6*g + b (where r,g,b are 0-5)
		cubeIndex := index - 16
		r := cubeIndex / 36
		g := (cubeIndex % 36) / 6
		b := cubeIndex % 6
		
		// Convert 0-5 range to 0-255 range using standard 6-level palette
		// Standard values: [0, 95, 135, 175, 215, 255]
		paletteValues := [6]uint8{0, 95, 135, 175, 215, 255}
		rVal := paletteValues[r]
		gVal := paletteValues[g]
		bVal := paletteValues[b]
		
		return Rgb{rVal, gVal, bVal}
	default:
		// 24-level grayscale ramp (232-255)
		// Formula: gray = 8 + (index - 232) * 10
		gray := uint8(8 + (index-232)*10)
		return Rgb{gray, gray, gray}
	}
}

// Hsl represents a color in HSL (Hue, Saturation, Lightness) color space.
type Hsl struct {
	H float64 // Hue: 0.0-1.0 (0°-360°)
	S float64 // Saturation: 0.0-1.0
	L float64 // Lightness: 0.0-1.0
}

// NewHsl creates a new HSL color.
func NewHsl(h, s, l float64) Hsl {
	return Hsl{H: h, S: s, L: l}
}

// ToHsl converts RGB color to HSL color space.
func (c Rgb) ToHsl() Hsl {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0
	
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min
	
	// Lightness
	l := (max + min) / 2.0
	
	if delta == 0 {
		// Achromatic (gray)
		return Hsl{H: 0, S: 0, L: l}
	}
	
	// Saturation
	var s float64
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2.0 - max - min)
	}
	
	// Hue
	var h float64
	switch max {
	case r:
		h = (g - b) / delta
		if g < b {
			h += 6.0
		}
	case g:
		h = (b-r)/delta + 2.0
	case b:
		h = (r-g)/delta + 4.0
	}
	h /= 6.0
	
	return Hsl{H: h, S: s, L: l}
}

// ToRgb converts HSL color to RGB color space.
func (hsl Hsl) ToRgb() Rgb {
	if hsl.S == 0 {
		// Achromatic (gray)
		gray := uint8(hsl.L * 255.0)
		return Rgb{gray, gray, gray}
	}
	
	hueToRgb := func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		if t < 1.0/6.0 {
			return p + (q-p)*6.0*t
		}
		if t < 1.0/2.0 {
			return q
		}
		if t < 2.0/3.0 {
			return p + (q-p)*(2.0/3.0-t)*6.0
		}
		return p
	}
	
	var q float64
	if hsl.L < 0.5 {
		q = hsl.L * (1.0 + hsl.S)
	} else {
		q = hsl.L + hsl.S - hsl.L*hsl.S
	}
	p := 2.0*hsl.L - q
	
	r := hueToRgb(p, q, hsl.H+1.0/3.0)
	g := hueToRgb(p, q, hsl.H)
	b := hueToRgb(p, q, hsl.H-1.0/3.0)
	
	return Rgb{
		R: uint8(r * 255.0),
		G: uint8(g * 255.0),
		B: uint8(b * 255.0),
	}
}

// ColorBlindnessType represents different types of color blindness.
type ColorBlindnessType uint8

const (
	ColorBlindnessDeuteranopia ColorBlindnessType = iota // Green-blind
	ColorBlindnessProtanopia                             // Red-blind
	ColorBlindnessTritanopia                             // Blue-blind
)

// IsSafeWith checks if two colors are distinguishable for people with color blindness.
func (c Rgb) IsSafeWith(other Rgb, cbType ColorBlindnessType) bool {
	// For deuteranopia (green-blind), red and green colors are problematic
	if cbType == ColorBlindnessDeuteranopia {
		// Check if colors are primarily red/green and would be confused
		cLum := c.Luminance()
		otherLum := other.Luminance()
		
		// If both colors have similar luminance but different R/G ratios, they're unsafe
		lumDiff := math.Abs(cLum - otherLum)
		if lumDiff < 0.1 { // Similar luminance
			// Check if they differ mainly in R/G channels
			rDiff := math.Abs(float64(c.R) - float64(other.R))
			gDiff := math.Abs(float64(c.G) - float64(other.G))
			if rDiff > 100 || gDiff > 100 { // Large R/G difference
				return false // Unsafe for deuteranopes
			}
		}
		
		// Use luminance contrast as backup
		return c.Contrast(other) >= 3.0
	}
	
	// For other color blindness types, use simpler simulation
	var c1, c2 Rgb
	switch cbType {
	case ColorBlindnessProtanopia:
		// Remove red sensitivity
		c1 = Rgb{0, c.G, c.B}
		c2 = Rgb{0, other.G, other.B}
	case ColorBlindnessTritanopia:
		// Remove blue sensitivity
		c1 = Rgb{c.R, c.G, 0}
		c2 = Rgb{other.R, other.G, 0}
	default:
		c1, c2 = c, other
	}
	
	return c1.Contrast(c2) >= 3.0
}

// Terminal control sequence generation functions

// BeginSynchronizedUpdate returns the ANSI sequence to begin synchronized updates.
// This prevents screen flickering during complex updates.
func BeginSynchronizedUpdate() string {
	return "\x1b[?2026h"
}

// EndSynchronizedUpdate returns the ANSI sequence to end synchronized updates.
func EndSynchronizedUpdate() string {
	return "\x1b[?2026l"
}

// WrapInSynchronizedUpdate wraps content in synchronized update sequences.
func WrapInSynchronizedUpdate(content string) string {
	return BeginSynchronizedUpdate() + content + EndSynchronizedUpdate()
}

// ClearScreen returns the ANSI sequence to clear the entire screen.
func ClearScreen() string {
	return "\x1b[2J"
}

// ClearLine returns the ANSI sequence to clear from cursor to end of line.
func ClearLine() string {
	return "\x1b[K"
}

// MoveTo returns the ANSI sequence to move cursor to specific position.
// row and col are 0-indexed, but ANSI sequences are 1-indexed.
func MoveTo(row, col int) string {
	return fmt.Sprintf("\x1b[%d;%dH", row+1, col+1)
}

// SaveCursor returns the ANSI sequence to save current cursor position (DECSC).
func SaveCursor() string {
	return "\x1b7"
}

// RestoreCursor returns the ANSI sequence to restore saved cursor position (DECRC).
func RestoreCursor() string {
	return "\x1b8"
}

// Attr represents text formatting attributes.
type Attr uint32

const (
	AttrNone          Attr = 0
	AttrBold          Attr = 1 << 0
	AttrDim           Attr = 1 << 1
	AttrItalic        Attr = 1 << 2
	AttrUnderline     Attr = 1 << 3
	AttrBlinking      Attr = 1 << 4
	AttrReverse       Attr = 1 << 5
	AttrHidden        Attr = 1 << 6
	AttrStrikethrough Attr = 1 << 7
	AttrDoubleUnderline Attr = 1 << 8
	AttrCurlyUnderline  Attr = 1 << 9
	AttrDottedUnderline Attr = 1 << 10
	AttrDashedUnderline Attr = 1 << 11
)

// Has checks if the attribute set contains the given attribute.
func (a Attr) Has(attr Attr) bool {
	return a&attr != 0
}

// Add adds an attribute to the set.
func (a Attr) Add(attr Attr) Attr {
	return a | attr
}

// Remove removes an attribute from the set.
func (a Attr) Remove(attr Attr) Attr {
	return a &^ attr
}

// Toggle toggles an attribute in the set.
func (a Attr) Toggle(attr Attr) Attr {
	return a ^ attr
}

// Mode represents a terminal mode.
type Mode uint16

const (
	ModeNone Mode = 0
	// ANSI modes
	ModeKeyboardAction          Mode = 2
	ModeInsert                  Mode = 4
	ModeReplace                 Mode = 4 | 0x100 // with high bit to distinguish
	ModeSendReceive             Mode = 12
	ModeAutomaticNewline        Mode = 20
	// Private modes (start at 0x200)
	ModeApplicationCursor       Mode = 0x200 + 1
	ModeApplicationKeypad       Mode = 0x200 + 2
	ModeAlternateScreen         Mode = 0x200 + 3
	ModeShowCursor              Mode = 0x200 + 25
	ModeSaveRestoreCursor       Mode = 0x200 + 1048
	ModeAlternateScreenBuffer   Mode = 0x200 + 1049
	ModeBracketedPaste          Mode = 0x200 + 2004
	ModeSynchronizedOutput      Mode = 0x200 + 2026
)

// IsPrivate checks if this is a private mode.
func (m Mode) IsPrivate() bool {
	return m >= 0x200
}

// CursorShape represents the shape of the cursor.
type CursorShape uint8

const (
	CursorShapeBlock CursorShape = iota
	CursorShapeUnderline
	CursorShapeBeam
)

// CursorStyle represents cursor display properties.
type CursorStyle struct {
	Shape    CursorShape
	Blinking bool
}

// LineClearMode specifies how to clear a line.
type LineClearMode uint8

const (
	LineClearRight LineClearMode = iota // Clear from cursor to end of line
	LineClearLeft                        // Clear from beginning to cursor
	LineClearAll                         // Clear entire line
)

// ClearMode specifies how to clear the screen.
type ClearMode uint8

const (
	ClearBelow ClearMode = iota // Clear from cursor to end of screen
	ClearAbove                  // Clear from beginning to cursor
	ClearAll                    // Clear entire screen
	ClearSaved                  // Clear saved lines (scrollback)
)

// TabulationClearMode specifies how to clear tab stops.
type TabulationClearMode uint8

const (
	TabClearCurrent TabulationClearMode = iota // Clear tab at current position
	TabClearAll                                 // Clear all tabs
)

// String returns the string representation of TabulationClearMode.
func (m TabulationClearMode) String() string {
	switch m {
	case TabClearCurrent:
		return "TabClearCurrent"
	case TabClearAll:
		return "TabClearAll"
	default:
		return "Unknown"
	}
}


// C0 defines C0 control characters (0x00-0x1F).
var C0 = struct {
	NUL byte // Null
	SOH byte // Start of Heading
	STX byte // Start of Text
	ETX byte // End of Text
	EOT byte // End of Transmission
	ENQ byte // Enquiry
	ACK byte // Acknowledge
	BEL byte // Bell
	BS  byte // Backspace
	HT  byte // Horizontal Tab
	LF  byte // Line Feed
	VT  byte // Vertical Tab
	FF  byte // Form Feed
	CR  byte // Carriage Return
	SO  byte // Shift Out
	SI  byte // Shift In
	DLE byte // Data Link Escape
	DC1 byte // Device Control 1 (XON)
	DC2 byte // Device Control 2
	DC3 byte // Device Control 3 (XOFF)
	DC4 byte // Device Control 4
	NAK byte // Negative Acknowledge
	SYN byte // Synchronous Idle
	ETB byte // End of Transmission Block
	CAN byte // Cancel
	EM  byte // End of Medium
	SUB byte // Substitute
	ESC byte // Escape
	FS  byte // File Separator
	GS  byte // Group Separator
	RS  byte // Record Separator
	US  byte // Unit Separator
}{
	NUL: 0x00, SOH: 0x01, STX: 0x02, ETX: 0x03,
	EOT: 0x04, ENQ: 0x05, ACK: 0x06, BEL: 0x07,
	BS: 0x08, HT: 0x09, LF: 0x0A, VT: 0x0B,
	FF: 0x0C, CR: 0x0D, SO: 0x0E, SI: 0x0F,
	DLE: 0x10, DC1: 0x11, DC2: 0x12, DC3: 0x13,
	DC4: 0x14, NAK: 0x15, SYN: 0x16, ETB: 0x17,
	CAN: 0x18, EM: 0x19, SUB: 0x1A, ESC: 0x1B,
	FS: 0x1C, GS: 0x1D, RS: 0x1E, US: 0x1F,
}

// C1 defines C1 control characters (0x80-0x9F).
var C1 = struct {
	PAD  byte // Padding Character
	HOP  byte // High Octet Preset
	BPH  byte // Break Permitted Here
	NBH  byte // No Break Here
	IND  byte // Index
	NEL  byte // Next Line
	SSA  byte // Start of Selected Area
	ESA  byte // End of Selected Area
	HTS  byte // Horizontal Tab Set
	HTJ  byte // Horizontal Tab with Justification
	VTS  byte // Vertical Tab Set
	PLD  byte // Partial Line Down
	PLU  byte // Partial Line Up
	RI   byte // Reverse Index
	SS2  byte // Single Shift 2
	SS3  byte // Single Shift 3
	DCS  byte // Device Control String
	PU1  byte // Private Use 1
	PU2  byte // Private Use 2
	STS  byte // Set Transmit State
	CCH  byte // Cancel Character
	MW   byte // Message Waiting
	SPA  byte // Start of Protected Area
	EPA  byte // End of Protected Area
	SOS  byte // Start of String
	SGCI byte // Single Graphic Character Introducer
	SCI  byte // Single Character Introducer
	CSI  byte // Control Sequence Introducer
	ST   byte // String Terminator
	OSC  byte // Operating System Command
	PM   byte // Privacy Message
	APC  byte // Application Program Command
}{
	PAD: 0x80, HOP: 0x81, BPH: 0x82, NBH: 0x83,
	IND: 0x84, NEL: 0x85, SSA: 0x86, ESA: 0x87,
	HTS: 0x88, HTJ: 0x89, VTS: 0x8A, PLD: 0x8B,
	PLU: 0x8C, RI: 0x8D, SS2: 0x8E, SS3: 0x8F,
	DCS: 0x90, PU1: 0x91, PU2: 0x92, STS: 0x93,
	CCH: 0x94, MW: 0x95, SPA: 0x96, EPA: 0x97,
	SOS: 0x98, SGCI: 0x99, SCI: 0x9A, CSI: 0x9B,
	ST: 0x9C, OSC: 0x9D, PM: 0x9E, APC: 0x9F,
}

// Hyperlink represents a terminal hyperlink.
type Hyperlink struct {
	ID  string // Optional identifier for the hyperlink
	URI string // The URI to link to
}

// ModifyOtherKeys represents the state of the modifyOtherKeys mode.
type ModifyOtherKeys uint8

const (
	ModifyOtherKeysDisabled ModifyOtherKeys = 0
	ModifyOtherKeysEnabled  ModifyOtherKeys = 1
	ModifyOtherKeysExtended ModifyOtherKeys = 2
)

// CharsetIndex identifies which graphic character set can be designated as G0-G3.
type CharsetIndex int

const (
	// G0 is the default set, designated as ASCII at startup
	G0 CharsetIndex = iota
	G1
	G2
	G3
)

// String returns the string representation of CharsetIndex
func (c CharsetIndex) String() string {
	switch c {
	case G0:
		return "G0"
	case G1:
		return "G1"
	case G2:
		return "G2"
	case G3:
		return "G3"
	default:
		return "Unknown"
	}
}

// StandardCharset represents standard or common character sets which can be designated as G0-G3.
type StandardCharset int

const (
	// StandardCharsetAscii is the default ASCII charset
	StandardCharsetAscii StandardCharset = iota
	// StandardCharsetSpecialLineDrawing is the special character and line drawing set
	StandardCharsetSpecialLineDrawing
)

// String returns the string representation of StandardCharset
func (s StandardCharset) String() string {
	switch s {
	case StandardCharsetAscii:
		return "Ascii"
	case StandardCharsetSpecialLineDrawing:
		return "SpecialCharacterAndLineDrawing"
	default:
		return "Unknown"
	}
}

// Map transforms a character according to the charset mapping.
// ASCII is the common case and does as little as possible.
func (s StandardCharset) Map(c rune) rune {
	switch s {
	case StandardCharsetAscii:
		return c
	case StandardCharsetSpecialLineDrawing:
		return mapSpecialLineDrawing(c)
	default:
		return c
	}
}

// mapSpecialLineDrawing maps characters for the special line drawing charset
func mapSpecialLineDrawing(c rune) rune {
	switch c {
	case '_':
		return ' '
	case '`':
		return '◆'
	case 'a':
		return '▒'
	case 'b':
		return '\u2409' // Symbol for horizontal tabulation
	case 'c':
		return '\u240c' // Symbol for form feed
	case 'd':
		return '\u240d' // Symbol for carriage return
	case 'e':
		return '\u240a' // Symbol for line feed
	case 'f':
		return '°'
	case 'g':
		return '±'
	case 'h':
		return '\u2424' // Symbol for newline
	case 'i':
		return '\u240b' // Symbol for vertical tabulation
	case 'j':
		return '┘'
	case 'k':
		return '┐'
	case 'l':
		return '┌'
	case 'm':
		return '└'
	case 'n':
		return '┼'
	case 'o':
		return '⎺'
	case 'p':
		return '⎻'
	case 'q':
		return '─'
	case 'r':
		return '⎼'
	case 's':
		return '⎽'
	case 't':
		return '├'
	case 'u':
		return '┤'
	case 'v':
		return '┴'
	case 'w':
		return '┬'
	case 'x':
		return '│'
	case 'y':
		return '≤'
	case 'z':
		return '≥'
	case '{':
		return 'π'
	case '|':
		return '≠'
	case '}':
		return '£'
	case '~':
		return '·'
	default:
		return c
	}
}

