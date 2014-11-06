// Unit tests for sensor implementation

package data

import "fmt"
import "io"
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
			t.Logf("%s (element[%d]): %v != %v", message, i, v, actual[i])
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
	if !one.IsZero() {
		t.Error("should be zero:", one)
	}

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

	b.Set(127)
	ExpectEquals(t, "bit 127", true, b.IsSet(127))
	if b.IsZero() {
		t.Error("should not be zero:", b)
	}

	b.Set(11, 12)
	ExpectEquals(t, "bit 127", true, b.IsSet(127))
	ExpectEquals(t, "num bits", 3, b.NumSetBits())

	b.Reset()
	ExpectEquals(t, "bit 127", false, b.IsSet(127))
	if !b.IsZero() {
		t.Error("should be zero:", b)
	}

	b.Set(222, 444, 888, 1023, 1024, 1331, 2047)
	ExpectEquals(t, "num bits", 7, b.NumSetBits())
	b.Set(111)
	ExpectEquals(t, "bit 111", true, b.IsSet(111))
	b.Unset(111)
	ExpectEquals(t, "bit 111", false, b.IsSet(111))
}

func TestSetAfterLength(t *testing.T) {
	b := NewBitset(60)
	b.Set(61, 62, 63)
	ExpectEquals(t, "after-length bits not set", 0, b.NumSetBits())

	b.SetRange(59, 70)
	ExpectEquals(t, "overflowing range bits not set", 1, b.NumSetBits())
	ExpectEquals(t, "bit 59", true, b.IsSet(59))
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
	b.Set(1, 3)

	ExpectEquals(t, "first integer", uint64(8+2), b.binary[0])

	ExpectEquals(t, "bit 1", true, b.IsSet(1))
	ExpectEquals(t, "bit 0", false, b.IsSet(0))
	ExpectEquals(t, "bit 3", true, b.IsSet(3))
	ExpectEquals(t, "two bits set", 2, b.NumSetBits())

	if !b.AllSet(1, 3) {
		t.Errorf("Indices 1 and 3 should be set.")
	}
}

func TestIteration(t *testing.T) {
	b := NewBitset(2048).Set(1, 33, 63, 64, 2000)
	dest := make([]int, b.NumSetBits())
	d := 0
	b.Foreach(func(i int) {
		dest[d] = i
		d++
	})
	other := NewBitset(2048).Set(dest...)
	if !other.Equals(*b) {
		t.Errorf("Iterator failed. Expected: %v, but got: %v", *b, *other)
	}
}

func TestAppend(t *testing.T) {
	src := NewBitset(5)
	src.Set(1)
	src.Set(4)

	dest := NewBitset(5)
	dest.appendAt(*src, dest.Len())

	ExpectEquals(t, "length", 10, dest.Len())
	if !dest.AllSet(6, 9) {
		t.Errorf("dest should have 6 and 9: %v", *dest)
	}

	dest.appendAt(*src, dest.Len())

	ExpectEquals(t, "length", 15, dest.Len())
	if !dest.AllSet(6, 9, 11, 14) {
		t.Errorf("dest should have 6 and 9: %v", *dest)
	}
}

func TestAppendLong(t *testing.T) {
	dest := NewBitset(20)
	dest.Set(10)
	src := NewBitset(60)
	src.Set(2)
	src.Set(22)
	src.Set(59)

	dest.appendAt(*src, dest.Len())
	ExpectEquals(t, "length", 80, dest.Len())
	if !dest.AllSet(10, 22, 42, 79) {
		t.Errorf("dest missing bits: %v", *dest)
	}
}

func AppendBenchmarkTemplate(b *testing.B, srcSize, destSize int,
	truncate bool) {
	src := NewBitset(srcSize)
	src.Set(2, srcSize/2, (srcSize-1)/2*2)
	b.ResetTimer()
	dest := NewBitset(destSize)
	for i := 0; i < b.N; i++ {
		dest.appendAt(*src, dest.Len())
		if truncate {
			dest.Truncate(destSize)
		} else {
			dest = NewBitset(destSize)
		}
	}
}

func BenchmarkAppendEven64(b *testing.B) {
	AppendBenchmarkTemplate(b, 64, 64, false)
}

func BenchmarkAppendEven64T(b *testing.B) {
	AppendBenchmarkTemplate(b, 64, 64, true)
}

func BenchmarkAppendEven2048(b *testing.B) {
	AppendBenchmarkTemplate(b, 2048, 2048, false)
}

func BenchmarkAppendEven2048T(b *testing.B) {
	AppendBenchmarkTemplate(b, 2048, 2048, true)
}

func BenchmarkAppendOdd64(b *testing.B) {
	AppendBenchmarkTemplate(b, 59, 64, false)
}

