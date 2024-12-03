package ksuid

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// lexographic ordering (based on Unicode table) is 0-9A-Za-z
	base62Characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	zeroString       = "00000000000000000000000000000000"
	offsetUppercase  = 10
	offsetLowercase  = 36
)

var (
	errShortBuffer = errors.New("the output buffer is too small to hold to decoded value")
)

// Converts a base 62 byte into the number value that it represents.
func base62Value(digit byte) byte {
	switch {
	case digit >= '0' && digit <= '9':
		return digit - '0'
	case digit >= 'A' && digit <= 'Z':
		return offsetUppercase + (digit - 'A')
	default:
		return offsetLowercase + (digit - 'a')
	}
}

// This function encodes the base 62 representation of the src KSUID in binary
// form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 24 bytes long and dst is 32 bytes long.
//
// Any unused bytes in dst will be set to the padding '0' byte.
func fastEncodeBase62(dst []byte, src []byte) {
	const srcBase = 4294967296
	const dstBase = 62

	// Create a temporary buffer for calculations
	var temp [33]byte

	// Initialize with zeros
	for i := range temp {
		temp[i] = '0'
	}

	// Split src into 6 4-byte words
	parts := [6]uint32{
		binary.BigEndian.Uint32(src[0:4]),
		binary.BigEndian.Uint32(src[4:8]),
		binary.BigEndian.Uint32(src[8:12]),
		binary.BigEndian.Uint32(src[12:16]),
		binary.BigEndian.Uint32(src[16:20]),
		binary.BigEndian.Uint32(src[20:24]),
	}

	// Process all parts
	partsSlice := parts[:]
	n := len(temp)

	for len(partsSlice) > 0 {
		var remainder uint64
		quotient := make([]uint32, 0, len(partsSlice))

		// Process each part
		for _, part := range partsSlice {
			acc := uint64(part) + remainder*srcBase
			digit := acc / dstBase
			remainder = acc % dstBase

			if len(quotient) > 0 || digit > 0 {
				quotient = append(quotient, uint32(digit))
			}
		}

		// Write the digit
		n--
		if n >= 0 {
			temp[n] = base62Characters[remainder]
		}

		partsSlice = quotient

	}
	// Pad with zeros if we have less than 32 characters
	// Pad only the remaining unwritten positions from 0 to n-1
	for i := 0; i < n; i++ {
		temp[i] = '0'
	}

	// Copy the entire result to dst, including leading zeros
	copy(dst, temp[:])
}

// This function appends the base 62 representation of the KSUID in src to dst,
// and returns the extended byte slice.
// The result is left-padded with '0' bytes to always append 27 bytes to the
// destination buffer.
func fastAppendEncodeBase62(dst []byte, src []byte) []byte {
	// Always allocate 33 bytes to handle overflow cases
	dst = reserve(dst, 33)
	n := len(dst)
	result := dst[n : n+33]

	// Encode into the full buffer
	fastEncodeBase62(result, src)

	fmt.Printf("After encoding: %s\n", string(result))

	// Find first non-zero character

	start := 0
	if start < len(result) && result[start] == '0' {
		start++
	}
	if start == len(result) {
		start = len(result) - 1
	}

	fmt.Printf("Start position: %d\n", start)

	// Create new slice with just the significant portion
	significant := result[start:]

	// Create final slice at the correct position
	final := dst[n : n+len(significant)]

	// Copy the significant digits
	copy(final, significant)

	return dst[:n+len(significant)]
}

// This function decodes the base 62 representation of the src KSUID to the
// binary form into dst.
//
// In order to support a couple of optimizations the function assumes that src
// is 27 bytes long and dst is 20 bytes long.
//
// Any unused bytes in dst will be set to zero.
func fastDecodeBase62(dst []byte, src []byte) error {
	const srcBase = 62
	const dstBase = 4294967296 // 2^32

	// Determine the actual size needed based on input length
	size := 32
	if len(src) == 33 {
		size = 33
	}

	// Convert each character to its base62 value
	parts := make([]byte, size)
	for i := 0; i < len(src) && i < size; i++ {
		parts[i] = base62Value(src[i])
	}

	// Process in 32-bit chunks
	n := len(dst)
	bp := parts[:]
	bq := make([]byte, size) // Make buffer same size as input

	for len(bp) > 0 {
		quotient := bq[:0]
		remainder := uint64(0)

		for _, c := range bp {
			value := uint64(c) + remainder*srcBase
			digit := value / dstBase
			remainder = value % dstBase

			if len(quotient) != 0 || digit != 0 {
				quotient = append(quotient, byte(digit))
			}
		}

		if n < 4 {
			return errShortBuffer
		}

		// Write the remainder as a 32-bit big-endian integer
		dst[n-4] = byte(remainder >> 24)
		dst[n-3] = byte(remainder >> 16)
		dst[n-2] = byte(remainder >> 8)
		dst[n-1] = byte(remainder)
		n -= 4
		bp = quotient
	}

	// Zero out any remaining bytes
	for i := 0; i < n; i++ {
		dst[i] = 0
	}
	return nil
}

// This function appends the base 62 decoded version of src into dst.
func fastAppendDecodeBase62(dst []byte, src []byte) []byte {
	dst = reserve(dst, byteLength)
	n := len(dst)
	fastDecodeBase62(dst[n:n+byteLength], src)
	return dst[:n+byteLength]
}

// Ensures that at least nbytes are available in the remaining capacity of the
// destination slice, if not, a new copy is made and returned by the function.
func reserve(dst []byte, nbytes int) []byte {
	c := cap(dst)
	n := len(dst)

	if avail := c - n; avail < nbytes {
		c *= 2
		if (c - n) < nbytes {
			c = n + nbytes
		}
		b := make([]byte, n, c)
		copy(b, dst)
		dst = b
	}

	return dst
}
