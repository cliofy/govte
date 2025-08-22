//! Terminal character and styling definitions
//! Go port of the Rust implementation - simplified from Zellij's implementation

package terminal

import (
	"fmt"
	"strings"
)

// TerminalCharacter represents a single terminal character with its styling
type TerminalCharacter struct {
	Character rune
	Width     int
	Styles    CharacterStyles
}

// NewTerminalCharacter creates a new terminal character with default styles
func NewTerminalCharacter(character rune) TerminalCharacter {
	width := runeWidth(character)
	return TerminalCharacter{
		Character: character,
		Width:     width,
		Styles:    DefaultCharacterStyles(),
	}
}

// NewStyledTerminalCharacter creates a new terminal character with specific styles
func NewStyledTerminalCharacter(character rune, styles CharacterStyles) TerminalCharacter {
	width := runeWidth(character)
	return TerminalCharacter{
		Character: character,
		Width:     width,
		Styles:    styles,
	}
}

// EmptyTerminalCharacter returns a space character with default styles
func EmptyTerminalCharacter() TerminalCharacter {
	return TerminalCharacter{
		Character: ' ',
		Width:     1,
		Styles:    DefaultCharacterStyles(),
	}
}

// CharacterStyles holds character styling attributes
type CharacterStyles struct {
	Foreground *AnsiCode
	Background *AnsiCode
	Bold       *AnsiCode
	Dim        *AnsiCode
	Italic     *AnsiCode
	Underline  *AnsiCode
	Blink      *AnsiCode
	Reverse    *AnsiCode
	Hidden     *AnsiCode
	Strike     *AnsiCode
}

// DefaultCharacterStyles returns default character styles (all nil)
func DefaultCharacterStyles() CharacterStyles {
	return CharacterStyles{}
}

// ToAnsiSequence converts styles to ANSI escape sequence
func (cs *CharacterStyles) ToAnsiSequence() string {
	var sequence strings.Builder

	// Handle text attributes
	if cs.Bold != nil && cs.Bold.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[1m")
	}
	if cs.Dim != nil && cs.Dim.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[2m")
	}
	if cs.Italic != nil && cs.Italic.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[3m")
	}
	if cs.Underline != nil && cs.Underline.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[4m")
	}
	if cs.Blink != nil && cs.Blink.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[5m")
	}
	if cs.Reverse != nil && cs.Reverse.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[7m")
	}
	if cs.Hidden != nil && cs.Hidden.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[8m")
	}
	if cs.Strike != nil && cs.Strike.Type == AnsiCodeTypeOn {
		sequence.WriteString("\x1b[9m")
	}

	// Handle colors
	if cs.Foreground != nil {
		sequence.WriteString(cs.Foreground.ToAnsiFgSequence())
	}
	if cs.Background != nil {
		sequence.WriteString(cs.Background.ToAnsiBgSequence())
	}

	return sequence.String()
}

// DiffersFrom checks if this style is different from another (for optimization)
func (cs *CharacterStyles) DiffersFrom(other *CharacterStyles) bool {
	return !cs.equals(other)
}

// equals compares two CharacterStyles for equality
func (cs *CharacterStyles) equals(other *CharacterStyles) bool {
	return ansiCodeEquals(cs.Foreground, other.Foreground) &&
		ansiCodeEquals(cs.Background, other.Background) &&
		ansiCodeEquals(cs.Bold, other.Bold) &&
		ansiCodeEquals(cs.Dim, other.Dim) &&
		ansiCodeEquals(cs.Italic, other.Italic) &&
		ansiCodeEquals(cs.Underline, other.Underline) &&
		ansiCodeEquals(cs.Blink, other.Blink) &&
		ansiCodeEquals(cs.Reverse, other.Reverse) &&
		ansiCodeEquals(cs.Hidden, other.Hidden) &&
		ansiCodeEquals(cs.Strike, other.Strike)
}

