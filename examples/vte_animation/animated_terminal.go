package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cliofy/govte"
)

// AnimatedTerminal 动画终端 - 包含VTE解析器和缓冲区
// 类似于Rust版本的AnimatedTerminal结构体
type AnimatedTerminal struct {
	// GoVTE 解析器
	parser *govte.Parser
	// 终端缓冲区
	buffer *TerminalBuffer
}

// NewAnimatedTerminal 创建一个新的动画终端
func NewAnimatedTerminal(width, height int) *AnimatedTerminal {
	return &AnimatedTerminal{
		parser: govte.NewParser(),
		buffer: NewTerminalBuffer(width, height),
	}
}

// Process 处理输入并更新缓冲区
// 相当于Rust版本的process方法
func (a *AnimatedTerminal) Process(input []byte) {
	a.parser.Advance(a.buffer, input)
}

// ProcessString 处理字符串输入的便捷方法
func (a *AnimatedTerminal) ProcessString(input string) {
	a.Process([]byte(input))
}

// Render 渲染当前缓冲区到终端
// 实现带边框的终端显示，类似Rust版本的render方法
func (a *AnimatedTerminal) Render() {
	width, height := a.buffer.GetDimensions()
	buffer := a.buffer.GetBuffer()
	
	// 隐藏光标避免闪烁
	fmt.Print("\x1b[?25l")
	
	// 使用绝对定位绘制顶部边框（第1行）
	fmt.Printf("\x1b[1;1H┌%s┐\x1b[K", strings.Repeat("─", width))
	
	// 使用绝对定位绘制每一行内容
	for i, line := range buffer {
		// 定位到第 i+2 行（因为第1行是顶部边框）
		fmt.Printf("\x1b[%d;1H│", i+2)
		for _, ch := range line {
			fmt.Printf("%c", ch)
		}
		fmt.Print("│\x1b[K") // 绘制右边框并清除行尾
	}
	
	// 使用绝对定位绘制底部边框
	bottomRow := height + 2
	fmt.Printf("\x1b[%d;1H└%s┘\x1b[K", bottomRow, strings.Repeat("─", width))
	
	// 恢复光标显示
	fmt.Print("\x1b[?25h")
	
	// 刷新输出
	os.Stdout.Sync()
}

// Clear 清空终端缓冲区
func (a *AnimatedTerminal) Clear() {
	a.buffer.Clear()
}

// GetBuffer 获取底层缓冲区（用于直接访问）
func (a *AnimatedTerminal) GetBuffer() *TerminalBuffer {
	return a.buffer
}

// GetDimensions 获取终端尺寸
func (a *AnimatedTerminal) GetDimensions() (int, int) {
	return a.buffer.GetDimensions()
}

// GetCursor 获取当前光标位置
func (a *AnimatedTerminal) GetCursor() (int, int) {
	return a.buffer.GetCursor()
}

// MoveCursor 移动光标到指定位置（便捷方法）
func (a *AnimatedTerminal) MoveCursor(row, col int) {
	cmd := fmt.Sprintf("\x1b[%d;%dH", row+1, col+1) // 转换为1基索引
	a.ProcessString(cmd)
}

// WriteAt 在指定位置写入文本（便捷方法）
func (a *AnimatedTerminal) WriteAt(row, col int, text string) {
	a.MoveCursor(row, col)
	a.ProcessString(text)
}

// WriteAtColored 在指定位置写入带颜色的文本
func (a *AnimatedTerminal) WriteAtColored(row, col int, text string, colorCode string) {
	a.MoveCursor(row, col)
	coloredText := fmt.Sprintf("%s%s\x1b[0m", colorCode, text)
	a.ProcessString(coloredText)
}

// ClearScreen 清屏（发送CSI序列）
func (a *AnimatedTerminal) ClearScreen() {
	a.ProcessString("\x1b[2J")
}

// ClearLine 清除当前行
func (a *AnimatedTerminal) ClearLine() {
	a.ProcessString("\x1b[K")
}

// SetTitle 设置窗口标题（便捷方法）
func (a *AnimatedTerminal) SetTitle(title string) {
	titleSeq := fmt.Sprintf("\x1b]0;%s\x07", title)
	fmt.Print(titleSeq)
}

// 颜色常量，方便使用
const (
	ColorReset   = "\x1b[0m"
	ColorRed     = "\x1b[31m"
	ColorGreen   = "\x1b[32m"
	ColorYellow  = "\x1b[33m"
	ColorBlue    = "\x1b[34m"
	ColorMagenta = "\x1b[35m"
	ColorCyan    = "\x1b[36m"
	ColorWhite   = "\x1b[37m"
	
	// 明亮颜色
	ColorBrightRed     = "\x1b[91m"
	ColorBrightGreen   = "\x1b[92m"
	ColorBrightYellow  = "\x1b[93m"
	ColorBrightBlue    = "\x1b[94m"
	ColorBrightMagenta = "\x1b[95m"
	ColorBrightCyan    = "\x1b[96m"
	ColorBrightWhite   = "\x1b[97m"
)

// PrintTitle 在终端顶部打印标题（用于演示间的转换）
func PrintTitle(title string) {
	fmt.Print("\x1b[H\x1b[J") // 清屏并回到左上角
	fmt.Printf("%s%s%s\n", ColorBrightCyan, title, ColorReset)
	os.Stdout.Sync()
}

// EnterAlternateScreen 进入alternate screen buffer
func EnterAlternateScreen() {
	fmt.Print("\x1b[s\x1b[?1049h\x1b[?25l\x1b[H\x1b[2J")
	os.Stdout.Sync()
}

// ExitAlternateScreen 退出alternate screen buffer
func ExitAlternateScreen() {
	fmt.Print("\x1b[?25h\x1b[?1049l\x1b[u")
	os.Stdout.Sync()
}