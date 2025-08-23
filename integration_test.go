package govte

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// IntegrationPerformer captures all actions for verification
type IntegrationPerformer struct {
	Actions []IntegrationAction
}

type IntegrationAction struct {
	Type string
	Data interface{}
}

func (p *IntegrationPerformer) Print(c rune) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "print",
		Data: c,
	})
}

func (p *IntegrationPerformer) Execute(b byte) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "execute",
		Data: b,
	})
}

func (p *IntegrationPerformer) Hook(params *Params, intermediates []byte, ignore bool, action rune) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "hook",
		Data: map[string]interface{}{
			"params":        params,
			"intermediates": intermediates,
			"ignore":        ignore,
			"action":        action,
		},
	})
}

func (p *IntegrationPerformer) Put(b byte) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "put",
		Data: b,
	})
}

func (p *IntegrationPerformer) Unhook() {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "unhook",
		Data: nil,
	})
}

func (p *IntegrationPerformer) OscDispatch(params [][]byte, bellTerminated bool) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "osc",
		Data: map[string]interface{}{
			"params":         params,
			"bellTerminated": bellTerminated,
		},
	})
}

func (p *IntegrationPerformer) CsiDispatch(params *Params, intermediates []byte, ignore bool, action rune) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "csi",
		Data: map[string]interface{}{
			"params":        params,
			"intermediates": intermediates,
			"ignore":        ignore,
			"action":        action,
		},
	})
}

func (p *IntegrationPerformer) EscDispatch(intermediates []byte, ignore bool, b byte) {
	p.Actions = append(p.Actions, IntegrationAction{
		Type: "esc",
		Data: map[string]interface{}{
			"intermediates": intermediates,
			"ignore":        ignore,
			"byte":          b,
		},
	})
}

// TestIntegrationCompleteTerminalSequence tests a complete terminal interaction
func TestIntegrationCompleteTerminalSequence(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Simulate a terminal session with various sequences
	sequences := [][]byte{
		[]byte("Hello, "),                   // Plain text
		[]byte("\x1b[31m"),                  // Set foreground red
		[]byte("World"),                     // More text
		[]byte("\x1b[0m"),                   // Reset
		[]byte("\r\n"),                      // Carriage return + line feed
		[]byte("\x1b[2J"),                   // Clear screen
		[]byte("\x1b[H"),                    // Cursor home
		[]byte("\x1b]0;Terminal Title\x07"), // Set window title
		[]byte("\x1b[?25l"),                 // Hide cursor
		[]byte("Line 1"),                    // Text
		[]byte("\x1b[2;1H"),                 // Move cursor to row 2, col 1
		[]byte("Line 2"),                    // Text
		[]byte("\x1b[?25h"),                 // Show cursor
		[]byte("\x1bP1$qm\x1b\\"),           // DECRQSS (DCS sequence)
		[]byte("\x1b[38:2:255:128:64m"),     // RGB color with subparameters
		[]byte("Colored"),                   // Text
		[]byte("\x1b[m"),                    // Reset
	}

	for _, seq := range sequences {
		parser.Advance(performer, seq)
	}

	// Verify we got the expected mix of actions
	printCount := 0
	csiCount := 0
	oscCount := 0
	dcsCount := 0
	executeCount := 0

	for _, action := range performer.Actions {
		switch action.Type {
		case "print":
			printCount++
		case "csi":
			csiCount++
		case "osc":
			oscCount++
		case "hook":
			dcsCount++
		case "unhook":
			dcsCount++
		case "execute":
			executeCount++
		}
	}

	assert.Greater(t, printCount, 20)     // "Hello, ", "World", "Line 1", "Line 2", "Colored"
	assert.GreaterOrEqual(t, csiCount, 8) // Multiple CSI sequences
	assert.Equal(t, 1, oscCount)          // One OSC sequence
	assert.Equal(t, 2, dcsCount)          // One hook + one unhook
	assert.Equal(t, 2, executeCount)      // CR and LF
}

