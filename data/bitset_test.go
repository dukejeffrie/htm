// Unit tests for sensor implementation

package data

import "fmt"
import "io"
import "math/rand"
import "strings"
import "testing"

func ExpectEquals(t *testing.T, message string, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Errorf("%s: expected %v, but was: %v", message, expected, actual)
	}
}

func TestConstruction(t *testing.T) {
	one := NewBitset(2)
	ExpectEquals(t, "one.length", 2, one.length)
	ExpectEquals(t, "one.binary.length", 1, len(one.binary))
	if !one.IsZero() {
		t.Error("should be zero:", one)
	}

	// Check that invalid arguments don't blow up.
	if one.IsSet(-1) || one.IsSet(one.Len()+1) {
		t.Error("These bits should never be set.")
	}

	cent := NewBitset(100)
	ExpectEquals(t, "cent.length", 100, cent.length)
	ExpectEquals(t, "cent.binary.length", 2, len(cent.binary))

	two := NewBitset(128)
	ExpectEquals(t, "two.binary.length", 2, len(two.binary))

	three := NewBitset(129)
	ExpectEquals(t, "three.binary.length", 3, len(three.binary))
}

func TestEquals(t *testing.T) {
	rand.Seed(10)
	b := NewBitset(64)
	for i := 0; i < 10; i++ {
		b.Set(rand.Intn(64))
		if !b.Equals(*b) {
			t.Errorf("Should be equal to itself: %v", *b)
		}
	}

	b2 := b.Clone()
	if !b.Equals(*b2) {
		t.Errorf("Should be equal to its clone: %v != %v", *b, *b2)
	}
	b2.SetRange(0, 11)
	if b.Equals(*b2) {
		t.Errorf("Bitsets should be different, but are not: %v == %v", *b, *b2)
	}

	b.Reset()
	b3 := NewBitset(b.Len() + 1)
	if b.Equals(*b3) {
		t.Error("Two zeroed bitsets with different lengths should not be equal:",
			*b, *b3)
	}
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
	ExpectEquals(t, "num bits", 3, b.DenseCount())
	if b.AllSet(0, 11, 12) {
		t.Errorf("Bit 0 should not be set: %v", *b)
	}
	b.Set(0)
	if !b.AllSet(0, 11, 12) {
		t.Errorf("Bits 0, 11 and 12 should be set: %v", *b)
	}

	b.Reset()
	ExpectEquals(t, "bit 127", false, b.IsSet(127))
	if !b.IsZero() {
		t.Error("should be zero:", b)
	}

	b.Set(222, 444, 888, 1023, 1024, 1331, 2047)
	ExpectEquals(t, "num bits", 7, b.NumSetBits())
	ExpectEquals(t, "num bits", 7, b.DenseCount())
	b.Set(111)
	ExpectEquals(t, "bit 111", true, b.IsSet(111))
	b.Unset(111)
	ExpectEquals(t, "bit 111", false, b.IsSet(111))

	b2 := NewBitset(b.Len()).Set(1, 2, 3)
	b2.ResetTo(*b2)
	if b.Equals(*b2) {
		t.Errorf("Should be equal, but are not: %v != %v", *b, *b2)
	}
}

func TestBitsetReset_DifferentLength(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have failed, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "Cannot ResetTo") {
			t.Error("Should panic with the ResetTo message, but got: %v", err)
		}
	}()
	b := NewBitset(100)
	b2 := NewBitset(101)
	b.ResetTo(*b2)
}

func TestBitsetAnd(t *testing.T) {
	b := NewBitset(2048).Set(20, 200, 2000)
	b2 := b.Clone()
	b.And(*b)
	if !b.Equals(*b2) {
		t.Errorf("Failed b & b == b. Expected %v, but got: %v", *b2, *b)
	}
	b.Set(1, 2, 3)
	b.And(*b2)

	if !b.Equals(*b2) {
		t.Errorf("Failed (b|[1,2,3]) & b == b. Expected %v, but got: %v", *b2, *b)
	}
}

func TestBitsetAnd_DifferentLength(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have failed, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "Cannot AND") {
			t.Error("Should panic with the AND message, but got: %v", err)
		}
	}()
	b := NewBitset(100)
	b2 := NewBitset(101)
	b.And(*b2)
}

func TestBitsetAndNot(t *testing.T) {
	b := NewBitset(2048).Set(20, 200, 2000)
	b2 := b.Clone()
	b.AndNot(*b)
	if !b.IsZero() {
		t.Errorf("Failed b &^ b == 0. Expected empty, but got: %v", *b)
	}
	b.ResetTo(*b2)
	b3 := b2.Clone().Unset(20)
	b.AndNot(*b3)

	if !b.Equals(*NewBitset(b.Len()).Set(20)) {
		t.Errorf("Failed (b|[1,2,3]) & b == b. Expected %v, but got: %v", *b2, *b)
	}
}

func TestBitsetAndNot_DifferentLength(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have failed, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "Cannot AndNot") {
			t.Error("Should panic with the AndNot message, but got: %v", err)
		}
	}()
	b := NewBitset(100)
	b2 := NewBitset(101)
	b.AndNot(*b2)
}

func TestBitsetOr(t *testing.T) {
	b := NewBitset(2048).Set(20, 200, 2000)
	b2 := b.Clone()
	b.Or(*b)
	if !b.Equals(*b2) {
		t.Errorf("Failed b | b == b. Expected %v, but got: %v", *b2, *b)
	}
	b3 := NewBitset(b.Len()).Set(10, 1000)
	b.Or(*b3)
	b2.Set(10, 1000)
	if !b.Equals(*b2) {
		t.Errorf("Failed b | c. Expected %v, but got: %v", *b2, *b)
	}
}

