// Code for the column and potential pool.

package htm

import "fmt"
import "math/rand"
import "github.com/dukejeffrie/htm/data"
import "github.com/dukejeffrie/htm/log"
import "github.com/dukejeffrie/htm/segment"

var columnSource = rand.NewSource(1979)
var columnRand = rand.New(columnSource)

// Keeps information about a column in a cortical region.
type Column struct {
	Index int
	// The bitset of cells that are active.
	active *data.Bitset
	// The bitset of cells that are predicted.
	predictive *data.Bitset

	// Cell that was selected for learning.
	learning       int
	learningTarget int

	// Proximal dendrite segment.
	proximal *segment.DendriteSegment

	// Per-cell distal segment group.
	distal []*segment.DistalSegmentGroup
}

func NewColumn(inputSize, height int) *Column {
	result := &Column{
		active:         data.NewBitset(height),
		predictive:     data.NewBitset(height),
		proximal:       segment.NewDendriteSegment(inputSize),
		learning:       -1,
		learningTarget: 0,
		distal:         make([]*segment.DistalSegmentGroup, height),
	}
	for i := 0; i < height; i++ {
		result.distal[i] = segment.NewDistalSegmentGroup()
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

func (c Column) Connected() data.Bitset {
	return c.proximal.Connected()
}

func (c Column) Active() data.Bitset {
	return *c.active
}

func (c Column) Predictive() data.Bitset {
	return *c.predictive
}

func (c Column) Distal(i int) segment.DistalSegmentGroup {
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

func (c *Column) LearnFromInput(input data.Bitset, minOverlap int) {
	c.proximal.Learn(input, !c.Active().IsZero(), minOverlap)
}

func (c *Column) Predict(activeState data.Bitset, minOverlap int) {
	c.predictive.Reset()
	for i := 0; i < c.Height(); i++ {
		if c.distal[i].HasActiveSegment(activeState, minOverlap) {
			c.predictive.Set(i)
		}
	}
}

func (c Column) FindBestSegment(state data.Bitset, minOverlap int, weak bool) (bestCell, bestSegment, bestOverlap int) {
	bestCell = -1
	bestSegment = -1
	bestOverlap = minOverlap - 1
	maxOverlap := state.NumSetBits()
	for i := 0; i < c.Height(); i++ {
		sIndex, sOverlap := c.distal[i].ComputeActive(state, minOverlap, weak)
		if sOverlap > bestOverlap {
			bestCell = i
			bestSegment = sIndex
			bestOverlap = sOverlap
			if bestOverlap == maxOverlap {
				break
			}
		}
	}
	if log.HtmLogger.Enabled() {
		if bestSegment >= 0 {
			s := c.distal[bestCell].Segment(bestSegment)
			log.HtmLogger.Printf("\t\tFind best segment for column=%d, state=%v, minOverlap=%d, weak=%t: (%d, %d)=%04d <= %v (overlap=%d)",
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
	if c.learning >= 0 {
		if log.HtmLogger.Enabled() {
			log.HtmLogger.Printf("\tAdapt prediction for cell %04d",
				c.CellId(c.learning))
		}
		c.distal[c.learning].ApplyAll(c.active.IsSet(c.learning))
	}
}

func (c *Column) LearnPrediction(state data.Bitset, minOverlap int) bool {
	c.learning = -1
	cell, sIndex, _ := c.FindBestSegment(state, minOverlap, false)
	if sIndex >= 0 {
		update := c.distal[cell].CreateUpdate(sIndex, state, minOverlap)
		c.distal[cell].AddUpdate(update)
		c.learning = cell
		if log.HtmLogger.Enabled() {
			log.HtmLogger.Printf("\t(%d, %d)=%04d: Learning prediction: %v",
				c.CellId(cell), c.Index, cell, *update)
		}
		return true
	}
	return false
}

func (c *Column) ConfirmPrediction(state data.Bitset) bool {
	if c.learning >= 0 && state.IsSet(c.CellId(c.learning)) {
		if log.HtmLogger.Enabled() {
			log.HtmLogger.Printf("\t(%d, %d)=%04d: Confirmed prediction.",
				c.Index, c.learning, c.CellId(c.learning))
		}
		return true
	}
	return false
}

func (c *Column) LearnSequence(learnState data.Bitset) {
	logEnabled := log.HtmLogger.Enabled()
	if learnState.IsZero() {
		// Select the learning cell, but don't increment the target.
		c.learning = c.learningTarget
		if logEnabled {
			log.HtmLogger.Printf("\t(%d, %d)=%04d: skip learning empty sequence",
				c.Index, c.learning, c.CellId(c.learning))
		}
		return
	}
	cell, sIndex, _ := c.FindBestSegment(learnState, 1, true)
	if sIndex >= 0 {
		if logEnabled {
			log.HtmLogger.Printf("\t(%d, %d)=%04d: Will reinforce segment: %v",
				c.Index, cell, c.CellId(cell), c.distal[cell].Segment(sIndex))
		}
	} else {
		if logEnabled {
			log.HtmLogger.Printf("\t(%d, %d)=%04d: Will learn a new segment.",
				c.Index, c.learningTarget, c.CellId(c.learningTarget))
		}
		cell = c.learningTarget
		c.learningTarget = (c.learningTarget + 1) % c.Height()
		sIndex = -1
	}
	c.learning = cell
	update := c.distal[cell].CreateUpdate(sIndex, learnState, 1)
	if logEnabled {
		log.HtmLogger.Printf("\tLearning sequence %v => (%d, %d)=%04d",
			update, c.Index, cell, c.CellId(cell))
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
