package ksuid

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

// uint96 represents an unsigned 96 bits little endian integer.
type uint96 [2]uint64 // [0] holds low 64 bits, [1] holds high 32 bits

func uint96Payload(ksuid KSUID) uint96 {
	return makeUint96FromPayload(ksuid[timestampLengthInBytes:])
}

func makeUint96(high uint32, low uint64) uint96 {
	return uint96{low, uint64(high)}
}

func makeUint96FromPayload(payload []byte) uint96 {
	return uint96{
		binary.BigEndian.Uint64(payload[4:]),         // low (8 bytes)
		uint64(binary.BigEndian.Uint32(payload[:4])), // high (4 bytes)
	}
}

func (v uint96) ksuid(timestamp uint64) (out KSUID) {
	binary.BigEndian.PutUint64(out[:8], timestamp)      // time (8 bytes)
	binary.BigEndian.PutUint32(out[8:12], uint32(v[1])) // high (4 bytes)
	binary.BigEndian.PutUint64(out[12:], v[0])          // low (8 bytes)
	return
}

func (v uint96) bytes() (out [12]byte) {
	binary.BigEndian.PutUint32(out[:4], uint32(v[1]))
	binary.BigEndian.PutUint64(out[4:], v[0])
	return
}

func (v uint96) String() string {
	return fmt.Sprintf("0x%08X%016X", uint32(v[1]), v[0])
}

func cmp96(x, y uint96) int {
	if x[1] < y[1] {
		return -1
	}
	if x[1] > y[1] {
		return 1
	}
	if x[0] < y[0] {
		return -1
	}
	if x[0] > y[0] {
		return 1
	}
	return 0
}

func add96(x, y uint96) (z uint96) {
	var c uint64
	z[0], c = bits.Add64(x[0], y[0], 0)
	z[1], _ = bits.Add64(x[1], y[1], c)
	z[1] &= 0xFFFFFFFF // Mask to 32 bits
	return
}

func sub96(x, y uint96) (z uint96) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], _ = bits.Sub64(x[1], y[1], b)
	z[1] &= 0xFFFFFFFF // Mask to 32 bits
	return
}

func incr96(x uint96) (z uint96) {
	var c uint64
	z[0], c = bits.Add64(x[0], 1, 0)
	z[1] = (x[1] + c) & 0xFFFFFFFF // Add carry and mask to 32 bits
	return
}
