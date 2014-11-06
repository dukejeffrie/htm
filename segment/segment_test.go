package segment

import "testing"
import "github.com/dukejeffrie/htm/data"

func TestNewDendriteSegment(t *testing.T) {
	connections := []int{1, 3, 5, 8, 13}
	s := NewDendriteSegment(64)
	s.Reset(connections...)
	for _, v := range connections {
		if !s.Connected().IsSet(v) {
			t.Errorf("Should be connected @%d: %v", v, s.Connected())
		}
		if s.Get(v) != s.Config().Initial {
			t.Errorf("Initial permanence should have been set for connection %v (actual=%v)", v, s.Get(v))
		}
	}
}

func TestBroadenSynapses(t *testing.T) {
	ds := NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13)
	input := data.NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < 1000 && ds.permanence[3] >= ds.Config().Minimum; i++ {
		ds.narrow(*input)
		t.Log(ds)
	}
	if ds.permanence[22] != 0 {
		t.Errorf("Permanence for non-connected should be zero: %v", ds.permanence)
	}
	input.Reset()
	input.Set(1, 8, 22)
	ds.broaden(*input, 0)
	t.Log(ds)
	if ds.permanence[1] <= ds.permanence[3] {
		t.Errorf("Permanence scores did not improve: %v", ds.permanence)
	}
	if ds.permanence[1] != ds.permanence[5] {
		t.Errorf("Permanence scores must be uniform: %v", ds.permanence)
	}
	if ds.permanence[8] != ds.Config().Minimum || ds.permanence[22] != ds.Config().Minimum {
		t.Errorf("Permanence for broadened synapse should be %f: %v", ds.Config().Minimum, ds.permanence)
	}
}

func BenchmarkBroadenSynapses(b *testing.B) {
	ds := NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13)
	input := data.NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < b.N; i++ {
		ds.broaden(*input, 0)
	}
}

func TestDistalSegmentGroup(t *testing.T) {
	group := NewDistalSegmentGroup()
	v1 := data.NewBitset(64).Set(1, 10)
	v2 := data.NewBitset(64).Set(2, 20)
	u1 := group.CreateUpdate(-1, *v1, 2)
	u2 := group.CreateUpdate(-1, *v2, 2)
	group.AddUpdate(u1)
	group.AddUpdate(u2)

	active := data.NewBitset(64).Set(10)

	if group.HasActiveSegment(*active, 1) {
		t.Errorf("Test is broken, group should be empty. %v", *group)
	}
	group.ApplyAll(true)
	if !group.HasActiveSegment(*active, 1) {
		t.Errorf("Expected at least one active segment: %v", *group)
	}
	sIndex, sOverlap := group.ComputeActive(*active, 1, false)
	if sIndex != 0 {
		t.Errorf("Unexpected active segment. Expected: %d, but got: %d. %v", 0, sIndex, *group)
	}
	if sOverlap != 1 {
		t.Errorf("Unexpected overlap. Expected: %d, but got: %d. %v", 1, sOverlap, *group)
	}

	// Disconnect bits 2 and 20 in segment @1.
	config := group.segments[1].Config()
	group.segments[1].Set(2, config.Minimum)
	group.segments[1].Set(20, config.Minimum)

	// Make active state weakly better connected to @1 but strongly to @0.
	active.Reset().Set(1, 2, 20)

	sIndex, _ = group.ComputeActive(*active, 1, true)
	if sIndex != 1 {
		t.Errorf("Unexpected active segment. Expected: %d, but got: %d. %v", 0, sIndex, *group)
	}
	sIndex, _ = group.ComputeActive(*active, 1, false)
	if sIndex != 0 {
		t.Errorf("Unexpected active segment. Expected: %d, but got: %d. %v", 0, sIndex, *group)
	}
}
