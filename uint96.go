package ksuid

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

// uint96 represents an unsigned 96 bits little endian integer.

// So there are two different endian considerations here:
// The internal array structure is little-endian (lowest bits in lowest index)
// The external byte serialization is big-endian (highest bits in lowest byte address)

type uint96 [3]uint32 // [0] holds low 32 bits, [1] holds middle 32 bits, [2] holds high 32 bits

func uint96Payload(ksuid KSUID) uint96 {
	return makeUint96FromPayload(ksuid[timestampLengthInBytes:])
}

// uint32(low): Takes the lowest 32 bits of the low uint64
// uint32(low >> 32): Shifts the low value right by 32 bits and takes the result, giving us the middle 32 bits
func makeUint96(high uint32, low uint64) uint96 {
	return uint96{
		uint32(low),       // lowest 32 bits
		uint32(low >> 32), // middle 32 bits
		high,              // highest 32 bits
	}
}

func makeUint96FromPayload(payload []byte) uint96 {
	return uint96{
		binary.BigEndian.Uint32(payload[8:]),  // low (4 bytes)
		binary.BigEndian.Uint32(payload[4:8]), // middle (4 bytes)
		binary.BigEndian.Uint32(payload[:4]),  // high (4 bytes)
	}
}

func (v uint96) ksuid(timestamp uint64) (out KSUID) {
	binary.BigEndian.PutUint64(out[:8], timestamp) // time (8 bytes)
	binary.BigEndian.PutUint32(out[8:12], v[2])    // high (4 bytes)
	binary.BigEndian.PutUint32(out[12:16], v[1])   // middle (4 bytes)
	binary.BigEndian.PutUint32(out[16:], v[0])     // low (4 bytes)
	return
}

// The external byte serialization is big-endian (highest bits in lowest byte address)
func (v uint96) bytes() (out [12]byte) {
	binary.BigEndian.PutUint32(out[:4], v[2])
	binary.BigEndian.PutUint32(out[4:8], v[1])
	binary.BigEndian.PutUint32(out[8:], v[0])
	return
}

func (v uint96) String() string {
	return fmt.Sprintf("0x%08X%08X%08X", v[2], v[1], v[0])
}

func cmp96(x, y uint96) int {
	if x[2] < y[2] {
		return -1
	}
	if x[2] > y[2] {
		return 1
	}
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
	var c uint32
	z[0], c = bits.Add32(x[0], y[0], 0)
	z[1], c = bits.Add32(x[1], y[1], c)
	z[2], _ = bits.Add32(x[2], y[2], c)
	return
}

func sub96(x, y uint96) (z uint96) {
	var b uint32
	z[0], b = bits.Sub32(x[0], y[0], 0)
	z[1], b = bits.Sub32(x[1], y[1], b)
	z[2], _ = bits.Sub32(x[2], y[2], b)
	return
}

func incr96(x uint96) (z uint96) {
	var c uint32
	z[0], c = bits.Add32(x[0], 1, 0)
	z[1], c = bits.Add32(x[1], c, 0)
	z[2], _ = bits.Add32(x[2], c, 0)
	return
}
