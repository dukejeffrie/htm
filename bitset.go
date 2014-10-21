// Bitset implementation for HTM.
//
// Bits are stored into a variable-length slice of 64-bit integers.

package htm

import "fmt"
import "bufio"
import "io"
import "strings"
import "sort"

// A bitset type.
type Bitset struct {
	// The binary representation of this bitset. Bits are stored LittleEndian but integers grow in orders of magnitude into 64-bit integers instead of bytes.
	binary []uint64
	// The valid length of the binary array. Whatever additional space is wasted.
	length int
}

func (b Bitset) IsSet(index int) bool {
	pos, offset := index/64, index%64
	return b.binary[pos]&(1<<uint64(offset)) != 0
}

func (b Bitset) ToIndexes(indices []int) []int {
	dest := 0
	base := 0
	for pos, el := range b.binary {
		base = pos * 64
		for i := 0; el > 0; i++ {
			if el&1 == 1 {
				indices[dest] = base + i
				dest++
			}
			el >>= 1
		}
	}
	return indices[0:dest]
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

func (b *Bitset) Reset() {
	for i := 0; i < len(b.binary); i++ {
		b.binary[i] = uint64(0)
	}
}

func (b *Bitset) Set(indices []int) {
	if len(indices) == 0 {
		return
	}
	sort.Ints(indices)
	idx := 0
	for pos, el := range b.binary {
		if idx >= len(indices) {
			return
		}
		if indices[idx] > b.Len() {
			panic(fmt.Sprintf(
				"index %v is larger than the length of this bitset (length=%v",
				indices[idx], b.Len()))
		}
		min_idx, max_idx := pos*64, (pos+1)*64
		for ; idx < len(indices) &&
			indices[idx] >= min_idx &&
			indices[idx] < max_idx; idx++ {
			el |= 1 << uint64(indices[idx]-min_idx)
		}
		b.binary[pos] = el
	}
}

// Sets all bits in the interval [start, end).
func (b *Bitset) SetRange(start, end int) {
	if end <= start {
		return
	}
	pos := start / 64
	for pos < len(b.binary) {
		min_idx, max_idx := pos*64, (pos+1)*64
		if end <= min_idx {
			return
		}
		mask := ^uint64(0)
		if start > min_idx {
			mask = mask << uint64(start-min_idx)
		}
		if end < max_idx {
			mask &= ^uint64(0) >> uint64(max_idx-end)
		}
		b.binary[pos] |= mask
		pos++
	}
}

func (b *Bitset) SetOne(index int) {
	pos, offset := index/64, index%64
	b.binary[pos] |= 1 << uint64(offset)
}

func (b *Bitset) ClearOne(index int) {
	pos, offset := index/64, index%64
	b.binary[pos] &= ^(1 << uint64(offset))
}

func (b Bitset) String() string {
	indices := make([]int, b.NumSetBits())
	indices = b.ToIndexes(indices)
	s := make([]string, len(indices))
	for i, v := range indices {
		s[i] = fmt.Sprintf("%04d", v)
	}
	return "[" + strings.Join(s, ",") + "]"
}

func (b Bitset) Len() int {
	return b.length
}

func (b Bitset) Equals(other Bitset) bool {
	if b.Len() != other.Len() {
		return false
	}
	for i, v := range b.binary {
		if v != other.binary[i] {
			return false
		}
	}
	return true
}

func (b *Bitset) CopyFrom(other Bitset) {
	if b.Len() != other.Len() {
		panic(fmt.Errorf(
			"Cannot AND bitsets of different length (%d != %d)", b.Len(), other.Len()))
	}
	for i, v := range other.binary {
		b.binary[i] = v
	}
}

func (b *Bitset) Or(other Bitset) {
	if b.Len() != other.Len() {
		panic(fmt.Errorf(
			"Cannot AND bitsets of different length (%d != %d)", b.Len(), other.Len()))
	}
	for i, v := range other.binary {
		b.binary[i] |= v
	}
}

func (b *Bitset) And(other Bitset) {
	if b.Len() != other.Len() {
		panic(fmt.Errorf(
			"Cannot AND bitsets of different length (%d != %d)", b.Len(), other.Len()))
	}
	for i, v := range other.binary {
		b.binary[i] &= v
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
	num := length / 64
	if (length % 64) > 0 {
		num++
	}
	result.binary = make([]uint64, num)
	result.length = length
	return result
}
