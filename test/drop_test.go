// This test is a sequence of numbers.

package test

import "github.com/dukejeffrie/htm"
import "github.com/dukejeffrie/htm/data"
import "github.com/dukejeffrie/htm/input"
import "github.com/dukejeffrie/htm/log"

import "fmt"
import "math/rand"
import "io"
import "os"
import "testing"

type Drop struct {
	Output          io.Writer
	rawInput        []int
	Sensor          *input.ScalarSensor
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
	d.rawInput = make([]int, 100)
	delta := int(d.Sensor.MaxValue - 0.1 - d.Sensor.MinValue)
	for i := 1; i <= len(d.rawInput); i++ {
		den := i * i
		d.rawInput[i-1] = int(d.Sensor.MinValue) + delta/den
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
	d.predicted = data.NewBitset(d.region3.PredictiveState().Len())
	d.patterns = make(map[string]string)
}

func (d *Drop) SetLearning(learning bool) {
	d.region0.Learning = false
	d.region1.Learning = false
	//d.region2.Learning = false
	d.region3.Learning = false
	log.HtmLogger.Print("Learning = false")
	fmt.Fprintf(d.Output, "\nLearning = %t\n", false)
}

func (d *Drop) AddNoise() {
	noise := data.NewBitset(d.region0.InputLength)
	noise.Set(rand.Intn(noise.Len()), rand.Intn(noise.Len()))

	d.region0.ConsumeInput(*noise)
	d.region1.ConsumeInput(d.region0.Output())
	d.region3.ConsumeInput(d.region1.Output())
}

func (d *Drop) Step() (recognized int) {
	idx := d.step % len(d.rawInput)
	err := d.Sensor.Encode(d.rawInput[idx])
	if err != nil {
		d.t.Fatal(err)
		return
	}
	d.step++

	recognized = 0
	//input.Value.Print(16, d.Output)
	d.region0.ConsumeInput(d.Sensor.Get())
	d.region1.ConsumeInput(d.region0.Output())
	d.region3.ConsumeInput(d.region1.Output())
	fmt.Fprintf(d.Output, "\n%d) Input(%v) = %v", d.step, d.Sensor.Raw(), d.Sensor.Get())
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
	d.predicted.ResetTo(d.region3.PredictiveState())
	return
}

func TestDrop(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	log.HtmLogger.SetEnabled(true)
	dec, err := input.NewScalarSensor(64, 2, 0, 1200000)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	out, err := os.Create("drop_output.txt")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	drop := Drop{
		Sensor: dec,
		Output: out,
		t:      t}
	drop.InitializeNetwork()
	fmt.Printf("%+v\n%+v\n%+v\n",
		drop.region0.RegionParameters,
		drop.region1.RegionParameters,
		drop.region3.RegionParameters)

	drop.Generate()
	lastLearned := 0
	numRecognized := 0
	drop.LearnQuietUntil = 100
	for numRecognized < 40 && drop.step-lastLearned < 1000 && drop.step < 10000 {
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
