// Implementation of the distal dendrite segment, which connects a cell in a region to any other cell in the same region.

package segment

import "fmt"
import "bytes"
import "github.com/dukejeffrie/htm/data"
import "math/rand"

var segmentSource = rand.NewSource(304050)
var segmentRand = rand.New(segmentSource)

type CycleHistory struct {
	events *data.Bitset
	cycle  int
}

func NewCycleHistory(length int) *CycleHistory {
	result := &CycleHistory{
		events: data.NewBitset(length),
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
	return fmt.Sprintf("Dendrite{activationAvg=%f, overlapAvg=%f, boost=%f, perm=%v}",
		ac, ov, ds.Boost, ds.PermanenceMap)
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

func (ds *DendriteSegment) Learn(input data.Bitset, active bool, minOverlap int) {
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

func (ds *DendriteSegment) broaden(input data.Bitset, minOverlap int) {
	threshold := ds.Config().Threshold
	newPermanence := ds.Config().Minimum + ds.Boost
	if newPermanence > threshold {
		newPermanence = threshold
	}
	overlapCount := 0
	input.Foreach(func(k int) {
		v := ds.Get(k)
		if v < newPermanence {
			ds.Set(k, newPermanence)
		} else if v >= threshold {
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

type DistalSegment struct {
	*PermanenceMap
	sequence bool
}

type SegmentUpdate struct {
	pos          int
	sequence     bool
	bitsToUpdate *data.Bitset
}

func (u SegmentUpdate) String() string {
	return fmt.Sprintf("@%d(seq=%t)%v", u.pos, u.sequence, *u.bitsToUpdate)
}

func NewSegmentUpdate(pos int, sequence bool, state *data.Bitset) *SegmentUpdate {
	result := &SegmentUpdate{
		pos:          pos,
		sequence:     sequence,
		bitsToUpdate: state,
	}
	return result
}

type DistalSegmentGroup struct {
	segments []*DistalSegment
	updates  []*SegmentUpdate
}

func (g DistalSegmentGroup) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d segments, %d updates", len(g.segments), len(g.updates))
	for _, s := range g.segments {
		fmt.Fprintf(&buf, "\n\t%v", s)
	}
	return buf.String()
}

func NewDistalSegmentGroup() *DistalSegmentGroup {
	return &DistalSegmentGroup{
		segments: make([]*DistalSegment, 0, 15),
		updates:  make([]*SegmentUpdate, 0, 10),
	}
}

func (g DistalSegmentGroup) ComputeActive(activeState data.Bitset, minOverlap int, weak bool) (resultIndex, resultOverlap int) {
	resultIndex = -1
	resultOverlap = -1
	hasSequence := false
	for i, s := range g.segments {
		if overlap := s.Overlap(activeState, weak); overlap >= minOverlap {
			if overlap > resultOverlap && (!hasSequence || s.sequence) {
				resultIndex = i
				resultOverlap = overlap
				if s.sequence {
					hasSequence = true
				}
			}
		}
	}
	return
}

func (g DistalSegmentGroup) HasActiveSegment(activeState data.Bitset, minOverlap int) bool {
	for _, s := range g.segments {
		if s.Connected().Overlap(activeState) >= minOverlap {
			return true
		}
	}
	return false
}

func (g DistalSegmentGroup) Segment(i int) DistalSegment {
	return *g.segments[i]
}

func (g *DistalSegmentGroup) CreateUpdate(sIndex int, activeState data.Bitset, minSynapses int) *SegmentUpdate {
	state := data.NewBitset(activeState.Len())
	if sIndex >= 0 {
		s := g.segments[sIndex]
		state.ResetTo(s.Connected())
	}
	state.Or(activeState)
	for num := state.NumSetBits(); num < minSynapses; num = state.NumSetBits() {
		// TODO(tms): optimize.
		indices := segmentRand.Perm(state.Len())[num:minSynapses]
		state.Set(indices...)
	}
	return NewSegmentUpdate(sIndex, sIndex == -1, state)
}

func (g *DistalSegmentGroup) AddUpdate(update *SegmentUpdate) {
	g.updates = append(g.updates, update)
}

func (g DistalSegmentGroup) HasUpdates() bool {
	return len(g.updates) > 0
}

func (g *DistalSegmentGroup) ApplyAll(positive bool) {
	for _, u := range g.updates {
		g.Apply(u, positive)
	}
	g.updates = g.updates[0:0]
}

func (g *DistalSegmentGroup) Apply(update *SegmentUpdate, positive bool) {
	var s *DistalSegment
	if update.pos == -1 {
		s = &DistalSegment{
			PermanenceMap: PermanenceMapFromBits(*update.bitsToUpdate),
			sequence:      update.sequence,
		}
		g.segments = append(g.segments, s)
	} else {
		s = g.segments[update.pos]
		if positive {
			s.narrow(*update.bitsToUpdate)
		} else {
			s.weaken(*update.bitsToUpdate)
			// TODO(tms): trim segments.
		}
	}
	/* TODO(tms): restore logger.
	if htmLogger != nil {
		htmLogger.Printf("\t\tAfter reinforcement (positive=%t) => %v", positive, *s.PermanenceMap)
	}*/
}
