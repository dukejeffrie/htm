// Sensorial region. Converts a real world data into a binary representation.

package htm

import "fmt"
import "math/rand"

// This is one input
type RawInput struct {
	Name     string
	IntValue int
}

type DecoderFunction func(in RawInput) (out Bitset, err error)

type Input struct {
	Name  string
	Value Bitset
}

func (i Input) String() string {
	return fmt.Sprintf("%s) %v", i.Name, i.Value)
}

type InputSource struct {
	Name    string
	Decoder DecoderFunction
	Source  chan *RawInput
}

func NewInputSource(name string, dec DecoderFunction) InputSource {
	return InputSource{
		Name:    name,
		Decoder: dec,
		Source:  make(chan *RawInput),
	}
}

func (c *InputSource) Next() (result *Input, err error) {
	next := <-c.Source
	if next == nil {
		return nil, nil
	}
	result = new(Input)
	result.Name = next.Name
	result.Value, err = c.Decoder(*next)
	return
}

func NewCategoryDecoder(n, w int) (DecoderFunction, error) {
	if n < w*4 {
		return nil, fmt.Errorf("Cannnot create category decoder: n = %d is too small (must be at least 4 times %d)", n, w)
	}
	bits := make(map[int]Bitset)
	return func(in RawInput) (out Bitset, err error) {
		key := in.IntValue
		var ok bool
		out, ok = bits[key]
		if !ok {
			out = *NewBitset(n)
			for i := 0; i < w; i++ {
				out.Set(rand.Intn(n))
			}
			// In the rare case where rand produces the same number twice, we'll be missing a few. Keep trying until we're good.
			for out.NumSetBits() < w {
				out.Set(rand.Intn(n))
			}
			bits[key] = out
		}
		return
	}, nil
}

func NewScalarDecoder(n, w, min, max int) (DecoderFunction, error) {
	bucket_size := (max - min) / (n - w + 1)
	if bucket_size < 1 {
		bucket_size = 1
	}
	return func(in RawInput) (out Bitset, err error) {
		val := in.IntValue
		bucket := (val - min) / bucket_size
		out = *NewBitset(n)
		out.SetRange(bucket, bucket+w)
		return out, nil
	}, nil
}
