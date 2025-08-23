// Package main demonstrates a simple example of using GoVTE
package main

import (
	"fmt"

	"github.com/cliofy/govte"
	"github.com/cliofy/govte/terminal"
)

func main() {
	fmt.Println("=== GoVTE Simple Example ===\n")

	// Example 1: Basic text parsing
	fmt.Println("1. Basic text parsing:")
	parser := govte.NewParser()
	term := terminal.NewTerminalBuffer(80, 24)

	input := []byte("Hello, GoVTE!")
	for _, b := range input {
		parser.Advance(term, []byte{b})
	}
	fmt.Printf("   Output: %s\n\n", term.GetDisplay())

	// Example 2: ANSI color codes
	fmt.Println("2. ANSI color codes:")
	coloredText := "\x1b[31mRed\x1b[0m \x1b[32mGreen\x1b[0m \x1b[34mBlue\x1b[0m"
	output := terminal.ParseBytesWithColors([]byte(coloredText), 80, 24)
	fmt.Printf("   Input:  %q\n", coloredText)
	fmt.Printf("   Output: %s\n\n", output)

	// Example 3: Cursor movement
	fmt.Println("3. Cursor movement and overwriting:")
	term2 := terminal.NewTerminalBuffer(80, 24)
	parser2 := govte.NewParser()

	// Write "Hello", move cursor back, overwrite with "World"
	sequence := []byte("Hello\r     \rWorld")
	for _, b := range sequence {
		parser2.Advance(term2, []byte{b})
	}
	fmt.Printf("   Sequence: %q\n", sequence)
	fmt.Printf("   Output:   %s\n\n", term2.GetDisplay())

	// Example 4: Text attributes
	fmt.Println("4. Text attributes:")
	styledText := "\x1b[1mBold\x1b[0m \x1b[4mUnderline\x1b[0m \x1b[7mReverse\x1b[0m"
	output2 := terminal.ParseBytesWithColors([]byte(styledText), 80, 24)
	fmt.Printf("   Input:  %q\n", styledText)
	fmt.Printf("   Output: %s\n\n", output2)

	// Example 5: 256-color and RGB
	fmt.Println("5. Extended colors (256-color and RGB):")
	extColors := "\x1b[38;5;196mColor 196\x1b[0m \x1b[38;2;255;100;0mRGB Orange\x1b[0m"
	output3 := terminal.ParseBytesWithColors([]byte(extColors), 80, 24)
	fmt.Printf("   Input:  %q\n", extColors)
	fmt.Printf("   Output: %s\n\n", output3)

	// Example 6: Custom Performer
	fmt.Println("6. Custom Performer implementation:")
	customParser := govte.NewParser()
	counter := &CharCounter{}

	testInput := []byte("Count\nthese\nlines!")
	for _, b := range testInput {
		customParser.Advance(counter, []byte{b})
	}
	fmt.Printf("   Input: %q\n", testInput)
	fmt.Printf("   Stats: %d characters, %d control codes\n",
		counter.charCount, counter.controlCount)
}

// CharCounter is a custom Performer that counts characters and control codes
type CharCounter struct {
	govte.NoopPerformer
	charCount    int
	controlCount int
}

func (c *CharCounter) Print(ch rune) {
	c.charCount++
}

func (c *CharCounter) Execute(b byte) {
	c.controlCount++
}
