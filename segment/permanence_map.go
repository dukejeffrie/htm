// Permanence map implementation.
//
// Neuron dendrites are modeled as a number of connections to some input (which is
// usually the output of another set of neurons. Even the data from the real world
// is encoded as a sensorial region).
//
// Each connection from a dendrite to its input, representing a synapse, is binary,
// i.e. it is either connected or disconnected. But the durability of the
// connection is variable: synapses that activate often are strengthened, while
// synapses that are not used are weakened. Beyond some threshold, we consider the
// synapses no longer connected.
//
// A permanence value is a float32 in the closed range [0.0, 1.0]. Numenta uses a
// quantized increment/decrement strategy and there is a threshold for connection.

package segment

import "fmt"
import "github.com/dukejeffrie/htm/data"

// Parameters for the permanence map.
type PermanenceConfiguration struct {
	// Threshold for the permanence map. Permanence values above the threshold are
	// considered connected.
	Threshold float32

	// Permanence value for new synapses. Must be slightly above Threshold if
	// you mean to follow the Numenta whitepaper.
	Initial float32

	// Minimal permanence value. Synapses that go below this value are removed from
	// the map to contain memory explosion.
	Minimum float32

	// The increment to use when a synapse is reinforced.
	Increment float32

	// The decrement to use when a synapse is weakened.
	Decrement float32
}

// Default permanence configuration.
var DefaultPermanenceConfig PermanenceConfiguration

func init() {
	DefaultPermanenceConfig = PermanenceConfiguration{
		Threshold: 0.6,
		Initial:   0.66,
		Minimum:   0.3,
		Increment: 0.05,
		Decrement: 0.05,
	}
}

// PermanenceMap structure.
type PermanenceMap struct {
	config         PermanenceConfiguration
	permanence     map[int]float32
	synapses       *data.Bitset
	receptiveField *data.Bitset
}

func (pm PermanenceMap) Config() PermanenceConfiguration {
	return pm.config
}

func NewPermanenceMap(numBits int) *PermanenceMap {
	result := &PermanenceMap{
		config:         DefaultPermanenceConfig,
		permanence:     make(map[int]float32),
		synapses:       data.NewBitset(numBits),
		receptiveField: data.NewBitset(numBits),
	}
	return result
}

func PermanenceMapFromBits(bits data.Bitset) (pm *PermanenceMap) {
	pm = NewPermanenceMap(bits.Len())
	bits.Foreach(func(k int) {
		pm.permanence[k] = pm.config.Initial
	})
	if pm.config.Initial > pm.config.Threshold {
		pm.synapses.Or(bits)
	}
	return pm
}

func (pm *PermanenceMap) Reset(connected ...int) {
	if len(pm.permanence) > 0 {
		pm.permanence = make(map[int]float32)
		pm.synapses.Reset()
	}
	for _, v := range connected {
		pm.permanence[v] = pm.config.Initial
	}
	pm.synapses.Set(connected...)
	pm.receptiveField.ResetTo(*pm.synapses)
}

func (pm PermanenceMap) Len() int {
	return pm.synapses.Len()
}

func (pm *PermanenceMap) Get(k int) (v float32) {
	v = pm.permanence[k]
	return
}

func (pm *PermanenceMap) Set(k int, v float32) {
	if v > 1.0 {
		v = 1.0
	} else if v < 0.0 {
		v = 0.0
	}
	if v < pm.config.Minimum {
		pm.synapses.Unset(k)
		pm.receptiveField.Unset(k)
		delete(pm.permanence, k)
		return
	}
	pm.permanence[k] = v
	pm.receptiveField.Set(k)
	if v >= pm.config.Threshold {
		pm.synapses.Set(k)
	} else {
		pm.synapses.Unset(k)
	}
}

func (pm PermanenceMap) String() string {
	return fmt.Sprintf("(%d/%dconnected, %+v)",
		pm.synapses.NumSetBits(),
		pm.receptiveField.NumSetBits(),
		pm.permanence)
}

func (pm PermanenceMap) Connected() data.Bitset {
	return *pm.synapses
}

func (pm PermanenceMap) ReceptiveField() data.Bitset {
	return *pm.receptiveField
}

func (pm PermanenceMap) Overlap(input data.Bitset, weak bool) int {
	if weak {
		return pm.receptiveField.Overlap(input)
	} else {
		return pm.synapses.Overlap(input)
	}
}

func (pm *PermanenceMap) narrow(input data.Bitset) {
	for k, v := range pm.permanence {
		if input.IsSet(k) {
			v += pm.config.Increment
		} else {
			v -= pm.config.Decrement
		}
		pm.Set(k, v)
	}
}

func (pm *PermanenceMap) weaken(input data.Bitset) {
	for k, v := range pm.permanence {
		if input.IsSet(k) {
			v -= pm.config.Decrement
		}
		pm.Set(k, v)
	}
}
