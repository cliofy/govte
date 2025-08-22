package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// DCSHandler represents a handler that can process DCS sequences
type DCSHandler struct {
	NoopHandler
	dcsSequences []DCSSequence
}

// DCSSequence captures information about a DCS sequence
type DCSSequence struct {
	Params        [][]uint16
	Intermediates []byte
	Ignore        bool
	Action        rune
	Data          []byte
	Completed     bool
}

// Hook implements Handler for DCS sequences
func (h *DCSHandler) Hook(params [][]uint16, intermediates []byte, ignore bool, action rune) {
	seq := DCSSequence{
		Params:        make([][]uint16, len(params)),
		Intermediates: make([]byte, len(intermediates)),
		Ignore:        ignore,
		Action:        action,
		Data:          []byte{},
		Completed:     false,
	}
	
	// Deep copy params
	for i, param := range params {
		seq.Params[i] = make([]uint16, len(param))
		copy(seq.Params[i], param)
	}
	
	// Copy intermediates
	copy(seq.Intermediates, intermediates)
	
	h.dcsSequences = append(h.dcsSequences, seq)
}

// Put implements Handler for DCS data
func (h *DCSHandler) Put(data []byte) {
	if len(h.dcsSequences) > 0 {
		lastIdx := len(h.dcsSequences) - 1
		h.dcsSequences[lastIdx].Data = append(h.dcsSequences[lastIdx].Data, data...)
	}
}

// Unhook implements Handler for DCS completion
func (h *DCSHandler) Unhook() {
	if len(h.dcsSequences) > 0 {
		lastIdx := len(h.dcsSequences) - 1
		h.dcsSequences[lastIdx].Completed = true
	}
}

func TestDCSBasicSequence(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS sequence: ESC P 1 $ q m ESC \
	// This is a DECRQSS (Device Control Request Status String) sequence
	sequence := "\x1bP1$qm\x1b\\"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1, "Should capture one DCS sequence")
	
	dcs := handler.dcsSequences[0]
	assert.Equal(t, [][]uint16{{1}}, dcs.Params, "Should have parameter 1")
	assert.Equal(t, []byte("$"), dcs.Intermediates, "Should have $ intermediate")
	assert.Equal(t, 'q', dcs.Action, "Should have q action")
	assert.Equal(t, []byte("m"), dcs.Data, "Should have m data")
	assert.True(t, dcs.Completed, "Should be completed")
	assert.False(t, dcs.Ignore, "Should not be ignored")
}

func TestDCSMultipleParameters(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS with multiple parameters: ESC P 1;2;3 $ t data ESC \
	sequence := "\x1bP1;2;3$ttest_data\x1b\\"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	expected := [][]uint16{{1}, {2}, {3}}
	assert.Equal(t, expected, dcs.Params, "Should have three parameters")
	assert.Equal(t, []byte("$"), dcs.Intermediates)
	assert.Equal(t, 't', dcs.Action)
	assert.Equal(t, []byte("test_data"), dcs.Data)
	assert.True(t, dcs.Completed)
}

func TestDCSWithSubParameters(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS with subparameters: ESC P 1:2:3;4 $ s data ESC \
	sequence := "\x1bP1:2:3;4$ssubparam_data\x1b\\"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	expected := [][]uint16{{1, 2, 3}, {4}}
	assert.Equal(t, expected, dcs.Params, "Should handle subparameters correctly")
	assert.Equal(t, []byte("$"), dcs.Intermediates)
	assert.Equal(t, 's', dcs.Action)
	assert.Equal(t, []byte("subparam_data"), dcs.Data)
	assert.True(t, dcs.Completed)
}

