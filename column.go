// Code for the column and potential pool.

package htm

import "fmt"
import "math/rand"
import "sort"

const (
	CONNECTION_THRESHOLD = 0.6
	INITIAL_PERMANENCE   = 0.6
)

var columnSource = rand.NewSource(1979)
var columnRand = rand.New(columnSource)

// Keeps information about a column in a cortical layer.
type Column struct {
	// The bitset of columns that are active
	active *Bitset
	// The bitset of columns that are predicted
	predicted *Bitset

	// Permanence map
	permanence map[int]float32
}

func NewColumn(height int) *Column {
	result := new(Column)
	result.active = NewBitset(height)
	result.predicted = NewBitset(height)
	result.permanence = make(map[int]float32)
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
func (c Column) Connected() []int {
	result := make([]int, 0, len(c.permanence))
	for k, v := range c.permanence {
		if v >= CONNECTION_THRESHOLD {
			result = append(result, k)
		}
	}
	sort.Ints(result)
	return result
}

func (c *Column) ResetConnections(connected []int) {
	for k, _ := range c.permanence {
		delete(c.permanence, k)
	}
	for _, v := range connected {
		c.permanence[v] = INITIAL_PERMANENCE
	}
}

func (c Column) GetOverlap(input []int, result *Bitset) {
	for _, v := range input {
		if c.permanence[v] >= CONNECTION_THRESHOLD {
			result.SetOne(v)
		}
	}
}
