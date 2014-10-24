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
	d.region0 = htm.NewRegion("0-drop", 100, 8, 0.05)
	d.region0.ResetForInput(64, 2)
	d.region3 = htm.NewRegion("final", 64, 1, 0.05)
	d.region3.ResetForInput(d.region0.Output().Len(), 5)
	d.patterns = make(map[string]string)
}

func (d *Drop) Step() {
	d.step++
	input, err := d.Input.Next()
	if err != nil {
		d.t.Error(err)
		d.Running = false
		return
	}

	fmt.Fprintf(d.Output, "\n>>> Step %d: %v\n", d.step, input)
	//input.Value.Print(16, d.Output)
	d.region0.ConsumeInput(input.Value)
	d.region3.ConsumeInput(d.region0.Output())
	val := d.region3.Output().String()
	if pat, ok := d.patterns[val]; ok {
		fmt.Fprintf(d.Output, "\nRecognized as %s", pat)
	} else {
		fmt.Fprintf(d.Output, "\nNew pattern for input: %s\n", input.Name)
		d.region0.Print(d.Output)
		d.region3.Print(d.Output)
		d.patterns[val] = input.Name
	}
}

func TestDrop(t *testing.T) {
	dec, err := htm.NewScalarDecoder(64, 2, 0, 12000)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	src := htm.NewInputSource("Drop", dec)
	drop := Drop{
		Running: true,
		Input:   src,
		Output:  os.Stdout,
		t:       t}
	drop.InitializeNetwork()
	go drop.Generate()
	for drop.step < 100 {
		drop.Step()
	}
	fmt.Fprintf(drop.Output, "\nDone.\n")
}
