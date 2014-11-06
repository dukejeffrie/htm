package htm

import "testing"
import "github.com/dukejeffrie/htm/data"

func BenchmarkColumnOverlap(b *testing.B) {
	c := NewColumn(64, 1)
	connections := []int{1, 3, 5, 8, 11}
	c.ResetConnections(connections)
	input := data.NewBitset(64).Set(1, 5, 22)
	result := data.NewBitset(64).Set(1, 5, 22, 60)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.Overlap(*input)
	}
}

func TestNewColumn(t *testing.T) {
	c := NewColumn(64, 4)
	if c.Height() != 4 {
		t.Errorf("Column has the wrong height: %d", c.Height())
	}
	if c.active.Len() != 4 {
		t.Errorf("Column creation failed: %+v", *c)
	}
	if c.predictive.Len() != 4 {
		t.Errorf("Column creation failed: %+v", *c)
	}
	if c.proximal.Len() != 64 {
		t.Errorf("Proximal dendrite creation failed: %v", *c.proximal)
	}
	if c.learning != -1 {
		t.Errorf("Learning cell should be undefined, but was %v.", c.learning)
	}
	if c.learningTarget != 0 {
		t.Errorf("Learning target should be 0, but was %v.", c.learningTarget)
	}
}

func TestPredict(t *testing.T) {
	c := NewColumn(64, 4)
	active1 := data.NewBitset(64).Set(2, 20)
	active2 := data.NewBitset(64).Set(11, 31)
	update := c.distal[1].CreateUpdate(-1, *active1, 2)
	c.distal[1].Apply(update, true)
	update = c.distal[2].CreateUpdate(-1, *active2, 2)
	c.distal[2].Apply(update, true)

	state := data.NewBitset(64).Set(2)
	c.Predict(*state, 1)
	pred := c.Predictive()
	if !pred.AllSet(1) || pred.NumSetBits() != 1 {
		t.Errorf("Should have predicted only cell 1, but got %v", pred)
		t.Log(*c.distal[1])
	}

	state.Set(11)
	c.Predict(*state, 1)
	pred = c.Predictive()
	if !pred.AllSet(1, 2) || pred.NumSetBits() != 2 {
		t.Errorf("Should have predicted cells 1 and 2, but got %v", pred)
		t.Log(*c.distal[1])
		t.Log(*c.distal[2])
	}
}

func TestFindBestSegment(t *testing.T) {
	c := NewColumn(64, 4)
	active1 := data.NewBitset(64).Set(2, 20, 22)
	active2 := data.NewBitset(64).Set(11, 31)
	update := c.distal[1].CreateUpdate(-1, *active1, 2)
	c.distal[1].Apply(update, true)
	update = c.distal[2].CreateUpdate(-1, *active2, 2)
	c.distal[2].Apply(update, true)

	state := data.NewBitset(64).Set(2)
	cell, s, overlap := c.FindBestSegment(*state, 1, false)
	if cell != 1 || s != 0 || overlap != 1 {
		t.Errorf("FindBestSegment(%v,minOverlap=%d): %d, %v, %d", *state, 1, cell, s, overlap)
	}
	cell, s, overlap = c.FindBestSegment(*state, 2, false)
	if cell != -1 || s != -1 {
		t.Errorf("FindBestSegment(%v,minOverlap=%d): %d, %v, %d", *state, 2, cell, s, overlap)
	}
	state.Set(22)
	cell, s, overlap = c.FindBestSegment(*state, 2, false)
	if cell != 1 || s != 0 || overlap != 2 {
		t.Errorf("FindBestSegment(%v,minOverlap=%d): %d, %v, %d", *state, 2, cell, s, overlap)
	}
	state.Reset().Set(2, 20, 11, 31)
}
