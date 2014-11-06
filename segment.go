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

type PermanenceMap struct {
	permanence map[int]float32
	synapses   *Bitset
}

func NewPermanenceMap(numBits int) *PermanenceMap {
	result := &PermanenceMap{
		permanence: make(map[int]float32),
		synapses:   NewBitset(numBits),
	}
	return result
}

func (pm *PermanenceMap) Reset(connected ...int) {
	if len(pm.permanence) > 0 {
		pm.permanence = make(map[int]float32)
		pm.synapses.Reset()
	}
	for _, v := range connected {
		pm.permanence[v] = INITIAL_PERMANENCE
	}
	pm.synapses.Set(connected...)
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
	if v < PERMANENCE_MIN {
		pm.synapses.Unset(k)
		delete(pm.permanence, k)
		return
	}
	pm.permanence[k] = v
	if v >= CONNECTION_THRESHOLD {
		pm.synapses.Set(k)
	} else {
		pm.synapses.Unset(k)
	}
}

func (pm PermanenceMap) String() string {
	return fmt.Sprintf("(%d connected, %+v)", pm.synapses.NumSetBits(), pm.permanence)
}

func (pm PermanenceMap) Connected() Bitset {
	return *pm.synapses
}

func (pm PermanenceMap) Overlap(input Bitset, weak bool) int {
	if weak {
		max := input.NumSetBits()
		count := 0
		for k, _ := range pm.permanence {
			if input.IsSet(k) {
				count++
				if count >= max {
					break
				}
			}
		}
		return count
	} else {
		return pm.synapses.Overlap(input)
	}
}

func (pm *PermanenceMap) narrow(input Bitset) {
	for k, v := range pm.permanence {
		if input.IsSet(k) {
			v += PERMANENCE_INC
		} else {
			v -= PERMANENCE_DEC
		}
		pm.Set(k, v)
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
	*PermanenceMap
	// Minimum firing rate for this segment.
	MinActivityRatio float32
	Boost            float32

	overlapHistory    *CycleHistory
	activationHistory *CycleHistory
}

func (ds DendriteSegment) String() string {
	ac, _ := ds.activationHistory.Average()
	ov, _ := ds.overlapHistory.Average()
	return fmt.Sprintf("Dendrite{activationAvg=%f, overlapAvg=%f, boost=%f, perm=%v, syn=%v}",
		ac, ov, ds.Boost, ds.permanence, ds.synapses)
}

func NewDendriteSegment(numBits int) *DendriteSegment {
	ds := &DendriteSegment{
		PermanenceMap:     NewPermanenceMap(numBits),
		MinActivityRatio:  0.02,
		Boost:             0,
		overlapHistory:    NewCycleHistory(100),
		activationHistory: NewCycleHistory(100),
	}
	return ds
}

func (ds *DendriteSegment) Learn(input Bitset, active bool, minOverlap int) {
	ds.activationHistory.Record(active)
	if active {
		ds.narrow(input)
		ds.Boost = 0.0
	} else {
		ds.broaden(input, minOverlap)
		if avg, ok := ds.activationHistory.Average(); ok && avg < ds.MinActivityRatio {
			ds.Boost *= 1.05
		}
	}
}

func (ds *DendriteSegment) broaden(input Bitset, minOverlap int) {
	newPermanence := PERMANENCE_MIN + ds.Boost
	if newPermanence > CONNECTION_THRESHOLD {
		newPermanence = CONNECTION_THRESHOLD
	}
	overlapCount := 0
	input.Foreach(func(k int) {
		v := ds.Get(k)
		if v < newPermanence {
			ds.Set(k, newPermanence)
		} else if v >= CONNECTION_THRESHOLD {
			overlapCount++
		}
	})
	ds.overlapHistory.Record(overlapCount >= minOverlap)
	if avg, ok := ds.overlapHistory.Average(); ok && avg < ds.MinActivityRatio {
		for k, v := range ds.permanence {
			ds.Set(k, v*1.01)
		}
	}
}
