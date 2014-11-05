package htm

import "fmt"
import "io"
import "testing"

func BenchmarkOverlap(b *testing.B) {
	c := NewColumn(1)
	connections := []int{1, 3, 5, 8, 11}
	c.ResetConnections(64, connections)
	input := NewBitset(64)
	input.Set(1, 5, 22)
	result := NewBitset(64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result.ResetTo(c.Connected())
		result.And(*input)
	}
}

func TestPrintColumn(t *testing.T) {
	col := NewColumn(5)
	col.ResetConnections(64, []int{1, 10, 11, 20})
	col.predicted.Set(3, 4)
	col.active.Set(1, 3)
	reader, writer := io.Pipe()

	go func() {
		werr := col.Print(col.Height(), writer)
		if werr != nil {
			t.Error(werr)
		}
		writer.Close()
	}()

	var s1 string
	if _, rerr := fmt.Fscan(reader, &s1); rerr != nil {
		t.Error(rerr)
	}
	expected := "-!-xo"
	if s1 != expected {
		t.Errorf("Print doesn't match: %v != %v", expected, s1)
	}
}
