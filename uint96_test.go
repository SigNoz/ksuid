package ksuid

import (
	"fmt"
	"testing"
)

func TestCmp96(t *testing.T) {
	tests := []struct {
		x uint96
		y uint96
		k int
	}{
		{
			x: makeUint96(0, 0),
			y: makeUint96(0, 0),
			k: 0,
		},
		{
			x: makeUint96(0, 1),
			y: makeUint96(0, 0),
			k: +1,
		},
		{
			x: makeUint96(0, 0),
			y: makeUint96(0, 1),
			k: -1,
		},
		{
			x: makeUint96(1, 0),
			y: makeUint96(0, 1),
			k: +1,
		},
		{
			x: makeUint96(0, 1),
			y: makeUint96(1, 0),
			k: -1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("cmp96(%s,%s)", test.x, test.y), func(t *testing.T) {
			if k := cmp96(test.x, test.y); k != test.k {
				t.Error(k, "!=", test.k)
			}
		})
	}
}

func TestAdd96(t *testing.T) {
	tests := []struct {
		x uint96
		y uint96
		z uint96
	}{
		{
			x: makeUint96(0, 0),
			y: makeUint96(0, 0),
			z: makeUint96(0, 0),
		},
		{
			x: makeUint96(0, 1),
			y: makeUint96(0, 0),
			z: makeUint96(0, 1),
		},
		{
			x: makeUint96(0, 0),
			y: makeUint96(0, 1),
			z: makeUint96(0, 1),
		},
		{
			x: makeUint96(1, 0),
			y: makeUint96(0, 1),
			z: makeUint96(1, 1),
		},
		{
			x: makeUint96(0, 1),
			y: makeUint96(1, 0),
			z: makeUint96(1, 1),
		},
		{
			x: makeUint96(0, 0xFFFFFFFFFFFFFFFF),
			y: makeUint96(0, 1),
			z: makeUint96(1, 0),
		},
		// Test 32-bit overflow masking
		{
			x: makeUint96(0xFFFFFFFF, 0),
			y: makeUint96(1, 0),
			z: makeUint96(0, 0), // Should wrap around due to 32-bit mask
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("add96(%s,%s)", test.x, test.y), func(t *testing.T) {
			if z := add96(test.x, test.y); z != test.z {
				t.Error(z, "!=", test.z)
			}
		})
	}
}

func TestSub96(t *testing.T) {
	tests := []struct {
		x uint96
		y uint96
		z uint96
	}{
		{
			x: makeUint96(0, 0),
			y: makeUint96(0, 0),
			z: makeUint96(0, 0),
		},
		{
			x: makeUint96(0, 1),
			y: makeUint96(0, 0),
			z: makeUint96(0, 1),
		},
		{
			x: makeUint96(0, 0),
			y: makeUint96(0, 1),
			z: makeUint96(0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF),
		},
		{
			x: makeUint96(1, 0),
			y: makeUint96(0, 1),
			z: makeUint96(0, 0xFFFFFFFFFFFFFFFF),
		},
		{
			x: makeUint96(0, 1),
			y: makeUint96(1, 0),
			z: makeUint96(0xFFFFFFFF, 1),
		},
		// Test 32-bit underflow masking
		{
			x: makeUint96(0, 0),
			y: makeUint96(1, 0),
			z: makeUint96(0xFFFFFFFF, 0),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sub96(%s,%s)", test.x, test.y), func(t *testing.T) {
			if z := sub96(test.x, test.y); z != test.z {
				t.Error(z, "!=", test.z)
			}
		})
	}
}

func TestIncr96(t *testing.T) {
	tests := []struct {
		x uint96
		z uint96
	}{
		{
			x: makeUint96(0, 0),
			z: makeUint96(0, 1),
		},
		{
			x: makeUint96(0, 0xFFFFFFFFFFFFFFFF),
			z: makeUint96(1, 0),
		},
		{
			x: makeUint96(0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF),
			z: makeUint96(0, 0), // Should wrap around due to 32-bit mask
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("incr96(%s)", test.x), func(t *testing.T) {
			if z := incr96(test.x); z != test.z {
				t.Error(z, "!=", test.z)
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
