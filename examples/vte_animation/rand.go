package main

import (
	"sync"
	"time"
)

// 简单的线性同余随机数生成器
// 类似Rust版本，避免引入额外依赖
type SimpleRand struct {
	seed uint64
	mu   sync.Mutex
}

// 全局随机数生成器实例
var globalRand = &SimpleRand{
	seed: uint64(time.Now().UnixNano()),
}

// Random 随机数生成器接口
type Random interface {
	Random() interface{}
}

// RandomInt 生成随机整数
func RandomInt() int {
	return int(RandomUint64())
}

// RandomUint64 生成随机uint64
func RandomUint64() uint64 {
	globalRand.mu.Lock()
	defer globalRand.mu.Unlock()
	
	// 使用线性同余生成器 (LCG)
	// 参数来自Numerical Recipes
	globalRand.seed = (globalRand.seed*1664525 + 1013904223) & 0xFFFFFFFF
	
	// 使用xorshift增加随机性
	x := globalRand.seed
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	
	globalRand.seed = x
	return x
}

// RandomFloat32 生成0.0到1.0之间的随机浮点数
func RandomFloat32() float32 {
	return float32(RandomUint64()) / float32(^uint64(0))
}

// RandomRange 生成指定范围内的随机整数 [min, max)
func RandomRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + int(RandomUint64())%(max-min)
}

// RandomBool 生成随机布尔值
func RandomBool() bool {
	return RandomUint64()%2 == 0
}

// RandomChoice 从切片中随机选择一个元素
func RandomChoice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	idx := RandomRange(0, len(slice))
	return slice[idx]
}

// RandomString 生成随机字符串
func RandomString(chars string, length int) string {
	if length <= 0 || len(chars) == 0 {
		return ""
	}
	
	result := make([]byte, length)
	charRunes := []rune(chars)
	
	for i := 0; i < length; i++ {
		idx := RandomRange(0, len(charRunes))
		result[i] = byte(charRunes[idx])
	}
	
	return string(result)
}

// Shuffle 随机打乱切片
func Shuffle[T any](slice []T) {
	for i := len(slice) - 1; i > 0; i-- {
		j := RandomRange(0, i+1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// SetSeed 设置随机数种子
func SetSeed(seed uint64) {
	globalRand.mu.Lock()
	defer globalRand.mu.Unlock()
	globalRand.seed = seed
}

// 特定用途的随机函数

// RandomMatrixChar 生成矩阵雨效果的随机字符
func RandomMatrixChar() rune {
	// 数字和一些日文片假名字符（类似电影《黑客帝国》）
	chars := []rune{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'ア', 'イ', 'ウ', 'エ', 'オ',
		'カ', 'キ', 'ク', 'ケ', 'コ',
		'サ', 'シ', 'ス', 'セ', 'ソ',
		'タ', 'チ', 'ツ', 'テ', 'ト',
		'ナ', 'ニ', 'ヌ', 'ネ', 'ノ',
	}
	return RandomChoice(chars)
}

// RandomASCII 生成可打印ASCII字符
func RandomASCII() rune {
	return rune(RandomRange(33, 127)) // 可打印ASCII范围
}

// RandomEmoji 生成随机emoji（用于演示）
func RandomEmoji() string {
	emojis := []string{
		"🌟", "⭐", "💫", "✨", "🔥", "💎", "🎯", "🚀",
		"🎉", "🎊", "🎈", "🎁", "🏆", "🥇", "🎖️", "🏅",
		"❤️", "💙", "💚", "💛", "💜", "🧡", "🖤", "🤍",
	}
	return RandomChoice(emojis)
}