package htm

import "fmt"
import "testing"

func TestConsumeInput(t *testing.T) {
	// 500 columns with 4 cells each.
	l := NewLayer("Single Layer", 50, 4)

	// 64-bit scalar input, 2 bits of real data.
	l.ResetForInput(64, 2)

	input := NewBitset(64)
	input.SetRange(0, 2)

	l.ConsumeInput(input)

	// Let's store all the top cells connections into the input bit, so we get them again the second time around.
	next_input := NewBitset(64)

	fmt.Printf("--- For input: %v", input)
	for _, el := range l.scratch.scores {
		col := l.columns[el.index]
		fmt.Printf("\n\t@%d(score=%d): %v", el.index, el.score, col)
		next_input.Or(col.Connected())
	}
	fmt.Println()
	last_scores := make([]ScoredElement, l.scratch.scores.Len())
	copy(last_scores, l.scratch.scores)

	l.ConsumeInput(NewBitset(64))
	if l.scratch.scores.Len() != 0 {
		t.Errorf("Leftovers in scratch: %v", l.scratch)
	}

	l.ConsumeInput(next_input)
	fmt.Printf("--- For input: %v", next_input)
	for _, el := range l.scratch.scores {
		col := l.columns[el.index]
		fmt.Printf("\n\t@%d(score=%d): %v", el.index, el.score, col)
	}
	fmt.Println()
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
}

func BenchmarkConsumeInput(b *testing.B) {
	// 500 columns with 4 cells each.
	l := NewLayer("Single Layer", 500, 4)

	// 2048-bit scalar input, 28 bits of real data.
	l.ResetForInput(2048, 28)

	input := NewBitset(2048)
	input.Set(columnRand.Perm(2048)[0:28])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.ConsumeInput(input)
	}
}
