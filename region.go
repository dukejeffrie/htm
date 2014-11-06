// A region is a group of cells of the same hierarchy within a region. It is called "local neighborhood" in Numenta code.

package htm

import "bufio"
import "container/heap"
import "flag"
import "fmt"
import "github.com/dukejeffrie/htm/data"
import "io"
import "log"
import "os"

var htmLogger *log.Logger
var htmLoggerEnabled = flag.Bool(
	"enable_htm_trace_log", false,
	"whether to enable trace logging of the htm execution.")

func initLogger() {
	if *htmLoggerEnabled {
		if htmLogger == nil {
			htmLogger = log.New(os.Stderr, "htm) ", 0)
		}
	} else {
		htmLogger = nil
	}
}

type ScoredElement struct {
	index int
	score float32
}
type TopN []ScoredElement

func (t TopN) Len() int {
	return len(t)
}

func (t TopN) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TopN) Less(i, j int) bool {
	return t[i].score < t[j].score
}

func (t *TopN) Push(x interface{}) {
	el := x.(ScoredElement)
	*t = append(*t, el)
}

func (t *TopN) Pop() interface{} {
	n := t.Len() - 1
	el := (*t)[n]
	*t = (*t)[0:n]
	return el
}

// Parameters to describe a region.
type RegionParameters struct {
	// The name of the region, used for debugging.
	Name string
	// Whether learning is on or off.
	Learning bool
	// Number of cells in each column.
	Height int
	// Number of columns in this region.
	Width int
	// Size of the input, in bits.
	InputLength int
	// Maximum number of columns that can fire.
	MaximumFiringColumns int
	// Minimum overlap between an input and a column's proximal dentrite to trigger
	// activation.
	MinimumInputOverlap int
}

type Region struct {
	RegionParameters
	columns              []*Column
	output               *data.Bitset
	active               *data.Bitset
	lastActive           *data.Bitset
	predictive           *data.Bitset
	lastPredictive       *data.Bitset
	learnActiveState     *data.Bitset
	learnActiveStateLast *data.Bitset
	learnPredictiveState *data.Bitset
	scores               TopN
}

// Creates a new named region with the given parameters.
func NewRegion(params RegionParameters) *Region {
	initLogger()
	result := &Region{
		RegionParameters:     params,
		columns:              make([]*Column, params.Width),
		output:               data.NewBitset(params.Width * params.Height),
		active:               data.NewBitset(params.Width * params.Height),
		lastActive:           data.NewBitset(params.Width * params.Height),
		predictive:           data.NewBitset(params.Width * params.Height),
		lastPredictive:       data.NewBitset(params.Width * params.Height),
		learnActiveState:     data.NewBitset(params.Width * params.Height),
		learnActiveStateLast: data.NewBitset(params.Width * params.Height),
		learnPredictiveState: data.NewBitset(params.Width * params.Height),
		scores:               make([]ScoredElement, 0, params.MaximumFiringColumns+1),
	}
	for i := 0; i < params.Width; i++ {
		result.columns[i] = NewColumn(params.InputLength, params.Height)
		result.columns[i].Index = i
	}
	if htmLogger != nil {
		htmLogger.Printf("Region created: %+v", params)
	}
	return result
}

func (l Region) Height() int {
	return l.RegionParameters.Height
}

func (l Region) Width() int {
	return l.RegionParameters.Width
}

func (l Region) Column(i int) Column {
	return *l.columns[i]
}

func (l Region) ActiveState() data.Bitset {
	return *l.active
}

func (l Region) PredictiveState() data.Bitset {
	return *l.predictive
}

func (l Region) LearningActiveState() data.Bitset {
	return *l.learnActiveState
}

func (l Region) LearningPredictiveState() data.Bitset {
	return *l.learnPredictiveState
}

func (l *Region) RandomizeColumns(w int) {
	perm := make([]int, w)
	for _, col := range l.columns {
		for i := 0; i < w; i++ {
			perm[i] = columnRand.Intn(l.InputLength)
		}
		col.ResetConnections(perm)
		col.SetBoost(columnRand.Float32() * 0.00001)
	}
}

func (l *Region) ResetColumnSynapses(i int, indices ...int) {
	col := l.columns[i]
	col.ResetConnections(indices)
	col.SetBoost(columnRand.Float32() * 0.00001)
}

func (l *Region) SensedInput() data.Bitset {
	dest := data.NewBitset(l.InputLength)
	for _, col := range l.columns {
		if !col.Active().IsZero() {
			dest.Or(col.Connected())
		}
	}
	return *dest
}

func (l *Region) FeedBack(output data.Bitset) *data.Bitset {
	dest := data.NewBitset(l.InputLength)
	output.Foreach(func(cellId int) {
		col := l.columns[cellId/l.Height()]
		cell := cellId % l.Height()
		if !col.Distal(cell).HasActiveSegment(output, l.MinimumInputOverlap) {
			// Not a predicted cell, must be active from fast-forward.
			dest.Or(col.Connected())
		}
	})
	return dest
}

func (l *Region) PredictedInput() data.Bitset {
	dest := data.NewBitset(l.InputLength)
	for _, col := range l.columns {
		if !col.Predictive().IsZero() {
			dest.Or(col.Connected())
		}
	}
	return *dest
}

