package data

// A moving window history tracker for boolean events. If you create a cycle of
// length N, the (N+1)th event will overwrite the 0th event. It is memory-efficient,
// requiring length/64+8 bytes of memory, and relatively fast to count the true
// events by using a Bitset underneath.
type CycleHistory struct {
	events *Bitset
	cycle  int
}

// Creates a new history object with the given history length.
func NewCycleHistory(length int) *CycleHistory {
	result := &CycleHistory{
		events: NewBitset(length),
		cycle:  -length,
	}
	return result
}

// Records a new event.
func (ch *CycleHistory) Record(event bool) {
	at := ch.cycle
	if at < 0 {
		at = ch.events.Len() + at
	}
	if event {
		ch.events.Set(at)
	} else {
		ch.events.Unset(at)
	}
	ch.cycle = (ch.cycle + 1) % ch.events.Len()
}

// Returns the average count of true occurrences in the last window, along with a
// boolean that says whether the value can be used, to be used with the ok idiom:
//
// if val, ok := cycle.Average(); ok && somePredicate(val) { ... }
//
// Currently, the only reason for ok to be false is when no event has ever been
// recorded.
func (ch CycleHistory) Average() (result float32, ok bool) {
	l := ch.events.Len()
	if ch.cycle < 0 {
		l += ch.cycle
	}
	if l == 0 {
		return
	}
	ok = true
	result = float32(ch.events.DenseCount()) / float32(l)
	return
}