func TestDCSIgnoreMode(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Create a DCS sequence with too many parameters to trigger ignore mode
	// This should test the parser's ability to handle parameter overflow
	longParams := "1;2;3;4;5;6;7;8;9;10;11;12;13;14;15;16;17;18;19;20;21;22;23;24;25;26;27;28;29;30;31;32;33;34;35"
	sequence := "\x1bP" + longParams + "$qm\x1b\\"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	// Should be in ignore mode due to parameter overflow
	assert.True(t, dcs.Ignore, "Should be in ignore mode due to parameter overflow")
	assert.Equal(t, 'q', dcs.Action)
	assert.Equal(t, []byte("m"), dcs.Data)
	assert.True(t, dcs.Completed)
}

func TestDCSBellTerminated(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS terminated with BEL instead of ST
	// ESC P 1 $ q m BEL
	sequence := "\x1bP1$qm\x07"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	assert.Equal(t, [][]uint16{{1}}, dcs.Params)
	assert.Equal(t, []byte("$"), dcs.Intermediates)
	assert.Equal(t, 'q', dcs.Action)
	assert.Equal(t, []byte("m"), dcs.Data)
	assert.True(t, dcs.Completed, "Should be completed with BEL termination")
}

func TestDCSStreamingData(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS sequence processed in chunks
	chunks := []string{
		"\x1bP",     // Start DCS
		"1$",        // Parameters and intermediate
		"q",         // Action (triggers Hook)
		"test",      // Data chunk 1
		"_data",     // Data chunk 2
		"\x1b\\",    // ST termination
	}
	
	for _, chunk := range chunks {
		processor.Advance(handler, []byte(chunk))
	}
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	assert.Equal(t, [][]uint16{{1}}, dcs.Params)
	assert.Equal(t, []byte("$"), dcs.Intermediates)
	assert.Equal(t, 'q', dcs.Action)
	assert.Equal(t, []byte("test_data"), dcs.Data, "Should accumulate data from multiple chunks")
	assert.True(t, dcs.Completed)
}

func TestDCSEmptySequence(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test minimal DCS: ESC P p ESC \
	sequence := "\x1bPp\x1b\\"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	assert.Empty(t, dcs.Params, "Should have no parameters")
	assert.Empty(t, dcs.Intermediates, "Should have no intermediates") 
	assert.Equal(t, 'p', dcs.Action)
	assert.Empty(t, dcs.Data, "Should have no data")
	assert.True(t, dcs.Completed)
}

func TestDCSWithControlCharacters(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS data containing control characters
	sequence := "\x1bP1$qm\x00\x01\x1f\x7f\x1b\\"
	
	processor.Advance(handler, []byte(sequence))
	
	assert.Len(t, handler.dcsSequences, 1)
	
	dcs := handler.dcsSequences[0]
	expected := []byte{'m', 0x00, 0x01, 0x1f, 0x7f}
	assert.Equal(t, expected, dcs.Data, "Should handle control characters in data")
	assert.True(t, dcs.Completed)
}

// Test DCS sequence cancellation
func TestDCSCancellation(t *testing.T) {
	processor := NewProcessor(&NoopHandler{})
	handler := &DCSHandler{}
	
	// Test DCS cancelled by CAN (0x18)
	sequence := "\x1bP1$qmdata\x18"
	
	processor.Advance(handler, []byte(sequence))
	
	// Should have received and processed the DCS sequence until cancellation
	assert.Len(t, handler.dcsSequences, 1, "Should have received one DCS sequence")
	
	dcs := handler.dcsSequences[0]
	// Verify the sequence was properly started
	assert.Equal(t, [][]uint16{{1}}, dcs.Params, "Should have parameter 1")
	assert.Equal(t, []byte("$"), dcs.Intermediates, "Should have $ intermediate")
	assert.Equal(t, 'q', dcs.Action, "Should have q action")
	
	// Verify data was received up to cancellation point (CAN should not be included)
	assert.Equal(t, "mdata", string(dcs.Data), "Should have received data before cancellation")
	
	// Unhook should have been called to clean up the sequence
	assert.True(t, dcs.Completed, "Sequence should be marked as ended (Unhook called)")
}