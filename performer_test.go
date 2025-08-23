package govte

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockPerformer is a test implementation of the Performer interface
type MockPerformer struct {
	printed       []rune
	executed      []byte
	csiDispatched []CSIDispatch
	escDispatched []ESCDispatch
	oscDispatched []OSCDispatch
	hookCalled    bool
	unhookCalled  bool
	putBytes      []byte
}

type CSIDispatch struct {
	params        *Params
	intermediates []byte
	ignore        bool
	action        rune
}

type ESCDispatch struct {
	intermediates []byte
	ignore        bool
	b             byte
}

type OSCDispatch struct {
	params         [][]byte
	bellTerminated bool
}

func (m *MockPerformer) Print(c rune) {
	m.printed = append(m.printed, c)
}

func (m *MockPerformer) Execute(b byte) {
	m.executed = append(m.executed, b)
}

func (m *MockPerformer) Hook(params *Params, intermediates []byte, ignore bool, action rune) {
	m.hookCalled = true
}

func (m *MockPerformer) Put(b byte) {
	m.putBytes = append(m.putBytes, b)
}

func (m *MockPerformer) Unhook() {
	m.unhookCalled = true
}

func (m *MockPerformer) OscDispatch(params [][]byte, bellTerminated bool) {
	m.oscDispatched = append(m.oscDispatched, OSCDispatch{
		params:         params,
		bellTerminated: bellTerminated,
	})
}

func (m *MockPerformer) CsiDispatch(params *Params, intermediates []byte, ignore bool, action rune) {
	// Make a copy of params to avoid reference issues
	paramsCopy := &Params{}
	if params != nil {
		// Copy the params data
		*paramsCopy = *params
	}

	m.csiDispatched = append(m.csiDispatched, CSIDispatch{
		params:        paramsCopy,
		intermediates: append([]byte(nil), intermediates...), // Copy intermediates too
		ignore:        ignore,
		action:        action,
	})
}

func (m *MockPerformer) EscDispatch(intermediates []byte, ignore bool, b byte) {
	m.escDispatched = append(m.escDispatched, ESCDispatch{
		intermediates: intermediates,
		ignore:        ignore,
		b:             b,
	})
}

func TestPerformerInterface(t *testing.T) {
	// 验证 MockPerformer 实现了 Performer 接口
	var _ Performer = (*MockPerformer)(nil)

	mock := &MockPerformer{}

	// 测试 Print
	mock.Print('A')
	mock.Print('B')
	assert.Equal(t, []rune{'A', 'B'}, mock.printed)

	// 测试 Execute
	mock.Execute(0x08) // Backspace
	mock.Execute(0x0A) // Line Feed
	assert.Equal(t, []byte{0x08, 0x0A}, mock.executed)

	// 测试 Hook 和 Unhook
	mock.Hook(nil, nil, false, 'p')
	assert.True(t, mock.hookCalled)

	mock.Unhook()
	assert.True(t, mock.unhookCalled)

	// 测试 Put
	mock.Put('x')
	mock.Put('y')
	assert.Equal(t, []byte{'x', 'y'}, mock.putBytes)

	// 测试 OscDispatch
	mock.OscDispatch([][]byte{[]byte("test")}, false)
	assert.Len(t, mock.oscDispatched, 1)
	assert.Equal(t, [][]byte{[]byte("test")}, mock.oscDispatched[0].params)
	assert.False(t, mock.oscDispatched[0].bellTerminated)

	// 测试 CsiDispatch
	params := &Params{}
	mock.CsiDispatch(params, []byte{}, false, 'H')
	assert.Len(t, mock.csiDispatched, 1)
	assert.Equal(t, 'H', mock.csiDispatched[0].action)

	// 测试 EscDispatch
	mock.EscDispatch([]byte{}, false, 'M')
	assert.Len(t, mock.escDispatched, 1)
	assert.Equal(t, byte('M'), mock.escDispatched[0].b)
}

func TestNoopPerformer(t *testing.T) {
	// 测试空实现（默认实现）
	noop := &NoopPerformer{}

	// 这些调用不应该 panic
	noop.Print('A')
	noop.Execute(0x08)
	noop.Hook(nil, nil, false, 'p')
	noop.Put('x')
	noop.Unhook()
	noop.OscDispatch(nil, false)
	noop.CsiDispatch(nil, nil, false, 'H')
	noop.EscDispatch(nil, false, 'M')

	// 测试通过意味着所有方法都可以安全调用
	assert.True(t, true, "NoopPerformer should not panic")
}
