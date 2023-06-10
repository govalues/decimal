package decimal

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

func TestFint_add(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
		wantOk         bool
	}{
		{0, 0, 0, true},
		{maxFint, 1, 0, false},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef, gotOk := x.add(y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.add(%v) = %v, %v, want %v, %v", x, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_mul(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
		wantOk         bool
	}{
		{0, 0, 0, true},
		{10, 10, 100, true},
		{maxFint, 2, 0, false},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef, gotOk := x.mul(y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.mul(%v) = %v, %v, want %v, %v", x, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_quo(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
		wantOk         bool
	}{
		{1, 0, 0, false},
		{1, 1, 1, true},
		{2, 4, 0, false},
		{20, 4, 5, true},
		{20, 3, 0, false},
		{maxFint, 2, 0, false},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef, gotOk := x.quo(y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.quo(%v) = %v, %v, want %v, %v", x, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_dist(t *testing.T) {
	cases := []struct {
		x, y, wantCoef fint
	}{
		{1, 0, 1},
		{1, 1, 0},
		{2, 4, 2},
		{20, 4, 16},
		{20, 3, 17},
		{maxFint, 2, 9_999_999_999_999_999_997},
	}
	for _, tt := range cases {
		x, y := tt.x, tt.y
		gotCoef := x.dist(y)
		if gotCoef != tt.wantCoef {
			t.Errorf("%v.dist(%v) = %v, want %v", x, y, gotCoef, tt.wantCoef)
		}
	}
}

func TestFint_lsh(t *testing.T) {
	cases := []struct {
		x        fint
		shift    int
		wantCoef fint
		wantOk   bool
	}{
		{0, 0, 0, true},
		{10, 0, 10, true},
		{0, 1, 0, true},
		{10, 1, 100, true},
		{1, 20, 0, false},
	}
	for _, tt := range cases {
		x := tt.x
		gotCoef, gotOk := x.lsh(tt.shift)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.lsh(%v) = %v, %v, want %v, %v", x, tt.shift, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_fsa(t *testing.T) {
	cases := []struct {
		x        fint
		shift    int
		y        byte
		wantCoef fint
		wantOk   bool
	}{
		{0, 0, 0, 0, true},
		{1, 1, 1, 11, true},
		{2, 2, 2, 202, true},
		{3, 3, 3, 3_003, true},
		{1, 20, 0, 0, false},
		{maxFint, 0, 1, 0, false},
	}
	for _, tt := range cases {
		x, y, shift := tt.x, tt.y, tt.shift
		gotCoef, gotOk := x.fsa(shift, y)
		if gotCoef != tt.wantCoef || gotOk != tt.wantOk {
			t.Errorf("%v.fsa(%v, %v) = %v, %v, want %v, %v", x, shift, y, gotCoef, gotOk, tt.wantCoef, tt.wantOk)
		}
	}
}

func TestFint_rshHalfEven(t *testing.T) {
	cases := []struct {
		x     fint
		shift int
		want  fint
	}{
		// Negative shift
		{1, -1, 1},
		// Rounding
		{1, 0, 1},
		{20, 1, 2},
		{18, 1, 2},
		{15, 1, 2},
		{12, 1, 1},
		{10, 1, 1},
		{8, 1, 1},
		{5, 1, 0},
		{2, 1, 0},
		{maxFint, 19, 1},
		// Large shifts
		{0, 17, 0},
		{0, 18, 0},
		{0, 19, 0},
		{0, 20, 0},
		{0, 21, 0},
		{1, 17, 0},
		{1, 18, 0},
		{1, 19, 0},
		{1, 20, 0},
		{1, 21, 0},
		{5_000_000_000_000_000_000, 17, 50},
		{5_000_000_000_000_000_000, 18, 5},
		{5_000_000_000_000_000_000, 19, 0},
		{5_000_000_000_000_000_000, 20, 0},
		{5_000_000_000_000_000_000, 21, 0},
		{5_000_000_000_000_000_001, 17, 50},
		{5_000_000_000_000_000_001, 18, 5},
		{5_000_000_000_000_000_001, 19, 1},
		{5_000_000_000_000_000_001, 20, 0},
		{5_000_000_000_000_000_001, 21, 0},
		{maxFint, 17, 100},
		{maxFint, 18, 10},
		{maxFint, 19, 1},
		{maxFint, 20, 0},
		{maxFint, 21, 0},
		{10_000_000_000_000_000_000, 17, 100},
		{10_000_000_000_000_000_000, 18, 10},
		{10_000_000_000_000_000_000, 19, 1},
		{10_000_000_000_000_000_000, 20, 0},
		{10_000_000_000_000_000_000, 21, 0},
		{14_999_999_999_999_999_999, 17, 150},
		{14_999_999_999_999_999_999, 18, 15},
		{14_999_999_999_999_999_999, 19, 1},
		{14_999_999_999_999_999_999, 20, 0},
		{14_999_999_999_999_999_999, 21, 0},
		{15_000_000_000_000_000_000, 17, 150},
		{15_000_000_000_000_000_000, 18, 15},
		{15_000_000_000_000_000_000, 19, 2},
		{15_000_000_000_000_000_000, 20, 0},
		{15_000_000_000_000_000_000, 21, 0},
		{math.MaxUint64, 17, 184},
		{math.MaxUint64, 18, 18},
		{math.MaxUint64, 19, 2},
		{math.MaxUint64, 20, 0},
		{math.MaxUint64, 21, 0},
	}
	for _, tt := range cases {
		got := tt.x.rshHalfEven(tt.shift)
		if got != tt.want {
			t.Errorf("%v.rshHalfEven(%v) = %v, want %v", tt.x, tt.shift, got, tt.want)
		}
	}
}

func TestFint_rshUp(t *testing.T) {
	cases := []struct {
		x     fint
		shift int
		want  fint
	}{
		// Negative shift
		{1, -1, 1},
		// Rounding
		{20, 1, 2},
		{18, 1, 2},
		{15, 1, 2},
		{12, 1, 2},
		{10, 1, 1},
		{8, 1, 1},
		{5, 1, 1},
		{2, 1, 1},
		// Large shifts
		{0, 17, 0},
		{0, 18, 0},
		{0, 19, 0},
		{0, 20, 0},
		{0, 21, 0},
		{1, 17, 1},
		{1, 18, 1},
		{1, 19, 1},
		{1, 20, 1},
		{1, 21, 1},
		{maxFint, 17, 100},
		{maxFint, 18, 10},
		{maxFint, 19, 1},
		{maxFint, 20, 1},
		{maxFint, 21, 1},
		{10_000_000_000_000_000_000, 17, 100},
		{10_000_000_000_000_000_000, 18, 10},
		{10_000_000_000_000_000_000, 19, 1},
		{10_000_000_000_000_000_000, 20, 1},
		{10_000_000_000_000_000_000, 21, 1},
		{math.MaxUint64, 17, 185},
		{math.MaxUint64, 18, 19},
		{math.MaxUint64, 19, 2},
		{math.MaxUint64, 20, 1},
		{math.MaxUint64, 21, 1},
	}
	for _, tt := range cases {
		got := tt.x.rshUp(tt.shift)
		if got != tt.want {
			t.Errorf("%v.rshUp(%v) = %v, want %v", tt.x, tt.shift, got, tt.want)
		}
	}
}

func TestFint_rshDown(t *testing.T) {
	cases := []struct {
		x     fint
		shift int
		want  fint
	}{
		// Negative shift
		{1, -1, 1},
		// Rounding
		{1, 0, 1},
		{20, 1, 2},
		{18, 1, 1},
		{15, 1, 1},
		{12, 1, 1},
		{10, 1, 1},
		{8, 1, 0},
		{5, 1, 0},
		{2, 1, 0},

		// Large shifts
		{0, 17, 0},
		{0, 18, 0},
		{0, 19, 0},
		{0, 20, 0},
		{0, 21, 0},
		{1, 17, 0},
		{1, 18, 0},
		{1, 19, 0},
		{1, 20, 0},
		{1, 21, 0},
		{maxFint, 17, 99},
		{maxFint, 18, 9},
		{maxFint, 19, 0},
		{maxFint, 20, 0},
		{maxFint, 21, 0},
		{10_000_000_000_000_000_000, 17, 100},
		{10_000_000_000_000_000_000, 18, 10},
		{10_000_000_000_000_000_000, 19, 1},
		{10_000_000_000_000_000_000, 20, 0},
		{10_000_000_000_000_000_000, 21, 0},
		{math.MaxUint64, 17, 184},
		{math.MaxUint64, 18, 18},
		{math.MaxUint64, 19, 1},
		{math.MaxUint64, 20, 0},
		{math.MaxUint64, 21, 0},
	}
	for _, tt := range cases {
		got := tt.x.rshDown(tt.shift)
		if got != tt.want {
			t.Errorf("%v.rshDown(%v) = %v, want %v", tt.x, tt.shift, got, tt.want)
		}
	}
}

func TestFint_prec(t *testing.T) {
	cases := []struct {
		x    fint
		want int
	}{
		{0, 0},
		{1, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{999, 3},
		{1_000, 4},
		{9_999, 4},
		{10_000, 5},
		{99_999, 5},
		{100_000, 6},
		{999_999, 6},
		{1_000_000, 7},
		{9_999_999, 7},
		{10_000_000, 8},
		{99_999_999, 8},
		{100_000_000, 9},
		{999_999_999, 9},
		{1_000_000_000, 10},
		{9_999_999_999, 10},
		{10_000_000_000, 11},
		{99_999_999_999, 11},
		{100_000_000_000, 12},
		{999_999_999_999, 12},
		{1_000_000_000_000, 13},
		{9_999_999_999_999, 13},
		{10_000_000_000_000, 14},
		{99_999_999_999_999, 14},
		{100_000_000_000_000, 15},
		{999_999_999_999_999, 15},
		{1_000_000_000_000_000, 16},
		{9_999_999_999_999_999, 16},
		{10_000_000_000_000_000, 17},
		{99_999_999_999_999_999, 17},
		{100_000_000_000_000_000, 18},
		{999_999_999_999_999_999, 18},
		{1_000_000_000_000_000_000, 19},
		{maxFint, 19},
		{10_000_000_000_000_000_000, 20},
		{math.MaxUint64, 20},
	}
	for _, tt := range cases {
		got := tt.x.prec()
		if got != tt.want {
			t.Errorf("%v.prec() = %v, want %v", tt.x, got, tt.want)
		}
	}
}

func TestFint_tzeros(t *testing.T) {
	cases := []struct {
		x    fint
		want int
	}{
		{0, 0},
		{1, 0},
		{9, 0},
		{10, 1},
		{99, 0},
		{100, 2},
		{999, 0},
		{1_000, 3},
		{9_999, 0},
		{10_000, 4},
		{99_999, 0},
		{100_000, 5},
		{999_999, 0},
		{1_000_000, 6},
		{9_999_999, 0},
		{10_000_000, 7},
		{99_999_999, 0},
		{100_000_000, 8},
		{999_999_999, 0},
		{1_000_000_000, 9},
		{9_999_999_999, 0},
		{10_000_000_000, 10},
		{99_999_999_999, 0},
		{100_000_000_000, 11},
		{999_999_999_999, 0},
		{1_000_000_000_000, 12},
		{9_999_999_999_999, 0},
		{10_000_000_000_000, 13},
		{99_999_999_999_999, 0},
		{100_000_000_000_000, 14},
		{999_999_999_999_999, 0},
		{1_000_000_000_000_000, 15},
		{9_999_999_999_999_999, 0},
		{10_000_000_000_000_000, 16},
		{99_999_999_999_999_999, 0},
		{100_000_000_000_000_000, 17},
		{999_999_999_999_999_999, 0},
		{1_000_000_000_000_000_000, 18},
		{1_000_000_000_000_000_001, 0},
		{maxFint, 0},
		{10_000_000_000_000_000_000, 19},
		{math.MaxUint64, 0},
	}
	for _, tt := range cases {
		got := tt.x.tzeros()
		if got != tt.want {
			t.Errorf("%v.tzeros() = %v, want %v", tt.x, got, tt.want)
		}
	}
}

func TestFint_hasPrec(t *testing.T) {
	cases := []struct {
		x    fint
		prec int
		want bool
	}{
		{0, 0, true},
		{0, 1, false},
		{0, 2, false},
		{1, 0, true},
		{1, 1, true},
		{1, 2, false},
		{9, 0, true},
		{9, 1, true},
		{9, 2, false},
		{10, 0, true},
		{10, 1, true},
		{10, 2, true},
		{10, 3, false},
		{99, 0, true},
		{99, 1, true},
		{99, 2, true},
		{99, 3, false},
		{100_000_000_000_000, 17, false},
		{1_000_000_000_000_000, 17, false},
		{10_000_000_000_000_000, 17, true},
		{100_000_000_000_000_000, 17, true},
		{1_000_000_000_000_000_000, 17, true},
		{1_000_000_000_000_000_000, 17, true},
		{1_000_000_000_000_000_000, 18, true},
		{1_000_000_000_000_000_000, 19, true},
		{1_000_000_000_000_000_000, 20, false},
		{1_000_000_000_000_000_000, 21, false},
		{maxFint, 17, true},
		{maxFint, 18, true},
		{maxFint, 19, true},
		{maxFint, 20, false},
		{maxFint, 21, false},
		{10_000_000_000_000_000_000, 17, true},
		{10_000_000_000_000_000_000, 18, true},
		{10_000_000_000_000_000_000, 19, true},
		{10_000_000_000_000_000_000, 20, true},
		{10_000_000_000_000_000_000, 21, false},
		{math.MaxUint64, 17, true},
		{math.MaxUint64, 18, true},
		{math.MaxUint64, 19, true},
		{math.MaxUint64, 20, true},
		{math.MaxUint64, 21, false},
	}
	for _, tt := range cases {
		got := tt.x.hasPrec(tt.prec)
		if got != tt.want {
			t.Errorf("%v.hasPrec(%v) = %v, want %v", tt.x, tt.prec, got, tt.want)
		}
	}
}

func TestSint_rshHalfEven(t *testing.T) {
	cases := []struct {
		z     string
		shift int
		want  string
	}{
		// Rounding
		{"1", 0, "1"},
		{"20", 1, "2"},
		{"18", 1, "2"},
		{"15", 1, "2"},
		{"12", 1, "1"},
		{"10", 1, "1"},
		{"8", 1, "1"},
		{"5", 1, "0"},
		{"2", 1, "0"},
		{"9999999999999999999", 19, "1"},

		// Large shifts
		{"0", 17, "0"},
		{"0", 18, "0"},
		{"0", 19, "0"},
		{"0", 20, "0"},
		{"0", 21, "0"},
		{"1", 17, "0"},
		{"1", 18, "0"},
		{"1", 19, "0"},
		{"1", 20, "0"},
		{"1", 21, "0"},
		{"5000000000000000000", 17, "50"},
		{"5000000000000000000", 18, "5"},
		{"5000000000000000000", 19, "0"},
		{"5000000000000000000", 20, "0"},
		{"5000000000000000000", 21, "0"},
		{"5000000000000000001", 17, "50"},
		{"5000000000000000001", 18, "5"},
		{"5000000000000000001", 19, "1"},
		{"5000000000000000001", 20, "0"},
		{"5000000000000000001", 21, "0"},
		{"9999999999999999999", 17, "100"},
		{"9999999999999999999", 18, "10"},
		{"9999999999999999999", 19, "1"},
		{"9999999999999999999", 20, "0"},
		{"9999999999999999999", 21, "0"},
		{"10000000000000000000", 17, "100"},
		{"10000000000000000000", 18, "10"},
		{"10000000000000000000", 19, "1"},
		{"10000000000000000000", 20, "0"},
		{"10000000000000000000", 21, "0"},
		{"14999999999999999999", 17, "150"},
		{"14999999999999999999", 18, "15"},
		{"14999999999999999999", 19, "1"},
		{"14999999999999999999", 20, "0"},
		{"14999999999999999999", 21, "0"},
		{"15000000000000000000", 17, "150"},
		{"15000000000000000000", 18, "15"},
		{"15000000000000000000", 19, "2"},
		{"15000000000000000000", 20, "0"},
		{"15000000000000000000", 21, "0"},
		{"18446744073709551615", 17, "184"},
		{"18446744073709551615", 18, "18"},
		{"18446744073709551615", 19, "2"},
		{"18446744073709551615", 20, "0"},
		{"18446744073709551615", 21, "0"},
	}
	for _, tt := range cases {
		got := mustParseSint(tt.z)
		got.rshHalfEven(got, tt.shift)
		want := mustParseSint(tt.want)
		if got.cmp(want) != 0 {
			t.Errorf("%v.rshHalfEven(%v) = %v, want %v", tt.z, tt.shift, got, want)
		}
	}
}

func TestSint_lsh(t *testing.T) {
	cases := []struct {
		z     string
		shift int
		want  string
	}{
		{"0", 1, "0"},
		{"1", 1, "10"},
		{"1", 20, "100000000000000000000"},
		{"1", 100, "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
	}
	for _, tt := range cases {
		got := mustParseSint(tt.z)
		got.lsh(got, tt.shift)
		want := mustParseSint(tt.want)
		if got.cmp(want) != 0 {
			t.Errorf("%v.lsh(%v) = %v, want %v", tt.z, tt.shift, got, want)
		}
	}
}

func TestSint_prec(t *testing.T) {
	cases := []struct {
		z    string
		want int
	}{
		{"0", 0},
		{"1", 1},
		{"9", 1},
		{"10", 2},
		{"99", 2},
		{"100", 3},
		{"999", 3},
		{"1000", 4},
		{"9999", 4},
		{"10000", 5},
		{"99999", 5},
		{"100000", 6},
		{"999999", 6},
		{"1000000", 7},
		{"9999999", 7},
		{"10000000", 8},
		{"99999999", 8},
		{"100000000", 9},
		{"999999999", 9},
		{"1000000000", 10},
		{"9999999999", 10},
		{"10000000000", 11},
		{"99999999999", 11},
		{"100000000000", 12},
		{"999999999999", 12},
		{"1000000000000", 13},
		{"9999999999999", 13},
		{"10000000000000", 14},
		{"99999999999999", 14},
		{"100000000000000", 15},
		{"999999999999999", 15},
		{"1000000000000000", 16},
		{"9999999999999999", 16},
		{"10000000000000000", 17},
		{"99999999999999999", 17},
		{"100000000000000000", 18},
		{"999999999999999999", 18},
		{"1000000000000000000", 19},
		{"9999999999999999999", 19},
		{"10000000000000000000", 20},
		{"99999999999999999999", 20},
		{"100000000000000000000", 21},
		{"999999999999999999999", 21},
		{"1000000000000000000000", 22},
		{"9999999999999999999999", 22},
		{"10000000000000000000000", 23},
		{"99999999999999999999999", 23},
		{"100000000000000000000000", 24},
		{"999999999999999999999999", 24},
		{"1000000000000000000000000", 25},
		{"9999999999999999999999999", 25},
		{"10000000000000000000000000", 26},
		{"99999999999999999999999999", 26},
		{"100000000000000000000000000", 27},
		{"999999999999999999999999999", 27},
		{"1000000000000000000000000000", 28},
		{"9999999999999999999999999999", 28},
		{"10000000000000000000000000000", 29},
		{"99999999999999999999999999999", 29},
		{"100000000000000000000000000000", 30},
		{"999999999999999999999999999999", 30},
		{"1000000000000000000000000000000", 31},
		{"9999999999999999999999999999999", 31},
		{"10000000000000000000000000000000", 32},
		{"99999999999999999999999999999999", 32},
		{"100000000000000000000000000000000", 33},
		{"999999999999999999999999999999999", 33},
		{"1000000000000000000000000000000000", 34},
		{"9999999999999999999999999999999999", 34},
		{"10000000000000000000000000000000000", 35},
		{"99999999999999999999999999999999999", 35},
		{"100000000000000000000000000000000000", 36},
		{"999999999999999999999999999999999999", 36},
		{"1000000000000000000000000000000000000", 37},
		{"9999999999999999999999999999999999999", 37},
		{"10000000000000000000000000000000000000", 38},
		{"99999999999999999999999999999999999999", 38},
		{"100000000000000000000000000000000000000", 39},
		{"999999999999999999999999999999999999999", 39},
		{"1000000000000000000000000000000000000000", 40},
		{"9999999999999999999999999999999999999999", 40},
		{"10000000000000000000000000000000000000000", 41},
		{"99999999999999999999999999999999999999999", 41},
		{"100000000000000000000000000000000000000000", 42},
		{"999999999999999999999999999999999999999999", 42},
		{"1000000000000000000000000000000000000000000", 43},
		{"9999999999999999999999999999999999999999999", 43},
		{"10000000000000000000000000000000000000000000", 44},
		{"99999999999999999999999999999999999999999999", 44},
		{"100000000000000000000000000000000000000000000", 45},
		{"999999999999999999999999999999999999999999999", 45},
		{"1000000000000000000000000000000000000000000000", 46},
		{"9999999999999999999999999999999999999999999999", 46},
		{"10000000000000000000000000000000000000000000000", 47},
		{"99999999999999999999999999999999999999999999999", 47},
		{"100000000000000000000000000000000000000000000000", 48},
		{"999999999999999999999999999999999999999999999999", 48},
		{"1000000000000000000000000000000000000000000000000", 49},
		{"9999999999999999999999999999999999999999999999999", 49},
		{"10000000000000000000000000000000000000000000000000", 50},
		{"99999999999999999999999999999999999999999999999999", 50},
		{"100000000000000000000000000000000000000000000000000", 51},
		{"999999999999999999999999999999999999999999999999999", 51},
		{"1000000000000000000000000000000000000000000000000000", 52},
		{"9999999999999999999999999999999999999999999999999999", 52},
		{"10000000000000000000000000000000000000000000000000000", 53},
		{"99999999999999999999999999999999999999999999999999999", 53},
		{"100000000000000000000000000000000000000000000000000000", 54},
		{"999999999999999999999999999999999999999999999999999999", 54},
		{"1000000000000000000000000000000000000000000000000000000", 55},
		{"9999999999999999999999999999999999999999999999999999999", 55},
		{"10000000000000000000000000000000000000000000000000000000", 56},
		{"99999999999999999999999999999999999999999999999999999999", 56},
		{"100000000000000000000000000000000000000000000000000000000", 57},
		{"999999999999999999999999999999999999999999999999999999999", 57},
		{"1000000000000000000000000000000000000000000000000000000000", 58},
		{"9999999999999999999999999999999999999999999999999999999999", 58},
		{"10000000000000000000000000000000000000000000000000000000000", 59},
		{"99999999999999999999999999999999999999999999999999999999999", 59},
		{"100000000000000000000000000000000000000000000000000000000000", 60},
		{"999999999999999999999999999999999999999999999999999999999999", 60},
		{"1000000000000000000000000000000000000000000000000000000000000", 61},
		{"9999999999999999999999999999999999999999999999999999999999999", 61},
		{"10000000000000000000000000000000000000000000000000000000000000", 62},
		{"99999999999999999999999999999999999999999999999999999999999999", 62},
		{"100000000000000000000000000000000000000000000000000000000000000", 63},
		{"999999999999999999999999999999999999999999999999999999999999999", 63},
		{"1000000000000000000000000000000000000000000000000000000000000000", 64},
		{"9999999999999999999999999999999999999999999999999999999999999999", 64},
		{"10000000000000000000000000000000000000000000000000000000000000000", 65},
		{"99999999999999999999999999999999999999999999999999999999999999999", 65},
		{"100000000000000000000000000000000000000000000000000000000000000000", 66},
		{"999999999999999999999999999999999999999999999999999999999999999999", 66},
		{"1000000000000000000000000000000000000000000000000000000000000000000", 67},
		{"9999999999999999999999999999999999999999999999999999999999999999999", 67},
		{"10000000000000000000000000000000000000000000000000000000000000000000", 68},
		{"99999999999999999999999999999999999999999999999999999999999999999999", 68},
		{"100000000000000000000000000000000000000000000000000000000000000000000", 69},
		{"999999999999999999999999999999999999999999999999999999999999999999999", 69},
		{"1000000000000000000000000000000000000000000000000000000000000000000000", 70},
		{"9999999999999999999999999999999999999999999999999999999999999999999999", 70},
		{"10000000000000000000000000000000000000000000000000000000000000000000000", 71},
		{"99999999999999999999999999999999999999999999999999999999999999999999999", 71},
		{"100000000000000000000000000000000000000000000000000000000000000000000000", 72},
		{"999999999999999999999999999999999999999999999999999999999999999999999999", 72},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000", 73},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999", 73},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000", 74},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999", 74},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000", 75},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999", 75},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000", 76},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999", 76},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000", 77},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999", 77},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000", 78},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999", 78},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000", 79},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999", 79},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000", 80},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999", 80},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000", 81},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999", 81},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000", 82},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999", 82},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000", 83},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999", 83},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000", 84},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 84},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 85},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 85},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 86},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 86},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 87},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 87},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 88},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 88},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 89},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 89},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 90},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 90},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 91},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 91},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 92},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 92},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 93},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 93},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 94},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 94},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 95},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 95},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 96},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 96},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 97},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 97},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 98},
		{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 98},
		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 99},
		{"999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 99},
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 100},
		{"9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 100},
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 101},
	}
	for _, tt := range cases {
		z := mustParseSint(tt.z)
		got := z.prec()
		if got != tt.want {
			t.Errorf("%q.prec() = %v, want %v", tt.z, got, tt.want)
		}
	}
}

