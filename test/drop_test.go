// This test is a sequence of numbers.

package test

import "github.com/dukejeffrie/htm"
import "github.com/dukejeffrie/htm/data"
import "github.com/dukejeffrie/htm/log"

import "fmt"
import "math/rand"
import "io"
import "os"
import "testing"

type Drop struct {
	Running         bool
	Input           htm.InputSource
	Output          io.Writer
	rawInput        []*htm.RawInput
	region0         *htm.Region
	region1         *htm.Region
	region2         *htm.Region
	region3         *htm.Region
	predicted       *data.Bitset
	step            int
	t               *testing.T
	patterns        map[string]string
	LearnQuietUntil int
}

func (d *Drop) Generate() {
	d.rawInput = make([]*htm.RawInput, 10)
	max := 100 * len(d.rawInput) * len(d.rawInput)
	for i := 1; i <= len(d.rawInput); i++ {
		den := i * i
		d.rawInput[i-1] = &htm.RawInput{
			Name:     fmt.Sprintf("%d", max/den),
			IntValue: max / den,
		}
	}
	n := 0
	for d.Running {
		d.Input.Source <- d.rawInput[n]
		n = (n + 1) % len(d.rawInput)
	}
}

func (d *Drop) InitializeNetwork() {
	params := htm.RegionParameters{
		Name:                "0-drop",
		Width:               640,
		Height:              8,
		MinimumInputOverlap: 1,
		InputLength:         64,
		Learning:            true,
	}
	params.MaximumFiringColumns = params.Width / 50
	d.step = 0
	d.predicted = data.NewBitset(params.InputLength)
	d.region0 = htm.NewRegion(params)
	d.region0.RandomizeColumns(params.InputLength / 2)

	params.Name = "1-drop"
	params.Width = 2000
	params.Height = 5
	params.InputLength = d.region0.Output().Len()
	params.MaximumFiringColumns = params.Width / 100
	d.region1 = htm.NewRegion(params)
	d.region1.RandomizeColumns(params.InputLength / 2)

	params.Name = "final"
	params.Width = 200
	params.Height = 3
	params.InputLength = d.region1.Output().Len()
	params.MaximumFiringColumns = 1
	d.region3 = htm.NewRegion(params)
	d.region3.RandomizeColumns(params.InputLength / 2)
	d.patterns = make(map[string]string)
}

func (d *Drop) SetLearning(learning bool) {
	d.region0.Learning = false
	d.region1.Learning = false
	//d.region2.Learning = false
	d.region3.Learning = false
	log.HtmLogger.Print("Learning = false")
}

func (d *Drop) AddNoise() {
	noise := data.NewBitset(d.region0.InputLength)
	noise.Set(rand.Intn(noise.Len()), rand.Intn(noise.Len()))

	d.region0.ConsumeInput(*noise)
	d.region1.ConsumeInput(d.region0.Output())
	d.region3.ConsumeInput(d.region1.Output())
}

func (d *Drop) Step() (recognized int) {
	d.step++
	input, err := d.Input.Next()
	if err != nil {
		d.t.Error(err)
		d.Running = false
		return
	}

	recognized = 0
	//input.Value.Print(16, d.Output)
	d.region0.ConsumeInput(input.Value)
	d.region1.ConsumeInput(d.region0.Output())
	d.region3.ConsumeInput(d.region1.Output())
	fmt.Fprintf(d.Output, "\n%d) Input(%s) = %v", d.step, input.Name, input.Value)
	if d.predicted.IsZero() {
		recognized = -1
		fmt.Fprintf(d.Output, ", new: %v", d.region3.ActiveState())
	} else if d.predicted.Equals(d.region3.ActiveState()) {
		fmt.Fprintf(d.Output, ", predicted: %v", *d.predicted)
		recognized = 1
	} else {
		fmt.Fprintf(d.Output, ", missed (expected = %v, got: %v)", *d.predicted, d.region3.ActiveState())
		recognized = 0
	}
	d.predicted = d.region3.PredictiveState().Clone()
	return
}

func TestDrop(t *testing.T) {
	log.HtmLogger.SetEnabled(true)
	dec, err := htm.NewScalarDecoder(64, 2, 0, 12000)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	src := htm.NewInputSource("Drop", dec)
	out, err := os.Create("drop_output.txt")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	drop := Drop{
		Running: true,
		Input:   src,
		Output:  out,
		t:       t}
	drop.InitializeNetwork()
	fmt.Printf("%+v\n%+v\n%+v\n",
		drop.region0.RegionParameters,
		drop.region1.RegionParameters,
		drop.region3.RegionParameters)

	go drop.Generate()
	lastLearned := 0
	numRecognized := 0
	drop.LearnQuietUntil = 100
	for numRecognized < 20 && drop.step-lastLearned < 1000 && drop.step < 1000000 {
		recognized := drop.Step()
		if recognized == 1 {
			numRecognized++
			fmt.Print("v")
		} else if recognized == 0 {
			numRecognized = 0
			fmt.Print("x")
		} else {
			numRecognized = 0
			lastLearned = drop.step
			fmt.Print(".")
		}
	}

	lastStep := drop.step
	drop.SetLearning(false)
	fmt.Printf("\nLearning = %t\n", false)

	numTested := 50
	for i := 0; i < numTested/3; i++ {
		drop.AddNoise()
	}
	numRecognized = 0
	for i := 0; i < numTested; i++ {
		recognized := drop.Step()
		if recognized == 1 {
			numRecognized++
			fmt.Print("v")
		} else if recognized == 0 {
			fmt.Print("x")
		} else {
			fmt.Print("x")
		}
	}

	fmt.Printf("\nDone after %d steps, predicting %d/%d steps.\n", lastStep, numRecognized, numTested)
	fmt.Fprintf(drop.Output, "\nDone after %d steps, predicting %d/%d patterns.\n", lastStep, numRecognized, numTested)
}
