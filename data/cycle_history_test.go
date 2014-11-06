package data

import "testing"

func TestCycleHistory(t *testing.T) {
	ch := NewCycleHistory(10)
	if avg, ok := ch.Average(); ok {
		t.Errorf("Should not be ok: %f", avg)
	}
	ch.Record(true)
	if avg, ok := ch.Average(); !ok || avg != 1.0 {
		t.Errorf("Should be %f average: %f, ok=%t", 1.0, avg, ok)
	}
	ch.Record(false)
	if avg, ok := ch.Average(); !ok || avg != 0.5 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.5, avg, ok, ch)
	}
	for i := 2; i < 10; i++ {
		ch.Record(false)
	}
	if avg, ok := ch.Average(); !ok || avg != 0.1 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.1, avg, ok, ch)
	}
	ch.Record(false)
	if avg, ok := ch.Average(); !ok || avg != 0.0 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.0, avg, ok, ch)
	}
	ch.Record(false)
	ch.Record(false)
	ch.Record(false)
	ch.Record(true)
	for i := 1; i < 10; i++ {
		ch.Record(false)
	}
	if avg, ok := ch.Average(); !ok || avg != 0.1 {
		t.Errorf("Should be %f average: %f, ok=%t, %v", 0.1, avg, ok, ch)
	}
}
