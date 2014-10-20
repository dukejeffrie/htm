package htm

import "testing"

func TestReset(t *testing.T) {
	c := NewColumn(1)
	connections := []int{1, 3, 5, 8, 11}
	c.ResetConnections(64, connections)
	connected := c.Connected().ToIndexes(make([]int, len(connections)))
	for i, v := range connections {
		if connected[i] != v {
			t.Errorf("Connection mismatch: %v != %v", v, connected[i])
		}
		if c.permanence[v] != INITIAL_PERMANENCE {
			t.Errorf("Initial permanence should have been set for connection %v", v)
		}
	}
}

func TestOverlap(t *testing.T) {
	c := NewColumn(1)
	connections := []int{1, 3, 5, 8, 11}
	c.ResetConnections(64, connections)
	input := NewBitset(64)
	input.Set([]int{1, 5, 22})
	result := NewBitset(64)
	c.Overlap(*input, result)
	if result.NumSetBits() != 2 {
		t.Errorf("Overlap score for columns 1, 5 should be 2: %v", result)
	}
}

func BenchmarkOverlap(b *testing.B) {
	c := NewColumn(1)
	connections := []int{1, 3, 5, 8, 11}
	c.ResetConnections(64, connections)
	input := NewBitset(64)
	input.Set([]int{1, 5, 22})
	result := NewBitset(64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Overlap(*input, result)
	}
}
