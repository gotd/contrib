package sequence

// Helps to understand what happens inside a box during execution.
// Used for testing.
type hooks interface {
	Transition(from, to mode)
	FastgapBegin()
	FastgapCollectorFinished()
	FastgapEnd()
	Apply(state int, updates []interface{})
}

type nopHooks struct{}

func (nopHooks) Transition(from, to mode)  {}
func (nopHooks) FastgapBegin()             {}
func (nopHooks) FastgapCollectorFinished() {}
func (nopHooks) FastgapEnd()               {}
func (nopHooks) Apply(int, []interface{})  {}
