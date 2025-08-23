package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cliofy/govte"
)

// LogPerformer logs all actions to stdout
type LogPerformer struct {
	govte.NoopPerformer
}

func (l *LogPerformer) Print(c rune) {
	fmt.Printf("[print] %q\n", c)
}

func (l *LogPerformer) Execute(b byte) {
	fmt.Printf("[execute] 0x%02x", b)
	switch b {
	case 0x08:
		fmt.Print(" (BS)")
	case 0x09:
		fmt.Print(" (HT)")
	case 0x0A:
		fmt.Print(" (LF)")
	case 0x0D:
		fmt.Print(" (CR)")
	case 0x1B:
		fmt.Print(" (ESC)")
	}
	fmt.Println()
}

func (l *LogPerformer) Hook(params *govte.Params, intermediates []byte, ignore bool, action rune) {
	fmt.Printf("[hook] params=%v, intermediates=%v, ignore=%v, action=%q\n",
		params, intermediates, ignore, action)
}

func (l *LogPerformer) Put(b byte) {
	fmt.Printf("[put] 0x%02x\n", b)
}

func (l *LogPerformer) Unhook() {
	fmt.Println("[unhook]")
}

func (l *LogPerformer) OscDispatch(params [][]byte, bellTerminated bool) {
	fmt.Printf("[osc_dispatch] params=%v, bell_terminated=%v\n", params, bellTerminated)
}

func (l *LogPerformer) CsiDispatch(params *govte.Params, intermediates []byte, ignore bool, action rune) {
	fmt.Printf("[csi_dispatch] params=%v, intermediates=%v, ignore=%v, action=%q\n",
		params, intermediates, ignore, action)
}

func (l *LogPerformer) EscDispatch(intermediates []byte, ignore bool, b byte) {
	fmt.Printf("[esc_dispatch] intermediates=%v, ignore=%v, byte=0x%02x\n",
		intermediates, ignore, b)
}

func main() {
	fmt.Println("=== GoVTE Parse Log ===")
	fmt.Println("Type or pipe input to see parsed actions")
	fmt.Println("Press Ctrl+D (Unix) or Ctrl+Z (Windows) to exit")
	fmt.Println()

	parser := govte.NewParser()
	performer := &LogPerformer{}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Bytes()
		fmt.Printf("Input: %q\n", input)

		// Parse the input using VTE parser
		parser.Advance(performer, input)

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
