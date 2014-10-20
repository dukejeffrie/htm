// Code for the column and potential pool.

package htm

import "fmt"
import "math/rand"

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
	boost      float32
	connected  *Bitset
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
func (c Column) Connected() Bitset {
	return *c.connected
}

func (c Column) Boost() float32 {
	return c.boost
}

func (c *Column) ResetConnections(num_bits int, connected []int) {
	c.connected = NewBitset(num_bits)
	for k, _ := range c.permanence {
		delete(c.permanence, k)
	}
	for _, v := range connected {
		c.permanence[v] = INITIAL_PERMANENCE
	}
	c.boost = columnRand.Float32() * 0.0001
	c.connected.Set(connected)
}

func (c Column) Overlap(input Bitset, result *Bitset) {
	result.CopyFrom(*c.connected)
	result.And(input)
}

func (c *Column) LearnFromInput(input *Bitset, score float32) {
	increment := float32(0.05)
	for k, v := range c.permanence {
		if input.IsSet(k) {
			c.permanence[k] += increment
		} else {
			c.permanence[k] -= increment
		}
		v2 := c.permanence[k]
		if v >= CONNECTION_THRESHOLD {
			if v2 < CONNECTION_THRESHOLD {
				c.connected.ClearOne(k)
			}
		} else if v2 >= CONNECTION_THRESHOLD {
			c.connected.SetOne(k)
		}
	}
}
