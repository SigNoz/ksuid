package ksuid

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// these don't use the optimized version
func TestBase10ToBase62AndBack(t *testing.T) {
	number := []byte{1, 2, 3, 4}
	encoded := base2base(number, 10, 62)
	decoded := base2base(encoded, 62, 10)

	if bytes.Compare(number, decoded) != 0 {
		t.Fatal(number, " != ", decoded)
	}
}

// these don't use the optimized version
func TestBase256ToBase62AndBack(t *testing.T) {
	number := []byte{255, 254, 253, 251}
	encoded := base2base(number, 256, 62)
	decoded := base2base(encoded, 62, 256)

	if bytes.Compare(number, decoded) != 0 {
		t.Fatal(number, " != ", decoded)
	}
}

// these don't use the optimized version
func TestEncodeAndDecodeBase62(t *testing.T) {
	helloWorld := []byte("hello world")
	encoded := encodeBase62(helloWorld)
	decoded := decodeBase62(encoded)

	if len(encoded) < len(helloWorld) {
		t.Fatal("length of encoded base62 string", encoded, "should be >= than raw bytes!")

	}

	if bytes.Compare(helloWorld, decoded) != 0 {
		t.Fatal(decoded, " != ", helloWorld)
	}
}

// these don't use the optimized version
func TestLexographicOrdering(t *testing.T) {
	unsortedStrings := make([]string, 256)
	for i := 0; i < 256; i++ {
		s := string(encodeBase62([]byte{0, byte(i)}))
		unsortedStrings[i] = strings.Repeat("0", 2-len(s)) + s
	}

	if !sort.StringsAreSorted(unsortedStrings) {
		sortedStrings := make([]string, len(unsortedStrings))
		for i, s := range unsortedStrings {
			sortedStrings[i] = s
		}
		sort.Strings(sortedStrings)

		t.Fatal("base62 encoder does not produce lexographically sorted output.",
			"expected:", sortedStrings,
			"actual:", unsortedStrings)
	}
}

// these don't use the optimized version
func TestBase62Value(t *testing.T) {
	s := base62Characters

	for i := range s {
		v := int(base62Value(s[i]))

		if v != i {
			t.Error("bad value:")
			t.Log("<<<", i)
			t.Log(">>>", v)
		}
	}
}

// these use the optimized version and compares the results with the generic version
func TestFastAppendEncodeBase62(t *testing.T) {
	for i := 0; i != 1000; i++ {
		id := New()

		b0 := id[:]
		generic := appendEncodeBase62(nil, b0)
		optimized := fastAppendEncodeBase62(nil, b0)

		s1 := string(leftpad(generic, '0', stringEncodedLength))
		s2 := string(optimized)

		if s1 != s2 {
			t.Error("bad base62 representation of", id)
			t.Log("<<<", s1, len(s1))
			t.Log(">>>", s2, len(s2))
		}
	}
}

func TestFastAppendDecodeBase62(t *testing.T) {
	for i := 0; i != 1000; i++ {
		id := New()
		b0 := leftpad(encodeBase62(id[:]), '0', stringEncodedLength)

		b1 := appendDecodeBase62(nil, []byte(string(b0))) // because it modifies the input buffer
		b2 := fastAppendDecodeBase62(nil, b0)

		if !bytes.Equal(leftpad(b1, 0, byteLength), b2) {
			t.Error("bad binary representation of", string(b0))
			t.Log("<<<", b1)
			t.Log(">>>", b2)
		}
	}
}

func TestEncodeDecodeBase62(t *testing.T) {
	for i := 0; i != 1000; i++ {
		// Create new ID
		id := New()
		original := id[:]

		// Encode using both methods
		normalEncoded := encodeBase62(original)
		fastEncoded := fastAppendEncodeBase62(nil, original)

		// Verify both encoded versions match
		if !bytes.Equal(normalEncoded, fastEncoded) {
			t.Error("encoded versions don't match")
			t.Log("normal encoded:", string(normalEncoded))
			t.Log("fast encoded:", string(fastEncoded))
		}

		// Decode both encoded versions
		normalDecoded := decodeBase62(normalEncoded)
		fastDecoded := fastAppendDecodeBase62(nil, fastEncoded)

		// Compare decoded results with original
		if !bytes.Equal(leftpad(normalDecoded, 0, byteLength), original) {
			t.Error("normal encode/decode failed to match original")
			t.Log("original:", original)
			t.Log("decoded:", normalDecoded)
		}

		if !bytes.Equal(fastDecoded, original) {
			t.Error("fast encode/decode failed to match original")
			t.Log("original:", original)
			t.Log("decoded:", fastDecoded)
		}

		// Compare both decoded versions match each other
		if !bytes.Equal(leftpad(normalDecoded, 0, byteLength), fastDecoded) {
			t.Error("normal and fast decode results don't match")
			t.Log("normal:", normalDecoded)
			t.Log("fast:", fastDecoded)
		}
	}
}

func BenchmarkAppendEncodeBase62(b *testing.B) {
	a := [stringEncodedLength]byte{}
	id := New()

	for i := 0; i != b.N; i++ {
		appendEncodeBase62(a[:0], id[:])
	}
}

func BenchmarkAppendFastEncodeBase62(b *testing.B) {
	a := [stringEncodedLength]byte{}
	id := New()

	for i := 0; i != b.N; i++ {
		fastAppendEncodeBase62(a[:0], id[:])
	}
}

