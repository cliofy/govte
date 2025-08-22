package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// DemoProgressBar 演示：动画进度条
// 复现Rust版本的demo_progress_bar函数
func DemoProgressBar() {
	PrintTitle("📊 VTE 动画进度条演示:")
	time.Sleep(500 * time.Millisecond)
	
	terminal := NewAnimatedTerminal(60, 5)
	
	// 标题
	terminal.WriteAt(0, 0, "Progress Bar Animation Demo")
	terminal.WriteAt(2, 4, "[")
	terminal.WriteAt(2, 55, "]")
	terminal.Render()
	time.Sleep(500 * time.Millisecond)
	
	// 动画进度条
	for i := 0; i <= 50; i++ {
		// 移动到进度条位置并绘制等号
		terminal.WriteAt(2, 5+i, "=")
		
		// 显示百分比
		percent := fmt.Sprintf("%d%%", i*2)
		terminal.WriteAt(3, 27, percent)
		
		terminal.Render()
		time.Sleep(100 * time.Millisecond)
	}
	
	// 完成信息
	terminal.WriteAtColored(4, 22, "Complete!", ColorGreen)
	terminal.Render()
	time.Sleep(1 * time.Second)
}

// DemoTypewriter 演示：打字机效果
// 复现Rust版本的demo_typewriter函数
func DemoTypewriter() {
	PrintTitle("⌨️  VTE 打字机效果演示:")
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
		// 移动光标到指定位置
		terminal.MoveCursor(msg.row, msg.col)
		
		// 逐字符显示
		for _, ch := range msg.text {
			terminal.ProcessString(string(ch))
			terminal.Render()
			time.Sleep(50 * time.Millisecond)
		}
		
		time.Sleep(300 * time.Millisecond)
	}
	
	time.Sleep(1 * time.Second)
}

// DemoMatrixRain 演示：矩阵雨效果
// 复现Rust版本的demo_matrix_rain函数
func DemoMatrixRain() {
	PrintTitle("💊 VTE 矩阵雨效果演示:")
	time.Sleep(500 * time.Millisecond)
	
	terminal := NewAnimatedTerminal(60, 15)
	
	// 每列的当前位置 - 随机初始化以避免整齐排列
	columns := make([]int, 60)
	for i := range columns {
		columns[i] = RandomRange(0, 15)
	}
	
	// 增加帧数以获得更好的视觉效果
	for frame := 0; frame < 80; frame++ {
		// 清屏
		terminal.ClearScreen()
		
		for col, row := range columns {
			// 只在有效范围内绘制
			if row >= 0 && row < 15 {
				// 绘制主字符（亮绿色，更突出）
				ch := RandomMatrixChar()
				terminal.WriteAtColored(row, col, string(ch), ColorBrightGreen)
			}
			
			// 绘制更长的拖尾效果
			for trailPos := 1; trailPos <= 3; trailPos++ {
				trailRow := row - trailPos
				if trailRow >= 0 && trailRow < 15 {
					trailChar := RandomMatrixChar()
					// 根据距离调整颜色深度
					if trailPos == 1 {
						terminal.WriteAtColored(trailRow, col, string(trailChar), ColorGreen)
					} else {
						// 更暗的绿色表示较老的拖尾
						terminal.WriteAtColored(trailRow, col, string(trailChar), "\x1b[2;32m") // 暗绿色
					}
				}
			}
		}
		
		terminal.Render()
		
		// 更新列位置 - 降低重置概率
		for i := range columns {
			if RandomFloat32() < 0.05 { // 从10%降低到5%
				columns[i] = -2 // 从屏幕上方开始
			} else {
				columns[i] = columns[i] + 1
				// 当字符移出屏幕底部时，从顶部重新开始
				if columns[i] > 18 {
					columns[i] = -RandomRange(1, 5) // 随机延迟重新开始
				}
			}
		}
		
		time.Sleep(80 * time.Millisecond) // 稍微加快动画速度
	}
}