// AddStyleFromAnsiParams applies SGR (Select Graphic Rendition) parameters
func (cs *CharacterStyles) AddStyleFromAnsiParams(params [][]uint16) {
	i := 0
	for i < len(params) {
		if len(params[i]) == 0 {
			i++
			continue
		}
		param := params[i][0]

		switch param {
		case 0: // Reset
			*cs = DefaultCharacterStyles()
		case 1: // Bold
			bold := AnsiCodeOn()
			cs.Bold = &bold
		case 2: // Dim
			dim := AnsiCodeOn()
			cs.Dim = &dim
		case 3: // Italic
			italic := AnsiCodeOn()
			cs.Italic = &italic
		case 4: // Underline
			underline := AnsiCodeOn()
			cs.Underline = &underline
		case 5, 6: // Blink
			blink := AnsiCodeOn()
			cs.Blink = &blink
		case 7: // Reverse
			reverse := AnsiCodeOn()
			cs.Reverse = &reverse
		case 8: // Hidden
			hidden := AnsiCodeOn()
			cs.Hidden = &hidden
		case 9: // Strike
			strike := AnsiCodeOn()
			cs.Strike = &strike
		// Reset individual attributes
		case 21: // Bold off
			reset := AnsiCodeReset()
			cs.Bold = &reset
		case 22: // Bold and dim off
			reset := AnsiCodeReset()
			cs.Bold = &reset
			cs.Dim = &reset
		case 23: // Italic off
			reset := AnsiCodeReset()
			cs.Italic = &reset
		case 24: // Underline off
			reset := AnsiCodeReset()
			cs.Underline = &reset
		case 25: // Blink off
			reset := AnsiCodeReset()
			cs.Blink = &reset
		case 27: // Reverse off
			reset := AnsiCodeReset()
			cs.Reverse = &reset
		case 28: // Hidden off
			reset := AnsiCodeReset()
			cs.Hidden = &reset
		case 29: // Strike off
			reset := AnsiCodeReset()
			cs.Strike = &reset
		// Foreground colors
		case 30, 31, 32, 33, 34, 35, 36, 37:
			color := AnsiCodeNamedColor(NamedColorFromAnsi(uint8(param)))
			cs.Foreground = &color
		case 38: // Extended foreground color
			consumed := cs.handleExtendedColor(params[i:], true)
			i += consumed - 1 // -1 because loop will increment
		case 39: // Default foreground
			reset := AnsiCodeReset()
			cs.Foreground = &reset
		// Background colors
		case 40, 41, 42, 43, 44, 45, 46, 47:
			color := AnsiCodeNamedColor(NamedColorFromAnsi(uint8(param - 10)))
			cs.Background = &color
		case 48: // Extended background color
			consumed := cs.handleExtendedColor(params[i:], false)
			i += consumed - 1 // -1 because loop will increment
		case 49: // Default background
			reset := AnsiCodeReset()
			cs.Background = &reset
		// Bright foreground colors
		case 90, 91, 92, 93, 94, 95, 96, 97:
			color := AnsiCodeNamedColor(NamedColorFromAnsi(uint8(param - 60)))
			cs.Foreground = &color
		// Bright background colors
		case 100, 101, 102, 103, 104, 105, 106, 107:
			color := AnsiCodeNamedColor(NamedColorFromAnsi(uint8(param - 60)))
			cs.Background = &color
		}
		i++
	}
}

// handleExtendedColor processes 38/48 (extended color) sequences
func (cs *CharacterStyles) handleExtendedColor(params [][]uint16, isForeground bool) int {
	if len(params) < 2 || len(params[1]) == 0 {
		return 1
	}

	colorType := params[1][0]
	switch colorType {
	case 2: // RGB color
		if len(params) < 5 {
			return 1
		}
		var r, g, b uint8 = 0, 0, 0
		if len(params[2]) > 0 {
			r = uint8(params[2][0])
		}
		if len(params[3]) > 0 {
			g = uint8(params[3][0])
		}
		if len(params[4]) > 0 {
			b = uint8(params[4][0])
		}
		color := AnsiCodeRgbCode(r, g, b)
		if isForeground {
			cs.Foreground = &color
		} else {
			cs.Background = &color
		}
		return 5
	case 5: // 256 color
		if len(params) < 3 || len(params[2]) == 0 {
			return 2
		}
		index := uint8(params[2][0])
		color := AnsiCodeColorIndex(index)
		if isForeground {
			cs.Foreground = &color
		} else {
			cs.Background = &color
		}
		return 3
	}
	return 1
}

// AnsiCodeType represents the type of ANSI color code
type AnsiCodeType int

const (
	AnsiCodeTypeOn AnsiCodeType = iota
	AnsiCodeTypeReset
	AnsiCodeTypeNamedColor
	AnsiCodeTypeRgb
	AnsiCodeTypeColorIndex
)

// AnsiCode represents ANSI color codes and attributes with proper data storage
type AnsiCode struct {
	Type       AnsiCodeType
	NamedColor NamedColor
	RGB        struct{ R, G, B uint8 }
	ColorIndex uint8
}

// AnsiCodeOn creates an "On" AnsiCode
func AnsiCodeOn() AnsiCode {
	return AnsiCode{Type: AnsiCodeTypeOn}
}

// AnsiCodeReset creates a "Reset" AnsiCode
func AnsiCodeReset() AnsiCode {
	return AnsiCode{Type: AnsiCodeTypeReset}
}

// AnsiCodeNamedColor creates an AnsiCode for a named color
func AnsiCodeNamedColor(color NamedColor) AnsiCode {
	return AnsiCode{
		Type:       AnsiCodeTypeNamedColor,
		NamedColor: color,
	}
}

// AnsiCodeRgbCode creates an AnsiCode for RGB color
func AnsiCodeRgbCode(r, g, b uint8) AnsiCode {
	return AnsiCode{
		Type: AnsiCodeTypeRgb,
		RGB:  struct{ R, G, B uint8 }{R: r, G: g, B: b},
	}
}

// AnsiCodeColorIndex creates an AnsiCode for 256-color index
func AnsiCodeColorIndex(index uint8) AnsiCode {
	return AnsiCode{
		Type:       AnsiCodeTypeColorIndex,
		ColorIndex: index,
	}
}

