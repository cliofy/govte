//! GoVTE 实时动画进度条示例
//! 
//! 展示如何使用GoVTE创建真正的动画效果，包括：
//! - 使用 \r 回车符实现行内更新
//! - 时间控制实现平滑动画
//! - 多种进度条样式
//! - ANSI颜色序列处理

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cliofy/govte"
)

// simpleProgressBar 简单进度条 - 直接使用GoVTE处理ANSI序列
func simpleProgressBar(durationSecs int) {
	fmt.Printf("简单进度条（%d秒）:\n", durationSecs)
	
	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()
	
	totalSteps := 100
	delayMs := time.Duration((durationSecs * 1000) / totalSteps)
	
	for i := 0; i <= totalSteps; i++ {
		// 计算进度条宽度（50个字符宽）
		filled := (i * 50) / totalSteps
		empty := 50 - filled
		
		// 构造进度条序列：\r + 进度条内容
		progressText := fmt.Sprintf("\r[%s%s] %d%%", 
			repeatString("=", filled),
			repeatString(" ", empty),
			i)
		
		// 使用GoVTE处理序列
		processor.Advance(handler, []byte(progressText))
		handler.Flush()
		
		if i < totalSteps {
			time.Sleep(delayMs * time.Millisecond)
		}
	}
	
	handler.PrintLineDirect(" 完成!")
}

// animatedProgressBar 带动画的进度条 - 有移动的指示器
func animatedProgressBar(durationSecs int) {
	fmt.Printf("\n带动画的进度条（%d秒）:\n", durationSecs)
	
	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()
	
	totalSteps := 100
	delayMs := time.Duration((durationSecs * 1000) / totalSteps)
	
	for i := 0; i <= totalSteps; i++ {
		spinner := GetSpinner(i, "braille")
		bar := handler.RenderUnicodeBar(i, 50)
		
		// 构造带动画的进度条
		progressText := fmt.Sprintf("\r%c %s", spinner, bar)
		
		processor.Advance(handler, []byte(progressText))
		handler.Flush()
		
		if i < totalSteps {
			time.Sleep(delayMs * time.Millisecond)
		}
	}
	
	handler.PrintLineDirect(" ✓")
}

// coloredProgressBar 彩色进度条 - 使用 ANSI 颜色代码通过GoVTE处理
func coloredProgressBar(durationSecs int) {
	fmt.Printf("\n彩色进度条（%d秒）:\n", durationSecs)
	
	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()
	
	totalSteps := 100
	delayMs := time.Duration((durationSecs * 1000) / totalSteps)
	
	for i := 0; i <= totalSteps; i++ {
		bar := handler.RenderColoredBar(i, 50)
		
		// 使用 \r 回车符进行行内更新
		progressText := fmt.Sprintf("\r%s", bar)
		
		processor.Advance(handler, []byte(progressText))
		handler.Flush()
		
		if i < totalSteps {
			time.Sleep(delayMs * time.Millisecond)
		}
	}
	
	handler.PrintLineDirect(" 完成!")
}

// multiProgressBars 多任务进度条 - 同时显示多个进度
func multiProgressBars() {
	fmt.Println("\n多任务进度条（模拟下载）:")
	
	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()
	
	// 保存光标位置 - 通过GoVTE处理
	processor.Advance(handler, []byte("\x1b[s"))
	
	// 准备显示区域
	handler.PrintLineDirect("文件 1: [                                                  ] 0%")
	handler.PrintLineDirect("文件 2: [                                                  ] 0%")
	handler.PrintLineDirect("文件 3: [                                                  ] 0%")
	handler.PrintLineDirect("总进度: [                                                  ] 0%")
	
	progress := [3]int{0, 0, 0}
	speeds := [3]int{3, 5, 2} // 不同的下载速度
	
	start := time.Now()
	
	for hasIncompleteTask(progress[:]) {
		// 更新每个进度
		for i := 0; i < 3; i++ {
			if progress[i] < 100 {
				progress[i] = min(progress[i] + speeds[i], 100)
			}
		}
		
		// 计算总进度
		totalProgress := (progress[0] + progress[1] + progress[2]) / 3
		
		// 恢复光标位置并更新显示 - 通过GoVTE处理ANSI序列
		processor.Advance(handler, []byte("\x1b[u")) // 恢复光标位置
		
		for i := 0; i < 3; i++ {
			// 向下移动光标到对应行
			moveSeq := fmt.Sprintf("\x1b[%dB", i+1)
			processor.Advance(handler, []byte(moveSeq))
			
			// 更新进度条
			bar := renderFileProgressBar(progress[i], 50)
			updateText := fmt.Sprintf("\r文件 %d: %s", i+1, bar)
			processor.Advance(handler, []byte(updateText))
			
			if i < 2 {
				processor.Advance(handler, []byte("\x1b[u")) // 恢复到起始位置
			}
		}
		
		// 显示总进度
		processor.Advance(handler, []byte("\x1b[u\x1b[4B")) // 恢复位置并移动到总进度行
		totalBar := renderTotalProgressBar(totalProgress, 50)
		totalText := fmt.Sprintf("\r总进度: %s", totalBar)
		processor.Advance(handler, []byte(totalText))
		
		handler.Flush()
		time.Sleep(100 * time.Millisecond)
		
		// 防止运行时间过长
		if time.Since(start).Seconds() > 10 {
			break
		}
	}
	
	handler.PrintLineDirect("\n\n✅ 所有下载完成!")
}

