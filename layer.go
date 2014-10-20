// A layer is a group of cells of the same hierarchy within a region.

package htm

import "container/heap"

type ScoredElement struct {
	index int
	score int
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
	columns          []*Column
	name             string
	scratch          Scratch
	maxFiringColumns int
	Learning         bool
}

// Creates a new named layer with this many columns.
func NewLayer(name string, width, height int) *Layer {
	maxFiringColumns := (width * 2 / 100) + 1
	result := &Layer{
		columns: make([]*Column, width),
		name:    name,
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
	permutation := columnRand.Perm(n)
	for i, col := range l.columns {
		col.ResetConnections(n, permutation[i:i+w])
	}
	if cap(l.scratch.input) < w {
		l.scratch.input = make([]int, w)
	}
	l.scratch.overlap = NewBitset(n)
	l.scratch.scores = l.scratch.scores[0:0]
}

func (l *Layer) ConsumeInput(input *Bitset) {
	l.scratch.scores = l.scratch.scores[0:0]
	for i, c := range l.columns {
		c.Overlap(*input, l.scratch.overlap)
		overlap_score := l.scratch.overlap.NumSetBits()
		if overlap_score > 0 {
			heap.Push(&l.scratch.scores, ScoredElement{i, overlap_score})
			if l.scratch.scores.Len() > l.maxFiringColumns {
				heap.Pop(&l.scratch.scores)
			}
		}
	}
	if l.Learning {
		l.Learn(input)
	}
}

func (l *Layer) Learn(input *Bitset) {
	// TODO(tms): implement
}
