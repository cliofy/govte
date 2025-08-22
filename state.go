package govte

import "fmt"

// State represents the current state of the VTE parser state machine
type State uint8

// State constants matching the VTE state machine
const (
	StateGround State = iota
	StateEscape
	StateEscapeIntermediate
	StateCSIEntry
	StateCSIParam
	StateCSIIntermediate
	StateCSIIgnore
	StateOSCString
	StateDCSEntry
	StateDCSParam
	StateDCSIntermediate
	StateDCSPassthrough
	StateDCSIgnore
	StateSOSPMApcString
)

// String returns the string representation of the state
func (s State) String() string {
	names := []string{
		"Ground",
		"Escape",
		"EscapeIntermediate",
		"CSIEntry",
		"CSIParam",
		"CSIIntermediate",
		"CSIIgnore",
		"OSCString",
		"DCSEntry",
		"DCSParam",
		"DCSIntermediate",
		"DCSPassthrough",
		"DCSIgnore",
		"SOSPMApcString",
	}

	if int(s) < len(names) {
		return names[s]
	}
	return fmt.Sprintf("Unknown(%d)", s)
}

// IsValid checks if the state is a valid state
func (s State) IsValid() bool {
	return s <= StateSOSPMApcString
}

// Transition determines the next state based on input byte
// This is a simplified version - full implementation will be in parser.go
func (s State) Transition(b byte) State {
	switch s {
	case StateGround:
		switch b {
		case 0x1B: // ESC
			return StateEscape
		}
	case StateEscape:
		switch b {
		case '[':
			return StateCSIEntry
		case ']':
			return StateOSCString
		case 'P':
			return StateDCSEntry
		case '_':
			return StateSOSPMApcString
		case 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F:
			// Intermediate characters
			return StateEscapeIntermediate
		}
		// For most other characters, return to ground
		return StateGround
	}
	
	// Default: stay in current state
	return s
}