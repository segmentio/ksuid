package ksuid

import (
	"fmt"
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
