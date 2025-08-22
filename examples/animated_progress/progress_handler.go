package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cliofy/govte"
)

// ProgressHandler 是一个专门用于进度条显示的Handler实现
// 它继承了NoopHandler，并重写了需要的方法来支持进度条渲染
type ProgressHandler struct {
	govte.NoopHandler
	output strings.Builder
	currentLine strings.Builder
}

// NewProgressHandler 创建一个新的ProgressHandler实例
func NewProgressHandler() *ProgressHandler {
	return &ProgressHandler{}
}

// Input 处理字符输出，将字符添加到当前行
func (h *ProgressHandler) Input(c rune) {
	h.currentLine.WriteRune(c)
}

// CarriageReturn 处理回车符，用于行内更新（\r）
func (h *ProgressHandler) CarriageReturn() {
	// 回车不换行，只是将光标移到行首
	// 下次输出会覆盖当前行内容
	fmt.Print("\r")
	h.currentLine.Reset()
}

// LineFeed 处理换行符，输出当前行并换行
func (h *ProgressHandler) LineFeed() {
	if h.currentLine.Len() > 0 {
		fmt.Print(h.currentLine.String())
		h.currentLine.Reset()
	}
	fmt.Println()
}

// Bell 处理响铃字符
func (h *ProgressHandler) Bell() {
	fmt.Print("\a") // 发送响铃字符到终端
}

// SetForeground 设置前景色
func (h *ProgressHandler) SetForeground(color govte.Color) {
	// 根据GoVTE的Color结构体输出对应的ANSI序列
	switch color.Type {
	case govte.ColorTypeNamed:
		switch color.Named {
		case govte.Red:
			fmt.Print("\x1b[31m")
		case govte.Green:
			fmt.Print("\x1b[32m")
		case govte.Yellow:
			fmt.Print("\x1b[33m")
		case govte.Blue:
			fmt.Print("\x1b[34m")
		case govte.Magenta:
			fmt.Print("\x1b[35m")
		case govte.Cyan:
			fmt.Print("\x1b[36m")
		case govte.White:
			fmt.Print("\x1b[37m")
		case govte.Black:
			fmt.Print("\x1b[30m")
		}
	case govte.ColorTypeIndexed:
		// 256色模式
		fmt.Printf("\x1b[38;5;%dm", color.Index)
	case govte.ColorTypeRgb:
		// RGB真彩色模式
		fmt.Printf("\x1b[38;2;%d;%d;%dm", color.Rgb.R, color.Rgb.G, color.Rgb.B)
	}
}

// ResetColors 重置颜色到默认值
func (h *ProgressHandler) ResetColors() {
	fmt.Print("\x1b[0m")
}

// ClearLine 清除行内容（用于进度条更新）
func (h *ProgressHandler) ClearLine(mode govte.LineClearMode) {
	switch mode {
	case govte.LineClearRight:
		fmt.Print("\x1b[K") // 清除从光标到行尾
	case govte.LineClearLeft:
		fmt.Print("\x1b[1K") // 清除从行首到光标
	case govte.LineClearAll:
		fmt.Print("\x1b[2K") // 清除整行
	}
}

// Flush 立即刷新输出缓冲区，确保进度条及时显示
func (h *ProgressHandler) Flush() {
	if h.currentLine.Len() > 0 {
		fmt.Print(h.currentLine.String())
		h.currentLine.Reset()
	}
	os.Stdout.Sync()
}

// PrintDirect 直接输出文本，绕过ANSI处理（用于简单输出）
func (h *ProgressHandler) PrintDirect(text string) {
	fmt.Print(text)
	os.Stdout.Sync()
}

// PrintLineDirect 直接输出一行文本
func (h *ProgressHandler) PrintLineDirect(text string) {
	fmt.Println(text)
	os.Stdout.Sync()
}

// 进度条渲染辅助函数

// RenderSimpleBar 渲染简单的ASCII进度条
func (h *ProgressHandler) RenderSimpleBar(progress, width int) string {
	filled := (progress * width) / 100
	empty := width - filled
	
	bar := "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"
	return fmt.Sprintf("%s %d%%", bar, progress)
}

// RenderUnicodeBar 渲染Unicode风格的进度条
func (h *ProgressHandler) RenderUnicodeBar(progress, width int) string {
	filled := (progress * width) / 100
	empty := width - filled
	
	var bar strings.Builder
	bar.WriteString("[")
	
	// 已完成部分
	bar.WriteString(strings.Repeat("█", filled))
	
	// 部分完成字符（如果有余数）
	remainder := (progress * width) % 100
	if remainder > 0 && filled < width {
		bar.WriteString("▓")
		empty--
	}
	
	// 未完成部分
	bar.WriteString(strings.Repeat("░", empty))
	bar.WriteString("]")
	
	return fmt.Sprintf("%s %d%%", bar.String(), progress)
}

// RenderColoredBar 渲染带颜色的进度条
func (h *ProgressHandler) RenderColoredBar(progress, width int) string {
	filled := (progress * width) / 100
	
	// 根据进度选择颜色
	var colorCode string
	if progress < 33 {
		colorCode = "\x1b[31m" // 红色
	} else if progress < 66 {
		colorCode = "\x1b[33m" // 黄色
	} else {
		colorCode = "\x1b[32m" // 绿色
	}
	
	var bar strings.Builder
	bar.WriteString("Progress: [")
	
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteString("=")
		} else if i == filled && progress < 100 {
			bar.WriteString(">")
		} else {
			bar.WriteString(" ")
		}
	}
	
	bar.WriteString("]")
	
	return fmt.Sprintf("%s%s %d%%\x1b[0m", colorCode, bar.String(), progress)
}

// GetSpinner 获取旋转器字符
func GetSpinner(step int, spinnerType string) rune {
	switch spinnerType {
	case "braille":
		spinners := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		return spinners[step%len(spinners)]
	case "blocks":
		spinners := []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}
		return spinners[step%len(spinners)]
	default:
		spinners := []rune{'|', '/', '-', '\\'}
		return spinners[step%len(spinners)]
	}
}