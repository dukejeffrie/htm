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
	layer0   *htm.Layer
	step     int
	t        *testing.T
}

func (d *Drop) Generate() {
	d.rawInput = make([]*htm.RawInput, 10)
	max := 100 * len(d.rawInput) * len(d.rawInput)
	for i := 0; i < len(d.rawInput); i++ {
		den := i * i
		if den == 0 {
			den = 1
		}
		d.rawInput[i] = &htm.RawInput{
			Name:     fmt.Sprintf("Drop[%d]=%d", i, max/den),
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
	d.layer0 = htm.NewLayer("0-drop", 100, 9, 0.05)
	d.layer0.ResetForInput(64, 2)
}

func (d *Drop) Step() {
	d.step++
	input, err := d.Input.Next()
	if err != nil {
		d.t.Error(err)
		d.Running = false
		return
	}

	fmt.Fprintf(d.Output, "\n\v>>> Step %d: %v\n", d.step, input)
	input.Value.Print(16, d.Output)
	d.layer0.ConsumeInput(input.Value)
	d.layer0.Print(d.Output)
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
	drop.Step()
	drop.Step()
	drop.Step()
}
