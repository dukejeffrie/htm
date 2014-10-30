// This test is a sequence of numbers.

package test

import "github.com/dukejeffrie/htm"

import "fmt"
import "io"
import "os"
import "testing"

type Drop struct {
	Running  bool
	Input    htm.InputSource
	Output   io.Writer
	rawInput []*htm.RawInput
	region0  *htm.Region
	region1  *htm.Region
	region2  *htm.Region
	region3  *htm.Region
	step     int
	t        *testing.T
	patterns map[string]string
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
		Name:                 "0-drop",
		Width:                2000,
		Height:               9,
		MaximumFiringColumns: 40,
		MinimumInputOverlap:  0,
		InputLength:          64,
		Learning:             true,
	}
	d.step = 0
	d.region0 = htm.NewRegion(params)
	d.region0.ResetForInput(d.region0.InputLength, 20)

	params.Name = "1-drop"
	params.Width = 200
	params.Height = 8
	params.InputLength = d.region0.Output().Len()
	params.MaximumFiringColumns = 4
	d.region1 = htm.NewRegion(params)
	d.region1.ResetForInput(d.region1.InputLength, 40)

	params.Name = "final"
	params.Width = 64
	params.Height = 1
	params.InputLength = d.region1.Output().Len()
	params.MaximumFiringColumns = 2
	d.region3 = htm.NewRegion(params)
	d.region3.ResetForInput(d.region3.InputLength, 5)
	d.patterns = make(map[string]string)
}

func (d *Drop) SetLearning(learning bool) {
	d.region0.Learning = false
	d.region1.Learning = false
	//d.region2.Learning = false
	d.region3.Learning = false
	fmt.Fprintf(d.Output, "\nLearning = %t\n", learning)
}

func (d *Drop) Step() (recognized int) {
	d.step++
	input, err := d.Input.Next()
	if err != nil {
		d.t.Error(err)
		d.Running = false
		return
	}

	fmt.Fprintf(d.Output, "\n%d) Input = %v", d.step, input)
	//input.Value.Print(16, d.Output)
	d.region0.ConsumeInput(input.Value)
	d.region1.ConsumeInput(d.region0.Output())
	d.region3.ConsumeInput(d.region1.Output())
	val := d.region3.Output().String()
	if pat, ok := d.patterns[val]; ok {
		if pat == input.Name {
			fmt.Fprintf(d.Output, ", recognized as %s.", pat)
			recognized = 1
		} else {
			fmt.Fprintf(d.Output, ", mislabeled as %s.", pat)
			recognized = 0
		}
	} else {
		recognized = -1
		fmt.Fprintf(d.Output, ", new pattern\n")
		d.region0.Print(d.Output)
		d.region3.Print(d.Output)
		if d.region3.Learning {
			d.patterns[val] = input.Name
		}
	}
	return
}

func TestDrop(t *testing.T) {
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
	fmt.Printf("%v\n%v\n%v\n",
		drop.region0.RegionParameters,
		drop.region1.RegionParameters,
		drop.region3.RegionParameters)

	go drop.Generate()
	lastLearned := 0
	numRecognized := 0
	for numRecognized < 10 && drop.step-lastLearned < 100 && drop.step < 1000000 {
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

	numRecognized = 0
	for i := 0; i < 10; i++ {
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

	fmt.Printf("\nDone after %d steps, recognizing %d patterns.\n", lastStep, numRecognized)
	fmt.Fprintf(drop.Output, "\nDone after %d steps, recognizing %d patterns.\n", lastStep, numRecognized)
}
