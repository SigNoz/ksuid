package ksuid

import (
	"fmt"
	"testing"
)

func TestMakeUint96(t *testing.T) {
	tests := []struct {
		name     string
		high     uint32
		low      uint64
		expected uint96
	}{
		{
			name:     "zero values",
			high:     0,
			low:      0,
			expected: uint96{0, 0, 0},
		},
		{
			name:     "only low bits set",
			high:     0,
			low:      0x123456789A,
			expected: uint96{0x3456789A, 0x12, 0x0},
		},
		{
			name:     "only high bits set",
			high:     0xABCDEF12,
			low:      0,
			expected: uint96{0, 0, 0xABCDEF12},
		},
		{
			name:     "all parts set",
			high:     0x12345678,
			low:      0xABCDEF0123456789,
			expected: uint96{0x23456789, 0xABCDEF01, 0x12345678},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := makeUint96(test.high, test.low)
			if result != test.expected {
				t.Errorf("makeUint96(%#x, %#x) = %v, want %v",
					test.high, test.low, result, test.expected)
			}
		})
	}
}

func TestCmp96(t *testing.T) {
	tests := []struct {
		x      uint96
		y      uint96
		result int
	}{
		{
			x:      makeUint96(0, 0),
			y:      makeUint96(0, 0),
			result: 0,
		},
		{
			x:      makeUint96(0, 1),
			y:      makeUint96(0, 0),
			result: +1,
		},
		{
			x:      makeUint96(0, 0),
			y:      makeUint96(0, 1),
			result: -1,
		},
		{
			x:      makeUint96(1, 0),
			y:      makeUint96(0, 1),
			result: +1,
		},
		{
			x:      makeUint96(0, 1),
			y:      makeUint96(1, 0),
			result: -1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("cmp96(%s,%s)", test.x, test.y), func(t *testing.T) {
			if result := cmp96(test.x, test.y); result != test.result {
				t.Error(result, "!=", test.result)
			}
		})
	}
}

func TestAdd96(t *testing.T) {
	tests := []struct {
		name   string
		x      uint96
		y      uint96
		result uint96
	}{
		{
			name:   "zero plus zero equals zero",
			x:      makeUint96(0, 0),
			y:      makeUint96(0, 0),
			result: makeUint96(0, 0),
		},
		{
			name:   "one plus zero equals one",
			x:      makeUint96(0, 1),
			y:      makeUint96(0, 0),
			result: makeUint96(0, 1),
		},
		{
			name:   "zero plus one equals one",
			x:      makeUint96(0, 0),
			y:      makeUint96(0, 1),
			result: makeUint96(0, 1),
		},
		{
			name:   "high one plus low one",
			x:      makeUint96(1, 0),
			y:      makeUint96(0, 1),
			result: makeUint96(1, 1),
		},
		{
			name:   "low one plus high one",
			x:      makeUint96(0, 1),
			y:      makeUint96(1, 0),
			result: makeUint96(1, 1),
		},
		{
			// x:   0x00000000_00000000_FFFFFFFF
			// y:   0x00000000_00000000_00000001
			// ----------------------------------------
			// sum: 0x00000000_00000001_00000000
			name:   "carry from low to middle word",
			x:      uint96{0xFFFFFFFF, 0, 0},
			y:      uint96{1, 0, 0},
			result: uint96{0, 1, 0},
		},
		{
			name:   "overflow wraps to zero",
			x:      uint96{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			y:      uint96{1, 0, 0},
			result: uint96{0, 0, 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if result := add96(test.x, test.y); result != test.result {
				t.Error(result, "!=", test.result)
			}
		})
	}
}

func TestSub96(t *testing.T) {
	tests := []struct {
		name   string
		x      uint96
		y      uint96
		result uint96
	}{
		{
			name:   "zero minus zero equals zero",
			x:      makeUint96(0, 0),
			y:      makeUint96(0, 0),
			result: makeUint96(0, 0),
		},
		{
			name:   "one minus zero equals one",
			x:      makeUint96(0, 1),
			y:      makeUint96(0, 0),
			result: makeUint96(0, 1),
		},
		{
			name:   "zero minus one equals max value",
			x:      makeUint96(0, 0),
			y:      makeUint96(0, 1),
			result: uint96{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			//    00,000001,000000
			// -  00,000000,000001
			// =  00,000000,FFFFFF
			name:   "borrow from middle word",
			x:      uint96{0, 1, 0},
			y:      uint96{1, 0, 0},
			result: uint96{0xFFFFFFFF, 0, 0},
		},
		{
			name:   "borrow from high word",
			x:      uint96{0, 0, 1},
			y:      uint96{1, 0, 0},
			result: uint96{0xFFFFFFFF, 0xFFFFFFFF, 0},
		},
		{
			name:   "chain of borrows",
			x:      uint96{0, 0, 1},
			y:      uint96{1, 1, 0},
			result: uint96{0xFFFFFFFF, 0xFFFFFFFE, 0},
		},
		{
			name:   "borrow across all words",
			x:      uint96{0, 0, 1},
			y:      uint96{1, 1, 1},
			result: uint96{0xFFFFFFFF, 0xFFFFFFFE, 0xFFFFFFFF},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if result := sub96(test.x, test.y); result != test.result {
				t.Errorf("%s: got %v, want %v", test.name, result, test.result)
			}
		})
	}
}

func TestIncr96(t *testing.T) {
	tests := []struct {
		name   string
		x      uint96
		result uint96
	}{
		{
			name:   "zero plus one equals one",
			x:      makeUint96(0, 0),
			result: makeUint96(0, 1),
		},
		{
			name:   "carry from low to middle word",
			x:      uint96{0xFFFFFFFF, 0, 0},
			result: uint96{0, 1, 0},
		},
		{
			name:   "overflow wraps to zero",
			x:      uint96{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			result: makeUint96(0, 0),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("incr96(%s)", test.x), func(t *testing.T) {
			if result := incr96(test.x); result != test.result {
				t.Error(result, "!=", test.result)
			}
		})
	}
}

func BenchmarkCmp96(b *testing.B) {
	x := makeUint96(0, 0)
	y := makeUint96(0, 0)

	for i := 0; i != b.N; i++ {
		cmp96(x, y)
	}
}

func BenchmarkAdd96(b *testing.B) {
	x := makeUint96(0, 0)
	y := makeUint96(0, 0)

	for i := 0; i != b.N; i++ {
		add96(x, y)
	}
}

func BenchmarkSub96(b *testing.B) {
	x := makeUint96(0, 0)
	y := makeUint96(0, 0)

	for i := 0; i != b.N; i++ {
		sub96(x, y)
	}
}

func BenchmarkIncr96(b *testing.B) {
	x := makeUint96(0, 0)

	for i := 0; i != b.N; i++ {
		incr96(x)
	}
}
