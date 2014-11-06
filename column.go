// Code for the column and potential pool.

package htm

import "fmt"
import "math/rand"

var columnSource = rand.NewSource(1979)
var columnRand = rand.New(columnSource)

// Keeps information about a column in a cortical region.
type Column struct {
	Index int
	// The bitset of cells that are active.
	active *Bitset
	// The bitset of cells that are predicted.
	predictive *Bitset

	// Cell that was selected for learning.
	learning       int
	learningTarget int

	// Proximal dendrite segment.
	proximal *DendriteSegment

	// Per-cell distal segment group.
	distal []*DistalSegmentGroup
}

func NewColumn(inputSize, height int) *Column {
	result := &Column{
		active:         NewBitset(height),
		predictive:     NewBitset(height),
		proximal:       NewDendriteSegment(inputSize),
		learning:       -1,
		learningTarget: 0,
		distal:         make([]*DistalSegmentGroup, height),
	}
	for i := 0; i < height; i++ {
		result.distal[i] = NewDistalSegmentGroup()
	}
	return result
}

func (c Column) Height() int {
	return len(c.distal)
}

func (c Column) String() string {
	return fmt.Sprintf("Col[%d]{active=%d,predictive=%d,learning=%d,connected=%v}",
		c.Index,
		c.active.NumSetBits(),
		c.predictive.NumSetBits(),
		c.learning,
		c.Connected())
}

func (c Column) Connected() Bitset {
	return c.proximal.Connected()
}

func (c Column) Active() Bitset {
	return *c.active
}

func (c Column) Predictive() Bitset {
	return *c.predictive
}

func (c Column) Distal(i int) DistalSegmentGroup {
	return *c.distal[i]
}

func (c Column) Boost() float32 {
	return c.proximal.Boost
}

func (c *Column) SetBoost(boost float32) {
	c.proximal.Boost = boost
}

func (c *Column) ResetConnections(connected []int) {
	c.proximal.Reset(connected...)
}

func (c *Column) LearnFromInput(input Bitset, minOverlap int) {
	c.proximal.Learn(input, !c.Active().IsZero(), minOverlap)
}

func (c *Column) Predict(activeState Bitset, minOverlap int) {
	c.predictive.Reset()
	for i := 0; i < c.Height(); i++ {
		if c.distal[i].HasActiveSegment(activeState, minOverlap) {
			// TODO(tms): if s.sequence == true
			c.predictive.Set(i)
		}
	}
}

func (c Column) FindBestSegment(state Bitset, minOverlap int, weak bool) (bestCell, bestSegment, bestOverlap int) {
	bestCell = -1
	bestSegment = -1
	bestOverlap = minOverlap - 1
	for i := 0; i < c.Height(); i++ {
		sIndex, sOverlap := c.distal[i].ComputeActive(state, minOverlap, weak)
		if sOverlap > bestOverlap {
			bestCell = i
			bestSegment = sIndex
			bestOverlap = sOverlap
		}
	}
	if htmLogger != nil {
		if bestSegment >= 0 {
			s := c.distal[bestCell].segments[bestSegment]
			htmLogger.Printf("\t\tFind best segment for column=%d, state=%v, minOverlap=%d, weak=%t: (%d, %d)=%04d <= %v (overlap=%d)",
				c.Index, state, minOverlap, weak, c.Index, bestCell, c.CellId(bestCell),
				s.Connected(), bestOverlap)
		}
	}
	return
}

func (c Column) LearningCellId() int {
	return c.CellId(c.learning)
}

func (c Column) CellId(i int) int {
	if i < 0 || i > c.Height() {
		panic(fmt.Errorf("Invalid index for cell: %d", i))
	}
	return c.Index*c.Height() + i
}

func (c *Column) AdaptSegments() {
	for i := 0; i < c.Height(); i++ {
		if c.distal[i].HasUpdates() {
			// TODO(tms): incorporate still predicted.
			if htmLogger != nil {
				htmLogger.Printf("\tAdapt prediction for cell %04d", c.CellId(i))
			}
			c.distal[i].ApplyAll(c.active.IsSet(i))
		}
	}
}

func (c *Column) LearnPrediction(state Bitset, minOverlap int) bool {
	c.learning = -1
	cell, sIndex, _ := c.FindBestSegment(state, minOverlap, false)
	if sIndex >= 0 {
		update := c.distal[cell].CreateUpdate(sIndex, state, minOverlap)
		c.distal[cell].AddUpdate(update)
		c.learning = cell
		if htmLogger != nil {
			htmLogger.Printf("\tLearning prediction %v => [%04d] (%d, %d)", update, c.CellId(cell), c.Index, cell)
		}
		return true
	}
	return false
}

func (c *Column) ConfirmPrediction(state Bitset) bool {
	if c.learning >= 0 && state.IsSet(c.CellId(c.learning)) {
		if htmLogger != nil {
			htmLogger.Printf("\t(%d, %d)=%04d: Confirmed prediction from lPredictive(t-1)", c.Index, c.learning, c.CellId(c.learning))
		}
		return true
	}
	return false
}

func (c *Column) LearnSequence(learnState Bitset) {
	if learnState.IsZero() {
		// Skip learning an empty sequence.
		if htmLogger != nil {
			htmLogger.Printf("\tSkip learning empty sequence.")
		}
		// Select the learning cell, but don't increment the target.
		c.learning = c.learningTarget
		return
	}
	cell, sIndex, _ := c.FindBestSegment(learnState, 1, true)
	if sIndex >= 0 {
		if htmLogger != nil {
			htmLogger.Printf("\t(%d, %d)=%04d: Will reinforce segment: %v", c.Index, cell, c.CellId(cell), c.distal[cell].Segment(sIndex))
		}
	} else {
		if htmLogger != nil {
			htmLogger.Printf("\t(%d, %d)=%04d: Will learn a new segment.", c.Index, c.learningTarget, c.CellId(c.learningTarget))
		}
		cell = c.learningTarget
		c.learningTarget = (c.learningTarget + 1) % c.Height()
		sIndex = -1
	}
	c.learning = cell
	update := c.distal[cell].CreateUpdate(sIndex, learnState, 1)
	if htmLogger != nil {
		htmLogger.Printf("\tLearning sequence %v => (%d, %d)=%04d", update, c.Index, cell, c.CellId(cell))
	}
	c.distal[cell].Apply(update, true)
}

func (c *Column) Activate() {
	// This is the inference part of the temporal pooler, Phase 1: if any cell is
	// predicted from the last step, we activate the predicted cells for this column.
	// Otherwise we activate all cells.
	if !c.predictive.IsZero() {
		c.active.ResetTo(*c.predictive)
	} else {
		// Bursting.
		c.active.SetRange(0, c.Height())
	}
}