// ToAnsiFgSequence converts to ANSI foreground color sequence
func (ac AnsiCode) ToAnsiFgSequence() string {
	switch ac.Type {
	case AnsiCodeTypeOn:
		return ""
	case AnsiCodeTypeReset:
		return "\x1b[39m"
	case AnsiCodeTypeNamedColor:
		return fmt.Sprintf("\x1b[%dm", ac.NamedColor.ToAnsiFg())
	case AnsiCodeTypeRgb:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", ac.RGB.R, ac.RGB.G, ac.RGB.B)
	case AnsiCodeTypeColorIndex:
		return fmt.Sprintf("\x1b[38;5;%dm", ac.ColorIndex)
	default:
		return ""
	}
}

// ToAnsiBgSequence converts to ANSI background color sequence
func (ac AnsiCode) ToAnsiBgSequence() string {
	switch ac.Type {
	case AnsiCodeTypeOn:
		return ""
	case AnsiCodeTypeReset:
		return "\x1b[49m"
	case AnsiCodeTypeNamedColor:
		return fmt.Sprintf("\x1b[%dm", ac.NamedColor.ToAnsiBg())
	case AnsiCodeTypeRgb:
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", ac.RGB.R, ac.RGB.G, ac.RGB.B)
	case AnsiCodeTypeColorIndex:
		return fmt.Sprintf("\x1b[48;5;%dm", ac.ColorIndex)
	default:
		return ""
	}
}

// NamedColor represents named ANSI colors
type NamedColor int

const (
	NamedColorBlack NamedColor = iota
	NamedColorRed
	NamedColorGreen
	NamedColorYellow
	NamedColorBlue
	NamedColorMagenta
	NamedColorCyan
	NamedColorWhite
	NamedColorBrightBlack
	NamedColorBrightRed
	NamedColorBrightGreen
	NamedColorBrightYellow
	NamedColorBrightBlue
	NamedColorBrightMagenta
	NamedColorBrightCyan
	NamedColorBrightWhite
	NamedColorCount // For bounds checking
)

// NamedColorFromAnsi converts from ANSI color code
func NamedColorFromAnsi(code uint8) NamedColor {
	switch code {
	case 30, 40:
		return NamedColorBlack
	case 31, 41:
		return NamedColorRed
	case 32, 42:
		return NamedColorGreen
	case 33, 43:
		return NamedColorYellow
	case 34, 44:
		return NamedColorBlue
	case 35, 45:
		return NamedColorMagenta
	case 36, 46:
		return NamedColorCyan
	case 37, 47:
		return NamedColorWhite
	case 90, 100:
		return NamedColorBrightBlack
	case 91, 101:
		return NamedColorBrightRed
	case 92, 102:
		return NamedColorBrightGreen
	case 93, 103:
		return NamedColorBrightYellow
	case 94, 104:
		return NamedColorBrightBlue
	case 95, 105:
		return NamedColorBrightMagenta
	case 96, 106:
		return NamedColorBrightCyan
	case 97, 107:
		return NamedColorBrightWhite
	default:
		return NamedColorWhite
	}
}

// ToAnsiFg converts to ANSI foreground color code
func (nc NamedColor) ToAnsiFg() uint8 {
	switch nc {
	case NamedColorBlack:
		return 30
	case NamedColorRed:
		return 31
	case NamedColorGreen:
		return 32
	case NamedColorYellow:
		return 33
	case NamedColorBlue:
		return 34
	case NamedColorMagenta:
		return 35
	case NamedColorCyan:
		return 36
	case NamedColorWhite:
		return 37
	case NamedColorBrightBlack:
		return 90
	case NamedColorBrightRed:
		return 91
	case NamedColorBrightGreen:
		return 92
	case NamedColorBrightYellow:
		return 93
	case NamedColorBrightBlue:
		return 94
	case NamedColorBrightMagenta:
		return 95
	case NamedColorBrightCyan:
		return 96
	case NamedColorBrightWhite:
		return 97
	default:
		return 37
	}
}

// ToAnsiBg converts to ANSI background color code
func (nc NamedColor) ToAnsiBg() uint8 {
	return nc.ToAnsiFg() + 10
}

// Helper functions

// runeWidth calculates the display width of a rune
func runeWidth(r rune) int {
	if r < 32 || r == 127 {
		return 0 // Control characters
	}
	if r < 127 {
		return 1 // ASCII
	}
	// Simplified width calculation - would use proper Unicode width library in production
	return 1
}

// ansiCodeEquals compares two AnsiCode pointers for equality
func ansiCodeEquals(a, b *AnsiCode) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.equals(*b)
}

// equals compares two AnsiCode structs for equality
func (ac AnsiCode) equals(other AnsiCode) bool {
	if ac.Type != other.Type {
		return false
	}
	switch ac.Type {
	case AnsiCodeTypeNamedColor:
		return ac.NamedColor == other.NamedColor
	case AnsiCodeTypeRgb:
		return ac.RGB == other.RGB
	case AnsiCodeTypeColorIndex:
		return ac.ColorIndex == other.ColorIndex
	default:
		return true // AnsiCodeTypeOn and AnsiCodeTypeReset
	}
}