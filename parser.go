package govte

import (
	"unicode/utf8"
)

const (
	// MaxIntermediates is the maximum number of intermediate bytes
	MaxIntermediates = 2
	// MaxOSCRaw is the maximum size of OSC string
	MaxOSCRaw = 1024
	// MaxOSCParams is the maximum number of OSC parameters
	MaxOSCParams = 16
)

// Parser is the VTE parser state machine
type Parser struct {
	state            State
	intermediates    []byte
	intermediateIdx  int
	params           *Params
	currentParam     uint16  // Current parameter being built
	hasCurrentParam  bool    // Whether we have a current parameter
	inSubparam       bool    // Whether we're in a subparameter group
	oscRaw           []byte
	oscParams        []int // Indices into oscRaw for parameter boundaries
	oscNumParams     int
	ignoring         bool
	pendingESC       bool    // For DCS passthrough ESC tracking
	partialUTF8      [4]byte
	partialUTF8Len   int
}

// NewParser creates a new VTE parser
func NewParser() *Parser {
	return &Parser{
		state:         StateGround,
		params:        NewParams(),
		intermediates: make([]byte, 0, MaxIntermediates),
		oscRaw:        make([]byte, 0, MaxOSCRaw),
		oscParams:     make([]int, 0, MaxOSCParams*2), // start,end pairs
	}
}

// State returns the current parser state
func (p *Parser) State() State {
	return p.state
}

// Advance processes input bytes through the state machine
func (p *Parser) Advance(performer Performer, bytes []byte) {
	i := 0
	
	// Handle partial UTF-8 from previous call
	if p.partialUTF8Len > 0 {
		consumed := p.advancePartialUTF8(performer, bytes)
		i += consumed
		// If we consumed some bytes, we might still be in Ground state
		// and need to continue processing remaining bytes
		if i >= len(bytes) {
			return
		}
	}
	
	for i < len(bytes) {
		switch p.state {
		case StateGround:
			i += p.advanceGround(performer, bytes[i:])
		case StateEscape:
			p.advanceEscape(performer, bytes[i])
			i++
		case StateEscapeIntermediate:
			p.advanceEscapeIntermediate(performer, bytes[i])
			i++
		case StateCSIEntry:
			p.advanceCSIEntry(performer, bytes[i])
			i++
		case StateCSIParam:
			p.advanceCSIParam(performer, bytes[i])
			i++
		case StateCSIIntermediate:
			p.advanceCSIIntermediate(performer, bytes[i])
			i++
		case StateCSIIgnore:
			p.advanceCSIIgnore(performer, bytes[i])
			i++
		case StateOSCString:
			p.advanceOSCString(performer, bytes[i])
			i++
		case StateDCSEntry:
			p.advanceDCSEntry(performer, bytes[i])
			i++
		case StateDCSParam:
			p.advanceDCSParam(performer, bytes[i])
			i++
		case StateDCSIntermediate:
			p.advanceDCSIntermediate(performer, bytes[i])
			i++
		case StateDCSPassthrough:
			p.advanceDCSPassthrough(performer, bytes[i])
			i++
		case StateDCSIgnore:
			p.advanceDCSIgnore(performer, bytes[i])
			i++
		case StateSOSPMApcString:
			p.advanceSOSPMApcString(performer, bytes[i])
			i++
		default:
			i++
		}
	}
}

// advanceGround handles the ground state
func (p *Parser) advanceGround(performer Performer, bytes []byte) int {
	for i, b := range bytes {
		switch {
		case b == 0x1B: // ESC
			p.state = StateEscape
			p.resetParams()
			return i + 1
		case b < 0x20: // C0 control
			performer.Execute(b)
		case b >= 0x20 && b < 0x7F: // Printable ASCII
			performer.Print(rune(b))
		case b >= 0x80: // UTF-8 or C1 control
			if b >= 0xC0 {
				// Start of UTF-8 sequence
				return i + p.handleUTF8(performer, bytes[i:])
			} else if b == 0x90 {
				// DCS
				p.state = StateDCSEntry
				p.resetParams()
				return i + 1
			} else if b == 0x9B {
				// CSI
				p.state = StateCSIEntry
				p.resetParams()
				return i + 1
			} else if b == 0x9D {
				// OSC
				p.state = StateOSCString
				p.resetParams()
				return i + 1
			} else {
				// Invalid UTF-8 continuation byte without start - print replacement character
				performer.Print(utf8.RuneError)
			}
		case b == 0x7F: // DEL - ignore
			// Do nothing
		}
	}
	return len(bytes)
}

