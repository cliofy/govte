package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// DemoProgressBar demo: animated progress bar
// Reproduces the Rust version demo_progress_bar function
func DemoProgressBar() {
	PrintTitle("📊 VTE Animated Progress Bar Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(60, 5)

	// Title
	terminal.WriteAt(0, 0, "Progress Bar Animation Demo")
	terminal.WriteAt(2, 4, "[")
	terminal.WriteAt(2, 55, "]")
	terminal.Render()
	time.Sleep(500 * time.Millisecond)

	// Animated progress bar
	for i := 0; i <= 50; i++ {
		// Move to progress bar position and draw equals sign
		terminal.WriteAt(2, 5+i, "=")

		// Display percentage
		percent := fmt.Sprintf("%d%%", i*2)
		terminal.WriteAt(3, 27, percent)

		terminal.Render()
		time.Sleep(100 * time.Millisecond)
	}

	// Completion message
	terminal.WriteAtColored(4, 22, "Complete!", ColorGreen)
	terminal.Render()
	time.Sleep(1 * time.Second)
}

// DemoTypewriter demo: typewriter effect
// Reproduces the Rust version demo_typewriter function
func DemoTypewriter() {
	PrintTitle("⌨️  VTE Typewriter Effect Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(60, 10)

	messages := []struct {
		row, col int
		text     string
	}{
		{1, 2, "Welcome to the VTE Animation Demo!"},
		{3, 2, "This simulates a typewriter effect..."},
		{5, 2, "Each character appears one by one."},
		{7, 2, "Using VTE parser for terminal control."},
		{9, 2, "Pretty cool, right? :)"},
	}

	terminal.Render()

	for _, msg := range messages {
		// Move cursor to specified position
		terminal.MoveCursor(msg.row, msg.col)

		// Display character by character
		for _, ch := range msg.text {
			terminal.ProcessString(string(ch))
			terminal.Render()
			time.Sleep(50 * time.Millisecond)
		}

		time.Sleep(300 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)
}

// DemoMatrixRain demo: matrix rain effect
// Reproduces the Rust version demo_matrix_rain function
func DemoMatrixRain() {
	PrintTitle("💊 VTE Matrix Rain Effect Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(60, 15)

	// Current position for each column - randomly initialized to avoid uniform arrangement
	columns := make([]int, 60)
	for i := range columns {
		columns[i] = RandomRange(0, 15)
	}

	// Increase frame count for better visual effects
	for frame := 0; frame < 80; frame++ {
		// Clear screen
		terminal.ClearScreen()

		for col, row := range columns {
			// Only draw within valid range
			if row >= 0 && row < 15 {
				// Draw main character (bright green, more prominent)
				ch := RandomMatrixChar()
				terminal.WriteAtColored(row, col, string(ch), ColorBrightGreen)
			}

			// Draw longer trail effect
			for trailPos := 1; trailPos <= 3; trailPos++ {
				trailRow := row - trailPos
				if trailRow >= 0 && trailRow < 15 {
					trailChar := RandomMatrixChar()
					// Adjust color depth based on distance
					if trailPos == 1 {
						terminal.WriteAtColored(trailRow, col, string(trailChar), ColorGreen)
					} else {
						// Darker green indicates older trail
						terminal.WriteAtColored(trailRow, col, string(trailChar), "\x1b[2;32m") // Dark green
					}
				}
			}
		}

		terminal.Render()

		// Update column positions - reduce reset probability
		for i := range columns {
			if RandomFloat32() < 0.05 { // Reduced from 10% to 5%
				columns[i] = -2 // Start from above the screen
			} else {
				columns[i] = columns[i] + 1
				// When character moves off bottom of screen, restart from top
				if columns[i] > 18 {
					columns[i] = -RandomRange(1, 5) // Random delay before restarting
				}
			}
		}

		time.Sleep(80 * time.Millisecond) // Slightly speed up animation
	}
}

// DemoLiveChart demo: live chart
// Reproduces the Rust version demo_live_chart function
func DemoLiveChart() {
	PrintTitle("📈 VTE Live Chart Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(60, 15)
	values := make([]int, 60)

	// Initialize values
	for i := range values {
		values[i] = 7
	}

	for frame := 0; frame < 100; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 0, "Live Chart (CPU Usage Simulation)")

		// Update data - use sine wave to simulate CPU usage
		// Scroll data to the left
		for i := 0; i < len(values)-1; i++ {
			values[i] = values[i+1]
		}

		// Generate new value
		sinValue := math.Sin(float64(frame) * 0.2)
		newVal := int((sinValue + 1.0) * 6.0) // Map [-1,1] to [0,12]
		if newVal > 12 {
			newVal = 12
		}
		values[59] = newVal

		// Draw chart
		for row := 0; row < 13; row++ {
			y := 13 - row // Invert Y-axis to draw chart from bottom up
			terminal.MoveCursor(y, 0)

			for colIdx, val := range values {
				if val >= row {
					terminal.WriteAtColored(y, colIdx, "#", ColorCyan)
				} else {
					terminal.WriteAt(y, colIdx, " ")
				}
			}
		}

		// Draw X-axis
		terminal.WriteAt(14, 0, strings.Repeat("─", 60))

		terminal.Render()
		time.Sleep(100 * time.Millisecond)
	}
}

// DemoWaveAnimation extra demo: wave animation (enhanced version)
func DemoWaveAnimation() {
	PrintTitle("🌊 VTE Wave Animation Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(60, 12)

	for frame := 0; frame < 80; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 25, "Wave Animation")

		// Generate waves
		for x := 0; x < 60; x++ {
			// Use sine wave to generate Y coordinates
			y1 := int(5.0 + 3.0*math.Sin(float64(x)*0.2+float64(frame)*0.3))
			y2 := int(6.0 + 2.0*math.Cos(float64(x)*0.15+float64(frame)*0.2))

			if y1 >= 0 && y1 < 12 {
				terminal.WriteAtColored(y1, x, "~", ColorBlue)
			}
			if y2 >= 0 && y2 < 12 && y2 != y1 {
				terminal.WriteAtColored(y2, x, "≈", ColorCyan)
			}
		}

		terminal.Render()
		time.Sleep(80 * time.Millisecond)
	}
}

// DemoSpiralAnimation extra demo: spiral animation (enhanced version)
func DemoSpiralAnimation() {
	PrintTitle("🌀 VTE Spiral Animation Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(60, 15)

	for frame := 0; frame < 60; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 25, "Spiral Animation")

		// Spiral center
		centerX, centerY := 30.0, 7.0

		// Draw spiral
		for i := 0; i < 100; i++ {
			angle := float64(i)*0.3 + float64(frame)*0.1
			radius := float64(i) * 0.15

			x := int(centerX + radius*math.Cos(angle))
			y := int(centerY + radius*math.Sin(angle)*0.5) // Compress Y-axis

			if x >= 0 && x < 60 && y >= 1 && y < 15 {
				// Choose character and color based on distance from center
				if radius < 3.0 {
					terminal.WriteAtColored(y, x, "*", ColorYellow)
				} else if radius < 6.0 {
					terminal.WriteAtColored(y, x, "●", ColorMagenta)
				} else {
					terminal.WriteAtColored(y, x, "○", ColorBlue)
				}
			}
		}

		terminal.Render()
		time.Sleep(120 * time.Millisecond)
	}
}

// DemoFireworks extra demo: fireworks animation (innovative feature)
func DemoFireworks() {
	PrintTitle("🎆 VTE Fireworks Animation Demo:")
	time.Sleep(500 * time.Millisecond)

	terminal := NewAnimatedTerminal(70, 20)

	// Fireworks particle structure
	type Particle struct {
		x, y   float64
		vx, vy float64
		life   int
		char   string
		color  string
	}

	var particles []Particle

	for frame := 0; frame < 200; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 28, "Fireworks Show!")

		// Launch new fireworks at intervals
		if frame%20 == 0 {
			// Fireworks explosion center
			explodeX := float64(RandomRange(15, 55))
			explodeY := float64(RandomRange(5, 15))

			// Generate particles
			colors := []string{ColorRed, ColorGreen, ColorBlue, ColorYellow, ColorMagenta, ColorCyan}
			chars := []string{"*", "●", "○", "◆", "◇", "▲"}
			color := RandomChoice(colors)

			for i := 0; i < 25; i++ {
				angle := float64(i) * 2.0 * math.Pi / 25.0
				speed := RandomFloat32()*3.0 + 1.0

				particles = append(particles, Particle{
					x:     explodeX,
					y:     explodeY,
					vx:    math.Cos(angle) * float64(speed),
					vy:    math.Sin(angle) * float64(speed) * 0.6, // Compress vertical speed
					life:  RandomRange(15, 30),
					char:  RandomChoice(chars),
					color: color,
				})
			}
		}

		// Update and draw particles
		newParticles := particles[:0]
		for i := range particles {
			p := &particles[i]

			// Update position
			p.x += p.vx * 0.5
			p.y += p.vy * 0.3
			p.vy += 0.1 // Gravity
			p.life--

			// Draw particle
			if p.life > 0 && p.x >= 0 && p.x < 70 && p.y >= 1 && p.y < 20 {
				terminal.WriteAtColored(int(p.y), int(p.x), p.char, p.color)
				newParticles = append(newParticles, *p)
			}
		}
		particles = newParticles

		terminal.Render()
		time.Sleep(80 * time.Millisecond)
	}
}
