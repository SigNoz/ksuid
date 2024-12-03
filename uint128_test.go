package ksuid

import (
	"fmt"
	"math"
	"testing"
)

func TestCmp128(t *testing.T) {
	tests := []struct {
		x uint128
		y uint128
		k int
	}{
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 0),
			k: 0,
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(0, 0),
			k: +1,
		},
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 1),
			k: -1,
		},
		{
			x: makeUint128(1, 0),
			y: makeUint128(0, 1),
			k: +1,
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(1, 0),
			k: -1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("cmp128(%s,%s)", test.x, test.y), func(t *testing.T) {
			if k := cmp128(test.x, test.y); k != test.k {
				t.Error(k, "!=", test.k)
			}
		})
	}
}

func TestAdd128(t *testing.T) {
	tests := []struct {
		x uint128
		y uint128
		z uint128
	}{
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 0),
			z: makeUint128(0, 0),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(0, 0),
			z: makeUint128(0, 1),
		},
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 1),
			z: makeUint128(0, 1),
		},
		{
			x: makeUint128(1, 0),
			y: makeUint128(0, 1),
			z: makeUint128(1, 1),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(1, 0),
			z: makeUint128(1, 1),
		},
		{
			x: makeUint128(0, 0xFFFFFFFFFFFFFFFF),
			y: makeUint128(0, 1),
			z: makeUint128(1, 0),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("add128(%s,%s)", test.x, test.y), func(t *testing.T) {
			if z := add128(test.x, test.y); z != test.z {
				t.Error(z, "!=", test.z)
			}
		})
	}
}

func TestSub128(t *testing.T) {
	tests := []struct {
		x uint128
		y uint128
		z uint128
	}{
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 0),
			z: makeUint128(0, 0),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(0, 0),
			z: makeUint128(0, 1),
		},
		{
			x: makeUint128(0, 0),
			y: makeUint128(0, 1),
			z: makeUint128(0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF),
		},
		{
			x: makeUint128(1, 0),
			y: makeUint128(0, 1),
			z: makeUint128(0, 0xFFFFFFFFFFFFFFFF),
		},
		{
			x: makeUint128(0, 1),
			y: makeUint128(1, 0),
			z: makeUint128(0xFFFFFFFFFFFFFFFF, 1),
		},
		{
			x: makeUint128(0, 0xFFFFFFFFFFFFFFFF),
			y: makeUint128(0, 1),
			z: makeUint128(0, 0xFFFFFFFFFFFFFFFE),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sub128(%s,%s)", test.x, test.y), func(t *testing.T) {
			if z := sub128(test.x, test.y); z != test.z {
				t.Error(z, "!=", test.z)
			}
		})
	}
}

func BenchmarkCmp128(b *testing.B) {
	x := makeUint128(0, 0)
	y := makeUint128(0, 0)

	for i := 0; i != b.N; i++ {
		cmp128(x, y)
	}
}

func BenchmarkAdd128(b *testing.B) {
	x := makeUint128(0, 0)
	y := makeUint128(0, 0)

	for i := 0; i != b.N; i++ {
		add128(x, y)
	}
}

func BenchmarkSub128(b *testing.B) {
	x := makeUint128(0, 0)
	y := makeUint128(0, 0)

	for i := 0; i != b.N; i++ {
		sub128(x, y)
	}
}

func TestUint128WithLargeTimestamp(t *testing.T) {
	// Test cases with 8-byte timestamps
	tests := []struct {
		name      string
		timestamp uint64
		payload   uint128
		expected  KSUID
	}{
		{
			name:      "max_timestamp_zero_payload",
			timestamp: math.MaxUint64,
			payload:   makeUint128(0, 0),
			expected: KSUID{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // max timestamp
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // payload hi
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // payload lo
			},
		},
		{
			name:      "zero_timestamp_max_payload",
			timestamp: 0,
			payload:   makeUint128(math.MaxUint64, math.MaxUint64),
			expected: KSUID{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // zero timestamp
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // payload hi
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // payload lo
			},
		},
		{
			name:      "large_timestamp_mixed_payload",
			timestamp: uint64(1 << 63), // Large timestamp
			payload:   makeUint128(0xAAAAAAAAAAAAAAAA, 0x5555555555555555),
			expected: KSUID{
				0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // large timestamp
				0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, // payload hi
				0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, // payload lo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ksuid method
			result := tt.payload.ksuid(tt.timestamp)
			if result != tt.expected {
				t.Errorf("ksuid() = %v, want %v", result, tt.expected)
			}

			// Test uint128Payload method
			extractedPayload := uint128Payload(result)
			if extractedPayload != tt.payload {
				t.Errorf("uint128Payload() = %v, want %v", extractedPayload, tt.payload)
			}

			// Verify timestamp
			if result.Timestamp() != tt.timestamp {
				t.Errorf("Timestamp() = %v, want %v", result.Timestamp(), tt.timestamp)
			}
		})
	}
}

func TestUint128PayloadRoundTrip(t *testing.T) {
	// Create a KSUID with known values
	id := KSUID{
		0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, // timestamp
		0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, // payload hi
		0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, // payload lo
	}

	// Extract payload
	payload := uint128Payload(id)

	// Convert back to KSUID
	timestamp := id.Timestamp()
	newID := payload.ksuid(timestamp)

	// Compare
	if newID != id {
		t.Errorf("Round trip failed:\nOriginal: %v\nGot:      %v", id, newID)
	}
}
