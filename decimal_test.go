package decimal

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"
)

func TestDecimal_ZeroValue(t *testing.T) {
	got := Decimal{}
	want := MustNew(0, 0)
	if got != want {
		t.Errorf("Decimal{} = %q, want %q", got, want)
	}
}

func TestDecimal_Size(t *testing.T) {
	d := Decimal{}
	got := unsafe.Sizeof(d)
	want := uintptr(16)
	if got != want {
		t.Errorf("unsafe.Sizeof(%q) = %v, want %v", d, got, want)
	}
}

func TestDecimal_Interfaces(t *testing.T) {
	var d any

	d = Decimal{}
	_, ok := d.(fmt.Stringer)
	if !ok {
		t.Errorf("%T does not implement fmt.Stringer", d)
	}
	_, ok = d.(fmt.Formatter)
	if !ok {
		t.Errorf("%T does not implement fmt.Formatter", d)
	}
	_, ok = d.(encoding.TextMarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.TextMarshaler", d)
	}
	_, ok = d.(encoding.BinaryMarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.BinaryMarshaler", d)
	}
	_, ok = d.(driver.Valuer)
	if !ok {
		t.Errorf("%T does not implement driver.Valuer", d)
	}

	d = &Decimal{}
	_, ok = d.(encoding.TextUnmarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.TextUnmarshaler", d)
	}
	_, ok = d.(encoding.BinaryUnmarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.BinaryUnmarshaler", d)
	}
	_, ok = d.(sql.Scanner)
	if !ok {
		t.Errorf("%T does not implement sql.Scanner", d)
	}
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			coef  int64
			scale int
			want  string
		}{
			{math.MinInt64, 0, "-9223372036854775808"},
			{math.MinInt64, 1, "-922337203685477580.8"},
			{math.MinInt64, 2, "-92233720368547758.08"},
			{math.MinInt64, 19, "-0.9223372036854775808"},
			{0, 0, "0"},
			{0, 1, "0.0"},
			{0, 2, "0.00"},
			{0, 3, "0.000"},
			{0, 19, "0.0000000000000000000"},
			{1, 0, "1"},
			{1, 1, "0.1"},
			{1, 2, "0.01"},
			{1, 19, "0.0000000000000000001"},
			{math.MaxInt64, 0, "9223372036854775807"},
			{math.MaxInt64, 1, "922337203685477580.7"},
			{math.MaxInt64, 2, "92233720368547758.07"},
			{math.MaxInt64, 19, "0.9223372036854775807"},
		}
		for _, tt := range tests {
			got, err := New(tt.coef, tt.scale)
			if err != nil {
				t.Errorf("New(%v, %v) failed: %v", tt.coef, tt.scale, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("New(%v, %v) = %q, want %q", tt.coef, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			coef  int64
			scale int
		}{
			"scale range 1": {math.MinInt64, -1},
			"scale range 2": {math.MaxInt64, -1},
			"scale range 3": {0, -1},
			"scale range 4": {1, -1},
			"scale range 5": {math.MinInt64, 20},
			"scale range 6": {math.MinInt64, 39},
			"scale range 7": {math.MaxInt64, 20},
			"scale range 8": {math.MaxInt64, 39},
		}
		for _, tt := range tests {
			_, err := New(tt.coef, tt.scale)
			if err == nil {
				t.Errorf("New(%v, %v) did not fail", tt.coef, tt.scale)
			}
		}
	})
}

func TestMustNew(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNew(0, -1) did not panic")
			}
		}()
		MustNew(0, -1)
	})
}

func TestNewFromInt64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			whole, frac int64
			scale       int
			want        string
		}{
			// Zeros
			{0, 0, 0, "0"},
			{0, 0, 19, "0"},

			// Negatives
			{-1, -1, 1, "-1.1"},
			{-1, -1, 2, "-1.01"},
			{-1, -1, 3, "-1.001"},
			{-1, -1, 18, "-1.000000000000000001"},

			// Positives
			{1, 1, 1, "1.1"},
			{1, 1, 2, "1.01"},
			{1, 1, 3, "1.001"},
			{1, 100000000, 9, "1.1"},
			{1, 1, 18, "1.000000000000000001"},
			{100000000000000000, 100000000000000000, 18, "100000000000000000.1"},
			{1, 1, 19, "1.000000000000000000"},
			{999999999999999999, 9, 1, "999999999999999999.9"},
			{999999999999999999, 99, 2, "1000000000000000000"},
			{math.MaxInt64, math.MaxInt32, 10, "9223372036854775807"},
			{math.MaxInt64, math.MaxInt64, 19, "9223372036854775808"},
		}
		for _, tt := range tests {
			got, err := NewFromInt64(tt.whole, tt.frac, tt.scale)
			if err != nil {
				t.Errorf("NewFromInt64(%v, %v, %v) failed: %v", tt.whole, tt.frac, tt.scale, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("NewFromInt64(%v, %v, %v) = %q, want %q", tt.whole, tt.frac, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			whole, frac int64
			scale       int
		}{
			"different signs 1": {-1, 1, 0},
			"fraction range 1":  {1, 1, 0},
			"scale range 1":     {1, 1, -1},
			"scale range 2":     {1, 0, -1},
			"scale range 3":     {1, 1, 20},
			"scale range 4":     {1, 0, 20},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := NewFromInt64(tt.whole, tt.frac, tt.scale)
				if err == nil {
					t.Errorf("NewFromInt64(%v, %v, %v) did not fail", tt.whole, tt.frac, tt.scale)
				}
			})
		}
	})
}

func TestNewFromFloat64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			f    float64
			want string
		}{
			// Zeros
			{-0, "0"},
			{0, "0"},
			{0.0, "0"},
			{0.00, "0"},
			{0.0000000000000000000, "0"},

			// Smallest non-zero
			{math.SmallestNonzeroFloat64, "0.0000000000000000000"},

			// Powers of 10
			{1e-20, "0.0000000000000000000"},
			{1e-19, "0.0000000000000000001"},
			{1e-5, "0.00001"},
			{1e-4, "0.0001"},
			{1e-3, "0.001"},
			{1e-2, "0.01"},
			{1e-1, "0.1"},
			{1e0, "1"},
			{1e1, "10"},
			{1e2, "100"},
			{1e3, "1000"},
			{1e4, "10000"},
			{1e5, "100000"},
			{1e18, "1000000000000000000"},
		}
		for _, tt := range tests {
			got, err := NewFromFloat64(tt.f)
			if err != nil {
				t.Errorf("NewFromFloat64(%v) failed: %v", tt.f, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("NewFromFloat64(%v) = %q, want %q", tt.f, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]float64{
			"overflow 1":      1e19,
			"overflow 2":      1e20,
			"overflow 3":      math.MaxFloat64,
			"overflow 4":      -1e19,
			"overflow 5":      -1e20,
			"overflow 6":      -math.MaxFloat64,
			"special value 1": math.NaN(),
			"special value 2": math.Inf(1),
			"special value 3": math.Inf(-1),
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := NewFromFloat64(tt)
				if err == nil {
					t.Errorf("NewFromFloat64(%v) did not fail", tt)
				}
			})
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			s         string
			wantNeg   bool
			wantCoef  uint64
			wantScale int
		}{
			{"-9999999999999999999.0", true, 9999999999999999999, 0},
			{"-9999999999999999999", true, 9999999999999999999, 0},
			{"-999999999999999999.9", true, 9999999999999999999, 1},
			{"-99999999999999999.99", true, 9999999999999999999, 2},
			{"-1000000000000000000.0", true, 1000000000000000000, 0},
			{"-0.9999999999999999999", true, 9999999999999999999, 19},
			{"-00000000000000000000000000000000000001", true, 1, 0},
			{"-1", true, 1, 0},
			{"-1.", true, 1, 0},
			{"-.1", true, 1, 1},
			{"-0.1", true, 1, 1},
			{"-0.01", true, 1, 2},
			{"-0.0000000000000000001", true, 1, 19},
			{"-00000000000000000000000000000000000000", false, 0, 0},
			{"+00000000000000000000000000000000000000", false, 0, 0},
			{"0", false, 0, 0},
			{"0.", false, 0, 0},
			{".0", false, 0, 1},
			{"0.0", false, 0, 1},
			{"0.00", false, 0, 2},
			{"0.000000000000000000", false, 0, 18},
			{"0.0000000000000000000", false, 0, 19},
			{"0.00000000000000000000", false, 0, 19},
			{"00000000000000000000000000000000000001", false, 1, 0},
			{"1", false, 1, 0},
			{"1.", false, 1, 0},
			{".1", false, 1, 1},
			{"0.1", false, 1, 1},
			{"0.01", false, 1, 2},
			{"0.0000000000000000001", false, 1, 19},
			{"1000000000000000000.0", false, 1000000000000000000, 0},
			{"9999999999999999999.0", false, 9999999999999999999, 0},
			{"9999999999999999999", false, 9999999999999999999, 0},
			{"999999999999999999.9", false, 9999999999999999999, 1},
			{"99999999999999999.99", false, 9999999999999999999, 2},
			{"0.9999999999999999999", false, 9999999999999999999, 19},

			// Rounding
			{"0.00000000000000000000000000000000000000", false, 0, 19},
			{"0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false, 0, 19},
			{"-0.00000000000000000000000000000000000001", false, 0, 19},
			{"0.00000000000000000000000000000000000001", false, 0, 19},
			{"-999999999999999999.99", true, 1000000000000000000, 0},
			{"0.123456789012345678901234567890", false, 1234567890123456789, 19},
			{"0.12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678", false, 1234567890123456789, 19},

			// Exponential notation
			{"0e9", false, 0, 0},
			{"0e-9", false, 0, 9},
			{"1.23e-12", false, 123, 14},
			{"1.23e-5", false, 123, 7},
			{"1.23e-4", false, 123, 6},
			{"1.23e-3", false, 123, 5},
			{"1.23e-2", false, 123, 4},
			{"1.23e-1", false, 123, 3},
			{"1.23e+0", false, 123, 2},
			{"1.23e+1", false, 123, 1},
			{"1.23e+2", false, 123, 0},
			{"1.23e+3", false, 1230, 0},
			{"1.23e+4", false, 12300, 0},
			{"1.23e+5", false, 123000, 0},
			{"1.23e+12", false, 1230000000000, 0},
			{"0.0e-38", false, 0, 19},
			{"0e-38", false, 0, 19},
			{"1e-2", false, 1, 2},
			{"1e-1", false, 1, 1},
			{"1e0", false, 1, 0},
			{"1e+1", false, 10, 0},
			{"1e+2", false, 100, 0},
			{"0.0000000000000000001e-19", false, 0, 19},
			{"0.0000000000000000001e19", false, 1, 0},
			{"1000000000000000000e-19", false, 1000000000000000000, 19},
			{"1000000000000000000e-38", false, 0, 19},
			{"10000000000000000000e-38", false, 1, 19},
			{"100000000000000000000e-38", false, 10, 19},
			{"10000000000000000000000000000000000000e-38", false, 1000000000000000000, 19},
			{"1e+18", false, 1000000000000000000, 0},
			{"0.0000000001e10", false, 1, 0},
			{"10000000000e-10", false, 10000000000, 10},
			{"4E9", false, 4000000000, 0},
			{"0.73e-7", false, 73, 9},
		}
		for _, tt := range tests {
			got, err := Parse(tt.s)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", tt.s, err)
				continue
			}
			if got.IsNeg() != tt.wantNeg {
				t.Errorf("Parse(%q).IsNeg() = %v, want %v", tt.s, got.IsNeg(), tt.wantNeg)
				continue
			}
			if got.Coef() != tt.wantCoef {
				t.Errorf("Parse(%q).Coef() = %v, want %v", tt.s, got.Coef(), tt.wantCoef)
				continue
			}
			if got.Scale() != tt.wantScale {
				t.Errorf("Parse(%q).Scale() = %v, want %v", tt.s, got.Scale(), tt.wantScale)
				continue
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			s     string
			scale int
		}{
			"missing digits 1":  {"", 0},
			"missing digits 2":  {"+", 0},
			"missing digits 3":  {"-", 0},
			"missing digits 4":  {".", 0},
			"missing digits 5":  {"..", 0},
			"missing digits 6":  {".e", 0},
			"missing digits 7":  {"e1", 0},
			"missing digits 8":  {"+e", 0},
			"missing digits 9":  {"-e", 0},
			"missing digits 10": {"e+", 0},
			"missing digits 11": {"e-", 0},
			"missing digits 12": {"e.0", 0},
			"missing digits 13": {"e+1", 0},
			"missing digits 14": {"e-1", 0},
			"invalid char 1":    {"a", 0},
			"invalid char 2":    {"1a", 0},
			"invalid char 3":    {"1.a", 0},
			"invalid char 4":    {" 1", 0},
			"invalid char 5":    {" +1", 0},
			"invalid char 6":    {" -1", 0},
			"invalid char 7":    {"1 ", 0},
			"invalid char 8":    {"+1 ", 0},
			"invalid char 9":    {"-1 ", 0},
			"invalid char 10":   {" 1 ", 0},
			"invalid char 11":   {" + 1", 0},
			"invalid char 12":   {" - 1", 0},
			"invalid char 13":   {"1,1", 0},
			"missing exp 1":     {"0.e", 0},
			"missing exp 2":     {"1e", 0},
			"missing exp 3":     {"1ee", 0},
			"exp range 1":       {"1e-331", 0},
			"exp range 2":       {"1e331", 0},
			"double sign 1":     {"++1", 0},
			"double sign 2":     {"--1", 0},
			"double sign 3":     {"+-1", 0},
			"double sign 4":     {"-+1", 0},
			"double sign 5":     {"-1.-1", 0},
			"double sign 6":     {"1.1-", 0},
			"double sign 7":     {"1e--1", 0},
			"double sign 8":     {"1e-+1", 0},
			"double sign 9":     {"1e+-1", 0},
			"double sign 10":    {"1e++1", 0},
			"double sign 11":    {"1e-1-", 0},
			"double sign 12":    {"-1-", 0},
			"double dot 1":      {"1.1.1", 0},
			"double dot 2":      {"..1", 0},
			"double dot 3":      {"1..1", 0},
			"double dot 4":      {".1.1", 0},
			"double dot 5":      {"1.1.", 0},
			"double dot 6":      {".1.", 0},
			"special value 1":   {"Inf", 0},
			"special value 2":   {"-infinity", 0},
			"special value 3":   {"NaN", 0},
			"overflow 1":        {"-10000000000000000000", 0},
			"overflow 2":        {"-99999999999999999990", 0},
			"overflow 3":        {"10000000000000000000", 0},
			"overflow 4":        {"99999999999999999990", 0},
			"overflow 5":        {"123456789012345678901234567890123456789", 0},
			"many digits":       {"0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 0},
			"scale 1":           {"0", MaxScale + 1},
			"scale 2":           {"10", MaxScale},
			"scale 3":           {"100", MaxScale - 1},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := ParseExact(tt.s, tt.scale)
				if err == nil {
					t.Errorf("ParseExact(%q, %v) did not fail", tt.s, tt.scale)
					return
				}
			})
		}
	})
}

