// Code for the column and potential pool.

package htm

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
	boost    float32
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
	return c.boost
}

func (c *Column) ResetConnections(num_bits int, connected []int) {
	c.proximal = NewDendriteSegment(num_bits, connected)
	c.boost = columnRand.Float32() * 0.0001
}

func (c *Column) LearnFromInput(input Bitset) {
	c.proximal.Learn(input)
}

func (c *Column) Activate() {
	if c.predicted.NumSetBits() > 0 {
		c.active.CopyFrom(*c.predicted)
	} else {
		// Bursting.
		c.active.SetRange(0, c.Height())
	}
}

func (c Column) ToByte(idx int) byte {
	if c.active.IsSet(idx) {
		if c.predicted.IsSet(idx) {
			return 'v'
		} else {
			return '!'
		}
	} else if c.predicted.IsSet(idx) {
		return 'o'
	}
	return '.'
}

func (c Column) Print(writer io.Writer) error {
	bytes := make([]byte, c.Height())
	for i := 0; i < c.Height(); i++ {
		bytes[i] = c.ToByte(i)
	}
	_, err := writer.Write(bytes)
	return err
}
