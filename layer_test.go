package htm

import "testing"

func TestConsumeInput(t *testing.T) {
	// 50 columns with 4 cells each, firing 2% of columns
	l := NewLayer("Single Layer", 50, 4, 0.02)

	// 64-bit input, 2 bits of real data.
	l.ResetForInput(64, 2)

	input := NewBitset(64)
	input.SetRange(0, 1)

	l.ConsumeInput(*input)

	output := l.Output()
	if output.Len() != 50*4 {
		t.Errorf("Weird output length (expected %d, got: %d).", 50*4, output.Len())
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
	l := NewLayer("Single Layer", 500, 4, 0.02)
	l.Learning = false
	l.ResetForInput(2048, 28)

	input := NewBitset(2048)
	input.Set(columnRand.Perm(2048)[0:28])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.ConsumeInput(*input)
	}
}

func BenchmarkProduceOutput(b *testing.B) {
	l := NewLayer("Single Layer", 500, 4, 0.02)
	l.Learning = false
	l.ResetForInput(2048, 28)

	input := NewBitset(2048)
	input.Set(columnRand.Perm(2048)[0:28])
	l.ConsumeInput(*input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Output()
		l.output.Truncate(0)
	}
}
