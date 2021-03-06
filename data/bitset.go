// Bitset implementation for HTM.
//
// Bits are stored into a variable-length slice of 64-bit integers.

package data

import "fmt"
import "bufio"
import "io"
import "strings"

const deBrujin64 = 0x0218a392cd3d5dbf

var deBrujin64Table [64]int

func init() {
	for index := 0; index < 64; index++ {
		n := uint64(1) << uint(index)
		deBrujin64Table[((n&-n)*deBrujin64)>>58] = index
	}
}

// A bitset type.
type Bitset struct {
	// The binary representation of this bitset. Bits are stored LittleEndian but integers grow in orders of magnitude into 64-bit integers instead of bytes.
	binary []uint64
	// The valid length of the binary array. Whatever additional space is wasted.
	length int
}

func (b Bitset) IsSet(index int) bool {
	if index < 0 || index > b.length {
		return false
	}
	pos, offset := index/64, index%64
	return b.binary[pos]&(1<<uint64(offset)) != 0
}

func (b Bitset) AllSet(indices ...int) bool {
	for _, v := range indices {
		if v > b.length || !b.IsSet(v) {
			return false
		}
	}
	return true
}

func (b Bitset) Foreach(f func(int)) {
	base := 0
	for pos, el := range b.binary {
		base = pos * 64
		for el > 0 {
			f(base + deBrujin64Table[((el&-el)*deBrujin64)>>58])
			el &= el - 1
		}
	}
}

// Iterates through the true indices in this bitset until it finds an index that
// satisfies f(). Return that index. If f(x) return false for all x, this method
// return b.Len().
func (b Bitset) Any(f func(int) bool) int {
	base := 0
	for pos, el := range b.binary {
		base = pos * 64
		for el > 0 {
			val := base + deBrujin64Table[((el&-el)*deBrujin64)>>58]
			if f(val) {
				return val
			}
			el &= el - 1
		}
	}
	return b.Len()
}

func (b Bitset) Overlap(other Bitset) int {
	// This algorithm behaves well only when the number of set bits is small.
	count := 0
	for i, el := range b.binary {
		v := el & other.binary[i]
		for ; v != 0; count++ {
			v &= v - 1
		}
	}
	return count
}

func (b Bitset) NumSetBits() (count int) {
	// This algorithm behaves well only when the number of set bits is small. If
	// You are counting the non-zero bits in a dense bitset, use DenseCount()
	// below.
	for _, el := range b.binary {
		for ; el != 0; count++ {
			// Clear the least significant bit.
			el &= el - 1
		}
	}
	return
}

func (b Bitset) DenseCount() (count int) {
	// Only fast when the bitset is rather dense. Prefer NumSetBits() above in
	// most cases.
	for _, el := range b.binary {
		el -= (el >> 1) & 0x5555555555555555
		el = (el & 0x3333333333333333) + ((el >> 2) & 0x3333333333333333)
		el = (((el + (el >> 4)) & 0xf0f0f0f0f0f0f0f) * 0x101010101010101) >> 56
		count += int(el)
	}
	return
}

func (b Bitset) IsZero() bool {
	for _, el := range b.binary {
		if el != 0 {
			return false
		}
	}
	return true
}

func (b *Bitset) Reset() *Bitset {
	for i := 0; i < len(b.binary); i++ {
		b.binary[i] = uint64(0)
	}
	return b
}

func (b *Bitset) Set(indices ...int) *Bitset {
	for _, v := range indices {
		if v >= b.length {
			panic(fmt.Errorf(
				"Attempt to write past end of bitset (%d > %d)", v, b.length))
		}
		if v < 0 {
			panic(fmt.Errorf(
				"Attempt to write before start of bitset (%d < %d)", v, 0))
		}

		b.binary[v/64] |= 1 << uint64(v%64)
	}
	return b
}

// Sets all bits in the interval [start, end).
func (b *Bitset) SetRange(start, end int) *Bitset {
	if end <= start {
		return b
	}
	if start < 0 {
		panic(fmt.Errorf(
			"Attempt to write before start of bitset (%d < %d)", start, 0))
	}
	if end > b.length {
		panic(fmt.Errorf(
			"Attempt to write past end of bitset (%d > %d)", end, b.length))
	}
	pos := start / 64
	for pos < len(b.binary) {
		minIdx, maxIdx := pos*64, (pos+1)*64
		if end <= minIdx {
			return b
		}
		mask := ^uint64(0)
		if start > minIdx {
			mask = mask << uint64(start-minIdx)
		}
		if end < maxIdx {
			mask &= ^uint64(0) >> uint64(maxIdx-end)
		}
		b.binary[pos] |= mask
		pos++
	}
	return b
}

