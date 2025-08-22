//! 捕获并渲染 TUI 程序输出的示例（使用新的 terminal 模块）
//!
//! 这个示例展示如何使用新的 TerminalBuffer 实现：
//! 1. 在伪终端 (PTY) 中启动 TUI 程序（如 htop）
//! 2. 捕获程序的输出流
//! 3. 使用 GoVTE 的 terminal 模块解析并渲染输出

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cliofy/govte"
	"github.com/cliofy/govte/terminal"
	"github.com/creack/pty"
	"golang.org/x/term"
)

// getTerminalSize 获取当前终端大小，如果失败则返回默认值
func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return 120, 40 // 默认尺寸
	}
	return width, height
}

// captureTUIOutput 捕获 TUI 程序的输出
func captureTUIOutput(program string, args []string, duration time.Duration) ([]byte, int, int, error) {
	fmt.Printf("正在启动 %s ...\n", program)

	// 获取当前终端大小
	width, height := getTerminalSize()
	fmt.Printf("检测到终端大小: %dx%d\n", width, height)

	// 创建命令
	cmd := exec.Command(program, args...)

	// 设置环境变量
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// 创建 PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("无法创建 PTY: %w", err)
	}
	defer ptmx.Close()

	// 设置 PTY 大小
	err = pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
	})
	if err != nil {
		log.Printf("警告: 无法设置 PTY 大小: %v", err)
	}

	fmt.Printf("程序已启动，PID: %d\n", cmd.Process.Pid)
	fmt.Printf("开始捕获输出（%.0f 秒）...\n", duration.Seconds())

	// 收集输出
	var output []byte
	buffer := make([]byte, 4096)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// 使用 goroutine 读取数据
	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 设置读取超时
				ptmx.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, err := ptmx.Read(buffer)
				if err != nil {
					if err != io.EOF && !os.IsTimeout(err) {
						log.Printf("读取错误: %v", err)
					}
					continue
				}
				if n > 0 {
					output = append(output, buffer[:n]...)
					// 显示捕获进度
					fmt.Printf("\r已捕获 %d 字节", len(output))
				}
			}
		}
	}()

	// 等待超时或完成
	<-ctx.Done()

	// 给读取 goroutine 一点时间完成
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	}

	fmt.Println("\n捕获完成，正在关闭程序...")

	// 尝试优雅地终止程序
	if cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Wait()
	}

	// 给程序一点时间清理
	time.Sleep(100 * time.Millisecond)

	return output, width, height, nil
}

// renderOutput 使用新的 TerminalBuffer 渲染输出
func renderOutput(data []byte, width, height int, withColors bool) string {
	parser := govte.NewParser()
	terminalBuffer := terminal.NewTerminalBuffer(width, height)

	// 解析所有数据
	for _, b := range data {
		parser.Advance(terminalBuffer, []byte{b})
	}

	fmt.Println("\n=== 渲染统计 ===")
	fmt.Printf("捕获字节数: %d\n", len(data))
	fmt.Printf("终端大小: %dx%d\n", width, height)

	cursorX, cursorY := terminalBuffer.CursorPosition()
	fmt.Printf("光标位置: (%d, %d)\n", cursorX+1, cursorY+1)

	fmt.Printf("彩色输出: %s\n", map[bool]string{true: "启用", false: "禁用"}[withColors])

	if withColors {
		return terminalBuffer.GetDisplayWithColors()
	}
	return terminalBuffer.GetDisplay()
}

func main() {
	fmt.Println("=== GoVTE TUI 程序捕获示例 ===")

	// 检查是否启用颜色输出
	enableColors := false
	for _, arg := range os.Args[1:] {
		if arg == "--colors" || arg == "-c" {
			enableColors = true
			break
		}
	}

	if enableColors {
		fmt.Println("🎨 已启用彩色输出模式")
	} else {
		fmt.Println("💡 提示: 使用 --colors 或 -c 参数启用彩色输出")
	}
	fmt.Println()

	// 尝试不同的 TUI 程序
	programs := []struct {
		name string
		args []string
	}{
		{"htop", []string{}},
		{"btm", []string{}},
		{"top", []string{}},
		{"ps", []string{"aux"}},
	}

	var capturedData []byte
	var usedProgram string
	var terminalWidth, terminalHeight int

	// 尝试找到可用的程序
	for _, prog := range programs {
		data, width, height, err := captureTUIOutput(prog.name, prog.args, 3*time.Second)
		if err != nil {
			fmt.Printf("无法运行 %s: %v\n", prog.name, err)
			if prog.name == "htop" {
				fmt.Println("提示: 请安装 htop (例如: apt install htop 或 brew install htop)")
			}
			continue
		}

		capturedData = data
		terminalWidth = width
		terminalHeight = height
		usedProgram = prog.name
		break
	}

	// 渲染捕获的输出
	if capturedData != nil {
		fmt.Printf("\n成功捕获 %s 的输出\n", usedProgram)
		fmt.Println("\n=== 最终渲染帧 ===")

		rendered := renderOutput(capturedData, terminalWidth, terminalHeight, enableColors)

		// 直接输出渲染结果（避免 Unicode 字符截断问题）
		lines := strings.Split(rendered, "\n")
		for _, line := range lines {
			fmt.Println(line)
		}

		// 可选：将原始数据保存到文件
		for _, arg := range os.Args[1:] {
			if arg == "--save" {
				filename := fmt.Sprintf("%s_capture.dat", usedProgram)
				err := os.WriteFile(filename, capturedData, 0644)
				if err != nil {
					log.Printf("保存文件失败: %v", err)
				} else {
					fmt.Printf("\n原始数据已保存到: %s\n", filename)
				}
				break
			}
		}
	} else {
		fmt.Println("\n错误: 无法捕获任何 TUI 程序的输出")
		fmt.Println("请确保至少安装了 htop、top 或 ps 中的一个")
		os.Exit(1)
	}
}
