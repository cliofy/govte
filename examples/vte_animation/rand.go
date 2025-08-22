package main

import (
	"sync"
	"time"
)

// Simple linear congruential random number generator
// Similar to Rust version, avoids introducing extra dependencies
type SimpleRand struct {
	seed uint64
	mu   sync.Mutex
}

// Global random number generator instance
var globalRand = &SimpleRand{
	seed: uint64(time.Now().UnixNano()),
}

// Random random number generator interface
type Random interface {
	Random() interface{}
}

// RandomInt generates random integer
func RandomInt() int {
	return int(RandomUint64())
}

// RandomUint64 generates random uint64
func RandomUint64() uint64 {
	globalRand.mu.Lock()
	defer globalRand.mu.Unlock()
	
	// Use linear congruential generator (LCG)
	// Parameters from Numerical Recipes
	globalRand.seed = (globalRand.seed*1664525 + 1013904223) & 0xFFFFFFFF
	
	// Use xorshift to increase randomness
	x := globalRand.seed
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	
	globalRand.seed = x
	return x
}

// RandomFloat32 generates random float between 0.0 and 1.0
func RandomFloat32() float32 {
	return float32(RandomUint64()) / float32(^uint64(0))
}

// RandomRange generates random integer in range [min, max)
func RandomRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + int(RandomUint64())%(max-min)
}

// RandomBool generates random boolean
func RandomBool() bool {
	return RandomUint64()%2 == 0
}

// RandomChoice randomly selects an element from slice
func RandomChoice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	idx := RandomRange(0, len(slice))
	return slice[idx]
}

// RandomString generates random string
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

// Shuffle randomly shuffles slice
func Shuffle[T any](slice []T) {
	for i := len(slice) - 1; i > 0; i-- {
		j := RandomRange(0, i+1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// SetSeed sets random seed
func SetSeed(seed uint64) {
	globalRand.mu.Lock()
	defer globalRand.mu.Unlock()
	globalRand.seed = seed
}

// Random functions for specific purposes

// RandomMatrixChar generates random character for matrix rain effect
func RandomMatrixChar() rune {
	// Numbers and some Japanese katakana characters (like in The Matrix movie)
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

// RandomASCII generates printable ASCII character
func RandomASCII() rune {
	return rune(RandomRange(33, 127)) // Printable ASCII range
}

// RandomEmoji generates random emoji (for demo)
func RandomEmoji() string {
	emojis := []string{
		"🌟", "⭐", "💫", "✨", "🔥", "💎", "🎯", "🚀",
		"🎉", "🎊", "🎈", "🎁", "🏆", "🥇", "🎖️", "🏅",
		"❤️", "💙", "💚", "💛", "💜", "🧡", "🖤", "🤍",
	}
	return RandomChoice(emojis)
}