// advanceEscape handles the escape state
func (p *Parser) advanceEscape(performer Performer, b byte) {
	switch {
	case b < 0x20: // Execute C0
		performer.Execute(b)
	case b >= 0x20 && b <= 0x2F: // Collect intermediate
		p.collectIntermediate(b)
		p.state = StateEscapeIntermediate
	case b >= 0x30 && b <= 0x4F: // ESC dispatch
		performer.EscDispatch(p.intermediates, p.ignoring, b)
		p.state = StateGround
	case b == 0x5B: // [
		p.state = StateCSIEntry
	case b == 0x5D: // ]
		p.state = StateOSCString
	case b == 0x50: // P
		p.state = StateDCSEntry
	case b == 0x58 || b == 0x5E || b == 0x5F: // X, ^, _
		p.state = StateSOSPMApcString
	case b >= 0x51 && b <= 0x57 || b >= 0x59 && b <= 0x5A || b == 0x5C || b >= 0x60 && b <= 0x7E:
		// ESC dispatch
		performer.EscDispatch(p.intermediates, p.ignoring, b)
		p.state = StateGround
	case b == 0x7F: // Ignore
		// Do nothing
	}
}

// advanceEscapeIntermediate handles escape intermediate state
func (p *Parser) advanceEscapeIntermediate(performer Performer, b byte) {
	switch {
	case b < 0x20:
		performer.Execute(b)
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
	case b >= 0x30 && b <= 0x7E:
		performer.EscDispatch(p.intermediates, p.ignoring, b)
		p.state = StateGround
	case b == 0x7F:
		// Ignore
	}
}

// advanceCSIEntry handles CSI entry state
func (p *Parser) advanceCSIEntry(performer Performer, b byte) {
	switch {
	case b < 0x20:
		performer.Execute(b)
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
		p.state = StateCSIIntermediate
	case b >= 0x30 && b <= 0x39:
		p.paramDigit(b)
		p.state = StateCSIParam
	case b == 0x3A:
		p.paramSubparam()
		p.state = StateCSIParam
	case b == 0x3B:
		p.paramSeparator()
		p.state = StateCSIParam
	case b >= 0x3C && b <= 0x3F:
		p.collectIntermediate(b)
		p.state = StateCSIParam
	case b >= 0x40 && b <= 0x7E:
		p.csiDispatch(performer, b)
		p.state = StateGround
	case b == 0x7F:
		// Ignore
	}
}

// advanceCSIParam handles CSI parameter state
func (p *Parser) advanceCSIParam(performer Performer, b byte) {
	switch {
	case b < 0x20:
		performer.Execute(b)
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
		p.state = StateCSIIntermediate
	case b >= 0x30 && b <= 0x39:
		p.paramDigit(b)
	case b == 0x3A:
		p.paramSubparam()
	case b == 0x3B:
		p.paramSeparator()
	case b >= 0x3C && b <= 0x3F:
		p.state = StateCSIIgnore
	case b >= 0x40 && b <= 0x7E:
		p.csiDispatch(performer, b)
		p.state = StateGround
	case b == 0x7F:
		// Ignore
	}
}

// advanceCSIIntermediate handles CSI intermediate state
func (p *Parser) advanceCSIIntermediate(performer Performer, b byte) {
	switch {
	case b < 0x20:
		performer.Execute(b)
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
	case b >= 0x30 && b <= 0x3F:
		p.state = StateCSIIgnore
	case b >= 0x40 && b <= 0x7E:
		p.csiDispatch(performer, b)
		p.state = StateGround
	case b == 0x7F:
		// Ignore
	}
}

// advanceCSIIgnore handles CSI ignore state
func (p *Parser) advanceCSIIgnore(performer Performer, b byte) {
	switch {
	case b < 0x20:
		performer.Execute(b)
	case b >= 0x20 && b <= 0x3F:
		// Ignore
	case b >= 0x40 && b <= 0x7E:
		p.state = StateGround
	case b == 0x7F:
		// Ignore
	}
}

