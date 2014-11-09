package input

import "github.com/dukejeffrie/htm/data"
import "testing"

func TestRoundEncoder(t *testing.T) {
	s, err := NewScalarSensor(6, 2, 0.0, 10.0)
	if err != nil {
		t.Fatal(err)
	}
	s.Encode(0)
	if !s.Get().AllSet(0, 1) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(1)
	if !s.Get().AllSet(0, 1) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(2)
	if !s.Get().AllSet(1, 2) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(3)
	if !s.Get().AllSet(1, 2) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(4)
	if !s.Get().AllSet(2, 3) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(5)
	if !s.Get().AllSet(2, 3) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(6)
	if !s.Get().AllSet(3, 4) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(7)
	if !s.Get().AllSet(3, 4) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(8)
	if !s.Get().AllSet(4, 5) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	s.Encode(9)
	if !s.Get().AllSet(4, 5) || s.Get().NumSetBits() != 2 {
		t.Errorf("Encode failed: %v", *s)
	}
	err = s.Encode(10)
	if err == nil {
		t.Errorf("Should have failed, but didn't: %v", *s)
	}
}

func TestIntEncoder(t *testing.T) {
	s, err := NewScalarSensor(64, 2, -100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if s.MinValue != -100 || s.MaxValue != 100 {
		t.Error("Initialization error:", *s)
	}
	if s.BucketSize != 200.0/63.0 {
		t.Errorf("Bucket size should be %f, but is %f", 200.0/63.0, s.BucketSize)
	}
	// First bucket is [0..2], which corresponds to [-100..-98).
	if err = s.Encode(-100); err != nil {
		t.Fatal(err)
	}
	b := data.NewBitset(64).SetRange(0, 2)
	if !s.Get().Equals(*b) {
		t.Errorf("Encode failed. Expected: %v, but got: %v", *b, s.Get())
	}
	s.Encode(-99)
	if !s.Get().Equals(*b) {
		t.Errorf("Encode(%d) failed. Expected: %v, but got: %v", -99, *b, s.Get())
	}
	s.Encode(-98)
	if !s.Get().Equals(*b) {
		t.Errorf("Encode(%d) failed. Expected: %v, but got: %v", -98, *b, s.Get())
	}

	v := s.DecodeInt(*b)
	if v != -99 {
		t.Errorf("Decode failed. Expected: %v, but got: %v", -99, v)
	}

	// Next bucket is [1..3], which corresponds to [-97..-95).
	b.Reset().SetRange(1, 3)
	v = s.DecodeInt(*b)
	if v != -96 {
		t.Errorf("Decode failed. Expected: %v, but got: %v", -96, v)
	}

	// Last bucket is [62..64], which corresponds to [98..100)
	b.Reset().SetRange(62, 64)
	s.Encode(99)
	if !s.Get().Equals(*b) {
		t.Errorf("Encode(%d) failed. Expected: %v, but got: %v (%v)", 98, *b, s.Get(), *s)
	}
	if v = s.DecodeInt(*b); v != 98 {
		t.Errorf("Decode failed. Expected: %v, but got: %v (%v)", 98, v, *s)
	}
}

func TestFloatEncoder(t *testing.T) {
	s, err := NewScalarSensor(64, 3, -100, 100)
	if err != nil {
		t.Fatal(err)
	}
	s.Encode(-100)
	b := s.Get().Clone()
	v := s.Decode(*b)
	if v != -100.0+s.BucketSize/2 {
		t.Errorf("Decode failed: %f (sensor=%v)", v, *s)
	}
}

func TestSparseEncoder(t *testing.T) {
	_, err := NewScalarSensor(2048, 3, -100, 100)
	if err == nil {
		t.Fatal("Sensor cannot be too sparse.")
	}
}

func TestCategoryEncoder(t *testing.T) {
	s, err := NewCategorySensor(64, 4, "A", "B", "C")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Encode("A"); err != nil {
		t.Error(err)
	}
	b := data.NewBitset(64).SetRange(0, 4)
	if !s.Get().Equals(*b) {
		t.Errorf("Encode failed. Expected: %v, but got: %v", *b, s.Get())
	}
	if v := s.Decode(*b); v != "A" {
		t.Errorf("Decode failed. Expected: %v, but got: %v (%v)", "A", v, *s)
	}
	if err = s.Encode("B"); err != nil {
		t.Error(err)
	}
	b.Reset().SetRange(4, 8)
	if !s.Get().Equals(*b) {
		t.Errorf("Encode failed. Expected: %v, but got: %v", *b, s.Get())
	}
	if v := s.Decode(*b); v != "B" {
		t.Errorf("Decode failed. Expected: %v, but got: %v (%v)", "B", v, *s)
	}

	err = s.Encode("Other")
	if err == nil {
		t.Error("Should have failed to encode unknown category \"Other\":", *s)
	}
}
