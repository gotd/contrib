package sequence

// Update struct.
type Update struct {
	// Pts, PtsCount
	// Qts, 1
	// Seq, (Seq-SeqStart+1)
	State, Count int
	Value        interface{}
}

func (u Update) start() int {
	return u.State - u.Count + 1
}

func (u Update) end() int {
	return u.State
}
