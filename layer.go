// A layer is a group of cells of the same hierarchy within a region. It is called "local neighborhood" in Numenta code.

package htm

import "fmt"
import "math"
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
	overlap *Bitset
	scores  TopN
}

type Layer struct {
	Name             string
	Learning         bool
	columns          []*Column
	output           *Bitset
	scratch          Scratch
	maxFiringColumns int
	outputIsDirty    bool
}

// Creates a new named layer with this many columns.
func NewLayer(name string, width, height int, firing_ratio float64) *Layer {
	maxFiringColumns := int(math.Ceil(float64(width)*firing_ratio)) + 1
	result := &Layer{
		columns: make([]*Column, width),
		output:  NewBitset(width * height),
		Name:    name,
		scratch: Scratch{
			input:  make([]int, 28),
			scores: make([]ScoredElement, 0, maxFiringColumns),
		},
		maxFiringColumns: maxFiringColumns,
		Learning:         true,
	}
	for i := 0; i < width; i++ {
		result.columns[i] = NewColumn(height)
	}
	return result
}

func (l Layer) Height() int {
	return l.columns[0].Height()
}

func (l Layer) Width() int {
	return len(l.columns)
}

func (l *Layer) ResetForInput(n, w int) {
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
	l.scratch.overlap = NewBitset(n)
	l.scratch.scores = l.scratch.scores[0:0]
}

func (l *Layer) ConsumeInput(input Bitset) {
	l.scratch.scores = l.scratch.scores[0:0]
	l.outputIsDirty = true
	for i, c := range l.columns {
		c.active.Reset()
		l.scratch.overlap.CopyFrom(c.Connected())
		l.scratch.overlap.And(input)
		overlap_score := l.scratch.overlap.NumSetBits()
		if overlap_score > 0 {
			score := float32(overlap_score) + c.Boost()
			heap.Push(&l.scratch.scores, ScoredElement{i, score})
			if l.scratch.scores.Len() > l.maxFiringColumns {
				heap.Pop(&l.scratch.scores)
			}
		}
	}
	for _, el := range l.scratch.scores {
		col := l.columns[el.index]
		col.Activate()
	}
	if l.Learning {
		l.Learn(input)
	}
}

func (l *Layer) Output() Bitset {
	if l.outputIsDirty {
		l.output.Reset()
		for _, el := range l.scratch.scores {
			col := l.columns[el.index]
			l.output.SetFromBitsetAt(col.Active(), el.index*col.Height())
		}
		l.outputIsDirty = false
	}
	return *l.output
}

func (l *Layer) Learn(input Bitset) {
	if len(l.scratch.scores) > 0 {
		for _, el := range l.scratch.scores {
			col := l.columns[el.index]
			col.LearnFromInput(input)
		}
	}
}

func (l Layer) Print(writer io.Writer) {
	fmt.Fprintf(writer, "\n=== %s (learning: %t) ===", l.Name, l.Learning)
	for i, col := range l.columns {
		fmt.Fprintf(writer, "\n%d. ", i)
		col.Print(writer)
	}
	fmt.Fprintln(writer)
}
