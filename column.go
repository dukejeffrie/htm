// Code for the column and potential pool.

package htm

import "bufio"
import "fmt"
import "io"
import "math/rand"

var columnSource = rand.NewSource(1979)
var columnRand = rand.New(columnSource)

// Keeps information about a column in a cortical region.
type Column struct {
	// The bitset of cells that are active.
	active *Bitset
	// The bitset of cells that are predicted.
	predicted *Bitset

	// Proximal dendrite segment.
	proximal *DendriteSegment
}

func NewColumn(height int) *Column {
	result := new(Column)
	result.active = NewBitset(height)
	result.predicted = NewBitset(height)
	return result
}

func (c Column) Height() int {
	return c.active.Len()
}

func (c Column) String() string {
	return fmt.Sprintf("Column{active=%d,predicted=%d,connected=%v}",
		c.active.NumSetBits(),
		c.predicted.NumSetBits(),
		c.Connected())
}

// Gets the indices of the connected synapses. The result is in ascending order.
func (c Column) Connected() Bitset {
	return c.proximal.Connected()
}

func (c Column) Active() Bitset {
	return *c.active
}

func (c Column) Boost() float32 {
	return c.proximal.Boost
}

func (c *Column) ResetConnections(numBits int, connected []int) {
	c.proximal = NewDendriteSegment(numBits, connected)
	c.proximal.Boost = columnRand.Float32() * 0.00001
}

func (c *Column) LearnFromInput(input Bitset, minOverlap int) {
	c.proximal.Learn(input, !c.Active().IsZero(), minOverlap)
}

func (c *Column) Activate() {
	// This is the inference part of the temporal pooler, Phase 1: if any cell is
	// predicted from the last step, we activate the predicted cells for this column.
	// Otherwise we activate all cells.
	if !c.predicted.IsZero() {
		c.active.ResetTo(*c.predicted)
	} else {
		// Bursting.
		c.active.SetRange(0, c.Height())
	}
}

func (c Column) ToRune(idx int) rune {
	if c.active.IsSet(idx) {
		if c.predicted.IsSet(idx) {
			return 'x'
		} else {
			return '!'
		}
	} else if c.predicted.IsSet(idx) {
		return 'o'
	}
	return '-'
}

func (c Column) Print(width int, writer io.Writer) error {
	buf := bufio.NewWriter(writer)
	for i := 0; i < c.Height(); i++ {
		if _, err := buf.WriteRune(c.ToRune(i)); err != nil {
			return err
		}
		if (i+1)%width == 0 {
			if _, err := buf.WriteRune('\n'); err != nil {
				return err
			}
		}
	}
	return buf.Flush()
}
