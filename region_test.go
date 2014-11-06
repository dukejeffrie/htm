package htm

import "testing"
import "github.com/dukejeffrie/htm/data"

func TestMinOverlap(t *testing.T) {
	l := NewRegion(RegionParameters{
		Name:                 "Single Region",
		Learning:             true,
		Height:               4,
		Width:                100,
		InputLength:          2048,
		MaximumFiringColumns: 40,
		MinimumInputOverlap:  1,
	})
	columnRand.Seed(0)
	l.RandomizeColumns(20)

	input := data.NewBitset(2048)
	input.SetRange(0, 20)

	l.ConsumeInput(*input)
	output := l.Output()

	if output.IsZero() {
		t.Errorf("Output is empty: %v", output)
	}
	t.Log(output)

	l.MinimumInputOverlap = 4

	l.ConsumeInput(*input)
	output = l.Output()
	if !output.IsZero() {
		t.Errorf("Output should be empty: %v", output)
	}
}

func TestRegionFeedBack(t *testing.T) {
	l := NewRegion(RegionParameters{
		Name:                 "Single Region",
		Learning:             true,
		Height:               4,
		Width:                500,
		InputLength:          64,
		MaximumFiringColumns: 5,
		MinimumInputOverlap:  2,
	})
	l.RandomizeColumns(32)
	inputA := data.NewBitset(64).Set(1, 5)
	inputB := data.NewBitset(64).Set(2, 10)

	maxAttempts := 77
	for !l.SensedInput().Equals(*inputB) && maxAttempts > 0 {
		maxAttempts--
		l.ConsumeInput(*inputA)
		l.ConsumeInput(*inputB)
	}
	oB := l.Output().Clone()
	l.ConsumeInput(*inputA)
	oA := l.Output().Clone()
	fA := l.FeedBack(*oA)
	if !fA.Equals(*inputA) {
		t.Errorf("Feedback for A=%v does not match: %v", *inputA, *fA)
	}
	fB := l.FeedBack(*oB)
	if !fB.Equals(*inputB) {
		t.Errorf("Feedback for B=%v does not match: %v", *inputB, *fB)
	}
}

func TestConsumeInput(t *testing.T) {
	// 50 columns with 4 cells each, firing 2% of columns
	l := NewRegion(RegionParameters{
		Name:                 "Single Region",
		Learning:             true,
		Height:               4,
		Width:                50,
		InputLength:          64,
		MaximumFiringColumns: 40,
		MinimumInputOverlap:  1,
	})

	// 64-bit input, 2 bits of real data.
	columnRand.Seed(1)
	l.RandomizeColumns(2)

	input := data.NewBitset(64)
	input.SetRange(0, 1)

	l.ConsumeInput(*input)

	output := l.Output()
	if output.Len() != 50*4 {
		t.Errorf("Weird output length (expected %d, got: %d).", 50*4, output.Len())
	}
	if output.IsZero() {
		t.Errorf("Output is empty: %v", output)
	}
	t.Log(output)

	// Let's store all the top cells connections into the input bit, so we get them again the second time around.
	nextInput := data.NewBitset(64)

	for _, el := range l.scores {
		col := l.columns[el.index]
		nextInput.Or(col.Connected())
	}
	lastScores := make([]ScoredElement, l.scores.Len())
	copy(lastScores, l.scores)

	l.ConsumeInput(*data.NewBitset(64))
	if l.scores.Len() != 0 {
		t.Errorf("Leftovers in scores: %v", l.scores)
	}
	output2 := l.Output()
	if output2.Len() != 50*4 {
		t.Errorf("Weird output length (expected %d, got: %d).", 50*4, output2.Len())
	}
	if !output2.IsZero() {
		t.Errorf("Empty input caused non-empty output: %v", output2)
	}

	l.ConsumeInput(*nextInput)
	for _, old := range lastScores {
		found := false
		for _, el := range l.scores {
			if old.index == el.index {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Column %d was expected, but is missing!\n%v",
				old.index, l.scores)
		}
	}

	output3 := l.Output()
	if output3.IsZero() {
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
	l := NewRegion(RegionParameters{
		Name:                 "Single Region",
		Learning:             true,
		Height:               4,
		Width:                50,
		InputLength:          2048,
		MaximumFiringColumns: 40,
		MinimumInputOverlap:  1,
	})
	l.Learning = false
	l.RandomizeColumns(28)

	input := data.NewBitset(2048)
	input.Set(columnRand.Perm(2048)[0:28]...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.ConsumeInput(*input)
	}
}
