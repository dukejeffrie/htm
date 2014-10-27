// Implementation of the distal dendrite segment, which connects a cell in a region to any other cell in the same region.

package htm

import "fmt"

const (
	CONNECTION_THRESHOLD = 0.6
	INITIAL_PERMANENCE   = 0.6
	PERMANENCE_MIN       = 0.3
	PERMANENCE_INC       = float32(0.05)
	PERMANENCE_DEC       = float32(0.05)
)

func Permanence(v float32) float32 {
	if v > 1.0 {
		return 1.0
	} else if v < 0.0 {
		return 0.0
	} else {
		return v
	}
}

type CycleHistory struct {
	events *Bitset
	cycle  int
}

func NewCycleHistory(length int) *CycleHistory {
	result := &CycleHistory{
		events: NewBitset(length),
		cycle:  -length,
	}
	return result
}

func (ch *CycleHistory) Record(event bool) {
	at := ch.cycle
	if at < 0 {
		at = ch.events.Len() + at
	}
	if event {
		ch.events.Set(at)
	} else {
		ch.events.Unset(at)
	}
	ch.cycle = (ch.cycle + 1) % ch.events.Len()
}

func (ch CycleHistory) Average() (float32, bool) {
	l := ch.events.Len()
	if ch.cycle < 0 {
		l += ch.cycle
	}
	return float32(ch.events.NumSetBits()) / float32(l), l != 0
}

type DendriteSegment struct {
	// Minimum firing rate for this segment.
	MinActivityRatio float32
	Boost            float32

	// Permanence map.
	permanence        map[int]float32
	synapses          *Bitset
	overlapHistory    *CycleHistory
	activationHistory *CycleHistory
}

func (ds DendriteSegment) String() string {
	ac, _ := ds.activationHistory.Average()
	ov, _ := ds.overlapHistory.Average()
	return fmt.Sprintf("Dendrite{activationAvg=%f, overlapAvg=%f, boost=%f, perm=%v, syn=%v}",
		ac, ov, ds.Boost, ds.permanence, ds.synapses)
}

func (ds DendriteSegment) Connected() Bitset {
	return *ds.synapses
}

func NewDendriteSegment(numBits int, connected []int) *DendriteSegment {
	ds := &DendriteSegment{
		MinActivityRatio:  0.02,
		Boost:             0,
		permanence:        make(map[int]float32, len(connected)),
		overlapHistory:    NewCycleHistory(100),
		activationHistory: NewCycleHistory(100),
		synapses:          NewBitset(numBits),
	}
	for _, v := range connected {
		ds.permanence[v] = INITIAL_PERMANENCE
	}
	ds.synapses.Set(connected...)
	return ds
}

func (ds *DendriteSegment) Learn(input Bitset, active bool, minOverlap int) {
	ds.activationHistory.Record(active)
	if active {
		ds.Narrow(input)
		ds.Boost = 0.0
	} else {
		ds.Broaden(input, minOverlap)
		if avg, ok := ds.activationHistory.Average(); ok && avg < ds.MinActivityRatio {
			ds.Boost *= 1.05
		}
	}
}

func (ds *DendriteSegment) SetPermanence(k int, v float32) {
	v = Permanence(v)
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
		ds.SetPermanence(k, v)
	}
}

func (ds *DendriteSegment) Broaden(input Bitset, minOverlap int) {
	overlapCount := 0
	newPermanence := PERMANENCE_MIN + ds.Boost
	if newPermanence > CONNECTION_THRESHOLD {
		newPermanence = CONNECTION_THRESHOLD
	}
	for _, k := range input.ToIndexes(make([]int, input.NumSetBits())) {
		v, ok := ds.permanence[k]
		if !ok || v < newPermanence {
			ds.SetPermanence(k, newPermanence)
		} else if v >= CONNECTION_THRESHOLD {
			overlapCount++
		}
	}
	ds.overlapHistory.Record(overlapCount >= minOverlap)
	if avg, ok := ds.overlapHistory.Average(); ok && avg < ds.MinActivityRatio {
		for k, v := range ds.permanence {
			ds.SetPermanence(k, v*1.01)
		}
	}
}
