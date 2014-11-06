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

func (b Bitset) Hash() int {
	h := b.binary[0]
	for pos := 1; pos < b.length; pos++ {
		h |= b.binary[pos]
	}
	return int(h)
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

func (b Bitset) NumSetBits() int {
	// This algorithm behaves well only when the number of set bits is small.
	count := 0
	for _, el := range b.binary {
		for ; el != 0; count++ {
			// Clear the least significant bit.
			el &= el - 1
		}
	}
	return count
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
		if v >= 0 && v < b.length {
			b.binary[v/64] |= 1 << uint64(v%64)
		}
	}
	return b
}

// Sets all bits in the interval [start, end).
func (b *Bitset) SetRange(start, end int) {
	if end <= start {
		return
	}
	if end > b.length {
		end = b.length
	}
	pos := start / 64
	for pos < len(b.binary) {
		minIdx, maxIdx := pos*64, (pos+1)*64
		if end <= minIdx {
			return
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
}

func (b *Bitset) Unset(indices ...int) {
	for _, v := range indices {
		if v >= 0 && v < b.length {
			b.binary[v/64] &= ^(1 << uint64(v%64))
		}
	}
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
			"Cannot copy from bitset of different length (%d != %d)", b.length, other.length))
	}
	copy(b.binary, other.binary)
}

func (b *Bitset) Or(other Bitset) {
	if b.length != other.length {
		panic(fmt.Errorf(
			"Cannot AND bitsets of different length (%d != %d)", b.length, other.length))
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

func (b *Bitset) SetFromBitsetAt(other Bitset, offset int) {
	if offset+other.length > b.length {
		panic(fmt.Errorf("AND operation would go past end! Needs %d bytes, has %d.",
			offset+other.length, b.length))
	}
	b.appendAt(other, offset)
}

func (b *Bitset) appendAt(other Bitset, offset int) {
	rem := uint64(offset % 64)
	l := offset + other.length
	num := (l-1)/64 + 1
	num -= len(b.binary)
	dest := offset / 64

	for src := 0; src < len(other.binary); src++ {
		el := other.binary[src]
		b.binary[dest] |= el << rem
		if num > 0 {
			b.binary = append(b.binary, el>>(64-rem))
			num--
		}
		dest++
	}
	if l > b.length {
		b.length = l
	}
}

func (b *Bitset) Truncate(width int) {
	if width < b.length {
		b.length = width
		b.binary = b.binary[0 : (b.length-1)/64]
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
