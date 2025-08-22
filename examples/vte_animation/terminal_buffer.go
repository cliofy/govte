package main

import (
	"github.com/cliofy/govte"
)

// TerminalBuffer 实现了终端缓冲区，类似Rust版本的TerminalBuffer
// 它实现了govte.Performer接口来处理VTE解析器的回调
type TerminalBuffer struct {
	// 屏幕缓冲区 - 二维字符数组
	buffer [][]rune
	// 光标位置
	cursorRow int
	cursorCol int
	// 终端尺寸
	width  int
	height int
}

// NewTerminalBuffer 创建一个新的终端缓冲区
func NewTerminalBuffer(width, height int) *TerminalBuffer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
		// 初始化为空格字符
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}
	
	return &TerminalBuffer{
		buffer:    buffer,
		cursorRow: 0,
		cursorCol: 0,
		width:     width,
		height:    height,
	}
}

// Clear 清空缓冲区并重置光标
func (t *TerminalBuffer) Clear() {
	for i := range t.buffer {
		for j := range t.buffer[i] {
			t.buffer[i][j] = ' '
		}
	}
	t.cursorRow = 0
	t.cursorCol = 0
}

// GetBuffer 获取缓冲区内容（用于渲染）
func (t *TerminalBuffer) GetBuffer() [][]rune {
	return t.buffer
}

// GetCursor 获取当前光标位置
func (t *TerminalBuffer) GetCursor() (int, int) {
	return t.cursorRow, t.cursorCol
}

// GetDimensions 获取终端尺寸
func (t *TerminalBuffer) GetDimensions() (int, int) {
	return t.width, t.height
}

// === 实现 govte.Performer 接口 ===

// Print 处理可打印字符
func (t *TerminalBuffer) Print(c rune) {
	if t.cursorRow < t.height && t.cursorCol < t.width {
		t.buffer[t.cursorRow][t.cursorCol] = c
		t.cursorCol++
		
		// 自动换行
		if t.cursorCol >= t.width {
			t.cursorCol = 0
			if t.cursorRow < t.height-1 {
				t.cursorRow++
			}
		}
	}
}

// Execute 处理控制字符
func (t *TerminalBuffer) Execute(b byte) {
	switch b {
	case 0x08: // BS - 退格
		if t.cursorCol > 0 {
			t.cursorCol--
		}
	case 0x0A: // LF - 换行
		if t.cursorRow < t.height-1 {
			t.cursorRow++
		}
	case 0x0D: // CR - 回车
		t.cursorCol = 0
	}
}

// Hook DCS序列开始（暂不实现）
func (t *TerminalBuffer) Hook(params *govte.Params, intermediates []byte, ignore bool, action rune) {
}

// Put DCS数据（暂不实现）
func (t *TerminalBuffer) Put(b byte) {
}

// Unhook DCS序列结束（暂不实现）
func (t *TerminalBuffer) Unhook() {
}

// OscDispatch 处理OSC序列（暂不实现）
func (t *TerminalBuffer) OscDispatch(params [][]byte, bellTerminated bool) {
}

// CsiDispatch 处理CSI序列（核心终端控制）
func (t *TerminalBuffer) CsiDispatch(params *govte.Params, intermediates []byte, ignore bool, action rune) {
	if ignore {
		return
	}
	
	// 将Params转换为[]uint16切片以便处理
	var paramsVec []uint16
	if params != nil {
		groups := params.Iter()
		for _, group := range groups {
			if len(group) > 0 {
				paramsVec = append(paramsVec, group[0])
			}
		}
	}
	
	switch action {
	case 'H', 'f': // CUP - 光标定位
		row := 1
		col := 1
		
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			row = int(paramsVec[0])
		}
		if len(paramsVec) > 1 && paramsVec[1] > 0 {
			col = int(paramsVec[1])
		}
		
		// 转换为0基索引并限制在有效范围内
		t.cursorRow = min(row-1, t.height-1)
		t.cursorCol = min(col-1, t.width-1)
		if t.cursorRow < 0 {
			t.cursorRow = 0
		}
		if t.cursorCol < 0 {
			t.cursorCol = 0
		}
		
	case 'J': // ED - 擦除显示
		if len(paramsVec) == 0 || paramsVec[0] == 0 {
			// 清除从光标到屏幕末尾
			for row := t.cursorRow; row < t.height; row++ {
				startCol := 0
				if row == t.cursorRow {
					startCol = t.cursorCol
				}
				for col := startCol; col < t.width; col++ {
					t.buffer[row][col] = ' '
				}
			}
		} else if paramsVec[0] == 1 {
			// 清除从屏幕开始到光标
			for row := 0; row <= t.cursorRow; row++ {
				endCol := t.width
				if row == t.cursorRow {
					endCol = t.cursorCol + 1
				}
				for col := 0; col < endCol; col++ {
					t.buffer[row][col] = ' '
				}
			}
		} else if paramsVec[0] == 2 {
			// 清除整个屏幕
			for row := range t.buffer {
				for col := range t.buffer[row] {
					t.buffer[row][col] = ' '
				}
			}
			t.cursorRow = 0
			t.cursorCol = 0
		}
		
	case 'K': // EL - 擦除行
		if t.cursorRow < t.height {
			if len(paramsVec) == 0 || paramsVec[0] == 0 {
				// 清除到行尾
				for col := t.cursorCol; col < t.width; col++ {
					t.buffer[t.cursorRow][col] = ' '
				}
			} else if paramsVec[0] == 1 {
				// 清除行首到光标
				for col := 0; col <= t.cursorCol && col < t.width; col++ {
					t.buffer[t.cursorRow][col] = ' '
				}
			} else if paramsVec[0] == 2 {
				// 清除整行
				for col := 0; col < t.width; col++ {
					t.buffer[t.cursorRow][col] = ' '
				}
			}
		}
		
	case 'A': // CUU - 光标上移
		lines := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			lines = int(paramsVec[0])
		}
		t.cursorRow = max(0, t.cursorRow-lines)
		
	case 'B': // CUD - 光标下移  
		lines := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			lines = int(paramsVec[0])
		}
		t.cursorRow = min(t.height-1, t.cursorRow+lines)
		
	case 'C': // CUF - 光标右移
		cols := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			cols = int(paramsVec[0])
		}
		t.cursorCol = min(t.width-1, t.cursorCol+cols)
		
	case 'D': // CUB - 光标左移
		cols := 1
		if len(paramsVec) > 0 && paramsVec[0] > 0 {
			cols = int(paramsVec[0])
		}
		t.cursorCol = max(0, t.cursorCol-cols)
	}
}

// EscDispatch 处理ESC序列（暂不实现）
func (t *TerminalBuffer) EscDispatch(intermediates []byte, ignore bool, b byte) {
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}