// advanceOSCString handles OSC string state
func (p *Parser) advanceOSCString(performer Performer, b byte) {
	switch {
	case b == 0x07: // BEL terminates
		p.oscDispatch(performer, true)
		p.state = StateGround
	case b == 0x1B: // ESC might be ST
		// Need to peek next byte for '\'
		p.oscPut(b)
	case b == '\\' && len(p.oscRaw) > 0 && p.oscRaw[len(p.oscRaw)-1] == 0x1B:
		// ESC \ (ST) terminates
		p.oscRaw = p.oscRaw[:len(p.oscRaw)-1] // Remove ESC
		p.oscDispatch(performer, false)
		p.state = StateGround
	case b >= 0x20 && b < 0x7F:
		p.oscPut(b)
	case b < 0x20 || b >= 0x80:
		// Invalid in OSC, but we'll collect it
		p.oscPut(b)
	}
}

// advanceDCSEntry handles DCS entry state
func (p *Parser) advanceDCSEntry(performer Performer, b byte) {
	switch {
	case b < 0x20:
		// Ignore
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
		p.state = StateDCSIntermediate
	case b >= 0x30 && b <= 0x39:
		p.paramDigit(b)
		p.state = StateDCSParam
	case b == 0x3A:
		p.paramSubparam()
		p.state = StateDCSParam
	case b == 0x3B:
		p.paramSeparator()
		p.state = StateDCSParam
	case b >= 0x3C && b <= 0x3F:
		p.collectIntermediate(b)
		p.state = StateDCSParam
	case b >= 0x40 && b <= 0x7E:
		// Finalize current parameter before Hook
		if p.hasCurrentParam {
			if p.inSubparam {
				if p.params.IsFull() {
					p.ignoring = true
				} else {
					p.params.Extend(p.currentParam)
				}
			} else {
				if p.params.IsFull() {
					p.ignoring = true
				} else {
					p.params.Push(p.currentParam)
				}
			}
		}
		performer.Hook(p.params, p.intermediates, p.ignoring, rune(b))
		p.state = StateDCSPassthrough
	case b == 0x7F:
		// Ignore
	}
}

// advanceDCSParam handles DCS parameter state
func (p *Parser) advanceDCSParam(performer Performer, b byte) {
	switch {
	case b < 0x20:
		// Ignore
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
		p.state = StateDCSIntermediate
	case b >= 0x30 && b <= 0x39:
		p.paramDigit(b)
	case b == 0x3A:
		p.paramSubparam()
	case b == 0x3B:
		p.paramSeparator()
	case b >= 0x3C && b <= 0x3F:
		p.state = StateDCSIgnore
	case b >= 0x40 && b <= 0x7E:
		// Finalize current parameter before Hook
		if p.hasCurrentParam {
			if p.inSubparam {
				if p.params.IsFull() {
					p.ignoring = true
				} else {
					p.params.Extend(p.currentParam)
				}
			} else {
				if p.params.IsFull() {
					p.ignoring = true
				} else {
					p.params.Push(p.currentParam)
				}
			}
		}
		performer.Hook(p.params, p.intermediates, p.ignoring, rune(b))
		p.state = StateDCSPassthrough
	case b == 0x7F:
		// Ignore
	}
}

// advanceDCSIntermediate handles DCS intermediate state
func (p *Parser) advanceDCSIntermediate(performer Performer, b byte) {
	switch {
	case b < 0x20:
		// Ignore
	case b >= 0x20 && b <= 0x2F:
		p.collectIntermediate(b)
	case b >= 0x30 && b <= 0x3F:
		p.state = StateDCSIgnore
	case b >= 0x40 && b <= 0x7E:
		// Finalize current parameter before Hook
		if p.hasCurrentParam {
			if p.inSubparam {
				if p.params.IsFull() {
					p.ignoring = true
				} else {
					p.params.Extend(p.currentParam)
				}
			} else {
				if p.params.IsFull() {
					p.ignoring = true
				} else {
					p.params.Push(p.currentParam)
				}
			}
		}
		performer.Hook(p.params, p.intermediates, p.ignoring, rune(b))
		p.state = StateDCSPassthrough
	case b == 0x7F:
		// Ignore
	}
}

