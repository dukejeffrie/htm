package htm

import "testing"

func TestNewDendriteSegment(t *testing.T) {
	connections := []int{1, 3, 5, 8, 13}
	s := NewDendriteSegment(64)
	s.Reset(connections...)
	for _, v := range connections {
		if !s.Connected().IsSet(v) {
			t.Errorf("Should be connected @%d: %v", v, s.Connected())
		}
		if s.Get(v) != INITIAL_PERMANENCE {
			t.Errorf("Initial permanence should have been set for connection %v (actual=%v)", v, s.Get(v))
		}
	}
}

func TestPermanence(t *testing.T) {
	m := NewPermanenceMap(64)
	m.Reset(1, 10, 20, 30, 40, 50, 60)
	m.Set(10, PERMANENCE_MIN)
	if m.Connected().IsSet(10) {
		t.Errorf("Should have disconnected bit 10: %v", *m)
	}
	m.Set(8, CONNECTION_THRESHOLD)
	if !m.Connected().IsSet(8) {
		t.Errorf("Should have connected bit 8: %v", *m)
	}
	input := NewBitset(64).Set(1, 10, 20, 21, 22, 23)
	if m.Overlap(*input, false) != 2 {
		t.Errorf("Bad strong overlap with input=%v: %v", *input, *m)
	}
	if m.Overlap(*input, true) != 3 {
		t.Errorf("Bad weak overlap with input=%v: %v", *input, *m)
	}
}

func TestNarrowSynapses(t *testing.T) {
	connections := []int{1, 3, 5, 8, 13}
	ds := NewPermanenceMap(64)
	ds.Reset(connections...)
	input := NewBitset(64)
	input.Set(1, 5, 22)
	ds.narrow(*input)
	ds.narrow(*input)
	t.Log(ds.permanence)
	if ds.permanence[1] == ds.permanence[3] {
		t.Errorf("Permanence scores did not improve: %v", ds.permanence)
	}
	if ds.permanence[1] != ds.permanence[5] {
		t.Errorf("Permanence scores must be uniform: %v", ds.permanence)
	}
	if ds.permanence[22] != 0 {
		t.Errorf("Permanence for non-connected should be zero: %v", ds.permanence)
	}
	if ds.Connected().NumSetBits() != 2 {
		t.Errorf("Should have kept only 2 connections: %v", ds.permanence)
	}
}

func BenchmarkNarrowSynapses(b *testing.B) {
	ds := NewPermanenceMap(64)
	ds.Reset(1, 3, 5, 8, 13)
	input := NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < b.N; i++ {
		ds.narrow(*input)
	}
}

func TestWeakenSynapses(t *testing.T) {
	pm := NewPermanenceMap(64)
	pm.Reset(1, 10, 20)
	pm.Set(30, CONNECTION_THRESHOLD+PERMANENCE_DEC)

	input := NewBitset(64).Set(10, 30)
	pm.weaken(*input)
	if pm.Get(10) != INITIAL_PERMANENCE-PERMANENCE_DEC {
		t.Errorf("Bad permanence value @%d after weaken: %v", 10, *pm)
	}
	if pm.Get(30) != CONNECTION_THRESHOLD {
		t.Errorf("Bad permanence value @%d after weaken: %v", 30, *pm)
	}
	if !pm.Connected().IsSet(30) {
		t.Errorf("Should be connected @30:", *pm)
	}
	pm.weaken(*input)
	if pm.Connected().IsSet(30) {
		t.Errorf("Should not be connected @30:", *pm)
	}
}

func TestBroadenSynapses(t *testing.T) {
	ds := NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13)
	input := NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < 1000 && ds.permanence[3] >= PERMANENCE_MIN; i++ {
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
	if ds.permanence[8] != PERMANENCE_MIN || ds.permanence[22] != PERMANENCE_MIN {
		t.Errorf("Permanence for broadened synapse should be %d: %v", PERMANENCE_MIN, ds.permanence)
	}
}

func BenchmarkBroadenSynapses(b *testing.B) {
	ds := NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13)
	input := NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < b.N; i++ {
		ds.broaden(*input, 0)
	}
}

func TestTrim(t *testing.T) {
	ds := NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13)
	input := NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < 10 && ds.permanence[5] >= CONNECTION_THRESHOLD; i++ {
		ds.narrow(*input)
	}
	if len(ds.permanence) > 2 {
		t.Errorf("Should have trimmed synaptic inputs: %v", ds)
	}
	if len(ds.permanence) != ds.Connected().NumSetBits() {
		t.Errorf("Should have updated synapses: %v", ds)
	}
	t.Log(ds.permanence, ds.Connected())
}

func TestCycleHistory(t *testing.T) {
	ch := NewCycleHistory(10)
	if avg, ok := ch.Average(); ok {
		t.Errorf("Should not be ok: %f", avg)
	}
	ch.Record(true)
	if avg, ok := ch.Average(); !ok || avg != 1.0 {
		t.Errorf("Should be %f average: %f, ok=%t", 1.0, avg, ok)
	}
	ch.Record(false)
	if avg, ok := ch.Average(); !ok || avg != 0.5 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.5, avg, ok, ch)
	}
	for i := 2; i < 10; i++ {
		ch.Record(false)
	}
	if avg, ok := ch.Average(); !ok || avg != 0.1 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.1, avg, ok, ch)
	}
	ch.Record(false)
	if avg, ok := ch.Average(); !ok || avg != 0.0 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.0, avg, ok, ch)
	}
	ch.Record(false)
	ch.Record(false)
	ch.Record(false)
	ch.Record(true)
	for i := 1; i < 10; i++ {
		ch.Record(false)
	}
	if avg, ok := ch.Average(); !ok || avg != 0.1 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.1, avg, ok, ch)
	}
}

func TestDistalSegmentGroup(t *testing.T) {
	group := NewDistalSegmentGroup()
	v1 := NewBitset(64).Set(1, 10)
	v2 := NewBitset(64).Set(2, 20)
	u1 := group.CreateUpdate(-1, *v1, 2)
	u2 := group.CreateUpdate(-1, *v2, 2)
	group.AddUpdate(u1)
	group.AddUpdate(u2)

	active := NewBitset(64).Set(10)

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
	group.segments[1].Set(2, PERMANENCE_MIN)
	group.segments[1].Set(20, PERMANENCE_MIN)

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
