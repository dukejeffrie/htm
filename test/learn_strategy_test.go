package test

import "fmt"
import "testing"
import "math/rand"
import "github.com/dukejeffrie/htm/data"
import "github.com/dukejeffrie/htm/segment"

func TryToLearn(t *testing.T, maxTries int, ds *segment.DendriteSegment,
	minOverlap int, inputs ...data.Bitset) int {
	result := 0
	overlap := data.NewBitset(inputs[0].Len())
	for !ds.Connected().Equals(inputs[0]) {
		input := inputs[result%len(inputs)]
		overlap.ResetTo(input)
		overlap.And(ds.Connected())
		active := overlap.NumSetBits() >= minOverlap || rand.Float32()+ds.Boost > 3.0
		ds.Learn(input, active, minOverlap)
		result++
		if active {
			fmt.Print("N")
		} else {
			fmt.Print("B")
		}
		if result > maxTries {
			t.Error(fmt.Sprintf("Failed after %d rounds.", maxTries))
			return result
		}
	}
	fmt.Println()
	t.Log(fmt.Sprintf("Learned in %d rounds.", result))
	return result
}

func TestLearn64PatternA_5(t *testing.T) {
	rand.Seed(1979)
	ds := segment.NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13, 21)
	patternA := data.NewBitset(64)
	patternA.Set(2, 4, 22, 24, 42, 44, 62)
	TryToLearn(t, 80, ds, 5, *patternA)
	t.Log(ds)
}

func TestLearn64PatternAB_5(t *testing.T) {
	rand.Seed(1979)
	ds := segment.NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13, 21)
	patternA := data.NewBitset(64)
	patternA.Set(2, 4, 22, 24, 42, 44, 62)
	patternB := data.NewBitset(64)
	patternB.Set(22, 23, 24, 25, 26)
	TryToLearn(t, 90, ds, 5, *patternA, *patternB)
	t.Log(ds.Connected())
}

func TestLearn64PatternABC_5(t *testing.T) {
	rand.Seed(1979)
	ds := segment.NewDendriteSegment(64)
	ds.Reset(1, 3, 5, 8, 13, 21)
	patternA := data.NewBitset(64)
	patternA.Set(2, 4, 22, 24, 42, 44, 62)
	patternB := data.NewBitset(64)
	patternB.Set(22, 23, 24, 25, 26)
	patternC := data.NewBitset(64)
	patternC.Set(3, 13, 21, 39, 47)
	TryToLearn(t, 110, ds, 5, *patternA, *patternB, *patternC)
	t.Log(ds.Connected())
}