func BenchmarkAppendDecodeBase62(b *testing.B) {
	a := [byteLength]byte{}
	id := []byte(New().String())

	for i := 0; i != b.N; i++ {
		b := [stringEncodedLength]byte{}
		copy(b[:], id)
		appendDecodeBase62(a[:0], b[:])
	}
}

func BenchmarkAppendFastDecodeBase62(b *testing.B) {
	a := [byteLength]byte{}
	id := []byte(New().String())

	for i := 0; i != b.N; i++ {
		fastAppendDecodeBase62(a[:0], id)
	}
}

// The functions bellow were the initial implementation of the base conversion
// algorithms, they were replaced by optimized versions later on. We keep them
// in the test files as a reference to ensure compatibility between the generic
// and optimized implementations.

func appendBase2Base(dst []byte, src []byte, inBase int, outBase int) []byte {
	off := len(dst)
	bs := src[:]
	bq := [stringEncodedLength]byte{}

	for len(bs) > 0 {
		length := len(bs)
		quotient := bq[:0]
		remainder := 0

		for i := 0; i != length; i++ {
			acc := int(bs[i]) + remainder*inBase
			d := acc/outBase | 0
			remainder = acc % outBase

			if len(quotient) > 0 || d > 0 {
				quotient = append(quotient, byte(d))
			}
		}

		// Appends in reverse order, the byte slice gets reversed before it's
		// returned by the function.
		dst = append(dst, byte(remainder))
		bs = quotient
	}

	reverse(dst[off:])
	return dst
}

func base2base(src []byte, inBase int, outBase int) []byte {
	return appendBase2Base(nil, src, inBase, outBase)
}

func appendEncodeBase62(dst []byte, src []byte) []byte {
	off := len(dst)
	dst = appendBase2Base(dst, src, 256, 62)
	for i, c := range dst[off:] {
		dst[off+i] = base62Characters[c]
	}
	return dst
}

func encodeBase62(in []byte) []byte {
	return appendEncodeBase62(nil, in)
}

func appendDecodeBase62(dst []byte, src []byte) []byte {
	// Kind of intrusive, we modify the input buffer... it's OK here, it saves
	// a memory allocation in Parse.
	for i, b := range src {
		// O(1)... technically. Has better real-world perf than a map
		src[i] = byte(strings.IndexByte(base62Characters, b))
	}
	return appendBase2Base(dst, src, 62, 256)
}

func decodeBase62(src []byte) []byte {
	return appendDecodeBase62(
		make([]byte, 0, len(src)*2),
		append(make([]byte, 0, len(src)), src...),
	)
}

func reverse(b []byte) {
	i := 0
	j := len(b) - 1

	for i < j {
		b[i], b[j] = b[j], b[i]
		i++
		j--
	}
}

func leftpad(b []byte, c byte, n int) []byte {
	if n -= len(b); n > 0 {
		for i := 0; i != n; i++ {
			b = append(b, c)
		}

		copy(b[n:], b)

		for i := 0; i != n; i++ {
			b[i] = c
		}
	}
	return b
}

func TestFastEncodeBase62OverflowHandling(t *testing.T) {
	// Create a source with 24 bytes (the expected input size for fastEncodeBase62)
	src := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFE,
	}

	if len(src) != byteLength {
		t.Fatalf("Test source must be exactly %d bytes, got %d", byteLength, len(src))
	}

	// Create properly sized buffers
	normalEncoded := encodeBase62(src)
	fastEncoded := fastAppendEncodeBase62(nil, src)

	fmt.Printf("Normal encoded: %v\n", string(normalEncoded))
	fmt.Printf("Fast encoded: %v\n", string(fastEncoded))

	// Compare encoded results
	if !bytes.Equal(normalEncoded, fastEncoded) {
		t.Errorf("Fast encoding differs from normal encoding\nNormal: %s\nFast:   %s",
			string(normalEncoded), string(fastEncoded))

		// Additional debug information
		t.Logf("Normal encoded length: %d", len(normalEncoded))
		t.Logf("Fast encoded length: %d", len(fastEncoded))
		t.Logf("Expected length: %d", stringEncodedLength)
		t.Logf("Normal first byte: %c", normalEncoded[0])
		t.Logf("Fast first byte: %c", fastEncoded[0])
	}

	// Decode both encoded versions
	normalDecoded := decodeBase62(normalEncoded)
	fastDecoded := make([]byte, byteLength)
	if err := fastDecodeBase62(fastDecoded, fastEncoded); err != nil {
		t.Fatalf("Fast decode failed: %v", err)
	}

	fmt.Printf("Normal decoded: %v\n", normalDecoded)
	fmt.Printf("Fast decoded: %v\n", fastDecoded)
	// Compare decoded results with original
	if !bytes.Equal(normalDecoded, src) {
		t.Errorf("Normal decode failed to match original\nOriginal: %v\nDecoded:  %v",
			src, normalDecoded)
	}

	if !bytes.Equal(fastDecoded, src) {
		t.Errorf("Fast decode failed to match original\nOriginal: %v\nDecoded:  %v",
			src, fastDecoded)
	}

	// Compare both decoded versions
	if !bytes.Equal(normalDecoded, fastDecoded) {
		t.Errorf("Decoded results don't match\nNormal: %v\nFast:   %v",
			normalDecoded, fastDecoded)
	}
}