// advanceDCSPassthrough handles DCS passthrough state
func (p *Parser) advanceDCSPassthrough(performer Performer, b byte) {
	switch {
	case b == 0x1B:
		// Might be ST, don't put ESC yet, wait for next byte
		p.pendingESC = true
		return
	case b == '\\' && p.pendingESC:
		// This is ST (ESC \)
		p.pendingESC = false
		performer.Unhook()
		p.state = StateGround
	case b == 0x07:
		// BEL terminates DCS
		performer.Unhook()
		p.state = StateGround
	case b >= 0x00 && b <= 0x06 || b >= 0x08 && b <= 0x17 || b == 0x19 || b >= 0x1C && b <= 0x7E:
		// If we had a pending ESC that wasn't part of ST, put it first
		if p.pendingESC {
			performer.Put(0x1B)
			p.pendingESC = false
		}
		performer.Put(b)
	case b == 0x18 || b == 0x1A:
		// CAN/SUB cancels DCS - call Unhook to allow handler cleanup, then Execute
		performer.Unhook()
		performer.Execute(b)
		p.state = StateGround
	case b == 0x7F:
		// Include DEL in data
		if p.pendingESC {
			performer.Put(0x1B)
			p.pendingESC = false
		}
		performer.Put(b)
	default:
		// For other bytes after ESC
		if p.pendingESC {
			performer.Put(0x1B)
			p.pendingESC = false
		}
		performer.Put(b)
	}
}

// advanceDCSIgnore handles DCS ignore state
func (p *Parser) advanceDCSIgnore(performer Performer, b byte) {
	switch {
	case b == 0x1B:
		// Might be ST
	case b == 0x18 || b == 0x1A:
		p.state = StateGround
	}
}

// advanceSOSPMApcString handles SOS/PM/APC string state
func (p *Parser) advanceSOSPMApcString(performer Performer, b byte) {
	// Simply ignore until ST
	if b == 0x1B {
		// Might be ST
	} else if b == '\\' {
		// If previous was ESC, this is ST
		p.state = StateGround
	}
}

// Helper methods

func (p *Parser) resetParams() {
	p.params.Clear()
	p.intermediates = p.intermediates[:0]
	p.intermediateIdx = 0
	p.ignoring = false
	p.oscRaw = p.oscRaw[:0]
	p.oscParams = p.oscParams[:0]
	p.oscNumParams = 0
	p.currentParam = 0
	p.hasCurrentParam = false
	p.inSubparam = false
}

func (p *Parser) collectIntermediate(b byte) {
	if len(p.intermediates) < MaxIntermediates {
		p.intermediates = append(p.intermediates, b)
	} else {
		p.ignoring = true
	}
}

func (p *Parser) paramDigit(b byte) {
	digit := uint16(b - '0')
	
	if !p.hasCurrentParam {
		// Start new parameter
		p.currentParam = digit
		p.hasCurrentParam = true
	} else {
		// Accumulate digits
		p.currentParam = p.currentParam*10 + digit
		if p.currentParam > 9999 {
			p.currentParam = 9999 // Cap at reasonable maximum
		}
	}
}

func (p *Parser) paramSeparator() {
	if p.hasCurrentParam {
		if p.inSubparam {
			// We're in a subparameter group, add the current value as a subparam
			if p.params.IsFull() {
				p.ignoring = true
			} else {
				p.params.Extend(p.currentParam)
			}
		} else {
			// Normal parameter
			if p.params.IsFull() {
				p.ignoring = true
			} else {
				p.params.Push(p.currentParam)
			}
		}
	} else if !p.inSubparam {
		// Empty parameter (e.g., ";;")
		if p.params.IsFull() {
			p.ignoring = true
		} else {
			p.params.Push(0)
		}
	}
	
	// Reset for next parameter group
	p.currentParam = 0
	p.hasCurrentParam = false
	p.inSubparam = false
}

