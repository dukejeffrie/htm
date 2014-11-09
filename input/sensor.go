// Sensors convert real world data into a sparse binary representation.

package input

import "math"
import "fmt"
import "github.com/dukejeffrie/htm/data"

type Sensor struct {
	// The number of bits this sensor produces for each input
	N int
	// The approximate number of bits that should be set for an input
	W int

	// The last sensed value, both as bits as raw input
	value *data.Bitset
	input interface{}
}

func NewSensor(n, w int) *Sensor {
	return &Sensor{n, w, data.NewBitset(n), nil}
}

func (s Sensor) String() string {
	return fmt.Sprint(">>", s.input, "=", *s.value)
}

func (s Sensor) Get() data.Bitset {
	return *s.value
}

func (s Sensor) Raw() interface{} {
	return s.input
}

type ScalarSensor struct {
	*Sensor
	MaxValue   float64
	MinValue   float64
	BucketSize float64
}

func (s ScalarSensor) String() string {
	return fmt.Sprint(*s.Sensor, "[", s.MinValue, "..", s.MaxValue, "/",
		s.BucketSize, "]")
}

func (s *ScalarSensor) Encode(value interface{}) error {
	s.input = value
	s.value.Reset()
	switch value := value.(type) {
	case int:
		return s.EncodeInt(value)
	case float64:
		return s.EncodeFloat(value)
	default:
		return fmt.Errorf("Cannot encode values of type %T (%v).", value, value)
	}
}

func (s *ScalarSensor) EncodeFloat(value float64) error {
	if value < s.MinValue || value >= s.MaxValue {
		return fmt.Errorf("Precondition failed: min (%f) <= value (%f) < max (%f).",
			s.MinValue, value, s.MaxValue)
	}

	bucket := int(math.Floor((value - s.MinValue) / s.BucketSize))
	s.value.SetRange(bucket, bucket+s.W)
	return nil
}

func (s *ScalarSensor) EncodeInt(value int) error {
	return s.EncodeFloat(float64(value))
}

func (s ScalarSensor) Decode(bits data.Bitset) interface{} {
	found := bits.Any(func(int) bool {
		return true
	})

	return (0.5+float64(found))*s.BucketSize + s.MinValue
}

func (s ScalarSensor) DecodeInt(bits data.Bitset) int {
	floatResult := s.Decode(bits).(float64)
	return int(math.Floor(floatResult))
}

func NewScalarSensor(n, w int, min, max float64) (*ScalarSensor, error) {
	BucketSize := (max - min) / float64(n-w+1)
	if BucketSize < 1.0 {
		err := fmt.Errorf("Not enough buckets. Increase range or decrease length.")
		return nil, err
	}
	result := ScalarSensor{
		Sensor:     NewSensor(n, w),
		MaxValue:   max,
		MinValue:   min,
		BucketSize: BucketSize,
	}
	return &result, nil
}

type CategorySensor struct {
	*Sensor
	categories map[string]int
	reverse    []string
}

func (s *CategorySensor) Encode(value interface{}) error {
	s.input = value
	s.value.Reset()
	switch value := value.(type) {
	case string:
		return s.EncodeString(value)
	default:
		return fmt.Errorf("Cannot encode values of type %T (%v).", value, value)
	}
}

func (s *CategorySensor) Decode(bits data.Bitset) interface{} {
	found := bits.Any(func(int) bool {
		return true
	})
	return s.reverse[found/s.W]
}

func (s *CategorySensor) EncodeString(cat string) error {
	id, ok := s.categories[cat]
	if !ok {
		return fmt.Errorf(
			"Unknown category \"%s\" in sensor: %v", cat, *s)
	}
	s.value.SetRange((id-1)*s.W, id*s.W)
	return nil
}

func NewCategorySensor(n, w int, categories ...string) (*CategorySensor, error) {
	result := &CategorySensor{
		Sensor:     NewSensor(n, w),
		categories: make(map[string]int, 64),
		reverse:    make([]string, len(categories)),
	}
	for i, c := range categories {
		result.categories[c] = i + 1
		result.reverse[i] = c
	}
	return result, nil
}
