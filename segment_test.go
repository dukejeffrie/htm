package htm

import "testing"

func TestNewDendriteSegment(t *testing.T) {
	connections := []int{1, 3, 5, 8, 13}
	s := NewDendriteSegment(64, connections)
	connected := s.Connected().ToIndexes(make([]int, len(connections)))
	for i, v := range connections {
		if connected[i] != v {
			t.Errorf("Connection mismatch: %v != %v", v, connected[i])
		}
		if s.permanence[v] != INITIAL_PERMANENCE {
			t.Errorf("Initial permanence should have been set for connection %v", v)
		}
	}
}

func TestLearnSynapses(t *testing.T) {
	connections := []int{1, 3, 5, 8, 13}
	ds := NewDendriteSegment(64, connections)
	input := NewBitset(64)
	input.Set([]int{1, 5, 22})
	ds.Learn(*input)
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
