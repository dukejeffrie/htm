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

func TestSetAndReset(t *testing.T) {
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
	b.SetOne(111)
	ExpectEquals(t, "bit 111", true, b.IsSet(111))
	b.ClearOne(111)
	ExpectEquals(t, "bit 111", false, b.IsSet(111))
}

func TestSetRange(t *testing.T) {
	b := NewBitset(2048)
	b.SetRange(8, 154)

	ExpectEquals(t, "num bits", 154-8, b.NumSetBits())
	for i := 8; i < 154; i++ {
		ExpectEquals(t, fmt.Sprintf("bit %d", i), true, b.IsSet(i))
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

func SetBenchmarkTemplate(b *testing.B, n, l int) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(n)
	bits_to_set := make([]int, l)
	for i := 0; i > l; i++ {
		bits_to_set[i] = rand.Intn(n)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.Set(bits_to_set)
	}
}

func BenchmarkSet10(b *testing.B) {
	SetBenchmarkTemplate(b, 2048, 10)
}
func BenchmarkSet100(b *testing.B) {
	SetBenchmarkTemplate(b, 2048, 100)
}
func BenchmarkSet1000(b *testing.B) {
	SetBenchmarkTemplate(b, 2048, 1000)
}

func SetOneBenchmarkTemplate(b *testing.B, n, w int) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(n)
	start := rand.Intn(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.SetOne(start)
	}
}

func BenchmarkSetOne10(b *testing.B) {
	SetOneBenchmarkTemplate(b, 2048, 10)
}
func BenchmarkSetOne100(b *testing.B) {
	SetOneBenchmarkTemplate(b, 2048, 100)
}
func BenchmarkSetOne1000(b *testing.B) {
	SetOneBenchmarkTemplate(b, 2048, 1000)
}

func SetRangeBenchmarkTemplate(b *testing.B, n, w int) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(n)
	start := rand.Intn(n - w)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.SetRange(start, w)
	}
}

func BenchmarkSetRange10(b *testing.B) {
	SetRangeBenchmarkTemplate(b, 2048, 10)
}
func BenchmarkSetRange100(b *testing.B) {
	SetRangeBenchmarkTemplate(b, 2048, 100)
}
func BenchmarkSetRange1000(b *testing.B) {
	SetRangeBenchmarkTemplate(b, 2048, 1000)
}
