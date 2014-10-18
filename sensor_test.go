// Unit tests for sensor calls.

package htm

import "fmt"
import "testing"

const (
	BLUE   = iota
	RED    = iota
	GREEN  = iota
	PURPLE = iota
)

type TestSource struct {
	*InputSource
	Sink chan<- *RawInput
}

func (s *TestSource) SendColor(name string, value int) (*Input, error) {
	s.Sink <- &RawInput{name, value}
	input, err := s.Next()
	fmt.Printf("For color %s(%d): %v, %v\n", name, value, input, err)
	return input, err
}

func GenerateColors(n, w int) (*TestSource, error) {
	source := make(chan *RawInput, 1)
	decoder, err := NewCategoryDecoder(n, w)
	if err != nil {
		return nil, err
	}
	delegate := &InputSource{
		Name:    "Test Colors",
		Decoder: decoder,
		Source:  source,
	}
	return &TestSource{delegate, source}, nil
}

func TestColors(t *testing.T) {
	sink, err := GenerateColors(2048, 28)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	b1, err := sink.SendColor("blue", BLUE)
	if err != nil {
		t.Error(err)
	}
	t.Log(b1)

	r1, err := sink.SendColor("red", RED)
	if err != nil {
		t.Error(err)
	}
	t.Log(r1)
	if b1.Value.Equals(r1.Value) {
		t.Errorf("blue and red must be different: %v == %v", b1, r1)
	}

	b2, err := sink.SendColor("blue", BLUE)
	if err != nil {
		t.Error(err)
	}
	if !b1.Value.Equals(b2.Value) {
		t.Errorf("blue values don't match: %v != %v", b1, b2)
	}

	r2, err := sink.SendColor("red", RED)
	if err != nil {
		t.Error(err)
	}
	t.Log(r2)
	if !r1.Value.Equals(r2.Value) {
		t.Errorf("red values don't match: %v != %v", r1, r2)
	}
}