func TestMustParse(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParse(\".\") did not panic")
			}
		}()
		MustParse(".")
	})
}

func TestDecimal_String(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			neg   bool
			coef  fint
			scale int
			want  string
		}{
			{true, maxCoef, 0, "-9999999999999999999"},
			{true, maxCoef, 1, "-999999999999999999.9"},
			{true, maxCoef, 2, "-99999999999999999.99"},
			{true, maxCoef, 3, "-9999999999999999.999"},
			{true, maxCoef, 19, "-0.9999999999999999999"},
			{true, 1, 0, "-1"},
			{true, 1, 1, "-0.1"},
			{true, 1, 2, "-0.01"},
			{true, 1, 19, "-0.0000000000000000001"},
			{false, 0, 0, "0"},
			{false, 0, 1, "0.0"},
			{false, 0, 2, "0.00"},
			{false, 0, 19, "0.0000000000000000000"},
			{false, 1, 0, "1"},
			{false, 1, 1, "0.1"},
			{false, 1, 2, "0.01"},
			{false, 1, 19, "0.0000000000000000001"},
			{false, maxCoef, 0, "9999999999999999999"},
			{false, maxCoef, 1, "999999999999999999.9"},
			{false, maxCoef, 2, "99999999999999999.99"},
			{false, maxCoef, 3, "9999999999999999.999"},
			{false, maxCoef, 19, "0.9999999999999999999"},

			// Exported constants
			{NegOne.neg, NegOne.coef, NegOne.Scale(), "-1"},
			{Zero.neg, Zero.coef, Zero.Scale(), "0"},
			{One.neg, One.coef, One.Scale(), "1"},
			{Two.neg, Two.coef, Two.Scale(), "2"},
			{Ten.neg, Ten.coef, Ten.Scale(), "10"},
			{Hundred.neg, Hundred.coef, Hundred.Scale(), "100"},
			{Thousand.neg, Thousand.coef, Thousand.Scale(), "1000"},
			{E.neg, E.coef, E.Scale(), "2.718281828459045235"},
			{Pi.neg, Pi.coef, Pi.Scale(), "3.141592653589793238"},
		}
		for _, tt := range tests {
			d, err := newSafe(tt.neg, tt.coef, tt.scale)
			if err != nil {
				t.Errorf("newDecimal(%v, %v, %v) failed: %v", tt.neg, tt.coef, tt.scale, err)
				continue
			}
			got := d.String()
			if got != tt.want {
				t.Errorf("newDecimal(%v, %v, %v).String() = %q, want %q", tt.neg, tt.coef, tt.scale, got, tt.want)
			}
		}
	})
}

func TestParseBCD(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			bcd  []byte
			want string
		}{
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x00}, "-9999999999999999999"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x01}, "-999999999999999999.9"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x02}, "-99999999999999999.99"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x03}, "-9999999999999999.999"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x19}, "-0.9999999999999999999"},
			{[]byte{0x1d, 0x00}, "-1"},
			{[]byte{0x1d, 0x01}, "-0.1"},
			{[]byte{0x1d, 0x02}, "-0.01"},
			{[]byte{0x1d, 0x19}, "-0.0000000000000000001"},
			{[]byte{0x0c, 0x00}, "0"},
			{[]byte{0x0c, 0x01}, "0.0"},
			{[]byte{0x0c, 0x02}, "0.00"},
			{[]byte{0x0c, 0x19}, "0.0000000000000000000"},
			{[]byte{0x1c, 0x00}, "1"},
			{[]byte{0x1c, 0x01}, "0.1"},
			{[]byte{0x1c, 0x02}, "0.01"},
			{[]byte{0x1c, 0x19}, "0.0000000000000000001"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x00}, "9999999999999999999"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x01}, "999999999999999999.9"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x02}, "99999999999999999.99"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x03}, "9999999999999999.999"},
			{[]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x19}, "0.9999999999999999999"},

			// Exported constants
			{[]byte{0x1d, 0x00}, NegOne.String()},
			{[]byte{0x0c, 0x00}, Zero.String()},
			{[]byte{0x1c, 0x00}, One.String()},
			{[]byte{0x2c, 0x00}, Two.String()},
			{[]byte{0x01, 0x0c, 0x00}, Ten.String()},
			{[]byte{0x10, 0x0c, 0x00}, Hundred.String()},
			{[]byte{0x01, 0x00, 0x0c, 0x00}, Thousand.String()},
			{[]byte{0x27, 0x18, 0x28, 0x18, 0x28, 0x45, 0x90, 0x45, 0x23, 0x5c, 0x18}, E.String()},
			{[]byte{0x31, 0x41, 0x59, 0x26, 0x53, 0x58, 0x97, 0x93, 0x23, 0x8c, 0x18}, Pi.String()},
		}
		for _, tt := range tests {
			got, err := parseBCD(tt.bcd)
			if err != nil {
				t.Errorf("parseBCD(% x) failed: %v", tt.bcd, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("parseBCD(% x) = %q, want %q", tt.bcd, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]byte{
			"empty":              {},
			"invalid nibble 1":   {0x0f},
			"invalid nibble 2":   {0xf0},
			"invalid nibble 3":   {0x0c, 0x0f},
			"invalid nibble 4":   {0x0c, 0xf0},
			"decimal overflow 1": {0x09, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x00},
			"decimal overflow 2": {0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x00},
			"no sign":            {0x00},
			"scale overflow":     {0x0c, 0x00, 0x00},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := parseBCD(tt)
				if err == nil {
					t.Errorf("parseBCD(% x) did not fail", tt)
				}
			})
		}
	})
}

func TestDecimal_BCD(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d    string
			want []byte
		}{
			{"-9999999999999999999", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x00}},
			{"-999999999999999999.9", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x01}},
			{"-99999999999999999.99", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x02}},
			{"-9999999999999999.999", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x03}},
			{"-0.9999999999999999999", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9d, 0x19}},
			{"-1", []byte{0x1d, 0x00}},
			{"-0.1", []byte{0x1d, 0x01}},
			{"-0.01", []byte{0x1d, 0x02}},
			{"-0.0000000000000000001", []byte{0x1d, 0x19}},
			{"0", []byte{0x0c, 0x00}},
			{"0.0", []byte{0x0c, 0x01}},
			{"0.00", []byte{0x0c, 0x02}},
			{"0.0000000000000000000", []byte{0x0c, 0x19}},
			{"1", []byte{0x1c, 0x00}},
			{"0.1", []byte{0x1c, 0x01}},
			{"0.01", []byte{0x1c, 0x02}},
			{"0.0000000000000000001", []byte{0x1c, 0x19}},
			{"9999999999999999999", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x00}},
			{"999999999999999999.9", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x01}},
			{"99999999999999999.99", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x02}},
			{"9999999999999999.999", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x03}},
			{"0.9999999999999999999", []byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9c, 0x19}},

			// Exported constants
			{NegOne.String(), []byte{0x1d, 0x00}},
			{Zero.String(), []byte{0x0c, 0x00}},
			{One.String(), []byte{0x1c, 0x00}},
			{Two.String(), []byte{0x2c, 0x00}},
			{Ten.String(), []byte{0x01, 0x0c, 0x00}},
			{Hundred.String(), []byte{0x10, 0x0c, 0x00}},
			{Thousand.String(), []byte{0x01, 0x00, 0x0c, 0x00}},
			{E.String(), []byte{0x27, 0x18, 0x28, 0x18, 0x28, 0x45, 0x90, 0x45, 0x23, 0x5c, 0x18}},
			{Pi.String(), []byte{0x31, 0x41, 0x59, 0x26, 0x53, 0x58, 0x97, 0x93, 0x23, 0x8c, 0x18}},
		}
		for _, tt := range tests {
			d, err := Parse(tt.d)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", tt.d, err)
				continue
			}
			got := d.bcd()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Parse(%q).bcd() = % x, want % x", tt.d, got, tt.want)
			}
		}
	})
}

func TestDecimal_Float64(t *testing.T) {
	tests := []struct {
		d         string
		wantFloat float64
		wantOk    bool
	}{
		{"9999999999999999999", 9999999999999999999, true},
		{"1000000000000000000", 1000000000000000000, true},
		{"1", 1, true},
		{"0.9999999999999999999", 0.9999999999999999999, true},
		{"0.0000000000000000001", 0.0000000000000000001, true},

		{"-9999999999999999999", -9999999999999999999, true},
		{"-1000000000000000000", -1000000000000000000, true},
		{"-1", -1, true},
		{"-0.9999999999999999999", -0.9999999999999999999, true},
		{"-0.0000000000000000001", -0.0000000000000000001, true},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		gotFloat, gotOk := d.Float64()
		if gotFloat != tt.wantFloat || gotOk != tt.wantOk {
			t.Errorf("%q.Float64() = [%v %v], want [%v %v]", d, gotFloat, gotOk, tt.wantFloat, tt.wantOk)
		}
	}
}

func TestDecimal_Int64(t *testing.T) {
	tests := []struct {
		d                   string
		scale               int
		wantWhole, wantFrac int64
		wantOk              bool
	}{
		// Zeros
		{"0", 0, 0, 0, true},
		{"0.0", 1, 0, 0, true},
		{"00.0", 1, 0, 0, true},
		{"0.00", 2, 0, 0, true},

		// Powers of 10
		{"1000", 0, 1000, 0, true},
		{"100", 0, 100, 0, true},
		{"10", 0, 10, 0, true},
		{"1", 0, 1, 0, true},
		{"0.1", 1, 0, 1, true},
		{"0.01", 2, 0, 1, true},
		{"0.001", 3, 0, 1, true},
		{"0.0001", 4, 0, 1, true},
		{"0.10", 2, 0, 10, true},
		{"0.100", 3, 0, 100, true},
		{"0.1000", 4, 0, 1000, true},

		// Signs
		{"0.1", 1, 0, 1, true},
		{"1.0", 1, 1, 0, true},
		{"1.1", 1, 1, 1, true},
		{"-0.1", 1, 0, -1, true},
		{"-1.0", 1, -1, 0, true},
		{"-1.1", 1, -1, -1, true},

		// Rounding
		{"5", 0, 5, 0, true},
		{"5", 1, 5, 0, true},
		{"5", 2, 5, 0, true},
		{"5", 3, 5, 0, true},
		{"0.5", 0, 0, 0, true},
		{"0.5", 1, 0, 5, true},
		{"0.5", 2, 0, 50, true},
		{"0.5", 3, 0, 500, true},
		{"0.05", 0, 0, 0, true},
		{"0.05", 1, 0, 0, true},
		{"0.05", 2, 0, 5, true},
		{"0.05", 3, 0, 50, true},
		{"0.005", 0, 0, 0, true},
		{"0.005", 1, 0, 0, true},
		{"0.005", 2, 0, 0, true},
		{"0.005", 3, 0, 5, true},
		{"0.51", 0, 1, 0, true},
		{"0.051", 1, 0, 1, true},
		{"0.0051", 2, 0, 1, true},
		{"0.00051", 3, 0, 1, true},
		{"0.9", 0, 1, 0, true},
		{"0.9", 1, 0, 9, true},
		{"0.9", 2, 0, 90, true},
		{"0.9", 3, 0, 900, true},
		{"0.9999999999999999999", 0, 1, 0, true},
		{"0.9999999999999999999", 1, 1, 0, true},
		{"0.9999999999999999999", 2, 1, 0, true},
		{"0.9999999999999999999", 3, 1, 0, true},

		// Edge cases
		{"9223372036854775807", 0, 9223372036854775807, 0, true},
		{"-9223372036854775808", 0, -9223372036854775808, 0, true},
		{"922337203685477580.8", 1, 922337203685477580, 8, true},
		{"-922337203685477580.9", 1, -922337203685477580, -9, true},
		{"9.223372036854775808", 18, 9, 223372036854775808, true},
		{"-9.223372036854775809", 18, -9, -223372036854775809, true},
		{"0.9223372036854775807", 19, 0, 9223372036854775807, true},
		{"-0.9223372036854775808", 19, 0, -9223372036854775808, true},

		// Failures
		{"9223372036854775808", 0, 0, 0, false},
		{"-9223372036854775809", 0, 0, 0, false},
		{"0.9223372036854775808", 19, 0, 0, false},
		{"-0.9223372036854775809", 19, 0, 0, false},
		{"9999999999999999999", 0, 0, 0, false},
		{"-9999999999999999999", 0, 0, 0, false},
		{"0.9999999999999999999", 19, 0, 0, false},
		{"-0.9999999999999999999", 19, 0, 0, false},
		{"0.1", -1, 0, 0, false},
		{"0.1", 20, 0, 0, false},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		gotWhole, gotFrac, gotOk := d.Int64(tt.scale)
		if gotWhole != tt.wantWhole || gotFrac != tt.wantFrac || gotOk != tt.wantOk {
			t.Errorf("%q.Int64(%v) = [%v %v %v], want [%v %v %v]", d, tt.scale, gotWhole, gotFrac, gotOk, tt.wantWhole, tt.wantFrac, tt.wantOk)
		}
	}
}