func (l *Region) ConsumeInput(input data.Bitset) {
	if htmLogger != nil {
		htmLogger.Printf("\n============ %s Consume(learning=%t, input=%v)",
			l.Name, l.Learning, input)
	}
	l.scores = l.scores[0:0]
	for i, c := range l.columns {
		c.active.Reset()
		overlapScore := c.Connected().Overlap(input)
		if overlapScore >= l.MinimumInputOverlap {
			score := float32(overlapScore) + c.Boost()
			heap.Push(&l.scores, ScoredElement{i, score})
			if l.scores.Len() > l.MaximumFiringColumns {
				heap.Pop(&l.scores)
			}
		}
	}

	// 1) For each active column, check for cells that are in a predictive state and
	// activate them. If no cells are in a predictive state, activate all the cells in
	// the column (burst).
	l.lastActive.ResetTo(*l.active)
	l.active.Reset()
	for _, el := range l.scores {
		col := l.columns[el.index]
		col.Activate()
		l.active.SetFromBitsetAt(col.Active(), el.index*col.Height())
	}

	// 2) Cells with active dendrite segments are put in the predictive state.
	l.lastPredictive.ResetTo(*l.predictive)
	l.predictive.Reset()
	for _, col := range l.columns {
		col.Predict(*l.active, l.MinimumInputOverlap)
		l.predictive.SetFromBitsetAt(col.Predictive(), col.Index*col.Height())
	}
	// The output for the next level is the union of active and predicted cells.
	l.output.ResetTo(*l.active)
	l.output.Or(*l.predictive)
	if htmLogger != nil {
		htmLogger.Printf("Inference finished.\n\tOutput(t): %v\n\tActive(t): %v\n\tPredictive(t):%v\n",
			*l.output, *l.active, *l.predictive)
	}

	if l.Learning {
		l.Learn(input)
	}
}

func (l *Region) Output() data.Bitset {
	return *l.output
}

func (l *Region) Learn(input data.Bitset) {
	// Temporal pooler learning. Learn states are a subsample of the full state, with
	// hand-picked bits comprised of one cell per column.

	// 3) process segment updates (yes, we do it before 1 and 2).
	for _, col := range l.columns {
		col.AdaptSegments()
	}

	// 1) Learn that the last active state predicts this active state.
	if htmLogger != nil {
		htmLogger.Printf("Learning actual sequences...\n\tlActive(t-1): %v\n\tlPredictive(t-1): %v\n",
			*l.learnActiveState, *l.learnPredictiveState)
	}

	l.learnActiveStateLast.ResetTo(*l.learnActiveState)
	l.learnActiveState.Reset()
	for _, el := range l.scores {
		col := l.columns[el.index]
		if !col.ConfirmPrediction(*l.learnPredictiveState) {
			col.LearnSequence(*l.learnActiveStateLast)
		}
		l.learnActiveState.Set(col.LearningCellId())
	}
	// 2) Select one cell per column to learn the transition from the current input to
	// the next input
	if htmLogger != nil {
		htmLogger.Printf("Learning predictions...\n\tActive(t): %v\n\tlActive(t): %v\n",
			*l.active, *l.learnActiveState)
	}
	l.learnPredictiveState.Reset()
	for _, col := range l.columns {
		if col.LearnPrediction(*l.learnActiveState, l.MinimumInputOverlap) {
			l.learnPredictiveState.Set(col.LearningCellId())
		}
	}

	if htmLogger != nil {
		htmLogger.Printf("Sequence learner finished.\n\tlPredictive(t): %v",
			*l.learnPredictiveState)
	}

	// Spatial pooler learning.
	for _, col := range l.columns {
		col.LearnFromInput(input, l.MinimumInputOverlap)
	}
}

func (l Region) ToRune(cellId int) (r rune) {
	if l.active.IsSet(cellId) {
		if l.lastPredictive.IsSet(cellId) {
			r = 'v'
		} else {
			r = '!'
		}
	} else if l.lastPredictive.IsSet(cellId) {
		r = 'o'
	} else {
		r = '-'
	}
	return
}

func (l Region) Print(w io.Writer) error {
	writer := bufio.NewWriter(w)
	fmt.Fprintf(writer, "\n=== %s (learning: %t) ===\n", l.Name, l.Learning)
	line := 0
	rS := 20
	rL := 80
	// TODO(tms): dynamic values for rL and rS

	tabFormat := fmt.Sprintf("%%-%dd ", rS)
	for j := 0; j < rL; j += rS {
		fmt.Fprintf(writer, tabFormat, (line*rL)+j)
	}
	for i := 0; i < l.Width()*l.Height(); i++ {
		if i%rL == 0 {
			writer.WriteRune('\n')
			if line > 0 {
				fmt.Fprintf(writer, tabFormat, line*rL)
				writer.WriteRune('\n')
			}
			line++
		} else if i%rS == 0 {
			writer.WriteRune(' ')
		}
		writer.WriteRune(l.ToRune(i))
	}
	writer.WriteRune('\n')
	return writer.Flush()
}