// TestIntegrationColoredOutput tests handling of colored terminal output
func TestIntegrationColoredOutput(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Simulate colored output like from ls --color
	input := []byte(
		"\x1b[0m\x1b[01;34mdir1\x1b[0m  " +
			"\x1b[01;32mexecutable\x1b[0m  " +
			"normal.txt  " +
			"\x1b[01;35mlink\x1b[0m\n")

	parser.Advance(performer, input)

	// Extract text and verify it matches expected output
	var text bytes.Buffer
	for _, action := range performer.Actions {
		if action.Type == "print" {
			text.WriteRune(action.Data.(rune))
		} else if action.Type == "execute" && action.Data.(byte) == '\n' {
			text.WriteByte('\n')
		}
	}

	expected := "dir1  executable  normal.txt  link\n"
	assert.Equal(t, expected, text.String())
}

// TestIntegrationProgressBar tests handling of progress bar animations
func TestIntegrationProgressBar(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Simulate a progress bar that updates in place
	frames := []string{
		"\r[          ] 0%",
		"\r[==        ] 20%",
		"\r[====      ] 40%",
		"\r[======    ] 60%",
		"\r[========  ] 80%",
		"\r[==========] 100%",
	}

	for _, frame := range frames {
		parser.Advance(performer, []byte(frame))
	}

	// Count carriage returns
	crCount := 0
	for _, action := range performer.Actions {
		if action.Type == "execute" && action.Data.(byte) == '\r' {
			crCount++
		}
	}

	assert.Equal(t, 6, crCount, "Should have 6 carriage returns for progress updates")
}

// TestIntegrationUTF8AndEmoji tests UTF-8 and emoji handling
func TestIntegrationUTF8AndEmoji(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Mix of different scripts and emojis
	input := []byte("Hello 你好 مرحبا שלום こんにちは 🌍🎉🚀")
	parser.Advance(performer, input)

	// Extract printed characters
	var printed []rune
	for _, action := range performer.Actions {
		if action.Type == "print" {
			printed = append(printed, action.Data.(rune))
		}
	}

	expected := []rune("Hello 你好 مرحبا שלום こんにちは 🌍🎉🚀")
	assert.Equal(t, expected, printed)
}

// TestIntegrationHyperlinks tests OSC 8 hyperlink sequences
func TestIntegrationHyperlinks(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// OSC 8 hyperlink format: ESC]8;id=ID;URI ST LINK_TEXT ESC]8;; ST
	link := "\x1b]8;id=link1;https://example.com\x07Click here\x1b]8;;\x07"
	parser.Advance(performer, []byte(link))

	// Verify OSC sequences were dispatched
	oscCount := 0
	for _, action := range performer.Actions {
		if action.Type == "osc" {
			oscCount++
		}
	}
	assert.Equal(t, 2, oscCount, "Should have 2 OSC sequences for hyperlink")

	// Verify link text was printed
	var text bytes.Buffer
	for _, action := range performer.Actions {
		if action.Type == "print" {
			text.WriteRune(action.Data.(rune))
		}
	}
	assert.Equal(t, "Click here", text.String())
}

// TestIntegrationComplexSGR tests complex SGR (Select Graphic Rendition) sequences
func TestIntegrationComplexSGR(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Complex SGR with multiple attributes
	sequences := []string{
		"\x1b[0m",                       // Reset
		"\x1b[1;4;31m",                  // Bold, underline, red
		"\x1b[38;5;128m",                // 256-color foreground
		"\x1b[38:2:100:150:200m",        // RGB foreground with subparameters
		"\x1b[48;2;50;75;100m",          // RGB background with semicolons
		"\x1b[1;3;4;5;7;9m",             // Multiple attributes
		"\x1b[21;22;23;24;25;27;28;29m", // Reset various attributes
	}

	for _, seq := range sequences {
		parser.Advance(performer, []byte(seq))
	}

	// Verify all CSI sequences were processed
	csiCount := 0
	for _, action := range performer.Actions {
		if action.Type == "csi" {
			csiCount++
			data := action.Data.(map[string]interface{})
			assert.Equal(t, 'm', data["action"], "All should be SGR sequences")
		}
	}
	assert.Equal(t, len(sequences), csiCount)
}

