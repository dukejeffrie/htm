// Implementation of the distal dendrite segment, which connects a cell in a region to any other cell in the same region.

package htm

const (
	CONNECTION_THRESHOLD = 0.6
	INITIAL_PERMANENCE   = 0.6
	PERMANENCE_INC       = float32(0.05)
	PERMANENCE_DEC       = float32(0.05)
)

type DendriteSegment struct {
	// Permanence map.
	permanence map[int]float32
	synapses   *Bitset
}

func (ds DendriteSegment) Connected() Bitset {
	return *ds.synapses
}

func NewDendriteSegment(num_bits int, connected []int) *DendriteSegment {
	ds := new(DendriteSegment)
	ds.permanence = make(map[int]float32, len(connected))
	for _, v := range connected {
		ds.permanence[v] = INITIAL_PERMANENCE
	}
	ds.synapses = NewBitset(num_bits)
	ds.synapses.Set(connected)
	return ds
}

func (ds *DendriteSegment) Learn(input Bitset) {
	for k, v := range ds.permanence {
		if input.IsSet(k) {
			ds.permanence[k] += PERMANENCE_INC
		} else {
			ds.permanence[k] -= PERMANENCE_DEC
		}
		v2 := ds.permanence[k]
		if v >= CONNECTION_THRESHOLD {
			if v2 < CONNECTION_THRESHOLD {
				ds.synapses.ClearOne(k)
			}
		} else if v2 >= CONNECTION_THRESHOLD {
			ds.synapses.SetOne(k)
		}
	}
}