// streamingProgress 实时数据流进度（模拟日志输出）
func streamingProgress() {
	fmt.Println("\n实时数据流（模拟日志处理）:")
	fmt.Println(repeatString("─", 60))
	
	processor := govte.NewProcessor(NewProgressHandler())
	handler := NewProgressHandler()
	
	messages := []string{
		"初始化系统...",
		"加载配置文件...",
		"连接数据库...",
		"验证权限...",
		"加载模块...",
		"启动服务...",
		"监听端口 8080...",
		"系统就绪",
	}
	
	for i, msg := range messages {
		// 显示处理中的动画
		for j := 0; j < 8; j++ {
			spinner := GetSpinner(j, "blocks")
			statusText := fmt.Sprintf("\r%c %s", spinner, msg)
			
			processor.Advance(handler, []byte(statusText))
			handler.Flush()
			time.Sleep(125 * time.Millisecond)
		}
		
		// 完成当前步骤
		progress := ((i + 1) * 100) / len(messages)
		completeText := fmt.Sprintf("\r✓ %s [%d%%]\n", msg, progress)
		processor.Advance(handler, []byte(completeText))
		handler.Flush()
		
		time.Sleep(200 * time.Millisecond)
	}
	
	handler.PrintLineDirect(repeatString("─", 60))
	handler.PrintLineDirect("🚀 系统启动完成!")
}

// 辅助函数

// repeatString 重复字符串n次（Go 1.21之前版本的strings.Repeat替代）
func repeatString(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// hasIncompleteTask 检查是否有未完成的任务
func hasIncompleteTask(progress []int) bool {
	for _, p := range progress {
		if p < 100 {
			return true
		}
	}
	return false
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderFileProgressBar 渲染文件下载进度条
func renderFileProgressBar(progress, width int) string {
	filled := (progress * width) / 100
	
	bar := "["
	for j := 0; j < width; j++ {
		if j < filled {
			if progress == 100 {
				bar += "\x1b[32m=\x1b[0m" // 完成时显示绿色
			} else {
				bar += "="
			}
		} else if j == filled && progress < 100 {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += fmt.Sprintf("] %d%%", progress)
	
	return bar
}

// renderTotalProgressBar 渲染总进度条
func renderTotalProgressBar(progress, width int) string {
	filled := (progress * width) / 100
	
	bar := "["
	for j := 0; j < width; j++ {
		if j < filled {
			bar += "\x1b[36m▓\x1b[0m" // 青色
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("] %d%%", progress)
	
	return bar
}

func main() {
	fmt.Println("=== 动画进度条示例 (GoVTE) ===\n")
	
	// 检查终端是否支持UTF-8
	if os.Getenv("LANG") == "" {
		fmt.Println("注意: 如果显示异常，请确保终端支持UTF-8编码")
		fmt.Println()
	}
	
	// 示例 1: 简单进度条
	simpleProgressBar(3)
	
	// 示例 2: 带动画的进度条
	animatedProgressBar(3)
	
	// 示例 3: 彩色进度条
	coloredProgressBar(3)
	
	// 示例 4: 多任务进度条
	multiProgressBars()
	
	// 示例 5: 实时数据流
	streamingProgress()
	
	fmt.Println("\n所有示例完成！")
	fmt.Println("\n本示例展示了GoVTE库的以下功能:")
	fmt.Println("• ANSI转义序列解析和处理")
	fmt.Println("• 终端光标控制和定位")
	fmt.Println("• 颜色序列处理和渲染") 
	fmt.Println("• 实时输出和缓冲区管理")
	fmt.Println("• Unicode字符支持")
}