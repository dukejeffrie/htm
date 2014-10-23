package htm

import "testing"

func TestMinOverlap(t *testing.T) {
	l := NewRegion("Single Region", 100, 4, 0.02)
	columnRand.Seed(0)
	l.ResetForInput(2048, 20)

	input := NewBitset(2048)
	input.SetRange(0, 20)

	l.ConsumeInput(*input)
	output := l.Output()

	if output.NumSetBits() == 0 {
		t.Errorf("Output is empty: %v", output)
	}
	t.Log(output)

	l.MinOverlap = 4
	l.ConsumeInput(*input)
	output = l.Output()
	if output.NumSetBits() != 0 {
		t.Errorf("Output should be empty: %v", output)
	}
}

func TestConsumeInput(t *testing.T) {
	// 50 columns with 4 cells each, firing 2% of columns
	l := NewRegion("Single Region", 50, 4, 0.02)

	// 64-bit input, 2 bits of real data.
	columnRand.Seed(1)
	l.ResetForInput(64, 2)

	input := NewBitset(64)
	input.SetRange(0, 1)

	l.ConsumeInput(*input)

	output := l.Output()
	if output.Len() != 50*4 {
		t.Errorf("Weird output length (expected %d, got: %d).", 50*4, output.Len())
	}
	if output.NumSetBits() == 0 {
		t.Errorf("Output is empty: %v", output)
	}
	t.Log(output)

	// Let's store all the top cells connections into the input bit, so we get them again the second time around.
	next_input := NewBitset(64)

	for _, el := range l.scratch.scores {
		col := l.columns[el.index]
		next_input.Or(col.Connected())
	}
	last_scores := make([]ScoredElement, l.scratch.scores.Len())
	copy(last_scores, l.scratch.scores)

	l.ConsumeInput(*NewBitset(64))
	if l.scratch.scores.Len() != 0 {
		t.Errorf("Leftovers in scratch: %v", l.scratch)
	}
	output2 := l.Output()
	if output2.Len() != 50*4 {
		t.Errorf("Weird output length (expected %d, got: %d).", 50*4, output2.Len())
	}
	if output2.NumSetBits() != 0 {
		t.Errorf("Empty input caused non-empty output: %v", output2)
	}

	l.ConsumeInput(*next_input)
	for _, old := range last_scores {
		found := false
		for _, el := range l.scratch.scores {
			if old.index == el.index {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Column %d was expected, but is missing!\n%v",
				old.index, l.scratch.scores)
		}
	}

	output3 := l.Output()
	if output3.NumSetBits() == 0 {
		t.Errorf("Output is empty: %v", output3)
	}
	if output.NumSetBits() != output3.NumSetBits() {
		t.Errorf("Output density very different after learning (expected: %d, but got %d).", output.NumSetBits(), output3.NumSetBits())
	}
	match := output3.Clone()
	match.And(output)
	if m := match.NumSetBits(); m > output.NumSetBits()+2 || m < output.NumSetBits()-2 {
		// This is not a real error, but we leave it here to draw attention.
		t.Errorf("Output slightly very different after learning.")
	}
	t.Log(output3)
}

func BenchmarkConsumeInput(b *testing.B) {
	l := NewRegion("Single Region", 500, 4, 0.02)
	l.Learning = false
	l.ResetForInput(2048, 28)

	input := NewBitset(2048)
	input.Set(columnRand.Perm(2048)[0:28])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.ConsumeInput(*input)
	}
}