func (p *Parser) paramSubparam() {
	if p.hasCurrentParam {
		if !p.inSubparam {
			// First colon - current value is the main parameter
			if p.params.IsFull() {
				p.ignoring = true
			} else {
				p.params.Push(p.currentParam)
				p.inSubparam = true
			}
		} else {
			// Subsequent colon - current value is a subparameter
			if p.params.IsFull() {
				p.ignoring = true
			} else {
				p.params.Extend(p.currentParam)
			}
		}
		p.currentParam = 0
		p.hasCurrentParam = false
	} else {
		// No current param means we have an empty position
		if !p.inSubparam {
			// Empty main parameter before colon (e.g., ":5")
			if p.params.IsFull() {
				p.ignoring = true
			} else {
				p.params.Push(0)
				p.inSubparam = true
			}
		} else {
			// Empty subparameter (e.g., the empty part in "38::128")
			if p.params.IsFull() {
				p.ignoring = true
			} else {
				p.params.Extend(0)
			}
		}
	}
}

func (p *Parser) csiDispatch(performer Performer, action byte) {
	// Finalize any pending parameter
	if p.hasCurrentParam {
		if p.inSubparam {
			// Last parameter is a subparameter
			p.params.Extend(p.currentParam)
		} else {
			// Last parameter is a regular parameter
			p.params.Push(p.currentParam)
		}
	}
	
	performer.CsiDispatch(p.params, p.intermediates, p.ignoring, rune(action))
	p.resetParams()
}

func (p *Parser) oscPut(b byte) {
	if len(p.oscRaw) < MaxOSCRaw {
		if b == ';' && p.oscNumParams < MaxOSCParams {
			// Mark parameter boundary
			p.oscParams = append(p.oscParams, len(p.oscRaw))
			p.oscNumParams++
		} else {
			p.oscRaw = append(p.oscRaw, b)
		}
	}
}

func (p *Parser) oscDispatch(performer Performer, bellTerminated bool) {
	// Parse OSC parameters
	params := make([][]byte, 0, p.oscNumParams+1)
	start := 0
	
	for _, end := range p.oscParams {
		if end > start && end <= len(p.oscRaw) {
			params = append(params, p.oscRaw[start:end])
			start = end
		}
	}
	
	// Add final parameter
	if start < len(p.oscRaw) {
		params = append(params, p.oscRaw[start:])
	}
	
	performer.OscDispatch(params, bellTerminated)
	p.resetParams()
}

// handleUTF8 processes UTF-8 encoded characters
func (p *Parser) handleUTF8(performer Performer, bytes []byte) int {
	if len(bytes) == 0 {
		return 0
	}
	
	r, size := utf8.DecodeRune(bytes)
	if r == utf8.RuneError {
		// Incomplete UTF-8, save for next call
		if size == 1 && !utf8.FullRune(bytes) {
			// Partial UTF-8 sequence - save all available bytes
			n := copy(p.partialUTF8[:], bytes)
			p.partialUTF8Len = n
			return len(bytes)
		}
		// Invalid UTF-8, print replacement character and skip
		performer.Print(utf8.RuneError)
		return 1
	}
	
	performer.Print(r)
	return size
}

// advancePartialUTF8 handles partial UTF-8 from previous call
func (p *Parser) advancePartialUTF8(performer Performer, bytes []byte) int {
	if len(bytes) == 0 {
		return 0
	}
	
	// Check if the first byte is a control character that should interrupt UTF-8
	if bytes[0] < 0x20 || bytes[0] == 0x7F || bytes[0] == 0x1B {
		// Control character interrupts partial UTF-8
		// Print replacement character for the incomplete UTF-8
		performer.Print(utf8.RuneError)
		p.partialUTF8Len = 0
		return 0 // Don't consume the control character
	}
	
	// Try to complete the partial UTF-8
	needed := utf8.UTFMax - p.partialUTF8Len
	n := min(needed, len(bytes))
	copy(p.partialUTF8[p.partialUTF8Len:], bytes[:n])
	
	r, size := utf8.DecodeRune(p.partialUTF8[:p.partialUTF8Len+n])
	if r != utf8.RuneError {
		// Successfully decoded a character
		performer.Print(r)
		// Calculate how many bytes from the input we used
		bytesFromInput := size - p.partialUTF8Len
		p.partialUTF8Len = 0
		return bytesFromInput
	}
	
	if size == 1 && !utf8.FullRune(p.partialUTF8[:p.partialUTF8Len+n]) {
		// Still incomplete
		p.partialUTF8Len += n
		return n
	}
	
	// Invalid UTF-8, print replacement character and reset
	performer.Print(utf8.RuneError)
	p.partialUTF8Len = 0
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}