// TestIntegrationTerminalModes tests various terminal mode changes
func TestIntegrationTerminalModes(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	modes := []struct {
		sequence string
		desc     string
	}{
		{"\x1b[?25h", "Show cursor"},
		{"\x1b[?25l", "Hide cursor"},
		{"\x1b[?1049h", "Alternate screen buffer"},
		{"\x1b[?1049l", "Normal screen buffer"},
		{"\x1b[?2004h", "Enable bracketed paste"},
		{"\x1b[?2004l", "Disable bracketed paste"},
		{"\x1b[?1h", "Application cursor keys"},
		{"\x1b[?1l", "Normal cursor keys"},
	}

	for _, mode := range modes {
		parser.Advance(performer, []byte(mode.sequence))
	}

	// Verify all mode changes were processed
	csiCount := 0
	for _, action := range performer.Actions {
		if action.Type == "csi" {
			csiCount++
			data := action.Data.(map[string]interface{})
			actionChar := data["action"].(rune)
			assert.True(t, actionChar == 'h' || actionChar == 'l',
				"Mode changes should use 'h' (set) or 'l' (reset)")
		}
	}
	assert.Equal(t, len(modes), csiCount)
}

// TestIntegrationRealWorldShellOutput tests parsing real shell command output
func TestIntegrationRealWorldShellOutput(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Simulate output from a shell session
	shellOutput := []byte(
		"\x1b[?2004h" + // Enable bracketed paste
			"$ \x1b[32mls\x1b[0m -la\r\n" +
			"total 48\r\n" +
			"drwxr-xr-x  6 user user 4096 Jan 1 12:00 \x1b[34m.\x1b[0m\r\n" +
			"drwxr-xr-x 10 user user 4096 Jan 1 11:00 \x1b[34m..\x1b[0m\r\n" +
			"-rw-r--r--  1 user user  220 Jan 1 10:00 .bashrc\r\n" +
			"$ \x1b[?2004l") // Disable bracketed paste

	parser.Advance(performer, shellOutput)

	// Basic sanity checks
	assert.NotEmpty(t, performer.Actions)

	// Count different action types
	stats := make(map[string]int)
	for _, action := range performer.Actions {
		stats[action.Type]++
	}

	assert.Greater(t, stats["print"], 50, "Should have printed many characters")
	assert.Greater(t, stats["csi"], 4, "Should have multiple CSI sequences")
	assert.Greater(t, stats["execute"], 5, "Should have CR and LF characters")
}

// TestIntegrationErrorRecovery tests parser recovery from malformed sequences
func TestIntegrationErrorRecovery(t *testing.T) {
	parser := NewParser()
	performer := &IntegrationPerformer{}

	// Mix of valid and potentially malformed sequences
	sequences := [][]byte{
		[]byte("Valid text"),
		[]byte("\x1b["),                  // Incomplete CSI
		[]byte("999999999999999999999m"), // Overflow attempt
		[]byte("More valid text"),
		[]byte("\x1b]"), // Incomplete OSC
		[]byte("\x07"),  // Stray BEL
		[]byte("Final text"),
	}

	// Parser should handle all without panicking
	for _, seq := range sequences {
		parser.Advance(performer, seq)
	}

	// Verify we still got the valid text
	var text bytes.Buffer
	for _, action := range performer.Actions {
		if action.Type == "print" {
			text.WriteRune(action.Data.(rune))
		}
	}

	result := text.String()
	assert.Contains(t, result, "Valid text")
	assert.Contains(t, result, "More valid text")
	assert.Contains(t, result, "Final text")
}

// BenchmarkIntegrationRealWorldParsing benchmarks parsing of realistic terminal output
func BenchmarkIntegrationRealWorldParsing(b *testing.B) {
	// Create a realistic terminal output with various sequences
	var buf bytes.Buffer
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&buf, "\x1b[%dm Line %d with color \x1b[0m\r\n", 31+(i%7), i)
		if i%10 == 0 {
			buf.WriteString("\x1b[2K") // Clear line
			buf.WriteString("\x1b[1A") // Move up
		}
		if i%20 == 0 {
			fmt.Fprintf(&buf, "\x1b]0;Window Title %d\x07", i) // Set title
		}
	}
	input := buf.Bytes()

	parser := NewParser()
	performer := &NoopPerformer{}

	b.SetBytes(int64(len(input)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parser.Advance(performer, input)
	}
}
