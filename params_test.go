package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParamsCreation(t *testing.T) {
	params := NewParams()
	assert.NotNil(t, params)
	assert.Equal(t, 0, params.Len())
	assert.True(t, params.IsEmpty())
}

func TestParamsPush(t *testing.T) {
	params := NewParams()

	// 添加单个参数
	params.Push(1)
	assert.Equal(t, 1, params.Len())
	assert.False(t, params.IsEmpty())

	// 添加更多参数
	params.Push(2)
	params.Push(3)
	assert.Equal(t, 3, params.Len())

	// 验证参数值
	iter := params.Iter()
	assert.Equal(t, []uint16{1}, iter[0])
	assert.Equal(t, []uint16{2}, iter[1])
	assert.Equal(t, []uint16{3}, iter[2])
}

func TestParamsSubParams(t *testing.T) {
	params := NewParams()

	// 添加带子参数的参数
	params.Push(1)
	params.Extend(2) // 子参数
	params.Extend(3) // 子参数

	params.Push(4)
	params.Extend(5) // 子参数

	iter := params.Iter()
	assert.Len(t, iter, 2, "Should have 2 main parameters")

	// 第一个参数有3个值（主参数 + 2个子参数）
	assert.Equal(t, []uint16{1, 2, 3}, iter[0])

	// 第二个参数有2个值（主参数 + 1个子参数）
	assert.Equal(t, []uint16{4, 5}, iter[1])
}

func TestParamsClear(t *testing.T) {
	params := NewParams()

	// 添加一些参数
	params.Push(1)
	params.Push(2)
	params.Push(3)
	assert.Equal(t, 3, params.Len())

	// 清除参数
	params.Clear()
	assert.Equal(t, 0, params.Len())
	assert.True(t, params.IsEmpty())
}

func TestParamsMaxCapacity(t *testing.T) {
	params := NewParams()

	// 填充到最大容量
	for i := 0; i < MaxParams; i++ {
		if !params.IsFull() {
			params.Push(uint16(i))
		}
	}

	assert.True(t, params.IsFull())
	assert.Equal(t, MaxParams, params.Len())

	// 尝试添加更多参数（应该被忽略或返回错误）
	// 具体行为取决于实现
}

func TestParamsIterator(t *testing.T) {
	params := NewParams()

	// 设置测试数据
	params.Push(1)
	params.Extend(10)
	params.Extend(100)
	params.Push(2)
	params.Push(3)
	params.Extend(30)

	// 使用迭代器
	iter := params.Iter()
	assert.Len(t, iter, 3)

	// 验证每个参数组
	assert.Equal(t, []uint16{1, 10, 100}, iter[0])
	assert.Equal(t, []uint16{2}, iter[1])
	assert.Equal(t, []uint16{3, 30}, iter[2])
}

func TestParamsString(t *testing.T) {
	params := NewParams()

	params.Push(1)
	params.Push(2)
	params.Extend(20)
	params.Push(3)

	// 测试字符串表示
	str := params.String()
	assert.Contains(t, str, "1")
	assert.Contains(t, str, "2")
	assert.Contains(t, str, "20")
	assert.Contains(t, str, "3")
}

func TestParamsEdgeCases(t *testing.T) {
	t.Run("Empty params iteration", func(t *testing.T) {
		params := NewParams()
		iter := params.Iter()
		assert.Empty(t, iter)
	})

	t.Run("Single param with no subparams", func(t *testing.T) {
		params := NewParams()
		params.Push(42)
		iter := params.Iter()
		assert.Len(t, iter, 1)
		assert.Equal(t, []uint16{42}, iter[0])
	})

	t.Run("Zero values", func(t *testing.T) {
		params := NewParams()
		params.Push(0)
		params.Push(0)
		assert.Equal(t, 2, params.Len())
		iter := params.Iter()
		assert.Equal(t, []uint16{0}, iter[0])
		assert.Equal(t, []uint16{0}, iter[1])
	})

	t.Run("Maximum value", func(t *testing.T) {
		params := NewParams()
		params.Push(65535) // Max uint16
		iter := params.Iter()
		assert.Equal(t, []uint16{65535}, iter[0])
	})
}