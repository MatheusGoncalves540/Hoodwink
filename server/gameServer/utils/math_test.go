package utils

import "testing"

func TestFloorDiv(t *testing.T) {
	cases := []struct {
		a, b, want int
	}{
		{3, 2, 1},
		{5, 2, 2},
		{-1, 2, -1},
		{-3, 2, -2},
		{-5, 2, -3},
		{0, 2, 0},
		{7, 2, 3},
	}

	for _, c := range cases {
		if got := FloorDiv(c.a, c.b); got != c.want {
			t.Fatalf("FloorDiv(%d, %d) = %d, quer %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCeilDiv(t *testing.T) {
	cases := []struct {
		a, b, want int
	}{
		{3, 2, 2},
		{5, 2, 3},
		{-1, 2, 0},
		{-3, 2, -1},
		{-5, 2, -2},
		{0, 2, 0},
		{7, 2, 4},
	}

	for _, c := range cases {
		if got := CeilDiv(c.a, c.b); got != c.want {
			t.Fatalf("CeilDiv(%d, %d) = %d, quer %d", c.a, c.b, got, c.want)
		}
	}
}
