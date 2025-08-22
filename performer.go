package govte

// Performer defines the interface for handling parser actions.
// This is the Go equivalent of the Rust Perform trait.
type Performer interface {
	// Print draws a character to the screen and updates states.
	Print(c rune)

	// Execute executes a C0 or C1 control function.
	Execute(b byte)

	// Hook is invoked when a final character arrives in first part of device control string.
	// The control function should be determined from the private marker, final character,
	// and executed with a parameter list.
	Hook(params *Params, intermediates []byte, ignore bool, action rune)

	// Put passes bytes as part of a device control string to the handler chosen in Hook.
	// C0 controls will also be passed to the handler.
	Put(b byte)

	// Unhook is called when a device control string is terminated.
	// The previously selected handler should be notified that the DCS has terminated.
	Unhook()

	// OscDispatch dispatches an operating system command.
	OscDispatch(params [][]byte, bellTerminated bool)

	// CsiDispatch is called when a final character has arrived for a CSI sequence.
	// The ignore flag indicates that either more than two intermediates arrived
	// or the number of parameters exceeded the maximum supported length.
	CsiDispatch(params *Params, intermediates []byte, ignore bool, action rune)

	// EscDispatch is called when the final character of an escape sequence has arrived.
	// The ignore flag indicates that more than two intermediates arrived.
	EscDispatch(intermediates []byte, ignore bool, b byte)
}

// NoopPerformer is a no-op implementation of Performer interface.
// It can be embedded in custom implementations to avoid implementing all methods.
type NoopPerformer struct{}

// Print implements Performer
func (n *NoopPerformer) Print(c rune) {}

// Execute implements Performer
func (n *NoopPerformer) Execute(b byte) {}

// Hook implements Performer
func (n *NoopPerformer) Hook(params *Params, intermediates []byte, ignore bool, action rune) {}

// Put implements Performer
func (n *NoopPerformer) Put(b byte) {}

// Unhook implements Performer
func (n *NoopPerformer) Unhook() {}

// OscDispatch implements Performer
func (n *NoopPerformer) OscDispatch(params [][]byte, bellTerminated bool) {}

// CsiDispatch implements Performer
func (n *NoopPerformer) CsiDispatch(params *Params, intermediates []byte, ignore bool, action rune) {}

// EscDispatch implements Performer
func (n *NoopPerformer) EscDispatch(intermediates []byte, ignore bool, b byte) {}

// Ensure NoopPerformer implements Performer
var _ Performer = (*NoopPerformer)(nil)