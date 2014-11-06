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
