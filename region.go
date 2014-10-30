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
	input   []int
	scores  TopN
}

type RegionParameters struct {
	Name                 string
	Learning             bool
	Height               int
	Width                int
	InputLength          int
	MaximumFiringColumns int
	MinimumInputOverlap  int
}

type Region struct {
	RegionParameters
	columns    []*Column
	output     *Bitset
	learnState *Bitset
	scratch    Scratch
}

// Creates a new named region with this many columns.
func NewRegion(params RegionParameters) *Region {
	result := &Region{
		RegionParameters: params,
		columns:          make([]*Column, params.Width),
		output:           NewBitset(params.Width * params.Height),
		learnState:       NewBitset(params.Width * params.Height),
		scratch: Scratch{
			input:      make([]int, 28),
			scores:     make([]ScoredElement, 0, params.MaximumFiringColumns+1),
		},
	}
	for i := 0; i < params.Width; i++ {
		result.columns[i] = NewColumn(params.Height)
	}
	return result
}

func (l Region) Height() int {
	return l.RegionParameters.Height
}

func (l Region) Width() int {
	return l.RegionParameters.Width
}

func (l *Region) ResetForInput(n, w int) {
	perm := make([]int, w)
	for _, col := range l.columns {
		for i := 0; i < w; i++ {
			perm[i] = columnRand.Intn(n)
		}
		col.ResetConnections(n, perm)
	}
	if cap(l.scratch.input) < w {
		l.scratch.input = make([]int, w)
	}
	l.scratch.scores = l.scratch.scores[0:0]
}

func (l *Region) ResetColumnSynapses(i int, indices ...int) {
	col := l.columns[i]
	col.ResetConnections(l.InputLength, indices)
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
