// Implementation of the distal dendrite segment, which connects a cell in a region to any other cell in the same region.

package htm

const (
	CONNECTION_THRESHOLD = 0.6
	INITIAL_PERMANENCE   = 0.6
	PERMANENCE_MIN       = 0.3
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
	ds.synapses.Set(connected...)
	return ds
}

func (ds *DendriteSegment) Learn(input Bitset, active bool) {
	if active {
		ds.Narrow(input)
	} else {
		ds.Broaden(input)
	}
}

func (ds *DendriteSegment) Narrow(input Bitset) {
	for k, v := range ds.permanence {
		if input.IsSet(k) {
			v += PERMANENCE_INC
			if v > 1.0 {
				v = 1.0
			}
		} else {
			v -= PERMANENCE_DEC
		}
		ds.permanence[k] = v
		if v >= CONNECTION_THRESHOLD {
			ds.synapses.Set(k)
		} else {
			ds.synapses.Unset(k)
			if v < PERMANENCE_MIN {
				delete(ds.permanence, k)
			}
		}
	}
}

func (ds *DendriteSegment) Broaden(input Bitset) {
	for _, k := range input.ToIndexes(make([]int, input.NumSetBits())) {
		v, ok := ds.permanence[k]
		if !ok || v < PERMANENCE_MIN {
			ds.permanence[k] = PERMANENCE_MIN
		}
	}
}