func BenchmarkAppendOdd64T(b *testing.B) {
	AppendBenchmarkTemplate(b, 59, 64, true)
}

func BenchmarkAppendOdd2048(b *testing.B) {
	AppendBenchmarkTemplate(b, 2047, 2048, false)
}

func BenchmarkAppendOdd2048T(b *testing.B) {
	AppendBenchmarkTemplate(b, 2047, 2048, true)
}

func SetFromBitsetAtBenchmarkTemplate(b *testing.B, srcSize int) {
	src := NewBitset(srcSize)
	for i := 0; i < srcSize/2; i++ {
		src.Set(rand.Intn(srcSize))
	}
	b.ResetTimer()
	dest := NewBitset(20480)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest.SetFromBitsetAt(*src, 16666)
	}
}

func BenchmarkSetFromBitsetAt64(b *testing.B) {
	SetFromBitsetAtBenchmarkTemplate(b, 64)
}

func BenchmarkSetFromBitsetAt2048(b *testing.B) {
	SetFromBitsetAtBenchmarkTemplate(b, 2048)
}

func TestPrintBitset(t *testing.T) {
	even := NewBitset(64)
	even.Set(2, 4, 8, 16, 32)
	odd := NewBitset(64)
	odd.Set(1, 3, 7, 15, 31)

	reader, writer := io.Pipe()

	go func() {
		if err := even.Print(16, writer); err != nil {
			t.Error(err)
		}
		fmt.Fprint(writer, "\n")
		if err := odd.Print(40, writer); err != nil {
			t.Error(err)
		}
		writer.Close()
	}()

	var e1, e2, e3, e4 string
	if _, err := fmt.Fscan(reader, &e1, &e2, &e3, &e4); err != nil {
		t.Error(err)
	}
	t.Log(e1)
	t.Log(e2)
	t.Log(e3)
	t.Log(e4)
	estr := e1 + e2 + e3 + e4
	ExpectEquals(t, "even string",
		"--x-x---x-------x---------------x-------------------------------", estr)

	var o1, o2 string
	if _, err := fmt.Fscan(reader, &o1, &o2); err != nil {
		t.Error(err)
	}
	ostr := o1 + o2
	ExpectEquals(t, "odd string",
		"-x-x---x-------x---------------x--------------------------------", ostr)
}

func SetBenchmarkTemplate(b *testing.B, n, l int) {
	rand.Seed(int64(b.N))
	bitset := NewBitset(n)
	bits := make([]int, l)
	for i := 0; i < l; i++ {
		bits[i] = rand.Intn(n)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bitset.Set(bits...)
	}
}

func BenchmarkSet1(b *testing.B) {
	SetBenchmarkTemplate(b, 2048, 1)
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

func ForeachBenchmarkTemplate(b *testing.B, n, l int) {
	rand.Seed(int64(1979))
	bitset := NewBitset(n)
	bits := make([]int, l)
	for i := 0; i < l; i++ {
		bits[i] = rand.Intn(n)
	}
	bitset.Set(bits...)
	b.ResetTimer()
	f := func(i int) {
		// do nothing.
	}
	for i := 0; i < b.N; i++ {
		bitset.Foreach(f)
	}
}

func BenchmarkForeach10(b *testing.B) {
	ForeachBenchmarkTemplate(b, 2048, 10)
}

func BenchmarkForeach100(b *testing.B) {
	ForeachBenchmarkTemplate(b, 2048, 100)
}

func BenchmarkForeach1000(b *testing.B) {
	ForeachBenchmarkTemplate(b, 2048, 1000)
}

func TestOverlapBitset(t *testing.T) {
	alpha := NewBitset(2048)
	alpha.Set(0, 2, 4, 10, 12, 14, 20, 22, 24)
	beta := NewBitset(2048)
	beta.Set(2, 12, 22, 32)
	ExpectEquals(t, "alpha & beta", 3, alpha.Overlap(*beta))
	beta.Set(4)
	ExpectEquals(t, "alpha & beta", 4, alpha.Overlap(*beta))
}

func BenchmarkNumSetBits_Few(b *testing.B) {
	all := NewBitset(2048).Set(0, 63, 127, 128, 255, 256, 383, 384, 400, 420, 500, 600, 700, 800, 900, 1000, 1100, 1200, 2047)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.NumSetBits()
	}
}

func BenchmarkNumSetBits_Half(b *testing.B) {
	all := NewBitset(2048)
	for i := 0; i < 1024; i++ {
		all.Set(i * 2)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.NumSetBits()
	}
}

func BenchmarkNumSetBits_All(b *testing.B) {
	all := NewBitset(2048)
	all.SetRange(0, 2048)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.NumSetBits()
	}
}
