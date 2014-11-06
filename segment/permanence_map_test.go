package segment

import "testing"
import "github.com/dukejeffrie/htm/data"

func TestPermanence(t *testing.T) {
	pm := NewPermanenceMap(64)
	pm.Reset(1, 10, 20, 30, 40, 50, 60)
	pm.Set(10, pm.Config().Minimum)
	if pm.Connected().IsSet(10) {
		t.Errorf("Should have disconnected bit 10: %v", *pm)
	}
	pm.Set(8, pm.Config().Threshold)
	if !pm.Connected().IsSet(8) {
		t.Errorf("Should have connected bit 8: %v", *pm)
	}
	input := data.NewBitset(64).Set(1, 10, 20, 21, 22, 23)
	if pm.Overlap(*input, false) != 2 {
		t.Errorf("Bad strong overlap with input=%v: %v", *input, *pm)
	}
	if pm.Overlap(*input, true) != 3 {
		t.Errorf("Bad weak overlap with input=%v: %v", *input, *pm)
	}
}

func TestNarrowSynapses(t *testing.T) {
	connections := []int{1, 3, 5, 8, 13}
	pm := NewPermanenceMap(64)
	pm.Reset(connections...)
	input := data.NewBitset(64)
	input.Set(1, 5, 22)
	pm.narrow(*input)
	pm.narrow(*input)
	t.Log(pm.permanence)
	if pm.permanence[1] == pm.permanence[3] {
		t.Errorf("Permanence scores did not improve: %v", pm.permanence)
	}
	if pm.permanence[1] != pm.permanence[5] {
		t.Errorf("Permanence scores must be uniform: %v", pm.permanence)
	}
	if pm.permanence[22] != 0 {
		t.Errorf("Permanence for non-connected should be zero: %v", pm.permanence)
	}
	if pm.Connected().NumSetBits() != 2 {
		t.Errorf("Should have kept only 2 connections: %v", pm.permanence)
	}
}

func BenchmarkNarrowSynapses(b *testing.B) {
	pm := NewPermanenceMap(64)
	pm.Reset(1, 3, 5, 8, 13)
	input := data.NewBitset(64)
	input.Set(1, 5, 22)
	for i := 0; i < b.N; i++ {
		pm.narrow(*input)
	}
}

func TestWeakenSynapses(t *testing.T) {
	pm := NewPermanenceMap(64)
	pm.Reset(1, 10, 20)
	pm.Set(30, pm.Config().Threshold+pm.Config().Decrement)

	input := data.NewBitset(64).Set(10, 30)
	pm.weaken(*input)
	if pm.Get(10) != pm.Config().Initial-pm.Config().Decrement {
		t.Errorf("Bad permanence value @%d after weaken: %v", 10, *pm)
	}
	if pm.Get(30) != pm.Config().Threshold {
		t.Errorf("Bad permanence value @%d after weaken: %v", 30, *pm)
	}
	if !pm.Connected().IsSet(30) {
		t.Error("Should be connected @30:", *pm)
	}
	pm.weaken(*input)
	if pm.Connected().IsSet(30) {
		t.Error("Should not be connected @30:", *pm)
	}
}
