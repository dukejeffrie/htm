// Tests for temporal pooling.

package test

import "bytes"
import "fmt"
import "math/rand"
import "testing"
import "github.com/dukejeffrie/htm"
import "github.com/dukejeffrie/htm/data"

type LoggerInterface interface {
	Errorf(format string, args ...interface{})
	Failed() bool
	FailNow()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

type TpTest struct {
	LoggerInterface
	inputA *data.Bitset
	inputB *data.Bitset

	layer0     *htm.Region
	expected0  *data.Bitset
	active0    *data.Bitset
	predicted0 *data.Bitset
	step       int
	Verify     bool
}

func NewTpTest(t LoggerInterface) *TpTest {
	params := htm.RegionParameters{
		Name:                 "0-tp",
		Learning:             true,
		Height:               8,
		Width:                64,
		InputLength:          64,
		MinimumInputOverlap:  1,
		MaximumFiringColumns: 2,
	}
	result := &TpTest{
		LoggerInterface: t,
		step:            0,
		inputA:          data.NewBitset(64),
		inputB:          data.NewBitset(64),
		layer0:          htm.NewRegion(params),
		expected0:       data.NewBitset(params.Width * params.Height),
		active0:         data.NewBitset(params.Width * params.Height),
		predicted0:      data.NewBitset(params.Width * params.Height),
		Verify:          true,
	}

	odds := make([]int, 32)
	evens := make([]int, 32)
	for j := 0; j < 32; j++ {
		odds[j] = j + 1
		evens[j] = j
	}

	result.inputA.Set(1)
	result.inputB.Set(8)
	for i := 0; i < params.Width; i++ {
		result.layer0.ResetColumnSynapses(i, i)
	}

	return result
}

func (tp *TpTest) Step(input data.Bitset) {
	tp.step++
	tp.layer0.ConsumeInput(input)
	if !tp.Verify {
		return
	}
	tp.checkPredicted0()
	tp.predicted0.Reset()
	tp.checkActive0()
	tp.active0.Reset()
	if tp.Failed() {
		tp.PrintRegion(*tp.layer0)
		tp.FailNow()
	}
}

func (tp *TpTest) checkActive0() {
	expected := *tp.active0
	actual := tp.layer0.ActiveState()
	if !actual.Equals(expected) {
		tp.Errorf("ACtive state does not match expected (overlap=%d/%d)\n\tActual: %v\n\t Expected: %v", actual.Overlap(expected), expected.NumSetBits(), actual, expected)
	}
}

func (tp *TpTest) checkPredicted0() {
	expected := *tp.predicted0
	pred := tp.layer0.PredictiveState()
	if !pred.Equals(expected) {
		tp.Errorf("Prediction does not match expected (overlap=%d/%d)\n\tPredicted: %v\n\t Expected: %v", pred.Overlap(expected), expected.NumSetBits(), pred, expected)
	}
}

func (tp *TpTest) checkPredictedInput(input data.Bitset) {
	if !tp.layer0.PredictedInput().Equals(input) {
		tp.Errorf("(t=%d) Bad predicted input. Expected: %v, but got: %v", tp.step,
			input, tp.layer0.PredictedInput())
		tp.PrintRegion(*tp.layer0)
		tp.FailNow()
	}
}

func (tp *TpTest) PrintRegion(r htm.Region, columns ...int) {
	var buf bytes.Buffer
	r.Print(&buf)
	for _, idx := range columns {
		c := r.Column(idx)
		fmt.Fprintf(&buf, "\n%v: ", c)
		for i := 0; i < c.Height(); i++ {
			fmt.Fprintf(&buf, "\n[%04d] %v", c.CellId(i), c.Distal(i))
		}
	}
	tp.Log(buf.String())
	tp.Logf("Predicted: %v", r.PredictiveState())
	tp.Logf("Sensed input: %v", r.SensedInput())
	tp.Logf("Predicted input: %v", r.PredictedInput())
}

func TestTp_LearnAAB(t *testing.T) {
	test := NewTpTest(t)
	test.Verify = false
	for i := 0; i < 100; i++ {
		test.Step(*test.inputA)
		test.Step(*test.inputA)
		test.Step(*test.inputB)
	}

	aOrB := data.NewBitset(test.layer0.InputLength)
	aOrB.ResetTo(*test.inputA)
	aOrB.Or(*test.inputB)

	// After ...,B comes A.
	test.checkPredictedInput(*test.inputA)

	// After ...,A comes A.
	test.Step(*test.inputA)
	test.checkPredictedInput(*test.inputA)

	// After ...,A,A comes A or B.
	test.Step(*test.inputA)
	test.checkPredictedInput(*aOrB)
}

func BenchmarkTp_AAB(b *testing.B) {
	test := NewTpTest(b)
	test.Verify = false
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		test.Step(*test.inputA)
		test.Step(*test.inputA)
		test.Step(*test.inputB)
	}
}

func TestTp_AAxB(t *testing.T) {
	test := NewTpTest(t)
	test.Verify = false
	random := data.NewBitset(test.layer0.InputLength)
	for i := 0; i < 8000; i++ {
		test.Step(*test.inputA)
		test.Step(*test.inputA)
		random.Reset().Set(rand.Intn(test.layer0.InputLength))
		test.Step(*random)
		test.Step(*test.inputB)
	}

	aOrB := data.NewBitset(test.layer0.InputLength)
	aOrB.ResetTo(*test.inputA)
	aOrB.Or(*test.inputB)

	// After ...,B comes A.
	test.checkPredictedInput(*test.inputA)

	// After ...,A comes A.
	test.Step(*test.inputA)
	test.checkPredictedInput(*test.inputA)

	// After ...,A,A comes anything.
	test.Step(*test.inputA)
	test.Log("Predict anything:", test.layer0.PredictedInput())

	// Then comes B.
	random.Reset().Set(rand.Intn(test.layer0.InputLength))
	test.Step(*random)
	test.checkPredictedInput(*test.inputB)
}

func TestTp_LearnA2B_Step(t *testing.T) {
	test := NewTpTest(t)

	// Show A, burst column 1.
	test.active0.SetRange(8*1, 8*1+8) // burst
	test.Step(*test.inputA)

	// Show B, burst column 8.
	test.active0.SetRange(8*8, 8*8+8) // burst
	test.Step(*test.inputB)

	// Now check that inputA activates a segment that predicts inputB for each of
	// the active cells in column 8. In the first step, only cell (8, 0) is active.
	dest := test.layer0.Column(8)
	lActive := test.layer0.LearningActiveState()
	if !lActive.IsSet(dest.CellId(0)) {
		test.Errorf("Expected cell 0 to be selected for learning: %v", lActive)
		return
	}

	// Show A, burst column 1, expect B
	test.active0.SetRange(8*1, 8*1+8) // burst
	test.predicted0.Set(test.layer0.Column(8).CellId(0))
	test.Step(*test.inputA)
	lActive = test.layer0.LearningActiveState()
	if !lActive.IsSet(test.layer0.Column(1).CellId(0)) {
		test.Errorf("Expected cell (1,0) to be active: %v", lActive)
	}

	// Show B, activate (8, 0), expect A
	test.active0.Set(test.layer0.Column(8).CellId(0))
	test.predicted0.Set(test.layer0.Column(1).CellId(0))
	test.Step(*test.inputB)
	lActive = test.layer0.LearningActiveState()
	if !lActive.IsSet(test.layer0.Column(8).CellId(0)) {
		test.Errorf("Expected cell (8,0) to be active: %v", lActive)
	}
	test.PrintRegion(*test.layer0, 1, 8)
}
