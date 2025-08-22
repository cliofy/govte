//! GoVTE 动画终端演示
//! 
//! 展示如何将 GoVTE 与实时更新结合，创建各种终端动画效果
//! 这是Rust版本vte_animation.rs的Go实现

package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// 检查终端支持
	if !checkTerminalSupport() {
		fmt.Println("警告: 终端可能不完全支持所需功能，显示效果可能异常")
		time.Sleep(2 * time.Second)
	}
	
	// 进入 alternate screen buffer，隐藏光标，并清屏
	EnterAlternateScreen()
	
	// 确保退出时恢复终端状态
	defer func() {
		ExitAlternateScreen()
		fmt.Println("感谢使用 GoVTE 动画演示！")
	}()
	
	// 在 alternate screen 中显示标题
	fmt.Print("\x1b[1;1H=== VTE 动画终端演示 (GoVTE) ===")
	fmt.Print("\x1b[2;1H使用 GoVTE 解析器实现各种终端动画效果")
	fmt.Print("\x1b[4;1H本演示包含以下动画:")
	fmt.Print("\x1b[5;3H• 📊 动画进度条")
	fmt.Print("\x1b[6;3H• ⌨️  打字机效果")
	fmt.Print("\x1b[7;3H• 💊 矩阵雨效果")
	fmt.Print("\x1b[8;3H• 📈 实时图表")
	fmt.Print("\x1b[9;3H• 🌊 波浪动画 (增强)")
	fmt.Print("\x1b[10;3H• 🌀 螺旋动画 (增强)")
	fmt.Print("\x1b[11;3H• 🎆 烟花动画 (创新)")
	fmt.Print("\x1b[13;1H按 Ctrl+C 可随时退出")
	fmt.Print("\x1b[15;1H正在启动演示...")
	os.Stdout.Sync()
	
	// 启动倒计时
	for i := 3; i > 0; i-- {
		fmt.Printf("\x1b[15;15H开始倒计时: %d", i)
		os.Stdout.Sync()
		time.Sleep(1 * time.Second)
	}
	
	// 运行各个演示
	runAllDemos()
	
	// 显示完成信息
	fmt.Print("\x1b[H\x1b[2J")
	fmt.Print("\x1b[10;20H✨ 所有演示完成！")
	fmt.Print("\x1b[12;15H本演示展示了GoVTE的强大功能:")
	fmt.Print("\x1b[13;17H• VTE解析器的ANSI序列处理")
	fmt.Print("\x1b[14;17H• 终端缓冲区管理和渲染")
	fmt.Print("\x1b[15;17H• 实时动画和视觉效果")
	fmt.Print("\x1b[16;17H• 光标控制和屏幕操作")
	fmt.Print("\x1b[18;20H感谢观看！3秒后自动退出...")
	os.Stdout.Sync()
	time.Sleep(3 * time.Second)
}

// runAllDemos 运行所有演示
func runAllDemos() {
	demos := []struct {
		name string
		fn   func()
	}{
		{"动画进度条", DemoProgressBar},
		{"打字机效果", DemoTypewriter},
		{"矩阵雨效果", DemoMatrixRain},
		{"实时图表", DemoLiveChart},
		{"波浪动画", DemoWaveAnimation},
		{"螺旋动画", DemoSpiralAnimation},
		{"烟花动画", DemoFireworks},
	}
	
	for i, demo := range demos {
		// 显示当前演示信息
		showDemoTransition(i+1, len(demos), demo.name)
		
		// 运行演示
		demo.fn()
		
		// 演示间的间隔
		if i < len(demos)-1 {
			showTransitionMessage("准备下一个演示...")
			time.Sleep(1 * time.Second)
		}
	}
}

// showDemoTransition 显示演示切换信息
func showDemoTransition(current, total int, name string) {
	fmt.Print("\x1b[H\x1b[2J") // 清屏
	fmt.Printf("\x1b[8;20H演示进度: %d/%d", current, total)
	fmt.Printf("\x1b[10;20H当前演示: %s%s%s", ColorBrightYellow, name, ColorReset)
	fmt.Print("\x1b[12;25H准备中...")
	os.Stdout.Sync()
	time.Sleep(800 * time.Millisecond)
}

// showTransitionMessage 显示过渡消息
func showTransitionMessage(message string) {
	fmt.Print("\x1b[H\x1b[2J")
	fmt.Printf("\x1b[10;%dH%s", (80-len(message))/2, message)
	os.Stdout.Sync()
}

// checkTerminalSupport 检查终端支持情况
func checkTerminalSupport() bool {
	// 检查基本环境变量
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}
	
	// 检查是否支持彩色
	colorTerm := os.Getenv("COLORTERM")
	if colorTerm == "" && term != "xterm-256color" && term != "screen-256color" {
		return false
	}
	
	return true
}

// 演示功能说明
func init() {
	// 设置随机种子
	SetSeed(uint64(time.Now().UnixNano()))
}