func (b *Bitset) Unset(indices ...int) *Bitset {
	for _, v := range indices {
		if v >= 0 && v < b.length {
			b.binary[v/64] &= ^(1 << uint64(v%64))
		}
	}
	return b
}

func (b Bitset) String() string {
	s := make([]string, 0, b.NumSetBits())
	b.Foreach(func(i int) {
		s = append(s, fmt.Sprintf("%04d", i))
	})
	return "[" + strings.Join(s, ",") + "]"
}

func (b Bitset) Len() int {
	return b.length
}

func (b Bitset) Equals(other Bitset) bool {
	if b.length != other.length {
		return false
	}
	for i, v := range b.binary {
		if v != other.binary[i] {
			return false
		}
	}
	return true
}

func (b *Bitset) ResetTo(other Bitset) {
	if b.length != other.length {
		panic(fmt.Errorf(
			"Cannot ResetTo bitset of different length (%d != %d)", b.length, other.length))
	}
	copy(b.binary, other.binary)
}

func (b *Bitset) Or(other Bitset) {
	if b.length != other.length {
		panic(fmt.Errorf(
			"Cannot OR bitsets of different length (%d != %d)", b.length, other.length))
	}
	for i, v := range other.binary {
		b.binary[i] |= v
	}
}

func (b *Bitset) And(other Bitset) {
	if b.length != other.length {
		panic(fmt.Errorf(
			"Cannot AND bitsets of different length (%d != %d)", b.length, other.length))
	}
	for i, v := range other.binary {
		b.binary[i] &= v
	}
}

func (b *Bitset) AndNot(other Bitset) {
	if b.length != other.length {
		panic(fmt.Errorf(
			"Cannot AndNot bitsets of different length (%d != %d)", b.length, other.length))
	}
	for i, v := range other.binary {
		b.binary[i] &= ^v
	}
}

func (b *Bitset) SetFromBitsetAt(other Bitset, offset int) {
	if offset+other.length > b.length {
		panic(fmt.Errorf("SetFromBitset() would go past end! Needs %d bits, has %d.",
			offset+other.length, b.length))
	}
	rem := uint64(offset % 64)
	dest := offset / 64

	vr := uint64(0)
	for _, el := range other.binary {
		b.binary[dest] |= (el << rem) | vr
		vr = el >> (64 - rem)
		dest++
	}
	if vr != 0 {
		b.binary[dest] |= vr
	}
}

func (b Bitset) Print(width int, writer io.Writer) (err error) {
	n := 0
	buf := bufio.NewWriter(writer)
	for _, v := range b.binary {
		for i := 0; i < 64; i++ {
			if n >= b.length {
				break
			}
			if v&(1<<uint64(i)) != 0 {
				if _, err = buf.WriteRune('x'); err != nil {
					return
				}
			} else {
				if _, err = buf.WriteRune('-'); err != nil {
					return
				}
			}
			n++
			if n%width == 0 {
				if _, err = buf.WriteRune('\n'); err != nil {
					return
				}
			}
		}
	}
	buf.Flush()
	return
}

// Creates a new bitset with enough storage for at least the given number of bits.
func NewBitset(length int) *Bitset {
	result := new(Bitset)
	num := (length-1)/64 + 1
	result.binary = make([]uint64, num)
	result.length = length
	return result
}

func (b Bitset) Clone() *Bitset {
	result := new(Bitset)
	result.binary = make([]uint64, len(b.binary))
	copy(result.binary, b.binary)
	result.length = b.length
	return result
}

// A bitset encoder.
type Encoder interface {
	// Encodes a value as a bitset. If the value cannot be encoded, returns an error.
	Encode(value interface{}) error

	// Gets the last successfully encoded value.
	Get() Bitset
}

// A bitset decoder.
type Decoder interface {
	// Decodes the bitset into a value in the original domain. The value may or may
	// not reflect precisely the encoded value (e.g. due to loss of precision)
	Decode(bits Bitset) interface{}
}
