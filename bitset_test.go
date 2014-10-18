// Unit tests for sensor implementation

package htm

import "fmt"
import "testing"
import "math/rand"

func ExpectEquals(t *testing.T, message string, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Errorf("%s: expected %v, but was: %v", message, expected, actual)
	}
}

func ExpectContentEquals(t *testing.T, message string, expected []int, actual []int) {
	if len(expected) != len(actual) {
		t.Errorf("%s: length differs (%v != %v), actual: %v",
			message, len(expected), len(actual), actual)
	}
	failed := false
	for i, v := range expected {
		if v != actual[i] {
			t.Logf("%s (element[%d]): %v != %v", message, v, actual[i])
			failed = true
		}
	}
	if failed {
		t.Fail()
	}
}

func TestConstruction(t *testing.T) {
	one := NewBitset(2)
	ExpectEquals(t, "one.length", 2, one.length)
	ExpectEquals(t, "one.binary.length", 1, len(one.binary))

	cent := NewBitset(100)
	ExpectEquals(t, "cent.length", 100, cent.length)
	ExpectEquals(t, "cent.binary.length", 2, len(cent.binary))

	two := NewBitset(128)
	ExpectEquals(t, "two.binary.length", 2, len(two.binary))

	three := NewBitset(129)
	ExpectEquals(t, "three.binary.length", 3, len(three.binary))
}

func TestSetVersusOr(t *testing.T) {
	b := NewBitset(2048)

	b.Set([]int{127})
	ExpectEquals(t, "bit 127", true, b.IsSet(127))

	b.Set([]int{11, 12})
	ExpectEquals(t, "bit 127", true, b.IsSet(127))
	ExpectEquals(t, "num bits", 3, b.NumSetBits())

	b.Reset()
	ExpectEquals(t, "bit 127", false, b.IsSet(127))

	b.Set([]int{222, 444, 888, 1023, 1024, 1331, 2047})
	ExpectEquals(t, "num bits", 7, b.NumSetBits())
}

func TestLocate(t *testing.T) {
	b := NewBitset(128)

	for i := 0; i < 64; i++ {
		pos, off := b.locate(i)
		ExpectEquals(t, fmt.Sprintf("pos[%d]", i), 0, pos)
		ExpectEquals(t, fmt.Sprintf("off[%d]", i), i, off)
	}
	for i := 64; i < 128; i++ {
		pos, off := b.locate(i)
		ExpectEquals(t, fmt.Sprintf("pos[%d]", i), 1, pos)
		ExpectEquals(t, fmt.Sprintf("off[%d]", i), i-64, off)
	}
}

func TestIndexing(t *testing.T) {
	b := NewBitset(64)
	b.Set([]int{1, 3})

	ExpectEquals(t, "first integer", uint64(8+2), b.binary[0])

	ExpectEquals(t, "bit 1", true, b.IsSet(1))
	ExpectEquals(t, "bit 0", false, b.IsSet(0))
	ExpectEquals(t, "bit 3", true, b.IsSet(3))
	ExpectEquals(t, "two bits set", 2, b.NumSetBits())

	sl := make([]int, 10)
	ExpectContentEquals(t, "indices 1 and 3", []int{1, 3}, b.ToIndexes(sl))
}

func BenchmarkSet64(b *testing.B) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(64)
	bits_to_set := []int{rand.Intn(64), rand.Intn(64), rand.Intn(64)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.Set(bits_to_set)
	}
}

func BenchmarkSet128(b *testing.B) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(128)
	bits_to_set := []int{rand.Intn(128), rand.Intn(128), rand.Intn(128),
		rand.Intn(128), rand.Intn(128), rand.Intn(128)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.Set(bits_to_set)
	}
}

func BenchmarkSet2048(b *testing.B) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(2048)
	bits_to_set := []int{rand.Intn(2048), rand.Intn(2048), rand.Intn(2048),
		rand.Intn(2048), rand.Intn(2048), rand.Intn(2048)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.Set(bits_to_set)
	}
}
