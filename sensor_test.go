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
	return input, err
}

func (s *TestSource) SendScalar(value int) (*Input, error) {
	s.Sink <- &RawInput{fmt.Sprint(value), value}
	input, err := s.Next()
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

func GenerateScalars(n, w, min, max int) (*TestSource, error) {
	source := make(chan *RawInput, 1)
	decoder, err := NewScalarDecoder(n, w, min, max)
	if err != nil {
		return nil, err
	}
	delegate := &InputSource{
		Name:    "Test Scalars",
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

func TestScalars(t *testing.T) {
	sink, _ := GenerateScalars(2048, 28, 0, 100)
	s1, _ := sink.SendScalar(4)
	b1 := NewBitset(2048)
	b1.SetRange(4, 4+28)
	if !s1.Value.Equals(*b1) {
		t.Errorf("Scalar mismatch: expected %v, but got: %v", b1, s1)
	}
}
