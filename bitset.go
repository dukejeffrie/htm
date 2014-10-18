// Bitset implementation for HTM.
//
// Bits are stored into a variable-length slice of 64-bit integers.

package htm

import "fmt"
import "strings"

// A bitset type.
type Bitset struct {
	// The binary representation of this bitset. Bits are stored LittleEndian but integers grow in orders of magnitude into 64-bit integers instead of bytes.
	binary []uint64
	// The valid length of the binary array. Whatever additional space is wasted.
	length int
}

func (b Bitset) locate(index int) (pos, offset int) {
	pos = index / 64
	offset = index % 64
	return
}

func (b Bitset) IsSet(index int) bool {
	pos, offset := b.locate(index)
	return b.binary[pos]&(1<<uint64(offset)) != 0
}

func (b Bitset) ToIndexes(indices []int) []int {
	dest := 0
	for pos, el := range b.binary {
		for i := uint64(0); i < 64; i++ {
			if el&(1<<i) != 0 {
				indices[dest] = pos*64 + int(i)
				dest++
			}
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
	idx := 0
	for pos, el := range b.binary {
		if idx >= len(indices) {
			continue
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
			fmt.Printf("Set(%d) = 1\n", indices[idx])
			el |= 1 << uint64(indices[idx]-min_idx)
		}
		b.binary[pos] = el
	}
}

func (b Bitset) String() string {
	s := make([]string, len(b.binary))
	for i, v := range b.binary {
		s[i] = fmt.Sprintf("%b", v)
	}
	return strings.Join(s, " ")
}

func (b Bitset) Len() int {
	return b.length
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