func TestSint_hasPrec(t *testing.T) {
	cases := []struct {
		z    string
		prec int
		want bool
	}{
		{"0", -1, true},
		{"0", 0, true},
		{"0", 1, false},
		{"1", 0, true},
		{"1", 1, true},
		{"1", 2, false},
		{"10", 1, true},
		{"10", 2, true},
		{"10", 3, false},

		{"100000000000000000", 19, false},  // 18 digits
		{"1000000000000000000", 19, true},  // 19 digits
		{"10000000000000000000", 19, true}, // 20 digits
		{"1000000000000000000", 18, true},  // 19 digits
		{"1000000000000000000", 19, true},  // 19 digits
		{"1000000000000000000", 20, false}, // 19 digits

		{"100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 100, false},  // 99 digits
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 100, true},  // 100 digits
		{"10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 100, true}, // 101 digits
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 99, true},   // 100 digits
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 100, true},  // 100 digits
		{"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 101, false}, // 100 digits
	}
	for _, tt := range cases {
		z := mustParseSint(tt.z)
		got := z.hasPrec(tt.prec)
		if got != tt.want {
			t.Errorf("%v.hasPrec(%v) = %v, want %v", tt.z, tt.prec, got, tt.want)
		}
	}
}

// mustParseSint converts string to big.Int.
func mustParseSint(s string) *sint {
	z, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(fmt.Errorf("mustParseSint(%q) failed: parsing error", s))
	}
	if z.Sign() < 0 {
		panic(fmt.Errorf("mustParseSint(%q) failed: negative number", s))
	}
	return (*sint)(z)
}
