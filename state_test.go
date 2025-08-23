package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateConstants(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected string
	}{
		{"Ground state", StateGround, "Ground"},
		{"Escape state", StateEscape, "Escape"},
		{"CSI Entry state", StateCSIEntry, "CSIEntry"},
		{"CSI Param state", StateCSIParam, "CSIParam"},
		{"CSI Intermediate state", StateCSIIntermediate, "CSIIntermediate"},
		{"CSI Ignore state", StateCSIIgnore, "CSIIgnore"},
		{"OSC String state", StateOSCString, "OSCString"},
		{"DCS Entry state", StateDCSEntry, "DCSEntry"},
		{"DCS Param state", StateDCSParam, "DCSParam"},
		{"DCS Intermediate state", StateDCSIntermediate, "DCSIntermediate"},
		{"DCS Passthrough state", StateDCSPassthrough, "DCSPassthrough"},
		{"DCS Ignore state", StateDCSIgnore, "DCSIgnore"},
		{"SOS PM APC String state", StateSOSPMApcString, "SOSPMApcString"},
		{"Escape Intermediate state", StateEscapeIntermediate, "EscapeIntermediate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestStateDefaultValue(t *testing.T) {
	var s State
	assert.Equal(t, StateGround, s, "Default state should be Ground")
}

func TestStateValidation(t *testing.T) {
	// 测试所有状态值都是有效的
	states := []State{
		StateGround,
		StateEscape,
		StateEscapeIntermediate,
		StateCSIEntry,
		StateCSIParam,
		StateCSIIntermediate,
		StateCSIIgnore,
		StateOSCString,
		StateDCSEntry,
		StateDCSParam,
		StateDCSIntermediate,
		StateDCSPassthrough,
		StateDCSIgnore,
		StateSOSPMApcString,
	}

	for _, state := range states {
		assert.True(t, state.IsValid(), "State %v should be valid", state)
	}

	// 测试无效状态
	invalidState := State(99)
	assert.False(t, invalidState.IsValid(), "State 99 should be invalid")
}

func TestStateTransitions(t *testing.T) {
	// 测试状态转换的基本规则
	tests := []struct {
		name        string
		from        State
		input       byte
		expectedTo  State
		shouldAllow bool
	}{
		{
			name:        "ESC from Ground",
			from:        StateGround,
			input:       0x1B, // ESC
			expectedTo:  StateEscape,
			shouldAllow: true,
		},
		{
			name:        "CSI from Escape",
			from:        StateEscape,
			input:       '[',
			expectedTo:  StateCSIEntry,
			shouldAllow: true,
		},
		{
			name:        "OSC from Escape",
			from:        StateEscape,
			input:       ']',
			expectedTo:  StateOSCString,
			shouldAllow: true,
		},
		{
			name:        "DCS from Escape",
			from:        StateEscape,
			input:       'P',
			expectedTo:  StateDCSEntry,
			shouldAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextState := tt.from.Transition(tt.input)
			if tt.shouldAllow {
				assert.Equal(t, tt.expectedTo, nextState)
			}
		})
	}
}