func TestDecimal_Scan(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		tests := []struct {
			f    float64
			want string
		}{
			{1e-20, "0.0000000000000000000"},
			{1e-19, "0.0000000000000000001"},
			{1e-5, "0.00001"},
			{1e-4, "0.0001"},
			{1e-3, "0.001"},
			{1e-2, "0.01"},
			{1e-1, "0.1"},
			{1e0, "1"},
			{1e1, "10"},
			{1e2, "100"},
			{1e3, "1000"},
			{1e4, "10000"},
			{1e5, "100000"},
			{1e18, "1000000000000000000"},
		}
		for _, tt := range tests {
			got := Decimal{}
			err := got.Scan(tt.f)
			if err != nil {
				t.Errorf("Scan(1.23456) failed: %v", err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Scan(%v) = %v, want %v", tt.f, got, want)
			}
		}
	})

	t.Run("int64", func(t *testing.T) {
		tests := []struct {
			i    int64
			want string
		}{
			{math.MinInt64, "-9223372036854775808"},
			{0, "0"},
			{math.MaxInt64, "9223372036854775807"},
		}
		for _, tt := range tests {
			got := Decimal{}
			err := got.Scan(tt.i)
			if err != nil {
				t.Errorf("Scan(%v) failed: %v", tt.i, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Scan(%v) = %v, want %v", tt.i, got, want)
			}
		}
	})

	t.Run("[]byte", func(t *testing.T) {
		tests := []struct {
			b    []byte
			want string
		}{
			{[]byte("-9223372036854775808"), "-9223372036854775808"},
			{[]byte("0"), "0"},
			{[]byte("9223372036854775807"), "9223372036854775807"},
		}
		for _, tt := range tests {
			got := Decimal{}
			err := got.Scan(tt.b)
			if err != nil {
				t.Errorf("Scan(%v) failed: %v", tt.b, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Scan(%v) = %v, want %v", tt.b, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []any{
			int8(123),
			int16(123),
			int32(123),
			int(123),
			uint8(123),
			uint16(123),
			uint32(123),
			uint(123),
			uint64(123),
			float32(123),
			nil,
		}
		for _, tt := range tests {
			got := Decimal{}
			err := got.Scan(tt)
			if err == nil {
				t.Errorf("Scan(%v) did not fail", tt)
			}
		}
	})
}

func TestDecimal_Format(t *testing.T) {
	tests := []struct {
		d, format, want string
	}{
		// %T verb
		{"12.34", "%T", "decimal.Decimal"},

		// %q verb
		{"12.34", "%q", "\"12.34\""},
		{"12.34", "%+q", "\"+12.34\""},
		{"12.34", "%.6q", "\"12.34\""}, // precision is ignored
		{"12.34", "%7q", "\"12.34\""},
		{"12.34", "%8q", " \"12.34\""},
		{"12.34", "%9q", "  \"12.34\""},
		{"12.34", "%10q", "   \"12.34\""},
		{"12.34", "%010q", "\"00012.34\""},
		{"12.34", "%+10q", "  \"+12.34\""},
		{"12.34", "%-10q", "\"12.34\"   "},

		// %s verb
		{"12.34", "%s", "12.34"},
		{"12.34", "%+s", "+12.34"},
		{"12.34", "%.6s", "12.34"}, // precision is ignored
		{"12.34", "%7s", "  12.34"},
		{"12.34", "%8s", "   12.34"},
		{"12.34", "%9s", "    12.34"},
		{"12.34", "%10s", "     12.34"},
		{"12.34", "%010s", "0000012.34"},
		{"12.34", "%+10s", "    +12.34"},
		{"12.34", "%-10s", "12.34     "},

		// %v verb
		{"12.34", "%v", "12.34"},
		{"12.34", "% v", " 12.34"},
		{"12.34", "%+v", "+12.34"},
		{"12.34", "%.6v", "12.34"}, // precision is ignored
		{"12.34", "%7v", "  12.34"},
		{"12.34", "%8v", "   12.34"},
		{"12.34", "%9v", "    12.34"},
		{"12.34", "%10v", "     12.34"},
		{"12.34", "%010v", "0000012.34"},
		{"12.34", "%+10v", "    +12.34"},
		{"12.34", "%-10v", "12.34     "},

		// %k verb
		{"12.34", "%k", "1234%"},
		{"12.34", "%+k", "+1234%"},
		{"12.34", "%.1k", "1234.0%"},
		{"12.34", "%.2k", "1234.00%"},
		{"12.34", "%.3k", "1234.000%"},
		{"12.34", "%.4k", "1234.0000%"},
		{"12.34", "%.5k", "1234.00000%"},
		{"12.34", "%.6k", "1234.000000%"},
		{"12.34", "%7k", "  1234%"},
		{"12.34", "%8k", "   1234%"},
		{"12.34", "%9k", "    1234%"},
		{"12.34", "%10k", "     1234%"},
		{"12.34", "%010k", "000001234%"},
		{"12.34", "%+10k", "    +1234%"},
		{"12.34", "%-10k", "1234%     "},
		{"2.3", "%k", "230%"},
		{"0.23", "%k", "23%"},
		{"0.023", "%k", "2.3%"},
		{"2.30", "%k", "230%"},
		{"0.230", "%k", "23.0%"},
		{"0.0230", "%k", "2.30%"},
		{"2.300", "%k", "230.0%"},
		{"0.2300", "%k", "23.00%"},
		{"0.02300", "%k", "2.300%"},

		// %f verb
		{"12.34", "%f", "12.34"},
		{"12.34", "%+f", "+12.34"},
		{"12.34", "%.1f", "12.3"},
		{"12.34", "%.2f", "12.34"},
		{"12.34", "%.3f", "12.340"},
		{"12.34", "%.4f", "12.3400"},
		{"12.34", "%.5f", "12.34000"},
		{"12.34", "%.6f", "12.340000"},
		{"12.34", "%7f", "  12.34"},
		{"12.34", "%8f", "   12.34"},
		{"12.34", "%9f", "    12.34"},
		{"12.34", "%10f", "     12.34"},
		{"12.34", "%010f", "0000012.34"},
		{"12.34", "%+10f", "    +12.34"},
		{"12.34", "%-10f", "12.34     "},
		{"12.34", "%.1f", "12.3"},
		{"0", "%.2f", "0.00"},
		{"0", "%5.2f", " 0.00"},
		{"9.996208266660", "%.2f", "10.00"},
		{"0.9996208266660", "%.2f", "1.00"},
		{"0.09996208266660", "%.2f", "0.10"},
		{"0.009996208266660", "%.2f", "0.01"},
		{"500.44", "%6.1f", " 500.4"},
		{"-404.040", "%-010.f", "-404      "},
		{"-404.040", "%-10.f", "-404      "},
		{"1", "%.20f", "1.00000000000000000000"},
		{"1.000000000000000000", "%.20f", "1.00000000000000000000"},
		{"9999999999999999999", "%.1f", "9999999999999999999.0"},
		{"9999999999999999999", "%.2f", "9999999999999999999.00"},
		{"9999999999999999999", "%.3f", "9999999999999999999.000"},

		// Wrong verbs
		{"12.34", "%b", "%!b(decimal.Decimal=12.34)"},
		{"12.34", "%e", "%!e(decimal.Decimal=12.34)"},
		{"12.34", "%E", "%!E(decimal.Decimal=12.34)"},
		{"12.34", "%g", "%!g(decimal.Decimal=12.34)"},
		{"12.34", "%G", "%!G(decimal.Decimal=12.34)"},
		{"12.34", "%x", "%!x(decimal.Decimal=12.34)"},
		{"12.34", "%X", "%!X(decimal.Decimal=12.34)"},

		// Errors
		{"9999999999999999999", "%k", "%!k(PANIC=Format method: formatting percent: computing [9999999999999999999 * 100]: the integer part of a decimal.Decimal can have at most 19 digits, but it has 21 digits: decimal overflow)"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := fmt.Sprintf(tt.format, d)
		if got != tt.want {
			t.Errorf("fmt.Sprintf(%q, %q) = %q, want %q", tt.format, tt.d, got, tt.want)
		}
	}
}

func TestDecimal_Prec(t *testing.T) {
	tests := []struct {
		d    string
		want int
	}{
		{"0000", 0},
		{"000", 0},
		{"00", 0},
		{"0", 0},
		{"0.000", 0},
		{"0.00", 0},
		{"0.0", 0},
		{"0", 0},
		{"0.0000000000000000001", 1},
		{"0.000000000000000001", 1},
		{"0.00000000000000001", 1},
		{"0.0000000000000001", 1},
		{"0.000000000000001", 1},
		{"0.00000000000001", 1},
		{"0.0000000000001", 1},
		{"0.000000000001", 1},
		{"0.00000000001", 1},
		{"0.0000000001", 1},
		{"0.000000001", 1},
		{"0.00000001", 1},
		{"0.0000001", 1},
		{"0.000001", 1},
		{"0.00001", 1},
		{"0.0001", 1},
		{"0.001", 1},
		{"0.01", 1},
		{"0.1", 1},
		{"1", 1},
		{"0.1000000000000000000", 19},
		{"0.100000000000000000", 18},
		{"0.10000000000000000", 17},
		{"0.1000000000000000", 16},
		{"0.100000000000000", 15},
		{"0.10000000000000", 14},
		{"0.1000000000000", 13},
		{"0.100000000000", 12},
		{"0.10000000000", 11},
		{"0.1000000000", 10},
		{"0.100000000", 9},
		{"0.10000000", 8},
		{"0.1000000", 7},
		{"0.100000", 6},
		{"0.10000", 5},
		{"0.1000", 4},
		{"0.100", 3},
		{"0.10", 2},
		{"0.1", 1},
		{"1", 1},
		{"10", 2},
		{"100", 3},
		{"1000", 4},
		{"10000", 5},
		{"100000", 6},
		{"1000000", 7},
		{"10000000", 8},
		{"100000000", 9},
		{"1000000000", 10},
		{"10000000000", 11},
		{"100000000000", 12},
		{"1000000000000", 13},
		{"10000000000000", 14},
		{"100000000000000", 15},
		{"1000000000000000", 16},
		{"10000000000000000", 17},
		{"100000000000000000", 18},
		{"1000000000000000000", 19},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Prec()
		if got != tt.want {
			t.Errorf("%q.Prec() = %v, want %v", tt.d, got, tt.want)
		}
	}
}

func TestDecimal_Rescale(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		// Zeros
		{"0", 0, "0"},
		{"0", 1, "0.0"},
		{"0", 2, "0.00"},
		{"0", 19, "0.0000000000000000000"},
		{"0.0", 1, "0.0"},
		{"0.00", 2, "0.00"},
		{"0.000000000", 19, "0.0000000000000000000"},
		{"0.000000000", 0, "0"},
		{"0.000000000", 1, "0.0"},
		{"0.000000000", 2, "0.00"},

		// Tests from GDA
		{"2.17", 0, "2"},
		{"2.17", 1, "2.2"},
		{"2.17", 2, "2.17"},
		{"2.17", 9, "2.170000000"},
		{"1.2345", 2, "1.23"},
		{"1.2355", 2, "1.24"},
		{"1.2345", 9, "1.234500000"},
		{"9.9999", 2, "10.00"},
		{"0.0001", 2, "0.00"},
		{"0.001", 2, "0.00"},
		{"0.009", 2, "0.01"},

		// Some extra tests
		{"0.03", 2, "0.03"},
		{"0.02", 2, "0.02"},
		{"0.01", 2, "0.01"},
		{"0.00", 2, "0.00"},
		{"-0.01", 2, "-0.01"},
		{"-0.02", 2, "-0.02"},
		{"-0.03", 2, "-0.03"},
		{"0.0049", 2, "0.00"},
		{"0.0051", 2, "0.01"},
		{"0.0149", 2, "0.01"},
		{"0.0151", 2, "0.02"},
		{"-0.0049", 2, "0.00"},
		{"-0.0051", 2, "-0.01"},
		{"-0.0149", 2, "-0.01"},
		{"-0.0151", 2, "-0.02"},
		{"0.0050", 2, "0.00"},
		{"0.0150", 2, "0.02"},
		{"0.0250", 2, "0.02"},
		{"0.0350", 2, "0.04"},
		{"-0.0050", 2, "0.00"},
		{"-0.0150", 2, "-0.02"},
		{"-0.0250", 2, "-0.02"},
		{"-0.0350", 2, "-0.04"},
		{"3.0448", 2, "3.04"},
		{"3.0450", 2, "3.04"},
		{"3.0452", 2, "3.05"},
		{"3.0956", 2, "3.10"},

		// Tests from Wikipedia
		{"1.8", 0, "2"},
		{"1.5", 0, "2"},
		{"1.2", 0, "1"},
		{"0.8", 0, "1"},
		{"0.5", 0, "0"},
		{"0.2", 0, "0"},
		{"-0.2", 0, "0"},
		{"-0.5", 0, "0"},
		{"-0.8", 0, "-1"},
		{"-1.2", 0, "-1"},
		{"-1.5", 0, "-2"},
		{"-1.8", 0, "-2"},

		// Negative scale
		{"1000000000000000000", -1, "1000000000000000000"},

		// Padding overflow
		{"1000000000000000000", 1, "1000000000000000000"},
		{"100000000000000000", 2, "100000000000000000.0"},
		{"10000000000000000", 3, "10000000000000000.00"},
		{"1000000000000000", 4, "1000000000000000.000"},
		{"100000000000000", 5, "100000000000000.0000"},
		{"10000000000000", 6, "10000000000000.00000"},
		{"1000000000000", 7, "1000000000000.000000"},
		{"1", 19, "1.000000000000000000"},
		{"0", 20, "0.0000000000000000000"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Rescale(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Rescale(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_Quantize(t *testing.T) {
	tests := []struct {
		d, e, want string
	}{
		{"0", "0", "0"},
		{"0", "0.0", "0.0"},
		{"0.0", "0", "0"},
		{"0.0", "0.0", "0.0"},

		{"0.0078", "0.00001", "0.00780"},
		{"0.0078", "0.0001", "0.0078"},
		{"0.0078", "0.001", "0.008"},
		{"0.0078", "0.01", "0.01"},
		{"0.0078", "0.1", "0.0"},
		{"0.0078", "1", "0"},

		{"-0.0078", "0.00001", "-0.00780"},
		{"-0.0078", "0.0001", "-0.0078"},
		{"-0.0078", "0.001", "-0.008"},
		{"-0.0078", "0.01", "-0.01"},
		{"-0.0078", "0.1", "0.0"},
		{"-0.0078", "1", "0"},

		{"0.6666666", "0.1", "0.7"},
		{"9.9999", "1.00", "10.00"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		e := MustParse(tt.e)
		got := d.Quantize(e)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Quantize(%q) = %q, want %q", d, e, got, want)
		}
	}
}

func TestDecimal_Pad(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		// Zeros
		{"0", 0, "0"},
		{"0", 1, "0.0"},
		{"0", 2, "0.00"},
		{"0", 19, "0.0000000000000000000"},
		{"0", 20, "0.0000000000000000000"},
		{"0.000000000", 0, "0.000000000"},
		{"0.000000000", 1, "0.000000000"},
		{"0.000000000", 2, "0.000000000"},
		{"0.000000000", 19, "0.0000000000000000000"},
		{"0.000000000", 20, "0.0000000000000000000"},

		// Tests from GDA
		{"2.17", 0, "2.17"},
		{"2.17", 1, "2.17"},
		{"2.17", 2, "2.17"},
		{"2.17", 9, "2.170000000"},
		{"1.2345", 2, "1.2345"},
		{"1.2355", 2, "1.2355"},
		{"1.2345", 9, "1.234500000"},
		{"9.9999", 2, "9.9999"},
		{"0.0001", 2, "0.0001"},
		{"0.001", 2, "0.001"},
		{"0.009", 2, "0.009"},

		// Negative scale
		{"1000000000000000000", -1, "1000000000000000000"},

		// Padding overflow
		{"1000000000000000000", 1, "1000000000000000000"},
		{"100000000000000000", 2, "100000000000000000.0"},
		{"10000000000000000", 3, "10000000000000000.00"},
		{"1000000000000000", 4, "1000000000000000.000"},
		{"100000000000000", 5, "100000000000000.0000"},
		{"10000000000000", 6, "10000000000000.00000"},
		{"1000000000000", 7, "1000000000000.000000"},
		{"-0.0000000000032", 63, "-0.0000000000032000000"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Pad(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Pad(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_Round(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		// Zeros
		{"0", -1, "0"},
		{"0", 0, "0"},
		{"0", 1, "0"},
		{"0", 2, "0"},
		{"0", 19, "0"},
		{"0.0", 1, "0.0"},
		{"0.00", 2, "0.00"},
		{"0.000000000", 19, "0.000000000"},
		{"0.000000000", 0, "0"},
		{"0.000000000", 1, "0.0"},
		{"0.000000000", 2, "0.00"},

		// Tests from GDA
		{"2.17", -1, "2"},
		{"2.17", 0, "2"},
		{"2.17", 1, "2.2"},
		{"2.17", 2, "2.17"},
		{"2.17", 9, "2.17"},
		{"1.2345", 2, "1.23"},
		{"1.2355", 2, "1.24"},
		{"1.2345", 9, "1.2345"},
		{"9.9999", 2, "10.00"},
		{"0.0001", 2, "0.00"},
		{"0.001", 2, "0.00"},
		{"0.009", 2, "0.01"},

		// Some extra tests
		{"0.03", 2, "0.03"},
		{"0.02", 2, "0.02"},
		{"0.01", 2, "0.01"},
		{"0.00", 2, "0.00"},
		{"-0.01", 2, "-0.01"},
		{"-0.02", 2, "-0.02"},
		{"-0.03", 2, "-0.03"},
		{"0.0049", 2, "0.00"},
		{"0.0050", 2, "0.00"},
		{"0.0051", 2, "0.01"},
		{"0.0149", 2, "0.01"},
		{"0.0150", 2, "0.02"},
		{"0.0151", 2, "0.02"},
		{"0.0250", 2, "0.02"},
		{"0.0350", 2, "0.04"},
		{"-0.0049", 2, "0.00"},
		{"-0.0051", 2, "-0.01"},
		{"-0.0050", 2, "0.00"},
		{"-0.0149", 2, "-0.01"},
		{"-0.0151", 2, "-0.02"},
		{"-0.0150", 2, "-0.02"},
		{"-0.0250", 2, "-0.02"},
		{"-0.0350", 2, "-0.04"},
		{"3.0448", 2, "3.04"},
		{"3.0450", 2, "3.04"},
		{"3.0452", 2, "3.05"},
		{"3.0956", 2, "3.10"},

		// Tests from Wikipedia
		{"1.8", 0, "2"},
		{"1.5", 0, "2"},
		{"1.2", 0, "1"},
		{"0.8", 0, "1"},
		{"0.5", 0, "0"},
		{"0.2", 0, "0"},
		{"-0.2", 0, "0"},
		{"-0.5", 0, "0"},
		{"-0.8", 0, "-1"},
		{"-1.2", 0, "-1"},
		{"-1.5", 0, "-2"},
		{"-1.8", 0, "-2"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Round(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Round(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_Trunc(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		// Zeros
		{"0", -1, "0"},
		{"0", 0, "0"},
		{"0", 1, "0"},
		{"0", 2, "0"},
		{"0", 19, "0"},
		{"0.0", 1, "0.0"},
		{"0.00", 2, "0.00"},
		{"0.000000000", 19, "0.000000000"},
		{"0.000000000", 0, "0"},
		{"0.000000000", 1, "0.0"},
		{"0.000000000", 2, "0.00"},

		// Tests from GDA
		{"2.17", 0, "2"},
		{"2.17", 1, "2.1"},
		{"2.17", 2, "2.17"},
		{"2.17", 9, "2.17"},
		{"1.2345", 2, "1.23"},
		{"1.2355", 2, "1.23"},
		{"1.2345", 9, "1.2345"},
		{"9.9999", 2, "9.99"},
		{"0.0001", 2, "0.00"},
		{"0.001", 2, "0.00"},
		{"0.009", 2, "0.00"},

		// Some extra tests
		{"0.03", 2, "0.03"},
		{"0.02", 2, "0.02"},
		{"0.01", 2, "0.01"},
		{"0.00", 2, "0.00"},
		{"-0.01", 2, "-0.01"},
		{"-0.02", 2, "-0.02"},
		{"-0.03", 2, "-0.03"},
		{"0.0049", 2, "0.00"},
		{"0.0051", 2, "0.00"},
		{"0.0149", 2, "0.01"},
		{"0.0151", 2, "0.01"},
		{"-0.0049", 2, "0.00"},
		{"-0.0051", 2, "-0.00"},
		{"-0.0149", 2, "-0.01"},
		{"-0.0151", 2, "-0.01"},
		{"0.0050", 2, "0.00"},
		{"0.0150", 2, "0.01"},
		{"0.0250", 2, "0.02"},
		{"0.0350", 2, "0.03"},
		{"-0.0050", 2, "0.00"},
		{"-0.0150", 2, "-0.01"},
		{"-0.0250", 2, "-0.02"},
		{"-0.0350", 2, "-0.03"},
		{"3.0448", 2, "3.04"},
		{"3.0450", 2, "3.04"},
		{"3.0452", 2, "3.04"},
		{"3.0956", 2, "3.09"},

		// Tests from Wikipedia
		{"1.8", 0, "1"},
		{"1.5", 0, "1"},
		{"1.2", 0, "1"},
		{"0.8", 0, "0"},
		{"0.5", 0, "0"},
		{"0.2", 0, "0"},
		{"-0.2", 0, "0"},
		{"-0.5", 0, "0"},
		{"-0.8", 0, "0"},
		{"-1.2", 0, "-1"},
		{"-1.5", 0, "-1"},
		{"-1.8", 0, "-1"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Trunc(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Trunc(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_Ceil(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		// Zeros
		{"0", -1, "0"},
		{"0", 0, "0"},
		{"0", 1, "0"},
		{"0", 2, "0"},
		{"0", 19, "0"},
		{"0.0", 1, "0.0"},
		{"0.00", 2, "0.00"},
		{"0.000000000", 19, "0.000000000"},
		{"0.000000000", 0, "0"},
		{"0.000000000", 1, "0.0"},
		{"0.000000000", 2, "0.00"},

		// Tests from GDA
		{"2.17", 0, "3"},
		{"2.17", 1, "2.2"},
		{"2.17", 2, "2.17"},
		{"2.17", 9, "2.17"},
		{"1.2345", 2, "1.24"},
		{"1.2355", 2, "1.24"},
		{"1.2345", 9, "1.2345"},
		{"9.9999", 2, "10.00"},
		{"0.0001", 2, "0.01"},
		{"0.001", 2, "0.01"},
		{"0.009", 2, "0.01"},
		{"-2.17", 0, "-2"},
		{"-2.17", 1, "-2.1"},
		{"-2.17", 2, "-2.17"},
		{"-2.17", 9, "-2.17"},
		{"-1.2345", 2, "-1.23"},
		{"-1.2355", 2, "-1.23"},
		{"-1.2345", 9, "-1.2345"},
		{"-9.9999", 2, "-9.99"},
		{"-0.0001", 2, "0.00"},
		{"-0.001", 2, "0.00"},
		{"-0.009", 2, "0.00"},

		// Some extra tests
		{"0.03", 2, "0.03"},
		{"0.02", 2, "0.02"},
		{"0.01", 2, "0.01"},
		{"0.00", 2, "0.00"},
		{"-0.01", 2, "-0.01"},
		{"-0.02", 2, "-0.02"},
		{"-0.03", 2, "-0.03"},
		{"0.0049", 2, "0.01"},
		{"0.0051", 2, "0.01"},
		{"0.0149", 2, "0.02"},
		{"0.0151", 2, "0.02"},
		{"-0.0049", 2, "0.00"},
		{"-0.0051", 2, "0.00"},
		{"-0.0149", 2, "-0.01"},
		{"-0.0151", 2, "-0.01"},
		{"0.0050", 2, "0.01"},
		{"0.0150", 2, "0.02"},
		{"0.0250", 2, "0.03"},
		{"0.0350", 2, "0.04"},
		{"-0.0050", 2, "0.00"},
		{"-0.0150", 2, "-0.01"},
		{"-0.0250", 2, "-0.02"},
		{"-0.0350", 2, "-0.03"},
		{"3.0448", 2, "3.05"},
		{"3.0450", 2, "3.05"},
		{"3.0452", 2, "3.05"},
		{"3.0956", 2, "3.10"},

		// Tests from Wikipedia
		{"1.8", 0, "2"},
		{"1.5", 0, "2"},
		{"1.2", 0, "2"},
		{"0.8", 0, "1"},
		{"0.5", 0, "1"},
		{"0.2", 0, "1"},
		{"-0.2", 0, "0"},
		{"-0.5", 0, "0"},
		{"-0.8", 0, "0"},
		{"-1.2", 0, "-1"},
		{"-1.5", 0, "-1"},
		{"-1.8", 0, "-1"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Ceil(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Ceil(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_Floor(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		// Zeros
		{"0", -1, "0"},
		{"0", 0, "0"},
		{"0", 1, "0"},
		{"0", 2, "0"},
		{"0", 19, "0"},
		{"0.0", 1, "0.0"},
		{"0.00", 2, "0.00"},
		{"0.000000000", 19, "0.000000000"},
		{"0.000000000", 0, "0"},
		{"0.000000000", 1, "0.0"},
		{"0.000000000", 2, "0.00"},

		// Tests from GDA
		{"2.17", 0, "2"},
		{"2.17", 1, "2.1"},
		{"2.17", 2, "2.17"},
		{"2.17", 9, "2.17"},
		{"1.2345", 2, "1.23"},
		{"1.2355", 2, "1.23"},
		{"1.2345", 9, "1.2345"},
		{"9.9999", 2, "9.99"},
		{"0.0001", 2, "0.00"},
		{"0.001", 2, "0.00"},
		{"0.009", 2, "0.00"},
		{"-2.17", 0, "-3"},
		{"-2.17", 1, "-2.2"},
		{"-2.17", 2, "-2.17"},
		{"-2.17", 9, "-2.17"},
		{"-1.2345", 2, "-1.24"},
		{"-1.2355", 2, "-1.24"},
		{"-1.2345", 9, "-1.2345"},
		{"-9.9999", 2, "-10.00"},
		{"-0.0001", 2, "-0.01"},
		{"-0.001", 2, "-0.01"},
		{"-0.009", 2, "-0.01"},

		// Some extra tests
		{"0.03", 2, "0.03"},
		{"0.02", 2, "0.02"},
		{"0.01", 2, "0.01"},
		{"0.00", 2, "0.00"},
		{"-0.01", 2, "-0.01"},
		{"-0.02", 2, "-0.02"},
		{"-0.03", 2, "-0.03"},
		{"0.0049", 2, "0.00"},
		{"0.0051", 2, "0.00"},
		{"0.0149", 2, "0.01"},
		{"0.0151", 2, "0.01"},
		{"-0.0049", 2, "-0.01"},
		{"-0.0051", 2, "-0.01"},
		{"-0.0149", 2, "-0.02"},
		{"-0.0151", 2, "-0.02"},
		{"0.0050", 2, "0.00"},
		{"0.0150", 2, "0.01"},
		{"0.0250", 2, "0.02"},
		{"0.0350", 2, "0.03"},
		{"-0.0050", 2, "-0.01"},
		{"-0.0150", 2, "-0.02"},
		{"-0.0250", 2, "-0.03"},
		{"-0.0350", 2, "-0.04"},
		{"3.0448", 2, "3.04"},
		{"3.0450", 2, "3.04"},
		{"3.0452", 2, "3.04"},
		{"3.0956", 2, "3.09"},

		// Tests from Wikipedia
		{"1.8", 0, "1"},
		{"1.5", 0, "1"},
		{"1.2", 0, "1"},
		{"0.8", 0, "0"},
		{"0.5", 0, "0"},
		{"0.2", 0, "0"},
		{"-0.2", 0, "-1"},
		{"-0.5", 0, "-1"},
		{"-0.8", 0, "-1"},
		{"-1.2", 0, "-2"},
		{"-1.5", 0, "-2"},
		{"-1.8", 0, "-2"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Floor(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Floor(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_MinScale(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d    string
			want int
		}{
			{"0", 0},
			{"0.0", 0},
			{"1", 0},
			{"1.000000000", 0},
			{"0.100000000", 1},
			{"0.010000000", 2},
			{"0.001000000", 3},
			{"0.000100000", 4},
			{"0.000010000", 5},
			{"0.000001000", 6},
			{"0.000000100", 7},
			{"0.000000010", 8},
			{"0.000000001", 9},
			{"0.000000000", 0},
			{"0.0000000000000000000", 0},
			{"0.1000000000000000000", 1},
			{"0.0000000000000000001", 19},
			{"0.9999999999999999999", 19},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got := d.MinScale()
			if got != tt.want {
				t.Errorf("%q.MinScale() = %v, want %v", d, got, tt.want)
			}
		}
	})
}

func TestDecimal_Trim(t *testing.T) {
	tests := []struct {
		d     string
		scale int
		want  string
	}{
		{"0.000000", 0, "0"},
		{"0.000000", 2, "0.00"},
		{"0.000000", 4, "0.0000"},
		{"0.000000", 6, "0.000000"},
		{"0.000000", 8, "0.000000"},
		{"-10.00", 0, "-10"},
		{"10.00", 0, "10"},
		{"0.000001", 0, "0.000001"},
		{"0.0000010", 0, "0.000001"},
		{"-0.000001", 0, "-0.000001"},
		{"-0.0000010", 0, "-0.000001"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Trim(tt.scale)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Trim(%v) = %q, want %q", d, tt.scale, got, want)
		}
	}
}

func TestDecimal_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, want string
		}{
			{"1", "1", "2"},
			{"2", "3", "5"},
			{"5.75", "3.3", "9.05"},
			{"5", "-3", "2"},
			{"-5", "-3", "-8"},
			{"-7", "2.5", "-4.5"},
			{"0.7", "0.3", "1.0"},
			{"1.25", "1.25", "2.50"},
			{"1.1", "0.11", "1.21"},
			{"1.234567890", "1.000000000", "2.234567890"},
			{"1.234567890", "1.000000110", "2.234568000"},

			{"0.9998", "0.0000", "0.9998"},
			{"0.9998", "0.0001", "0.9999"},
			{"0.9998", "0.0002", "1.0000"},
			{"0.9998", "0.0003", "1.0001"},

			{"999999999999999999", "1", "1000000000000000000"},
			{"99999999999999999", "1", "100000000000000000"},
			{"9999999999999999", "1", "10000000000000000"},
			{"999999999999999", "1", "1000000000000000"},
			{"99999999999999", "1", "100000000000000"},
			{"9999999999999", "1", "10000000000000"},
			{"999999999999", "1", "1000000000000"},
			{"99999999999", "1", "100000000000"},
			{"9999999999", "1", "10000000000"},
			{"999999999", "1", "1000000000"},
			{"99999999", "1", "100000000"},
			{"9999999", "1", "10000000"},
			{"999999", "1", "1000000"},
			{"99999", "1", "100000"},
			{"9999", "1", "10000"},
			{"999", "1", "1000"},
			{"99", "1", "100"},
			{"9", "1", "10"},

			{"100000000000", "0.00000000", "100000000000.0000000"},
			{"100000000000", "0.00000001", "100000000000.0000000"},

			{"0.0", "0", "0.0"},
			{"0.00", "0", "0.00"},
			{"0.000", "0", "0.000"},
			{"0.0000000", "0", "0.0000000"},
			{"0", "0.0", "0.0"},
			{"0", "0.00", "0.00"},
			{"0", "0.000", "0.000"},
			{"0", "0.0000000", "0.0000000"},

			{"9999999999999999999", "0.4", "9999999999999999999"},
			{"-9999999999999999999", "-0.4", "-9999999999999999999"},
			{"1", "-9999999999999999999", "-9999999999999999998"},
			{"9999999999999999999", "-1", "9999999999999999998"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			got, err := d.Add(e)
			if err != nil {
				t.Errorf("%q.Add(%q) failed: %v", d, e, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Add(%q) = %q, want %q", d, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e  string
			scale int
		}{
			"overflow 1": {"9999999999999999999", "1", 0},
			"overflow 2": {"9999999999999999999", "0.6", 0},
			"overflow 3": {"-9999999999999999999", "-1", 0},
			"overflow 4": {"-9999999999999999999", "-0.6", 0},
			"scale 1":    {"1", "1", MaxScale},
			"scale 2":    {"0", "0", MaxScale + 1},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			_, err := d.AddExact(e, tt.scale)
			if err == nil {
				t.Errorf("%q.AddExact(%q, %v) did not fail", d, e, tt.scale)
			}
		}
	})
}

func TestDecimal_Sub(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, want string
		}{
			// Signs
			{"5", "3", "2"},
			{"3", "5", "-2"},
			{"-5", "-3", "-2"},
			{"-3", "-5", "2"},
			{"-5", "3", "-8"},
			{"-3", "5", "-8"},
			{"5", "-3", "8"},
			{"3", "-5", "8"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			got, err := d.Sub(e)
			if err != nil {
				t.Errorf("%q.Sub(%q) failed: %v", d, e, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Sub(%q) = %q, want %q", d, e, got, want)
			}
		}
	})
}

func TestDecimal_SubAbs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, want string
		}{
			// Signs
			{"5", "3", "2"},
			{"3", "5", "2"},
			{"-5", "-3", "2"},
			{"-3", "-5", "2"},
			{"-5", "3", "8"},
			{"-3", "5", "8"},
			{"5", "-3", "8"},
			{"3", "-5", "8"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			got, err := d.SubAbs(e)
			if err != nil {
				t.Errorf("%q.SubAbs(%q) failed: %v", d, e, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.SubAbs(%q) = %q, want %q", d, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e string
		}{
			"overflow 1": {"1", "-9999999999999999999"},
			"overflow 2": {"9999999999999999999", "-1"},
			"overflow 3": {"9999999999999999999", "-9999999999999999999"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			_, err := d.SubAbs(e)
			if err == nil {
				t.Errorf("%q.SubAbs(%q) did not fail", d, e)
			}
		}
	})
}

func TestDecimal_Mul(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, want string
		}{
			{"2", "2", "4"},
			{"2", "3", "6"},
			{"5", "1", "5"},
			{"5", "2", "10"},
			{"1.20", "2", "2.40"},
			{"1.20", "0", "0.00"},
			{"1.20", "-2", "-2.40"},
			{"-1.20", "2", "-2.40"},
			{"-1.20", "0", "0.00"},
			{"-1.20", "-2", "2.40"},
			{"5.09", "7.1", "36.139"},
			{"2.5", "4", "10.0"},
			{"2.50", "4", "10.00"},
			{"0.70", "1.05", "0.7350"},
			{"1.000000000", "1", "1.000000000"},
			{"1.23456789", "1.00000000", "1.2345678900000000"},
			{"1.000000000000000000", "1.000000000000000000", "1.000000000000000000"},
			{"1.000000000000000001", "1.000000000000000001", "1.000000000000000002"},
			{"9.999999999999999999", "9.999999999999999999", "99.99999999999999998"},
			{"0.0000000000000000001", "0.0000000000000000001", "0.0000000000000000000"},
			{"0.0000000000000000001", "0.9999999999999999999", "0.0000000000000000001"},
			{"0.0000000000000000003", "0.9999999999999999999", "0.0000000000000000003"},
			{"0.9999999999999999999", "0.9999999999999999999", "0.9999999999999999998"},
			{"6963.788300835654596", "0.001436", "10.00000000000000000"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			got, err := d.Mul(e)
			if err != nil {
				t.Errorf("%q.Mul(%q) failed: %v", d, e, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Mul(%q) = %q, want %q", d, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e  string
			scale int
		}{
			"overflow 1": {"10000000000", "1000000000", 0},
			"overflow 2": {"1000000000000000000", "10", 0},
			"overflow 3": {"4999999999999999995", "-2.000000000000000002", 0},
			"scale 1":    {"1", "1", MaxScale},
			"scale 2":    {"0", "0", MaxScale + 1},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			_, err := d.MulExact(e, tt.scale)
			if err == nil {
				t.Errorf("%q.MulExact(%q, %v) did not fail", d, e, tt.scale)
			}
		}
	})
}

func TestDecimal_FMA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, f, want string
		}{
			// Signs
			{"2", "3", "4", "10"},
			{"2", "3", "-4", "2"},
			{"2", "-3", "4", "-2"},
			{"2", "-3", "-4", "-10"},
			{"-2", "3", "4", "-2"},
			{"-2", "3", "-4", "-10"},
			{"-2", "-3", "4", "10"},
			{"-2", "-3", "-4", "2"},

			// Addition tests
			{"1", "1", "1", "2"},
			{"1", "2", "3", "5"},
			{"1", "5.75", "3.3", "9.05"},
			{"1", "5", "-3", "2"},
			{"1", "-5", "-3", "-8"},
			{"1", "-7", "2.5", "-4.5"},
			{"1", "0.7", "0.3", "1.0"},
			{"1", "1.25", "1.25", "2.50"},
			{"1", "1.1", "0.11", "1.21"},
			{"1", "1.234567890", "1.000000000", "2.234567890"},
			{"1", "1.234567890", "1.000000110", "2.234568000"},
			{"1", "0.9998", "0.0000", "0.9998"},
			{"1", "0.9998", "0.0001", "0.9999"},
			{"1", "0.9998", "0.0002", "1.0000"},
			{"1", "0.9998", "0.0003", "1.0001"},
			{"1", "999999999999999999", "1", "1000000000000000000"},
			{"1", "99999999999999999", "1", "100000000000000000"},
			{"1", "9999999999999999", "1", "10000000000000000"},
			{"1", "999999999999999", "1", "1000000000000000"},
			{"1", "99999999999999", "1", "100000000000000"},
			{"1", "9999999999999", "1", "10000000000000"},
			{"1", "999999999999", "1", "1000000000000"},
			{"1", "99999999999", "1", "100000000000"},
			{"1", "9999999999", "1", "10000000000"},
			{"1", "999999999", "1", "1000000000"},
			{"1", "99999999", "1", "100000000"},
			{"1", "9999999", "1", "10000000"},
			{"1", "999999", "1", "1000000"},
			{"1", "99999", "1", "100000"},
			{"1", "9999", "1", "10000"},
			{"1", "999", "1", "1000"},
			{"1", "99", "1", "100"},
			{"1", "9", "1", "10"},
			{"1", "100000000000", "0.00000000", "100000000000.0000000"},
			{"1", "100000000000", "0.00000001", "100000000000.0000000"},
			{"1", "0.0", "0", "0.0"},
			{"1", "0.00", "0", "0.00"},
			{"1", "0.000", "0", "0.000"},
			{"1", "0.0000000", "0", "0.0000000"},
			{"1", "0", "0.0", "0.0"},
			{"1", "0", "0.00", "0.00"},
			{"1", "0", "0.000", "0.000"},
			{"1", "0", "0.0000000", "0.0000000"},
			{"1", "9999999999999999999", "0.4", "9999999999999999999"},
			{"1", "-9999999999999999999", "-0.4", "-9999999999999999999"},
			{"1", "1", "-9999999999999999999", "-9999999999999999998"},
			{"1", "9999999999999999999", "-1", "9999999999999999998"},

			// Multiplication tests
			{"2", "2", "0", "4"},
			{"2", "3", "0", "6"},
			{"5", "1", "0", "5"},
			{"5", "2", "0", "10"},
			{"1.20", "2", "0", "2.40"},
			{"1.20", "0", "0", "0.00"},
			{"1.20", "-2", "0", "-2.40"},
			{"-1.20", "2", "0", "-2.40"},
			{"-1.20", "0", "0", "0.00"},
			{"-1.20", "-2", "0", "2.40"},
			{"5.09", "7.1", "0", "36.139"},
			{"2.5", "4", "0", "10.0"},
			{"2.50", "4", "0", "10.00"},
			{"0.70", "1.05", "0", "0.7350"},
			{"1.000000000", "1", "0", "1.000000000"},
			{"1.23456789", "1.00000000", "0", "1.2345678900000000"},
			{"1.000000000000000000", "1.000000000000000000", "0", "1.000000000000000000"},
			{"1.000000000000000001", "1.000000000000000001", "0", "1.000000000000000002"},
			{"9.999999999999999999", "9.999999999999999999", "0", "99.99999999999999998"},
			{"0.0000000000000000001", "0.0000000000000000001", "0", "0.0000000000000000000"},
			{"0.0000000000000000001", "0.9999999999999999999", "0", "0.0000000000000000001"},
			{"0.0000000000000000003", "0.9999999999999999999", "0", "0.0000000000000000003"},
			{"0.9999999999999999999", "0.9999999999999999999", "0", "0.9999999999999999998"},
			{"6963.788300835654596", "0.001436", "0", "10.00000000000000000"},

			// Tests from GDA
			{"27583489.6645", "2582471078.04", "2593183.42371", "71233564292579696.34"},
			{"24280.355566", "939577.397653", "2032.013252", "22813275328.80506589"},
			{"7848976432", "-2586831.2281", "137903.517909", "-20303977342780612.62"},
			{"56890.388731", "35872030.4255", "339337.123410", "2040774094814.077745"},
			{"7533543.57445", "360317763928", "5073392.31638", "2714469575205049785"},
			{"437484.00601", "598906432790", "894450638.442", "262011986336578659.5"},
			{"203258304486", "-8628278.8066", "153127.446727", "-1753769320861850379"},
			{"42560533.1774", "-3643605282.86", "178277.96377", "-155073783526334663.6"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			f := MustParse(tt.f)
			got, err := d.FMA(e, f)
			if err != nil {
				t.Errorf("%q.FMA(%q, %q) failed: %v", d, e, f, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.FMA(%q, %q) = %q, want %q", d, e, f, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e, f string
			scale   int
		}{
			"overflow 1": {"1", "9999999999999999999", "1", 0},
			"overflow 2": {"1", "9999999999999999999", "0.6", 0},
			"overflow 3": {"1", "-9999999999999999999", "-1", 0},
			"overflow 4": {"1", "-9999999999999999999", "-0.6", 0},
			"overflow 5": {"10000000000", "1000000000", "0", 0},
			"overflow 6": {"1000000000000000000", "10", "0", 0},
			"scale 1":    {"1", "1", "1", MaxScale},
			"scale 2":    {"0", "0", "0", MaxScale + 1},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			f := MustParse(tt.f)
			_, err := d.FMAExact(e, f, tt.scale)
			if err == nil {
				t.Errorf("%q.FMAExact(%q, %q, %v) did not fail", d, e, f, tt.scale)
			}
		}
	})
}

func TestDecimal_Pow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d     string
			power int
			want  string
		}{
			// Zeros
			{"0", 0, "1"},
			{"0", 1, "0"},
			{"0", 2, "0"},

			// Ones
			{"-1", -2, "1"},
			{"-1", -1, "-1"},
			{"-1", 0, "1"},
			{"-1", 1, "-1"},
			{"-1", 2, "1"},

			// One tenths
			{"0.1", -18, "1000000000000000000"},
			{"0.1", -10, "10000000000"},
			{"0.1", -9, "1000000000"},
			{"0.1", -8, "100000000"},
			{"0.1", -7, "10000000"},
			{"0.1", -6, "1000000"},
			{"0.1", -5, "100000"},
			{"0.1", -4, "10000"},
			{"0.1", -3, "1000"},
			{"0.1", -2, "100"},
			{"0.1", -1, "10"},
			{"0.1", 0, "1"},
			{"0.1", 1, "0.1"},
			{"0.1", 2, "0.01"},
			{"0.1", 3, "0.001"},
			{"0.1", 4, "0.0001"},
			{"0.1", 5, "0.00001"},
			{"0.1", 6, "0.000001"},
			{"0.1", 7, "0.0000001"},
			{"0.1", 8, "0.00000001"},
			{"0.1", 9, "0.000000001"},
			{"0.1", 10, "0.0000000001"},
			{"0.1", 18, "0.000000000000000001"},
			{"0.1", 19, "0.0000000000000000001"},
			{"0.1", 20, "0.0000000000000000000"},
			{"0.1", 40, "0.0000000000000000000"},

			// Negative one tenths
			{"-0.1", -18, "1000000000000000000"},
			{"-0.1", -10, "10000000000"},
			{"-0.1", -9, "-1000000000"},
			{"-0.1", -8, "100000000"},
			{"-0.1", -7, "-10000000"},
			{"-0.1", -6, "1000000"},
			{"-0.1", -5, "-100000"},
			{"-0.1", -4, "10000"},
			{"-0.1", -3, "-1000"},
			{"-0.1", -2, "100"},
			{"-0.1", -1, "-10"},
			{"-0.1", 0, "1"},
			{"-0.1", 1, "-0.1"},
			{"-0.1", 2, "0.01"},
			{"-0.1", 3, "-0.001"},
			{"-0.1", 4, "0.0001"},
			{"-0.1", 5, "-0.00001"},
			{"-0.1", 6, "0.000001"},
			{"-0.1", 7, "-0.0000001"},
			{"-0.1", 8, "0.00000001"},
			{"-0.1", 9, "-0.000000001"},
			{"-0.1", 10, "0.0000000001"},
			{"-0.1", 18, "0.000000000000000001"},
			{"-0.1", 19, "-0.0000000000000000001"},
			{"-0.1", 20, "0.0000000000000000000"},
			{"-0.1", 40, "0.0000000000000000000"},

			// Twos
			{"2", -64, "0.0000000000000000001"},
			{"2", -63, "0.0000000000000000001"},
			{"2", -32, "0.0000000002328306437"},
			{"2", -16, "0.0000152587890625"},
			{"2", -9, "0.001953125"},
			{"2", -8, "0.00390625"},
			{"2", -7, "0.0078125"},
			{"2", -6, "0.015625"},
			{"2", -5, "0.03125"},
			{"2", -4, "0.0625"},
			{"2", -3, "0.125"},
			{"2", -2, "0.25"},
			{"2", -1, "0.5"},
			{"2", 0, "1"},
			{"2", 1, "2"},
			{"2", 2, "4"},
			{"2", 3, "8"},
			{"2", 4, "16"},
			{"2", 5, "32"},
			{"2", 6, "64"},
			{"2", 7, "128"},
			{"2", 8, "256"},
			{"2", 9, "512"},
			{"2", 16, "65536"},
			{"2", 32, "4294967296"},
			{"2", 63, "9223372036854775808"},

			// Negative twos
			{"-2", -64, "0.0000000000000000001"},
			{"-2", -63, "-0.0000000000000000001"},
			{"-2", -32, "0.0000000002328306437"},
			{"-2", -16, "0.0000152587890625"},
			{"-2", -9, "-0.001953125"},
			{"-2", -8, "0.00390625"},
			{"-2", -7, "-0.0078125"},
			{"-2", -6, "0.015625"},
			{"-2", -5, "-0.03125"},
			{"-2", -4, "0.0625"},
			{"-2", -3, "-0.125"},
			{"-2", -2, "0.25"},
			{"-2", -1, "-0.5"},
			{"-2", 0, "1"},
			{"-2", 1, "-2"},
			{"-2", 2, "4"},
			{"-2", 3, "-8"},
			{"-2", 4, "16"},
			{"-2", 5, "-32"},
			{"-2", 6, "64"},
			{"-2", 7, "-128"},
			{"-2", 8, "256"},
			{"-2", 9, "-512"},
			{"-2", 16, "65536"},
			{"-2", 32, "4294967296"},
			{"-2", 63, "-9223372036854775808"},

			// Squares
			{"-3", 2, "9"},
			{"-2", 2, "4"},
			{"-1", 2, "1"},
			{"0", 2, "0"},
			{"1", 2, "1"},
			{"2", 2, "4"},
			{"3", 2, "9"},
			{"4", 2, "16"},
			{"5", 2, "25"},
			{"6", 2, "36"},
			{"7", 2, "49"},
			{"8", 2, "64"},
			{"9", 2, "81"},
			{"10", 2, "100"},
			{"11", 2, "121"},
			{"12", 2, "144"},
			{"13", 2, "169"},
			{"14", 2, "196"},

			// Cubes
			{"-3", 3, "-27"},
			{"-2", 3, "-8"},
			{"-1", 3, "-1"},
			{"0", 3, "0"},
			{"1", 3, "1"},
			{"2", 3, "8"},
			{"3", 3, "27"},
			{"4", 3, "64"},
			{"5", 3, "125"},
			{"6", 3, "216"},
			{"7", 3, "343"},
			{"8", 3, "512"},
			{"9", 3, "729"},
			{"10", 3, "1000"},
			{"11", 3, "1331"},
			{"12", 3, "1728"},
			{"13", 3, "2197"},
			{"14", 3, "2744"},

			// Interest accrual
			{"1.1", 60, "304.4816395414180996"},         // no error
			{"1.01", 600, "391.5833969993197743"},       // no error
			{"1.001", 6000, "402.2211245663552923"},     // no error
			{"1.0001", 60000, "403.3077910727185433"},   // no error
			{"1.00001", 600000, "403.4166908911542153"}, // no error

			// Captured during fuzzing
			{"0.85", -267, "7000786514887173013"},
			{"-0.9223372036854775808", -128, "31197.15320234751783"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Pow(tt.power)
			if err != nil {
				t.Errorf("%q.Pow(%d) failed: %v", d, tt.power, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Pow(%d) = %q, want %q", d, tt.power, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d            string
			power, scale int
		}{
			"overflow 1": {"2", 64, 0},
			"overflow 2": {"0.5", -64, 0},
			"overflow 3": {"10", 19, 0},
			"overflow 4": {"0.1", -19, 0},
			"overflow 5": {"0.0000000000000000001", -3, 0},
			"overflow 6": {"0.0000000000000000001", -3, 1},
			"zero 1":     {"0", -1, 0},
			"scale 1":    {"1", 1, MaxScale},
			"scale 2":    {"1", 1, -1},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(tt.d)
				_, err := d.PowExact(tt.power, tt.scale)
				if err == nil {
					t.Errorf("%q.PowExact(%d, %d) did not fail", d, tt.power, tt.scale)
				}
			})
		}
	})
}

func TestDecimal_Sqrt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Zeros
			{"0", "0"},
			{"0.0", "0"},
			{"0.00", "0.0"},
			{"0.000", "0.0"},
			{"0.0000", "0.00"},

			// Numbers
			{"0", "0"},
			{"1", "1"},
			{"2", "1.414213562373095049"},
			{"3", "1.732050807568877294"},
			{"4", "2"},
			{"5", "2.236067977499789696"},
			{"6", "2.449489742783178098"},
			{"7", "2.645751311064590591"},
			{"8", "2.828427124746190098"},
			{"9", "3"},
			{"10", "3.162277660168379332"},
			{"11", "3.316624790355399849"},
			{"12", "3.464101615137754587"},
			{"13", "3.605551275463989293"},
			{"14", "3.741657386773941386"},
			{"15", "3.872983346207416885"},
			{"16", "4"},
			{"17", "4.12310562561766055"},
			{"18", "4.242640687119285146"},
			{"19", "4.358898943540673552"},
			{"20", "4.472135954999579393"},
			{"21", "4.582575694955840007"},
			{"22", "4.690415759823429555"},
			{"23", "4.795831523312719542"},
			{"24", "4.898979485566356196"},
			{"25", "5"},

			// Edge cases
			{"0.0000000000000000001", "0.000000000316227766"},
			{"9999999999999999999", "3162277660.168379332"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Sqrt()
			if err != nil {
				t.Errorf("%q.Sqrt() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Sqrt() = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]string{
			"negative": "-1",
		}
		for name, d := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(d)
				_, err := d.Sqrt()
				if err == nil {
					t.Errorf("%q.Sqrt() did not fail", d)
				}
			})
		}
	})
}

func TestDecimal_Abs(t *testing.T) {
	tests := []struct {
		d, want string
	}{
		{"1", "1"},
		{"-1", "1"},
		{"1.00", "1.00"},
		{"-1.00", "1.00"},
		{"0", "0"},
		{"0.0", "0.0"},
		{"0.00", "0.00"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Abs()
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Abs() = %q, want %q", d, got, want)
		}
	}
}

func TestDecimal_CopySign(t *testing.T) {
	tests := []struct {
		d, e, want string
	}{
		{"10", "1", "10"},
		{"10", "0", "10"},
		{"10", "-1", "-10"},
		{"0", "1", "0"},
		{"0", "0", "0"},
		{"0", "-1", "0"},
		{"-10", "1", "10"},
		{"-10", "0", "10"},
		{"-10", "-1", "-10"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		e := MustParse(tt.e)
		got := d.CopySign(e)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.CopySign(%q) = %q, want %q", d, e, got, want)
		}
	}
}

func TestDecimal_Neg(t *testing.T) {
	tests := []struct {
		d, want string
	}{
		{"1", "-1"},
		{"-1", "1"},
		{"1.00", "-1.00"},
		{"-1.00", "1.00"},
		{"0", "0"},
		{"0.0", "0.0"},
		{"0.00", "0.00"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		got := d.Neg()
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Neg() = %q, want %q", d, got, want)
		}
	}
}

func TestDecimal_Quo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, want string
		}{
			// Zeros
			{"0", "1.000", "0"},
			{"0.0", "1.000", "0"},
			{"0.00", "1.000", "0"},
			{"0.000", "1.000", "0"},
			{"0.0000", "1.000", "0.0"},
			{"0.00000", "1.000", "0.00"},

			{"0.000", "1", "0.000"},
			{"0.000", "1.0", "0.00"},
			{"0.000", "1.00", "0.0"},
			{"0.000", "1.000", "0"},
			{"0.000", "1.0000", "0"},
			{"0.000", "1.00000", "0"},

			// Ones
			{"1", "1.000", "1"},
			{"1.0", "1.000", "1"},
			{"1.00", "1.000", "1"},
			{"1.000", "1.000", "1"},
			{"1.0000", "1.000", "1.0"},
			{"1.00000", "1.000", "1.00"},

			{"1.000", "1", "1.000"},
			{"1.000", "1.0", "1.00"},
			{"1.000", "1.00", "1.0"},
			{"1.000", "1.000", "1"},
			{"1.000", "1.0000", "1"},
			{"1.000", "1.00000", "1"},

			// Simple cases
			{"1", "1", "1"},
			{"2", "1", "2"},
			{"1", "2", "0.5"},
			{"2", "2", "1"},
			{"0", "1", "0"},
			{"0", "2", "0"},
			{"1.5", "3", "0.5"},
			{"3", "3", "1"},
			{"9999999999999999999", "1", "9999999999999999999"},
			{"9999999999999999999", "9999999999999999999", "1"},

			// Signs
			{"2.4", "1", "2.4"},
			{"2.4", "-1", "-2.4"},
			{"-2.4", "1", "-2.4"},
			{"-2.4", "-1", "2.4"},

			// Scales
			{"2.40", "1", "2.40"},
			{"2.400", "1", "2.400"},
			{"2.4", "2", "1.2"},
			{"2.400", "2", "1.200"},

			// 1 divided by digits
			{"1", "1", "1"},
			{"1", "2", "0.5"},
			{"1", "3", "0.3333333333333333333"},
			{"1", "4", "0.25"},
			{"1", "5", "0.2"},
			{"1", "6", "0.1666666666666666667"},
			{"1", "7", "0.1428571428571428571"},
			{"1", "8", "0.125"},
			{"1", "9", "0.1111111111111111111"},

			// 2 divided by digits
			{"2", "1", "2"},
			{"2", "2", "1"},
			{"2", "3", "0.6666666666666666667"},
			{"2", "4", "0.5"},
			{"2", "5", "0.4"},
			{"2", "6", "0.3333333333333333333"},
			{"2", "7", "0.2857142857142857143"},
			{"2", "8", "0.25"},
			{"2", "9", "0.2222222222222222222"},

			// 2 divided by 3
			{"0.0000000000000000002", "3", "0.0000000000000000001"},
			{"0.0000000000000000002", "3.000000000000000000", "0.0000000000000000001"},
			{"2", "3", "0.6666666666666666667"},
			{"2.000000000000000000", "3", "0.6666666666666666667"},
			{"2", "3.000000000000000000", "0.6666666666666666667"},
			{"2.000000000000000000", "3.000000000000000000", "0.6666666666666666667"},
			{"0.0000000000000000002", "0.0000000000000000003", "0.6666666666666666667"},
			{"2", "0.0000000000000000003", "6666666666666666667"},
			{"2.000000000000000000", "0.0000000000000000003", "6666666666666666667"},

			// Interest accrual
			{"0.0001", "365", "0.0000002739726027397"}, // no error
			{"0.0001", "366", "0.0000002732240437158"}, // no error

			// Captured during fuzzing
			{"9223372036854775807", "-9223372036854775808", "-0.9999999999999999999"},
			{"0.000000000000000001", "20", "0.000000000000000000"},
			{"105", "0.999999999999999990", "105.0000000000000011"},
			{"0.05", "999999999999999954", "0.0000000000000000001"},
			{"9.99999999999999998", "185", "0.0540540540540540539"},
			{"7", "2.000000000000000002", "3.499999999999999997"},
			{"0.000000009", "999999999999999999", "0.000000000"},
			{"0.0000000000000000001", "9999999999999999999", "0.0000000000000000000"},
			{"9999999999999999999", "2", "5000000000000000000"},
			{"9999999999999999999", "5000000000000000000", "2"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			got, err := d.Quo(e)
			if err != nil {
				t.Errorf("%q.Quo(%q) failed: %v", d, e, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Quo(%q) = %q, want %q", d, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e  string
			scale int
		}{
			"zero 1":     {"1", "0", 0},
			"overflow 1": {"9999999999999999999", "0.001", 0},
			"scale 1":    {"1", "1", MaxScale},
			"scale 2":    {"0", "1", MaxScale + 1},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			_, err := d.QuoExact(e, tt.scale)
			if err == nil {
				t.Errorf("%q.QuoExact(%q, %v) did not fail", d, e, tt.scale)
			}
		}
	})
}

func TestDecimal_Inv(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			{"0.1", "10"},
			{"1", "1"},
			{"10", "0.1"},
			{"2", "0.5"},
			{"2.0", "0.5"},
			{"2.00", "0.5"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Inv()
			if err != nil {
				t.Errorf("%q.Inv() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Inv() = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d string
		}{
			"zero 1":     {"0"},
			"overflow 1": {"0.0000000000000000001"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			_, err := d.Inv()
			if err == nil {
				t.Errorf("%q.Inv() did not fail", d)
			}
		}
	})
}

func TestDecimal_QuoRem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, wantQuo, wantRem string
		}{
			// Zeros
			{"0", "1.000", "0", "0.000"},
			{"0.0", "1.000", "0", "0.000"},
			{"0.00", "1.000", "0", "0.000"},
			{"0.000", "1.000", "0", "0.000"},
			{"0.0000", "1.000", "0", "0.0000"},
			{"0.00000", "1.000", "0", "0.00000"},

			{"0.000", "1", "0", "0.000"},
			{"0.000", "1.0", "0", "0.000"},
			{"0.000", "1.00", "0", "0.000"},
			{"0.000", "1.000", "0", "0.000"},
			{"0.000", "1.0000", "0", "0.0000"},
			{"0.000", "1.00000", "0", "0.00000"},

			// Ones
			{"1", "1.000", "1", "0.000"},
			{"1.0", "1.000", "1", "0.000"},
			{"1.00", "1.000", "1", "0.000"},
			{"1.000", "1.000", "1", "0.000"},
			{"1.0000", "1.000", "1", "0.0000"},
			{"1.00000", "1.000", "1", "0.00000"},

			{"1.000", "1", "1", "0.000"},
			{"1.000", "1.0", "1", "0.000"},
			{"1.000", "1.00", "1", "0.000"},
			{"1.000", "1.000", "1", "0.000"},
			{"1.000", "1.0000", "1", "0.0000"},
			{"1.000", "1.00000", "1", "0.00000"},

			// Signs
			{"2.4", "1", "2", "0.4"},
			{"2.4", "-1", "-2", "0.4"},
			{"-2.4", "1", "-2", "-0.4"},
			{"-2.4", "-1", "2", "-0.4"},

			// Scales
			{"2.40", "1", "2", "0.40"},
			{"2.400", "1", "2", "0.400"},
			{"2.4", "2", "1", "0.4"},
			{"2.400", "2", "1", "0.400"},

			// 1 divided by digits
			{"1", "1", "1", "0"},
			{"1", "2", "0", "1"},
			{"1", "3", "0", "1"},
			{"1", "4", "0", "1"},
			{"1", "5", "0", "1"},
			{"1", "6", "0", "1"},
			{"1", "7", "0", "1"},
			{"1", "8", "0", "1"},
			{"1", "9", "0", "1"},

			// 2 divided by digits
			{"2", "1", "2", "0"},
			{"2", "2", "1", "0"},
			{"2", "3", "0", "2"},
			{"2", "4", "0", "2"},
			{"2", "5", "0", "2"},
			{"2", "6", "0", "2"},
			{"2", "7", "0", "2"},
			{"2", "8", "0", "2"},
			{"2", "9", "0", "2"},

			// Other tests
			{"12345", "4.999", "2469", "2.469"},
			{"12345", "4.99", "2473", "4.73"},
			{"12345", "4.9", "2519", "1.9"},
			{"12345", "5", "2469", "0"},
			{"12345", "5.1", "2420", "3.0"},
			{"12345", "5.01", "2464", "0.36"},
			{"12345", "5.001", "2468", "2.532"},

			{"41", "21", "1", "20"},
			{"4.2", "3.1000003", "1", "1.0999997"},
			{"1.000000000000000000", "0.000000000000000003", "333333333333333333", "0.000000000000000001"},
			{"1.000000000000000001", "0.000000000000000003", "333333333333333333", "0.000000000000000002"},
			{"3", "0.9999999999999999999", "3", "0.0000000000000000003"},
			{"0.9999999999999999999", "3", "0", "0.9999999999999999999"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			gotQuo, gotRem, err := d.QuoRem(e)
			if err != nil {
				t.Errorf("%q.QuoRem(%q) failed: %v", d, e, err)
				continue
			}
			wantQuo := MustParse(tt.wantQuo)
			wantRem := MustParse(tt.wantRem)
			if gotQuo != wantQuo || gotRem != wantRem {
				t.Errorf("%q.QuoRem(%q) = (%q, %q), want (%q, %q)", d, e, gotQuo, gotRem, wantQuo, wantRem)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e string
		}{
			"zero 1":     {"1", "0"},
			"overflow 1": {"9999999999999999999", "0.0000000000000000001"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			_, _, err := d.QuoRem(e)
			if err == nil {
				t.Errorf("%q.QuoRem(%q) did not fail", d, e)
			}
		}
	})
}

func TestDecimal_Cmp(t *testing.T) {
	tests := []struct {
		d, e string
		want int
	}{
		{"-2", "-2", 0},
		{"-2", "-1", -1},
		{"-2", "0", -1},
		{"-2", "1", -1},
		{"-2", "2", -1},
		{"-1", "-2", 1},
		{"-1", "-1", 0},
		{"-1", "0", -1},
		{"-1", "1", -1},
		{"-1", "2", -1},
		{"0", "-2", 1},
		{"0", "-1", 1},
		{"0", "0", 0},
		{"0", "1", -1},
		{"0", "2", -1},
		{"1", "-2", 1},
		{"1", "-1", 1},
		{"1", "0", 1},
		{"1", "1", 0},
		{"1", "2", -1},
		{"2", "-2", 1},
		{"2", "-1", 1},
		{"2", "0", 1},
		{"2", "1", 1},
		{"2", "2", 0},
		{"2", "2.0", 0},
		{"2", "2.00", 0},
		{"2", "2.000", 0},
		{"2", "2.0000", 0},
		{"2", "2.00000", 0},
		{"2", "2.000000", 0},
		{"2", "2.0000000", 0},
		{"2", "2.00000000", 0},
		{"9999999999999999999", "0.9999999999999999999", 1},
		{"0.9999999999999999999", "9999999999999999999", -1},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		e := MustParse(tt.e)
		got := d.Cmp(e)
		if got != tt.want {
			t.Errorf("%q.Cmp(%q) = %v, want %v", d, e, got, tt.want)
		}
	}
}

func TestDecimal_Max(t *testing.T) {
	tests := []struct {
		d, e, want string
	}{
		{"-2", "-2", "-2"},
		{"-2", "-1", "-1"},
		{"-2", "0", "0"},
		{"-2", "1", "1"},
		{"-2", "2", "2"},
		{"-1", "-2", "-1"},
		{"-1", "-1", "-1"},
		{"-1", "0", "0"},
		{"-1", "1", "1"},
		{"-1", "2", "2"},
		{"0", "-2", "0"},
		{"0", "-1", "0"},
		{"0", "0", "0"},
		{"0", "1", "1"},
		{"0", "2", "2"},
		{"1", "-2", "1"},
		{"1", "-1", "1"},
		{"1", "0", "1"},
		{"1", "1", "1"},
		{"1", "2", "2"},
		{"2", "-2", "2"},
		{"2", "-1", "2"},
		{"2", "0", "2"},
		{"2", "1", "2"},
		{"2", "2", "2"},
		{"0.000", "0.0", "0.0"},
		{"0.0", "0.000", "0.0"},
		{"-0.000", "-0.0", "0.0"},
		{"-0.0", "-0.000", "0.0"},
		{"1.23", "1.2300", "1.23"},
		{"1.2300", "1.23", "1.23"},
		{"-1.23", "-1.2300", "-1.23"},
		{"-1.2300", "-1.23", "-1.23"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		e := MustParse(tt.e)
		got := d.Max(e)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Max(%q) = %q, want %q", d, e, got, want)
		}
	}
}

func TestDecimal_Min(t *testing.T) {
	tests := []struct {
		d, e, want string
	}{
		{"-2", "-2", "-2"},
		{"-2", "-1", "-2"},
		{"-2", "0", "-2"},
		{"-2", "1", "-2"},
		{"-2", "2", "-2"},
		{"-1", "-2", "-2"},
		{"-1", "-1", "-1"},
		{"-1", "0", "-1"},
		{"-1", "1", "-1"},
		{"-1", "2", "-1"},
		{"0", "-2", "-2"},
		{"0", "-1", "-1"},
		{"0", "0", "0"},
		{"0", "1", "0"},
		{"0", "2", "0"},
		{"1", "-2", "-2"},
		{"1", "-1", "-1"},
		{"1", "0", "0"},
		{"1", "1", "1"},
		{"1", "2", "1"},
		{"2", "-2", "-2"},
		{"2", "-1", "-1"},
		{"2", "0", "0"},
		{"2", "1", "1"},
		{"2", "2", "2"},
		{"0.000", "0.0", "0.000"},
		{"0.0", "0.000", "0.000"},
		{"-0.000", "-0.0", "0.000"},
		{"-0.0", "-0.000", "0.000"},
		{"1.23", "1.2300", "1.2300"},
		{"1.2300", "1.23", "1.2300"},
		{"-1.23", "-1.2300", "-1.2300"},
		{"-1.2300", "-1.23", "-1.2300"},
	}
	for _, tt := range tests {
		d := MustParse(tt.d)
		e := MustParse(tt.e)
		got := d.Min(e)
		want := MustParse(tt.want)
		if got != want {
			t.Errorf("%q.Min(%q) = %q, want %q", d, e, got, want)
		}
	}
}

func TestDecimal_Clamp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, min, max, want string
		}{
			{"0", "-2", "-1", "-1"},
			{"0", "-1", "1", "0"},
			{"0", "1", "2", "1"},
			{"0.000", "0.0", "0.000", "0.000"},
			{"0.000", "0.000", "0.0", "0.000"},
			{"0.0", "0.0", "0.000", "0.0"},
			{"0.0", "0.000", "0.0", "0.0"},
			{"0.000", "0.000", "1", "0.000"},
			{"0.000", "0.0", "1", "0.0"},
			{"0.0", "0.000", "1", "0.0"},
			{"0.0", "0.0", "1", "0.0"},
			{"0.000", "-1", "0.000", "0.000"},
			{"0.000", "-1", "0.0", "0.000"},
			{"0.0", "-1", "0.000", "0.000"},
			{"0.0", "-1", "0.0", "0.0"},
			{"1.2300", "1.2300", "2", "1.2300"},
			{"1.2300", "1.23", "2", "1.23"},
			{"1.23", "1.2300", "2", "1.23"},
			{"1.23", "1.23", "2", "1.23"},
			{"1.2300", "1", "1.2300", "1.2300"},
			{"1.2300", "1", "1.23", "1.2300"},
			{"1.23", "1", "1.2300", "1.2300"},
			{"1.23", "1", "1.23", "1.23"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			min := MustParse(tt.min)
			max := MustParse(tt.max)
			got, err := d.Clamp(min, max)
			if err != nil {
				t.Errorf("%q.Clamp(%q, %q) failed: %v", d, min, max, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Clamp(%q, %q) = %q, want %q", d, min, max, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			d, min, max string
		}{
			{"0", "1", "-1"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			min := MustParse(tt.min)
			max := MustParse(tt.max)
			_, err := d.Clamp(min, max)
			if err == nil {
				t.Errorf("%q.Clamp(%q, %q) did not fail", d, min, max)
			}
		}
	})
}

func TestNullDecimal_Interfaces(t *testing.T) {
	var n any = NullDecimal{}
	_, ok := n.(driver.Valuer)
	if !ok {
		t.Errorf("%T does not implement driver.Valuer", n)
	}

	n = &NullDecimal{}
	_, ok = n.(sql.Scanner)
	if !ok {
		t.Errorf("%T does not implement sql.Scanner", n)
	}
}

func TestNullDecimal_Scan(t *testing.T) {
	t.Run("[]byte", func(t *testing.T) {
		tests := []string{"."}
		for _, tt := range tests {
			got := NullDecimal{}
			err := got.Scan([]byte(tt))
			if err == nil {
				t.Errorf("Scan(%q) did not fail", tt)
			}
		}
	})
}

/******************************************************
* Fuzzing
******************************************************/

var corpus = []struct {
	neg   bool
	scale int
	coef  uint64
}{
	// zero
	{false, 0, 0},

	// positive
	{false, 0, 1},
	{false, 0, 3},
	{false, 0, 9999999999999999999},
	{false, 19, 3},
	{false, 19, 1},
	{false, 19, 9999999999999999999},

	// negative
	{true, 0, 1},
	{true, 0, 3},
	{true, 0, 9999999999999999999},
	{true, 19, 1},
	{true, 19, 3},
	{true, 19, 9999999999999999999},
}

func FuzzParse(f *testing.F) {
	for _, c := range corpus {
		for s := 0; s <= MaxScale; s++ {
			d, err := newSafe(c.neg, fint(c.coef), c.scale)
			if err != nil {
				continue
			}
			f.Add(d.String(), s)
		}
	}

	f.Fuzz(
		func(t *testing.T, num string, scale int) {
			got, err := parseFint(num, scale)
			if err != nil {
				t.Skip()
				return
			}
			want, err := parseBint(num, scale)
			if err != nil {
				t.Errorf("parseBint(%q) failed: %v", num, err)
				return
			}
			if got.CmpTotal(want) != 0 {
				t.Errorf("parseBint(%q) = %q, whereas parseFint(%q) = %q", num, want, num, got)
			}
		},
	)
}

func FuzzBCD(f *testing.F) {
	for _, c := range corpus {
		d, err := newSafe(c.neg, fint(c.coef), c.scale)
		if err != nil {
			continue
		}
		f.Add(d.bcd())
	}

	f.Fuzz(
		func(t *testing.T, bcd []byte) {
			_, err := parseBCD(bcd)
			if err != nil {
				t.Skip()
			}
		},
	)
}

func FuzzDecimal_String(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			want, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			s := want.String()
			got, err := Parse(s)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", s, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("Parse(%q) = %v, want %v", s, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_BCD(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			want, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			s := want.bcd()
			got, err := parseBCD(s)
			if err != nil {
				t.Errorf("parseBCD(% x) failed: %v", s, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("parseBCD(% x) = %v, want %v", s, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Int64(f *testing.F) {
	for _, d := range corpus {
		for s := 0; s <= MaxScale; s++ {
			f.Add(d.neg, d.scale, d.coef, s)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, scale int) {
			want, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}

			w, f, ok := want.Int64(scale)
			if !ok {
				t.Skip()
				return
			}

			got, err := NewFromInt64(w, f, scale)
			if err != nil {
				t.Errorf("NewFromInt64(%v, %v, %v) failed: %v", w, f, scale, err)
				return
			}

			want = want.Round(scale)
			if got.Cmp(want) != 0 {
				t.Errorf("NewFromInt64(%v, %v, %v) = %v, want %v", w, f, scale, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Float64(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64) {
			want, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil || want.Prec() > 17 {
				t.Skip()
				return
			}

			f, ok := want.Float64()
			if !ok {
				t.Errorf("%q.Float64() failed", want)
				return
			}

			got, err := NewFromFloat64(f)
			if err != nil {
				t.Logf("%q.Float64() = %v", want, f)
				t.Errorf("NewFromFloat64(%v) failed: %v", f, err)
				return
			}

			if got.Cmp(want) != 0 {
				t.Errorf("NewFromFloat64(%v) = %v, want %v", f, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Mul(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := 0; s <= MaxScale; s++ {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
			if scale < 0 || MaxScale < scale {
				t.Skip()
				return
			}
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}
			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.mulFint(e, scale)
			if err != nil {
				if errors.Is(err, errDecimalOverflow) {
					t.Skip() // Decimal overflow is an expected error in fast multiplication
				} else {
					t.Errorf("mulFint(%q, %q, %v) failed: %v", d, e, scale, err)
				}
				return
			}

			want, err := d.mulBint(e, scale)
			if err != nil {
				t.Errorf("mulBint(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}
			if got.CmpTotal(want) != 0 {
				t.Errorf("mulBint(%q, %q, %v) = %q, whereas mulFint(%q, %q, %v) = %q", d, e, scale, want, d, e, scale, got)
			}
		},
	)
}

func FuzzDecimal_FMA(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for _, g := range corpus {
				for s := 0; s <= MaxScale; s++ {
					f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, g.neg, g.scale, g.coef, s)
				}
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, gneg bool, gscale int, gcoef uint64, scale int) {
			if scale < 0 || MaxScale < scale {
				t.Skip()
				return
			}
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}
			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}
			g, err := newSafe(gneg, fint(gcoef), gscale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.fmaFint(e, g, scale)
			if err != nil {
				if errors.Is(err, errDecimalOverflow) {
					t.Skip() // Decimal overflow is an expected error in fast fused multiplication-addition
				} else {
					t.Errorf("fmaFint(%q, %q, %q, %v) failed: %v", d, e, g, scale, err)
				}
				return
			}

			want, err := d.fmaBint(e, g, scale)
			if err != nil {
				t.Errorf("fmaBint(%q, %q, %q, %v) failed: %v", d, e, g, scale, err)
				return
			}
			if got.CmpTotal(want) != 0 {
				t.Errorf("fmaBint(%q, %q, %q, %v) = %q, whereas fmaFint(%q, %q, %q, %v) = %q", d, e, g, scale, want, d, e, g, scale, got)
			}
		},
	)
}

func FuzzDecimal_Add(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := 0; s <= MaxScale; s++ {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
			if scale < 0 || MaxScale < scale {
				t.Skip()
				return
			}
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}
			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.addFint(e, scale)
			if err != nil {
				if errors.Is(err, errDecimalOverflow) {
					t.Skip() // Decimal overflow is an expected error in fast addition
				} else {
					t.Errorf("addFint(%q, %q, %v) failed: %v", d, e, scale, err)
				}
				return
			}

			want, err := d.addBint(e, scale)
			if err != nil {
				t.Errorf("addBint(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}
			if got.CmpTotal(want) != 0 {
				t.Errorf("addBint(%q, %q, %v) = %q, whereas addFint(%q, %q, %v) = %q", d, e, scale, want, d, e, scale, got)
			}
		},
	)
}

func FuzzDecimal_Quo(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := 0; s <= MaxScale; s++ {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
			if scale < 0 || MaxScale < scale {
				t.Skip()
				return
			}
			if ecoef == 0 {
				t.Skip()
				return
			}
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}
			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.quoFint(e, scale)
			if err != nil {
				switch {
				case errors.Is(err, errDecimalOverflow):
					t.Skip() // Decimal overflow is an expected error in fast division
				case errors.Is(err, errInexactDivision):
					t.Skip() // Inexact division is an expected error in fast division
				default:
					t.Errorf("quoFint(%q, %q, %v) failed: %v", d, e, scale, err)
				}
				return
			}

			want, err := d.quoBint(e, scale)
			if err != nil {
				t.Errorf("quoBint(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}
			if got.Cmp(want) != 0 {
				t.Errorf("quoBint(%q, %q, %v) = %q, whereas quoFint(%q, %q, %v) = %q", d, e, scale, want, d, e, scale, got)
			}
		},
	)
}

func FuzzDecimal_QuoRem(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64) {
			if ecoef == 0 {
				t.Skip()
				return
			}
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}
			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}

			gotQ, gotR, err := d.quoRemFint(e)
			if err != nil {
				switch {
				case errors.Is(err, errDecimalOverflow):
					t.Skip() // Decimal overflow is an expected error in fast division
				default:
					t.Errorf("quoRemFint(%q, %q) failed: %v", d, e, err)
				}
				return
			}

			wantQ, wantR, err := d.quoRemBint(e)
			if err != nil {
				t.Errorf("quoRemBint(%q, %q) failed: %v", d, e, err)
				return
			}

			if gotQ.Cmp(wantQ) != 0 || gotR.Cmp(wantR) != 0 {
				t.Errorf("quoRemBint(%q, %q) = (%q, %q), whereas quoRemFint(%q, %q) = (%q, %q)", d, e, wantQ, wantR, d, e, gotQ, gotR)
			}
		},
	)
}

func FuzzDecimal_Cmp(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64) {
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}

			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.cmpFint(e)
			if err != nil {
				if errors.Is(err, errDecimalOverflow) {
					t.Skip() // Decimal overflow is an expected error in fast comparison
				} else {
					t.Errorf("cmpFint(%q, %q) failed: %v", d, e, err)
				}
				return
			}

			want := d.cmpBint(e)
			if got != want {
				t.Errorf("cmpBint(%q, %q) = %v, whereas cmpFint(%q, %q) = %v", d, e, want, d, e, got)
				return
			}
		},
	)
}

func FuzzDecimal_Sqrt(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			if neg {
				t.Skip()
				return
			}
			if scale < 0 || MaxScale < scale {
				t.Skip()
				return
			}
			want, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}
			d, err := want.Sqrt()
			if err != nil {
				t.Errorf("%q.Sqrt() failed: %v", want, err)
				return
			}
			got, err := d.Pow(2)
			if err != nil {
				if errors.Is(err, errDecimalOverflow) {
					t.Skip() // Decimal overflow is an expected error here
				} else {
					t.Errorf("%q.Pow(2) failed: %v", d, err)
				}
				return
			}
			if cmp, err := cmpULP(got, want, 3); err != nil {
				t.Errorf("cmpULP(%q, %q) failed: %v", got, want, err)
			} else if cmp != 0 {
				t.Errorf("%q.Sqrt().Pow(2) = %q, want %q", want, got, want)
				return
			}
		},
	)
}

// cmpULP compares decimals and returns 0 if they are within specified number of ULPs.
func cmpULP(d, e Decimal, ulps int) (int, error) {
	n, err := New(int64(ulps), 0)
	if err != nil {
		return 0, err
	}
	dist, err := d.SubAbs(e)
	if err != nil {
		return 0, err
	}
	ulp := d.ULP().Min(e.ULP())
	tlr, err := ulp.Mul(n)
	if err != nil {
		return 0, err
	}
	if dist.Cmp(tlr) <= 0 {
		return 0, nil
	}
	return d.Cmp(e), nil
}

func FuzzDecimal_CmpSub(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64) {
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}

			e, err := newSafe(eneg, fint(ecoef), escale)
			if err != nil {
				t.Skip()
				return
			}

			got := d.Cmp(e)
			f, err := d.Sub(e)
			if err != nil {
				if errors.Is(err, errDecimalOverflow) {
					t.Skip() // Decimal overflow is an expected error in subtraction
				} else {
					t.Errorf("%q.Sub(%q) failed: %v", d, e, err)
				}
				return
			}
			want := f.Sign()
			if got != want {
				t.Errorf("%q.Cmp(%q) = %v, whereas %q.Sub(%q).Sign() = %v", d, e, got, d, e, want)
				return
			}
		},
	)
}

func FuzzDecimal_New(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			got, err := newFromFint(neg, fint(coef), scale, 0)
			if err != nil {
				t.Skip()
				return
			}
			want, err := newFromBint(neg, newBintFromUint64(coef), scale, 0)
			if err != nil {
				t.Errorf("newDecimalFromBint(%v, %v, %v, 0) failed: %v", neg, coef, scale, err)
				return
			}
			if got.CmpTotal(want) != 0 {
				t.Errorf("newDecimalFromFint(%v, %v, %v, 0) = %q, whereas newDecimalFromBint(%v, %v, %v, 0) = %q", neg, coef, scale, got, neg, coef, scale, want)
			}
		},
	)
}

// newBintFromUint64 converts uint64 to *big.Int.
func newBintFromUint64(u uint64) *bint {
	z := new(big.Int)
	z.SetUint64(u)
	return (*bint)(z)
}

func FuzzDecimal_Pad(f *testing.F) {
	for _, d := range corpus {
		for s := 0; s <= MaxScale; s++ {
			f.Add(d.neg, d.scale, d.coef, s)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, scale int) {
			want, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}
			got := want.Pad(scale)
			if got.Cmp(want) != 0 {
				t.Errorf("%q.Pad(%v) = %q", want, scale, got)
				return
			}
			if got.Scale() > MaxScale {
				t.Errorf("%q.Pad(%v).Scale() = %v", want, scale, got.Scale())
				return
			}
		},
	)
}