func TestBitsetOr_DifferentLength(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have failed, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "Cannot OR") {
			t.Error("Should panic with the OR message, but got: %v", err)
		}
	}()
	b := NewBitset(100)
	b2 := NewBitset(101)
	b.Or(*b2)
}

func TestSetAfterLength(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have panicked, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "past end") {
			t.Error("Unexpected error:", err)
		}
	}()

	NewBitset(60).Set(61, 62, 63)
}

func TestSetBeforeStart(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have panicked, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "before start") {
			t.Error("Unexpected error:", err)
		}
	}()

	NewBitset(60).Set(-1)
}

func TestSetRange(t *testing.T) {
	b := NewBitset(2048)
	b.SetRange(0, 0)
	if !b.IsZero() {
		t.Error("Accepted invalid range [0,0): %v", *b)
	}
	b.SetRange(8, 154)

	ExpectEquals(t, "num bits", 154-8, b.NumSetBits())
	ExpectEquals(t, "num bits", 154-8, b.DenseCount())
	for i := 8; i < 154; i++ {
		ExpectEquals(t, fmt.Sprintf("bit %d", i), true, b.IsSet(i))
	}
}

func TestSetRange_AfterLength(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have panicked, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "past end") {
			t.Error("Unexpected error:", err)
		}
	}()
	NewBitset(2048).SetRange(2000, 2049)
}

func TestSetRange_BeforeStart(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have panicked, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "before start") {
			t.Error("Unexpected error:", err)
		}
	}()

	NewBitset(60).SetRange(-1, 0)
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

func TestSetFromBitset(t *testing.T) {
	src := NewBitset(5)
	src.Set(1)
	src.Set(4)

	dest := NewBitset(10)
	dest.SetFromBitsetAt(*src, 5)

	if !dest.AllSet(6, 9) {
		t.Errorf("dest should have 6 and 9: %v", *dest)
	}

	dest = NewBitset(128 + 64).Set(0)
	src = NewBitset(64).Set(0)
	dest.SetFromBitsetAt(*src, 128)

	if !dest.AllSet(0, 128) {
		t.Errorf("dest should have 0 and 128: %v", *dest)
	}

	src.Set(63)
	dest.SetFromBitsetAt(*src, 60)

	expected := NewBitset(dest.Len()).Set(0, 60, 123, 128)
	if !dest.Equals(*expected) {
		t.Errorf("dest did not match. Expected %v, but got: %v",
			*expected, *dest)
	}
}

func TestSetFromBitset_NotEnoughSpace(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Should have failed, but didn't.")
		} else if !strings.Contains(fmt.Sprint(err), "past end") {
			t.Error("Should panic with the \"past end\" message, but got: %v", err)
		}
	}()

	src := NewBitset(10)
	dest := NewBitset(10)

	dest.SetFromBitsetAt(*src, 1)
}

func SetFromBitsetBenchmarkTemplate(b *testing.B, srcSize, destSize int,
	truncate bool) {
	src := NewBitset(srcSize)
	src.Set(0, srcSize/2, (srcSize-1)/2*2)
	dest := NewBitset(destSize + srcSize)
	dest.Set(0, destSize/2, destSize-1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest.SetFromBitsetAt(*src, destSize)
	}
}

func BenchmarkSetFromBitsetEven64(b *testing.B) {
	SetFromBitsetBenchmarkTemplate(b, 64, 64, false)
}

func BenchmarkSetFromBitsetEven2048(b *testing.B) {
	SetFromBitsetBenchmarkTemplate(b, 2048, 2048, false)
}

func BenchmarkSetFromBitsetOdd64(b *testing.B) {
	SetFromBitsetBenchmarkTemplate(b, 59, 64, false)
}

func BenchmarkSetFromBitsetOdd2048(b *testing.B) {
	SetFromBitsetBenchmarkTemplate(b, 2047, 2048, false)
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

func BenchmarkCountNumSetBits_Few(b *testing.B) {
	all := NewBitset(2048).Set(0, 63, 127, 128, 255, 256, 383, 384, 400, 420, 500, 600, 700, 800, 900, 1000, 1100, 1200, 2047)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.NumSetBits()
	}
}

func BenchmarkCountNumSetBits_Half(b *testing.B) {
	all := NewBitset(2048)
	for i := 0; i < 1024; i++ {
		all.Set(i * 2)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.NumSetBits()
	}
}

func BenchmarkCountNumSetBits_All(b *testing.B) {
	all := NewBitset(2048)
	all.SetRange(0, 2048)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.NumSetBits()
	}
}

func BenchmarkDenseCount_Few(b *testing.B) {
	all := NewBitset(2048).Set(0, 63, 127, 128, 255, 256, 383, 384, 400, 420, 500, 600, 700, 800, 900, 1000, 1100, 1200, 2047)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.DenseCount()
	}
}

func BenchmarkDenseCount_Half(b *testing.B) {
	all := NewBitset(2048)
	for i := 0; i < 1024; i++ {
		all.Set(i * 2)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.DenseCount()
	}
}

func BenchmarkDenseCount_All(b *testing.B) {
	all := NewBitset(2048)
	all.SetRange(0, 2048)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all.DenseCount()
	}
}