// DemoLiveChart 演示：实时图表
// 复现Rust版本的demo_live_chart函数
func DemoLiveChart() {
	PrintTitle("📈 VTE 实时图表演示:")
	time.Sleep(500 * time.Millisecond)
	
	terminal := NewAnimatedTerminal(60, 15)
	values := make([]int, 60)
	
	// 初始化数值
	for i := range values {
		values[i] = 7
	}
	
	for frame := 0; frame < 100; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 0, "Live Chart (CPU Usage Simulation)")
		
		// 更新数据 - 使用正弦波模拟CPU使用率
		// 向左滚动数据
		for i := 0; i < len(values)-1; i++ {
			values[i] = values[i+1]
		}
		
		// 生成新数值
		sinValue := math.Sin(float64(frame) * 0.2)
		newVal := int((sinValue+1.0)*6.0) // 将[-1,1]映射到[0,12]
		if newVal > 12 {
			newVal = 12
		}
		values[59] = newVal
		
		// 绘制图表
		for row := 0; row < 13; row++ {
			y := 13 - row // 反转Y轴，使图表从下往上绘制
			terminal.MoveCursor(y, 0)
			
			for colIdx, val := range values {
				if val >= row {
					terminal.WriteAtColored(y, colIdx, "#", ColorCyan)
				} else {
					terminal.WriteAt(y, colIdx, " ")
				}
			}
		}
		
		// 绘制X轴
		terminal.WriteAt(14, 0, strings.Repeat("─", 60))
		
		terminal.Render()
		time.Sleep(100 * time.Millisecond)
	}
}

// DemoWaveAnimation 额外演示：波浪动画（增强版）
func DemoWaveAnimation() {
	PrintTitle("🌊 VTE 波浪动画演示:")
	time.Sleep(500 * time.Millisecond)
	
	terminal := NewAnimatedTerminal(60, 12)
	
	for frame := 0; frame < 80; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 25, "Wave Animation")
		
		// 生成波浪
		for x := 0; x < 60; x++ {
			// 使用正弦波生成Y坐标
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

// DemoSpiralAnimation 额外演示：螺旋动画（增强版）
func DemoSpiralAnimation() {
	PrintTitle("🌀 VTE 螺旋动画演示:")
	time.Sleep(500 * time.Millisecond)
	
	terminal := NewAnimatedTerminal(60, 15)
	
	for frame := 0; frame < 60; frame++ {
		terminal.ClearScreen()
		terminal.WriteAt(0, 25, "Spiral Animation")
		
		// 螺旋中心
		centerX, centerY := 30.0, 7.0
		
		// 绘制螺旋
		for i := 0; i < 100; i++ {
			angle := float64(i)*0.3 + float64(frame)*0.1
			radius := float64(i) * 0.15
			
			x := int(centerX + radius*math.Cos(angle))
			y := int(centerY + radius*math.Sin(angle)*0.5) // 压缩Y轴
			
			if x >= 0 && x < 60 && y >= 1 && y < 15 {
				// 根据距离中心的远近选择字符和颜色
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

// DemoFireworks 额外演示：烟花动画（创新功能）
func DemoFireworks() {
	PrintTitle("🎆 VTE 烟花动画演示:")
	time.Sleep(500 * time.Millisecond)
	
	terminal := NewAnimatedTerminal(70, 20)
	
	// 烟花粒子结构
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
		
		// 每隔一段时间发射新烟花
		if frame%20 == 0 {
			// 烟花爆炸中心
			explodeX := float64(RandomRange(15, 55))
			explodeY := float64(RandomRange(5, 15))
			
			// 生成粒子
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
					vy:    math.Sin(angle) * float64(speed) * 0.6, // 压缩垂直速度
					life:  RandomRange(15, 30),
					char:  RandomChoice(chars),
					color: color,
				})
			}
		}
		
		// 更新和绘制粒子
		newParticles := particles[:0]
		for i := range particles {
			p := &particles[i]
			
			// 更新位置
			p.x += p.vx * 0.5
			p.y += p.vy * 0.3
			p.vy += 0.1 // 重力
			p.life--
			
			// 绘制粒子
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