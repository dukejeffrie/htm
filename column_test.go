package htm

import "fmt"
import "testing"

type Layer struct {
	columns []*Column
	name    string
}

// Creates a new named layer with this many columns.
func NewLayer(name string, width, height int) *Layer {
	result := &Layer{
		columns: make([]*Column, width),
		name:    name,
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
		col.ResetConnections(permutation[i : i+w])
	}
}

func TestPotentialPool(t *testing.T) {
	// 5 columns with 4 cells each.
	layer0 := NewLayer("Single Layer", 5, 4)

	// 64-bit scalar input, 2 bits of real data.
	layer0.ResetForInput(64, 2)

	for i, c := range layer0.columns {
		fmt.Printf("Column %d: %v", i, c)
	}
	//t.Fail()
}
