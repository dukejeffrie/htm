// A region is a group of cells of the same hierarchy within a region. It is called "local neighborhood" in Numenta code.

package htm

import "fmt"
import "container/heap"
import "io"

type ScoredElement struct {
	index int
	score float32
}
type TopN []ScoredElement

func (t TopN) Len() int {
	return len(t)
}

func (t TopN) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TopN) Less(i, j int) bool {
	return t[i].score < t[j].score
}

func (t *TopN) Push(x interface{}) {
	el := x.(ScoredElement)
	*t = append(*t, el)
}

func (t *TopN) Pop() interface{} {
	n := t.Len() - 1
	el := (*t)[n]
	*t = (*t)[0:n]
	return el
}

type Scratch struct {
	input  []int
	scores TopN
}

// Parameters to describe a region.
type RegionParameters struct {
	// The name of the region, used for debugging.
	Name string
	// Whether learning is on or off.
	Learning bool
	// Number of cells in each column.
	Height int
	// Number of columns in this region.
	Width int
	// Size of the input, in bits.
	InputLength int
	// Maximum number of columns that can fire.
	MaximumFiringColumns int
	// Minimum overlap between an input and a column's proximal dentrite to trigger
	// activation.
	MinimumInputOverlap int
}

type Region struct {
	RegionParameters
	columns    []*Column
	output     *Bitset
	learnState *Bitset
	scratch    Scratch
}

// Creates a new named region with the given parameters.
func NewRegion(params RegionParameters) *Region {
	result := &Region{
		RegionParameters: params,
		columns:          make([]*Column, params.Width),
		output:           NewBitset(params.Width * params.Height),
		learnState:       NewBitset(params.Width * params.Height),
		scratch: Scratch{
			input:  make([]int, params.InputLength),
			scores: make([]ScoredElement, 0, params.MaximumFiringColumns+1),
		},
	}
	for i := 0; i < params.Width; i++ {
		result.columns[i] = NewColumn(params.InputLength, params.Height)
	}
	return result
}

func (l Region) Height() int {
	return l.RegionParameters.Height
}

func (l Region) Width() int {
	return l.RegionParameters.Width
}

func (l *Region) RandomizeColumns(w int) {
	perm := make([]int, w)
	for _, col := range l.columns {
		for i := 0; i < w; i++ {
			perm[i] = columnRand.Intn(l.InputLength)
		}
		col.ResetConnections(perm)
		col.SetBoost(columnRand.Float32() * 0.00001)
	}
}

func (l *Region) ResetColumnSynapses(i int, indices ...int) {
	col := l.columns[i]
	col.ResetConnections(indices)
	col.SetBoost(columnRand.Float32() * 0.00001)
}

func (l *Region) ConsumeInput(input Bitset) {
	l.scratch.scores = l.scratch.scores[0:0]
	for i, c := range l.columns {
		c.active.Reset()
		overlapScore := c.Connected().Overlap(input)
		if overlapScore >= l.MinimumInputOverlap {
			score := float32(overlapScore) + c.Boost()
			heap.Push(&l.scratch.scores, ScoredElement{i, score})
			if l.scratch.scores.Len() > l.MaximumFiringColumns {
				heap.Pop(&l.scratch.scores)
			}
		}
	}
	l.output.Reset()
	for _, el := range l.scratch.scores {
		col := l.columns[el.index]
		col.Activate()
		l.output.SetFromBitsetAt(col.Active(), el.index*col.Height())
	}
	if l.Learning {
		l.Learn(input)
	}
}

func (l *Region) Output() Bitset {
	return *l.output
}

func (l *Region) Learn(input Bitset) {
	for _, col := range l.columns {
		col.LearnFromInput(input, l.MinimumInputOverlap)
	}
}

func (l Region) Print(writer io.Writer) {
	fmt.Fprintf(writer, "\n=== %s (learning: %t) ===\n", l.Name, l.Learning)
	for i := 8; i < len(l.columns) && i <= 80; i += 8 {
		fmt.Fprintf(writer, "%8d", i)
	}
	fmt.Fprintln(writer)
	rem := 80
	for _, c := range l.columns {
		c.Print(rem, writer)
		rem -= c.Height()
		if rem <= 0 {
			rem = 80
		}
	}
	fmt.Fprintln(writer)
}
