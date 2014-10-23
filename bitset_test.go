// Unit tests for sensor implementation

package htm

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

func TestAppend(t *testing.T) {
	scratch := make([]int, 20)

	src := NewBitset(5)
	src.SetOne(1)
	src.SetOne(4)

	dest := NewBitset(5)
	dest.Append(*src)

	ExpectEquals(t, "length", 10, dest.Len())
	ExpectContentEquals(t, fmt.Sprintf("dest: %v", dest), []int{6, 9}, dest.ToIndexes(scratch))

	dest.Append(*src)

	ExpectEquals(t, "length", 15, dest.Len())
	ExpectContentEquals(t, fmt.Sprintf("dest: %v", dest), []int{6, 9, 11, 14}, dest.ToIndexes(scratch))
}

func TestAppendLong(t *testing.T) {
	scratch := make([]int, 20)
	dest := NewBitset(20)
	dest.SetOne(10)
	src := NewBitset(60)
	src.SetOne(2)
	src.SetOne(22)
	src.SetOne(59)

	dest.Append(*src)
	ExpectEquals(t, "length", 80, dest.Len())
	ExpectContentEquals(t, fmt.Sprintf("dest: %v", dest), []int{10, 22, 42, 79}, dest.ToIndexes(scratch))
}

func AppendBenchmarkTemplate(b *testing.B, src_size, dest_size int,
	truncate bool) {
	src := NewBitset(src_size)
	src.Set([]int{2, src_size / 2, (src_size - 1) / 2 * 2})
	b.ResetTimer()
	dest := NewBitset(dest_size)
	for i := 0; i < b.N; i++ {
		dest.Append(*src)
		if truncate {
			dest.Truncate(dest_size)
		} else {
			dest = NewBitset(dest_size)
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

func SetFromBitsetAtBenchmarkTemplate(b *testing.B, src_size int) {
	src := NewBitset(src_size)
	src.Set([]int{1, 3, 5})
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
	even.Set([]int{2, 4, 8, 16, 32})
	odd := NewBitset(64)
	odd.Set([]int{1, 3, 7, 15, 31})

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

func ToIndexesBenchmarkTemplate(b *testing.B, n, l int) {
	rand.Seed(int64(1979))
	bitset := NewBitset(n)
	bits_to_set := make([]int, l)
	for i := 0; i > l; i++ {
		bits_to_set[i] = rand.Intn(n)
	}
	bitset.Set(bits_to_set)
	b.ResetTimer()
	output := make([]int, bitset.NumSetBits())
	for i := 0; i < b.N; i++ {
		bitset.ToIndexes(output)
	}
}

func BenchmarkToIndexes10(b *testing.B) {
	ToIndexesBenchmarkTemplate(b, 2048, 10)
}

func BenchmarkToIndexes100(b *testing.B) {
	ToIndexesBenchmarkTemplate(b, 2048, 100)
}

func BenchmarkToIndexes1000(b *testing.B) {
	ToIndexesBenchmarkTemplate(b, 2048, 1000)
}
