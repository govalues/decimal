package decimal

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
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
	_, ok = d.(json.Marshaler)
	if !ok {
		t.Errorf("%T does not implement json.Marshaler", d)
	}
	_, ok = d.(encoding.TextMarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.TextMarshaler", d)
	}
	_, ok = d.(encoding.BinaryMarshaler)
	if !ok {
		t.Errorf("%T does not implement encoding.BinaryMarshaler", d)
	}
	// Uncomment when Go 1.24 is minimum supported version.
	// _, ok = d.(encoding.TextAppender)
	// if !ok {
	// 	t.Errorf("%T does not implement encoding.TextAppender", d)
	// }
	// _, ok = d.(encoding.BinaryAppender)
	// if !ok {
	// 	t.Errorf("%T does not implement encoding.BinaryAppender", d)
	// }
	_, ok = d.(driver.Valuer)
	if !ok {
		t.Errorf("%T does not implement driver.Valuer", d)
	}

	d = &Decimal{}
	_, ok = d.(json.Unmarshaler)
	if !ok {
		t.Errorf("%T does not implement json.Unmarshaler", d)
	}
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
			value int64
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
			got, err := New(tt.value, tt.scale)
			if err != nil {
				t.Errorf("New(%v, %v) failed: %v", tt.value, tt.scale, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("New(%v, %v) = %q, want %q", tt.value, tt.scale, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			value int64
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
			_, err := New(tt.value, tt.scale)
			if err == nil {
				t.Errorf("New(%v, %v) did not fail", tt.value, tt.scale)
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

			// Powers of ten
			{1e-21, "0.0000000000000000000"},
			{1e-20, "0.0000000000000000000"},
			{1e-19, "0.0000000000000000001"},
			{1e-18, "0.000000000000000001"},
			{1e-17, "0.00000000000000001"},
			{1e-16, "0.0000000000000001"},
			{1e-15, "0.000000000000001"},
			{1e-14, "0.00000000000001"},
			{1e-13, "0.0000000000001"},
			{1e-12, "0.000000000001"},
			{1e-11, "0.00000000001"},
			{1e-10, "0.0000000001"},
			{1e-9, "0.000000001"},
			{1e-8, "0.00000001"},
			{1e-7, "0.0000001"},
			{1e-6, "0.000001"},
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
			{1e6, "1000000"},
			{1e7, "10000000"},
			{1e8, "100000000"},
			{1e9, "1000000000"},
			{1e10, "10000000000"},
			{1e11, "100000000000"},
			{1e12, "1000000000000"},
			{1e13, "10000000000000"},
			{1e14, "100000000000000"},
			{1e15, "1000000000000000"},
			{1e16, "10000000000000000"},
			{1e17, "100000000000000000"},
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
			{"-999999999999999999.99", true, 1000000000000000000, 0},
			{"0.00000000000000000000000000000000000000", false, 0, 19},
			{"0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false, 0, 19},
			{"-0.00000000000000000000000000000000000001", false, 0, 19},
			{"0.00000000000000000000000000000000000001", false, 0, 19},
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

func TestDecimalUnmarshalText(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		d := Decimal{}
		err := d.UnmarshalText([]byte("1.1.1"))
		if err == nil {
			t.Errorf("UnmarshalText(\"1.1.1\") did not fail")
		}
	})
}

func TestDecimalUnmarshalBinary(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		d := Decimal{}
		err := d.UnmarshalBinary([]byte("1.1.1"))
		if err == nil {
			t.Errorf("UnmarshalBinary(\"1.1.1\") did not fail")
		}
	})
}

func TestDecimalUnmarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			s    string
			want string
		}{
			{"null", "0"},
			{"\"-9999999999999999999.0\"", "-9999999999999999999"},
			{"\"-9999999999999999999\"", "-9999999999999999999"},
			{"\"-999999999999999999.9\"", "-999999999999999999.9"},
			{"\"-99999999999999999.99\"", "-99999999999999999.99"},
			{"\"-1000000000000000000.0\"", "-1000000000000000000"},
			{"\"-0.9999999999999999999\"", "-0.9999999999999999999"},
		}
		for _, tt := range tests {
			var got Decimal
			err := got.UnmarshalJSON([]byte(tt.s))
			if err != nil {
				t.Errorf("UnmarshalJSON(%q) failed: %v", tt.s, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("UnmarshalJSON(%q) = %q, want %q", tt.s, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		d := Decimal{}
		err := d.UnmarshalJSON([]byte("\"-1.1.1\""))
		if err == nil {
			t.Errorf("UnmarshalJSON(\"-1.1.1\") did not fail")
		}
	})
}

func TestDecimalUnmarshalBSONValue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			typ  byte
			data []byte
			want string
		}{
			{1, []byte{0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0xf0, 0x3f}, "1.0001220703125"},
			{2, []byte{0x15, 0x0, 0x0, 0x0, 0x2d, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x0}, "-9999999999999999999"},
			{10, nil, "0"},
			{16, []byte{0xff, 0xff, 0xff, 0x7f}, "2147483647"},
			{18, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, "9223372036854775807"},
			{19, []byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "1.265"},
		}
		for _, tt := range tests {
			got := Decimal{}
			err := got.UnmarshalBSONValue(tt.typ, tt.data)
			if err != nil {
				t.Errorf("UnmarshalBSONValue(%v, % x) failed: %v", tt.typ, tt.data, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("UnmarshalBSONValue(%v, % x) = %q, want %q", tt.typ, tt.data, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		d := Decimal{}
		err := d.UnmarshalBSONValue(12, nil)
		if err == nil {
			t.Errorf("UnmarshalBSONValue(12, nil) did not fail")
		}
	})
}

func TestParseBSONFloat64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			b    []byte
			want string
		}{
			// Zero
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0"},

			// Negative zero
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0"},

			// Integers
			{[]byte{0x2a, 0x1b, 0xf5, 0xf4, 0x10, 0x22, 0xb1, 0xc3}, "-1234567892123200000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0xbf}, "-1"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x3f}, "1"},
			{[]byte{0x2a, 0x1b, 0xf5, 0xf4, 0x10, 0x22, 0xb1, 0x43}, "1234567892123200000"},

			// Floats
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0xf0, 0xbf}, "-1.0001220703125"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0xf0, 0x3f}, "1.0001220703125"},
		}

		for _, tt := range tests {
			got, err := parseBSONFloat64(tt.b)
			if err != nil {
				t.Errorf("parseBSONFloat64(%v) failed: %v", tt.b, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("parseBSONFloat64(%v) = %q, want %q", tt.b, got, want)
			}

		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]byte{
			"length 1": {},
			"length 2": {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			"length 3": {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			"nan 1":    {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf8, 0x7f},
			"nan 2":    {0x12, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf8, 0x7f},
			"inf 1":    {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x7f},
			"inf 2":    {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0xff},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := parseBSONFloat64(tt)
				if err == nil {
					t.Errorf("parseBSONFloat64(%v) did not fail", tt)
				}
			})
		}
	})
}

func TestParseBSONString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			b    []byte
			want string
		}{
			{[]byte{0x15, 0x0, 0x0, 0x0, 0x2d, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x0}, "-9999999999999999999"},
			{[]byte{0x17, 0x0, 0x0, 0x0, 0x2d, 0x30, 0x2e, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x31, 0x0}, "-0.0000000000000000001"},
			{[]byte{0x16, 0x0, 0x0, 0x0, 0x30, 0x2e, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x0}, "0.0000000000000000000"},
			{[]byte{0x2, 0x0, 0x0, 0x0, 0x30, 0x0}, "0"},
			{[]byte{0x16, 0x0, 0x0, 0x0, 0x30, 0x2e, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x31, 0x0}, "0.0000000000000000001"},
			{[]byte{0x2, 0x0, 0x0, 0x0, 0x31, 0x0}, "1"},
			{[]byte{0x3, 0x0, 0x0, 0x0, 0x30, 0x31, 0x0}, "1"},
			{[]byte{0x15, 0x0, 0x0, 0x0, 0x33, 0x2e, 0x31, 0x34, 0x31, 0x35, 0x39, 0x32, 0x36, 0x35, 0x33, 0x35, 0x38, 0x39, 0x37, 0x39, 0x33, 0x32, 0x33, 0x38, 0x0}, "3.141592653589793238"},
			{[]byte{0x3, 0x0, 0x0, 0x0, 0x31, 0x30, 0x0}, "10"},
			{[]byte{0x14, 0x0, 0x0, 0x0, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x0}, "9999999999999999999"},
		}
		for _, tt := range tests {
			got, err := parseBSONString(tt.b)
			if err != nil {
				t.Errorf("parseBSONString(%v) failed: %v", tt.b, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("parseBSONString(%v) = %q, want %q", tt.b, got, want)
			}

		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]byte{
			"length 1":     {},
			"length 2":     {0x0, 0x0, 0x0, 0x0},
			"length 3":     {0x2, 0x0, 0x0, 0x0, 0x30},
			"terminator 1": {0x2, 0x0, 0x0, 0x0, 0x30, 0x30},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := parseBSONString(tt)
				if err == nil {
					t.Errorf("parseBSONString(%v) did not fail", tt)
				}
			})
		}
	})
}

func TestParseBSONInt32(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			b    []byte
			want string
		}{
			{[]byte{0x0, 0x0, 0x0, 0x80}, "-2147483648"},
			{[]byte{0xff, 0xff, 0xff, 0xff}, "-1"},
			{[]byte{0x0, 0x0, 0x0, 0x0}, "0"},
			{[]byte{0x1, 0x0, 0x0, 0x0}, "1"},
			{[]byte{0xff, 0xff, 0xff, 0x7f}, "2147483647"},
		}
		for _, tt := range tests {
			got, err := parseBSONInt32(tt.b)
			if err != nil {
				t.Errorf("parseBSONInt32(%v) failed: %v", tt.b, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("parseBSONInt32(%v) = %q, want %q", tt.b, got, want)
			}

		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]byte{
			"length 1": {},
			"length 2": {0x0, 0x0, 0x0},
			"length 3": {0x0, 0x0, 0x0, 0x0, 0x0},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := parseBSONInt32(tt)
				if err == nil {
					t.Errorf("parseBSONInt32(%v) did not fail", tt)
				}
			})
		}
	})
}

func TestParseBSONInt64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			b    []byte
			want string
		}{
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "-9223372036854775808"},
			{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "-1"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "1"},
			{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, "9223372036854775807"},
		}
		for _, tt := range tests {
			got, err := parseBSONInt64(tt.b)
			if err != nil {
				t.Errorf("parseBSONInt64(%v) failed: %v", tt.b, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("parseBSONInt64(%v) = %q, want %q", tt.b, got, want)
			}

		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]byte{
			"length 1": {},
			"length 2": {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			"length 3": {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := parseBSONInt64(tt)
				if err == nil {
					t.Errorf("parseBSONInt64(%v) did not fail", tt)
				}
			})
		}
	})
}

func TestParseIEEEDecimal128(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			b    []byte
			want string
		}{
			// Zeros
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7a, 0x2b}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2a, 0x30}, "0.00000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x30}, "0.0000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2e, 0x30}, "0.000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0x30}, "0.00000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0x30}, "0.0000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "0.000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x36, 0x30}, "0.00000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0x30}, "0.0000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "0.000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "0.00"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}, "0.0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x42, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x44, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x46, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x48, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4a, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4e, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x50, 0x30}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "0"},

			// Negative zeroes
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2e, 0xb0}, "0.000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0xb0}, "0.00000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0xb0}, "0.0000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0xb0}, "0.000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x36, 0xb0}, "0.00000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0xb0}, "0.0000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0xb0}, "0.000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0xb0}, "0.00"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0xb0}, "0.0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x46, 0xb0}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0xb0}, "0"},

			// Overflows with 0 coefficient
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x5f}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f}, "0"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0xdf}, "0"},

			// Underflows
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0a, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x0a, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x81, 0xef, 0xac, 0x85, 0x5b, 0x41, 0x6d, 0x2d, 0xee, 0x04, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x81, 0xef, 0xac, 0x85, 0x5b, 0x41, 0x6d, 0x2d, 0xee, 0x04, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x81, 0xef, 0xac, 0x85, 0x5b, 0x41, 0x6d, 0x2d, 0xee, 0x04, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x81, 0xef, 0xac, 0x85, 0x5b, 0x41, 0x6d, 0x2d, 0xee, 0x04, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0x20, 0x3b, 0x9d, 0xb5, 0x05, 0x6f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x24}, "0.0000000000000000000"},
			{[]byte{0x0, 0x0, 0xfe, 0xd8, 0x3f, 0x4e, 0x7c, 0x9f, 0xe4, 0xe2, 0x69, 0xe3, 0x8a, 0x5b, 0xcd, 0x17}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x02, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x02, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x02, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x02, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x42, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x42, 0x80}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x72, 0x28}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0a, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x01, 0x0, 0x0, 0x0, 0x0a, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x0a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x0a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0x0a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x0a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0x3c, 0x17, 0x25, 0x84, 0x19, 0xd7, 0x10, 0xc4, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x24}, "0.0000000000000000000"},
			{[]byte{0xff, 0xff, 0xff, 0xff, 0x09, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0x0, 0x0}, "0.0000000000000000000"},
			{[]byte{0xff, 0xff, 0xff, 0xff, 0x09, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0x0, 0x80}, "0.0000000000000000000"},
			{[]byte{0xff, 0xff, 0xff, 0xff, 0x63, 0x8e, 0x8d, 0x37, 0xc0, 0x87, 0xad, 0xbe, 0x09, 0xed, 0x01, 0x0}, "0.0000000000000000000"},

			// Powers of 10
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-1"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0xb0}, "-1.0"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0xb0}, "-0.1"},
			{[]byte{0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0xb0}, "-0.0000000100"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x26, 0x30}, "0.0000000000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x28, 0x30}, "0.000000000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2a, 0x30}, "0.00000000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x30}, "0.0000000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2e, 0x30}, "0.000000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0x30}, "0.00000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0x30}, "0.0000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "0.000010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x36, 0x30}, "0.00010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0x30}, "0.0010"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "0.010"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}, "0.1"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "0.10"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0xa, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0xfc, 0x2f}, "0.1000000000000000000"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "1"},
			{[]byte{0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "1.00"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}, "1.0"},
			{[]byte{0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}, "10.0"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "10"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x42, 0x30}, "100"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x46, 0x30}, "1000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x44, 0x30}, "1000"},
			{[]byte{0xe8, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "1000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x46, 0x30}, "10000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x48, 0x30}, "100000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4a, 0x30}, "1000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x30}, "10000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4e, 0x30}, "100000000"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "1000000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x50, 0x30}, "1000000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "10000000000"},
			{[]byte{0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "100000000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x54, 0x30}, "100000000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x56, 0x30}, "1000000000000"},
			{[]byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x30}, "10000000000000"},

			// Integers
			{[]byte{0xee, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-750"},
			{[]byte{0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-123"},
			{[]byte{0x4c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-76"},
			{[]byte{0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-12"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-1"},
			{[]byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "2"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "7"},
			{[]byte{0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "9"},
			{[]byte{0xc, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "12"},
			{[]byte{0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "17"},
			{[]byte{0x13, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "19"},
			{[]byte{0x14, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "20"},
			{[]byte{0x1d, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "29"},
			{[]byte{0x1e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "30"},
			{[]byte{0x27, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "39"},
			{[]byte{0x28, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "40"},
			{[]byte{0x2c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "44"},
			{[]byte{0x31, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "49"},
			{[]byte{0x32, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "50"},
			{[]byte{0x3b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "59"},
			{[]byte{0x3c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "60"},
			{[]byte{0x45, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "69"},
			{[]byte{0x46, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "70"},
			{[]byte{0x47, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "71"},
			{[]byte{0x48, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "72"},
			{[]byte{0x49, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "73"},
			{[]byte{0x4a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "74"},
			{[]byte{0x4b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "75"},
			{[]byte{0x4c, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "76"},
			{[]byte{0x4d, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "77"},
			{[]byte{0x4e, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "78"},
			{[]byte{0x4f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "79"},
			{[]byte{0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "123"},
			{[]byte{0x8, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "520"},
			{[]byte{0x9, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "521"},
			{[]byte{0x9, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "777"},
			{[]byte{0xa, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "778"},
			{[]byte{0x13, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "787"},
			{[]byte{0x1f, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "799"},
			{[]byte{0x6d, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "877"},
			{[]byte{0x78, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "888"},
			{[]byte{0x79, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "889"},
			{[]byte{0x82, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "898"},
			{[]byte{0x83, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "899"},
			{[]byte{0xd3, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "979"},
			{[]byte{0xdc, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "988"},
			{[]byte{0xdd, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "989"},
			{[]byte{0xe2, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "994"},
			{[]byte{0xe3, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "995"},
			{[]byte{0xe5, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "997"},
			{[]byte{0xe6, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "998"},
			{[]byte{0xe7, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "999"},
			{[]byte{0x30, 0x75, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "30000"},
			{[]byte{0x90, 0x94, 0xd, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "890000"},

			{[]byte{0xfe, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "4294967294"},
			{[]byte{0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "4294967295"},
			{[]byte{0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "4294967296"},
			{[]byte{0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "4294967297"},

			{[]byte{0x1, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-2147483649"},
			{[]byte{0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-2147483648"},
			{[]byte{0xff, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-2147483647"},
			{[]byte{0xfe, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-2147483646"},
			{[]byte{0xfe, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "2147483646"},
			{[]byte{0xff, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "2147483647"},
			{[]byte{0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "2147483648"},
			{[]byte{0x1, 0x0, 0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "2147483649"},

			// 1265 multiplied by powers of 10
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10, 0x30}, "0.0000000000000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0x30}, "0.0000000000000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14, 0x30}, "0.0000000000000000001"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x16, 0x30}, "0.0000000000000000013"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x18, 0x30}, "0.0000000000000000126"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x28, 0x30}, "0.000000001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2a, 0x30}, "0.00000001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x30}, "0.0000001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2e, 0x30}, "0.000001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0x30}, "0.00001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0x30}, "0.0001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "0.001265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x36, 0x30}, "0.01265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0x30}, "0.1265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "1.265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "12.65"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}, "126.5"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "1265"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x42, 0x30}, "12650"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x44, 0x30}, "126500"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x46, 0x30}, "1265000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x48, 0x30}, "12650000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4a, 0x30}, "126500000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x30}, "1265000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4e, 0x30}, "12650000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x50, 0x30}, "126500000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "1265000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x54, 0x30}, "12650000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x56, 0x30}, "126500000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x30}, "1265000000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5a, 0x30}, "12650000000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5c, 0x30}, "126500000000000000"},
			{[]byte{0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5e, 0x30}, "1265000000000000000"},

			// 7 multiplied by powers of 10
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x26, 0x30}, "0.0000000000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x28, 0x30}, "0.000000000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2a, 0x30}, "0.00000000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2c, 0x30}, "0.0000000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2e, 0x30}, "0.000000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0x30}, "0.00000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x32, 0x30}, "0.0000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "0.000007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x36, 0x30}, "0.00007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0x30}, "0.0007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "0.007"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "0.07"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}, "0.7"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "7"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x42, 0x30}, "70"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x44, 0x30}, "700"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x46, 0x30}, "7000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x48, 0x30}, "70000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4a, 0x30}, "700000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4c, 0x30}, "7000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x4e, 0x30}, "70000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x50, 0x30}, "700000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "7000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x52, 0x30}, "7000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x54, 0x30}, "70000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x56, 0x30}, "700000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x58, 0x30}, "7000000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5a, 0x30}, "70000000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5c, 0x30}, "700000000000000"},
			{[]byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5e, 0x30}, "7000000000000000"},

			// Sequences of digits
			{[]byte{0x18, 0x5c, 0xa, 0xce, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0xb0}, "-345678.5432"},
			{[]byte{0x39, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-12345"},
			{[]byte{0xd2, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}, "-1234"},
			{[]byte{0x39, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0xb0}, "-123.45"},
			{[]byte{0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0xb0}, "-1.23"},
			{[]byte{0x15, 0xcd, 0x5b, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x30}, "0.0000000123456789"},
			{[]byte{0x15, 0xcd, 0x5b, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x22, 0x30}, "0.000000123456789"},
			{[]byte{0xf2, 0xaf, 0x96, 0x7e, 0xd0, 0x5c, 0x82, 0xde, 0x32, 0x97, 0xff, 0x6f, 0xde, 0x3c, 0xf0, 0x2f}, "0.0000001234567890123"},
			{[]byte{0x15, 0xcd, 0x5b, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x24, 0x30}, "0.00000123456789"},
			{[]byte{0x15, 0xcd, 0x5b, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x26, 0x30}, "0.0000123456789"},
			{[]byte{0x40, 0xef, 0x5a, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2a, 0x30}, "0.00123400000"},
			{[]byte{0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "0.123"},
			{[]byte{0x78, 0xdf, 0xd, 0x86, 0x48, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x22, 0x30}, "0.123456789012344"},
			{[]byte{0x79, 0xdf, 0xd, 0x86, 0x48, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x22, 0x30}, "0.123456789012345"},
			{[]byte{0xf2, 0xaf, 0x96, 0x7e, 0xd0, 0x5c, 0x82, 0xde, 0x32, 0x97, 0xff, 0x6f, 0xde, 0x3c, 0xfc, 0x2f}, "0.1234567890123456789"},
			{[]byte{0x7b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "1.23"},
			{[]byte{0xd2, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}, "1.234"},
			{[]byte{0x39, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}, "123.45"},
			{[]byte{0xd2, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "1234"},
			{[]byte{0x39, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "12345"},
			{[]byte{0x18, 0x5c, 0xa, 0xce, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x38, 0x30}, "345678.5432"},
			{[]byte{0x6a, 0xf9, 0xb, 0x7c, 0x50, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "345678.543210"},
			{[]byte{0xf1, 0x98, 0x67, 0xc, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x36, 0x30}, "345678.54321"},
			{[]byte{0x6a, 0x19, 0x56, 0x25, 0x22, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "2345678.543210"},
			{[]byte{0x6a, 0xb9, 0xc8, 0x73, 0x3a, 0xb, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "12345678.543210"},
			{[]byte{0x40, 0xaf, 0xd, 0x86, 0x48, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "123456789.000000"},
			{[]byte{0x80, 0x91, 0xf, 0x86, 0x48, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x34, 0x30}, "123456789.123456"},
			{[]byte{0x80, 0x91, 0xf, 0x86, 0x48, 0x70, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}, "123456789123456"},
		}

		for _, tt := range tests {
			got, err := parseIEEEDecimal128(tt.b)
			if err != nil {
				t.Errorf("parseIEEEDecimal128([% x]) failed: %v", tt.b, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("parseIEEEDecimal128([% x]) = %q, want %q", tt.b, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]byte{
			"length 1":    {},
			"length 2":    {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			"length 3":    {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			"inf 1":       {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x78},
			"inf 2":       {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf8},
			"nan 1":       {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7c},
			"nan 2":       {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfc},
			"nan 3":       {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7e},
			"nan 4":       {0x12, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x7e},
			"nan 5":       {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe},
			"overflow 1":  {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x41, 0x30},
			"overflow 2":  {0x0, 0x0, 0x0, 0x0, 0xa, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0xcc, 0x37},
			"overflow 3":  {0x0, 0x0, 0x0, 0x0, 0xa, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0xfe, 0x5f},
			"overflow 4":  {0x0, 0x0, 0x0, 0x0, 0xa, 0x5b, 0xc1, 0x38, 0x93, 0x8d, 0x44, 0xc6, 0x4d, 0x31, 0xfe, 0xdf},
			"overflow 5":  {0x0, 0x0, 0x0, 0x0, 0x81, 0xef, 0xac, 0x85, 0x5b, 0x41, 0x6d, 0x2d, 0xee, 0x4, 0xfe, 0x5f},
			"overflow 6":  {0x0, 0x0, 0x0, 0x10, 0x61, 0x2, 0x25, 0x3e, 0x5e, 0xce, 0x4f, 0x20, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 7":  {0x0, 0x0, 0x0, 0x40, 0xea, 0xed, 0x74, 0x46, 0xd0, 0x9c, 0x2c, 0x9f, 0xc, 0x0, 0xfe, 0x5f},
			"overflow 8":  {0x0, 0x0, 0x0, 0x4a, 0x48, 0x1, 0x14, 0x16, 0x95, 0x45, 0x8, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 9":  {0x0, 0x0, 0x0, 0x80, 0x26, 0x4b, 0x91, 0xc0, 0x22, 0x20, 0xbe, 0x37, 0x7e, 0x0, 0xfe, 0x5f},
			"overflow 10": {0x0, 0x0, 0x0, 0x80, 0x7f, 0x1b, 0xcf, 0x85, 0xb2, 0x70, 0x59, 0xc8, 0xa4, 0x3c, 0xfe, 0x5f},
			"overflow 11": {0x0, 0x0, 0x0, 0x80, 0x7f, 0x1b, 0xcf, 0x85, 0xb2, 0x70, 0x59, 0xc8, 0xa4, 0x3c, 0xfe, 0xdf},
			"overflow 12": {0x0, 0x0, 0x0, 0xa0, 0xca, 0x17, 0x72, 0x6d, 0xae, 0xf, 0x1e, 0x43, 0x1, 0x0, 0xfe, 0x5f},
			"overflow 13": {0x0, 0x0, 0x0, 0xa1, 0xed, 0xcc, 0xce, 0x1b, 0xc2, 0xd3, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 14": {0x0, 0x0, 0x0, 0xe4, 0xd2, 0xc, 0xc8, 0xdc, 0xd2, 0xb7, 0x52, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 15": {0x0, 0x0, 0x0, 0xe8, 0x3c, 0x80, 0xd0, 0x9f, 0x3c, 0x2e, 0x3b, 0x3, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 16": {0x0, 0x0, 0x10, 0x63, 0x2d, 0x5e, 0xc7, 0x6b, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 17": {0x0, 0x0, 0x40, 0xb2, 0xba, 0xc9, 0xe0, 0x19, 0x1e, 0x2, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 18": {0x0, 0x0, 0x64, 0xa7, 0xb3, 0xb6, 0xe0, 0xd, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 19": {0x0, 0x0, 0x80, 0xf6, 0x4a, 0xe1, 0xc7, 0x2, 0x2d, 0x15, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 20": {0x0, 0x0, 0x8a, 0x5d, 0x78, 0x45, 0x63, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 21": {0x0, 0x0, 0xa0, 0xde, 0xc5, 0xad, 0xc9, 0x35, 0x36, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 22": {0x0, 0x0, 0xc1, 0x6f, 0xf2, 0x86, 0x23, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 23": {0x0, 0x0, 0xe8, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 24": {0x0, 0x10, 0xa5, 0xd4, 0xe8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 25": {0x0, 0x40, 0x7a, 0x10, 0xf3, 0x5a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 26": {0x0, 0x80, 0xc6, 0xa4, 0x7e, 0x8d, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 27": {0x0, 0xa0, 0x72, 0x4e, 0x18, 0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 28": {0x0, 0xca, 0x9a, 0x3b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 29": {0x0, 0xe1, 0xf5, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 30": {0x0, 0xe4, 0xb, 0x54, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 31": {0x0, 0xe8, 0x76, 0x48, 0x17, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 32": {0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf4, 0x30},
			"overflow 33": {0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfc, 0x5f},
			"overflow 34": {0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 35": {0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x6, 0x31},
			"overflow 36": {0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe, 0x38},
			"overflow 37": {0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1e, 0x5f},
			"overflow 38": {0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf4, 0x30},
			"overflow 39": {0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 40": {0x10, 0x27, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 41": {0x40, 0x42, 0xf, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 42": {0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf4, 0x30},
			"overflow 43": {0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 44": {0x79, 0xd9, 0xe0, 0xf9, 0x76, 0x3a, 0xda, 0x42, 0x9d, 0x2, 0x0, 0x0, 0x0, 0x0, 0x58, 0x30},
			"overflow 45": {0x80, 0x96, 0x98, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 46": {0xa0, 0x86, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 47": {0xc7, 0x71, 0x1c, 0xc7, 0xb5, 0x48, 0xf3, 0x77, 0xdc, 0x80, 0xa1, 0x31, 0xc8, 0x36, 0x40, 0x30},
			"overflow 48": {0xe8, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xfe, 0x5f},
			"overflow 49": {0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x60, 0x30},
			"overflow 50": {0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x62, 0x30},
			"overflow 51": {0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x64, 0x30},
			"overflow 52": {0xf1, 0x4, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x68, 0x30},
			"overflow 53": {0xf2, 0xaf, 0x96, 0x7e, 0xd0, 0x5c, 0x82, 0xde, 0x32, 0x97, 0xff, 0x6f, 0xde, 0x3c, 0x40, 0x30},
			"overflow 54": {0xf2, 0xaf, 0x96, 0x7e, 0xd0, 0x5c, 0x82, 0xde, 0x32, 0x97, 0xff, 0x6f, 0xde, 0x3c, 0x40, 0xb0},
			"overflow 55": {0xf2, 0xaf, 0x96, 0x7e, 0xd0, 0x5c, 0x82, 0xde, 0x32, 0x97, 0xff, 0x6f, 0xde, 0x3c, 0xfe, 0x5f},
			"overflow 56": {0xf2, 0xaf, 0x96, 0x7e, 0xd0, 0x5c, 0x82, 0xde, 0x32, 0x97, 0xff, 0x6f, 0xde, 0x3c, 0xfe, 0xdf},
			"overflow 57": {0xff, 0xff, 0xff, 0xff, 0x63, 0x8e, 0x8d, 0x37, 0xc0, 0x87, 0xad, 0xbe, 0x9, 0xed, 0x41, 0x30},
			"overflow 58": {0xff, 0xff, 0xff, 0xff, 0x63, 0x8e, 0x8d, 0x37, 0xc0, 0x87, 0xad, 0xbe, 0x9, 0xed, 0xff, 0x5f},
			"overflow 59": {0xff, 0xff, 0xff, 0xff, 0x63, 0x8e, 0x8d, 0x37, 0xc0, 0x87, 0xad, 0xbe, 0x9, 0xed, 0xff, 0xdf},
			"overflow 60": {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x40, 0x30},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				_, err := parseIEEEDecimal128(tt)
				if err == nil {
					t.Errorf("parseIEEEDecimal128([% x]) did not fail", tt)
				}
			})
		}
	})
}

func TestDecimal_IEEEDecimal128(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d    string
			want []byte
		}{
			{"-9999999999999999999", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}},
			{"-999999999999999999.9", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0xb0}},
			{"-99999999999999999.99", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0xb0}},
			{"-9999999999999999.999", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0xb0}},
			{"-0.9999999999999999999", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1a, 0xb0}},
			{"-1", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}},
			{"-0.1", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0xb0}},
			{"-0.01", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0xb0}},
			{"-0.0000000000000000001", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1a, 0xb0}},
			{"0", []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{"0.0", []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}},
			{"0.00", []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}},
			{"0.0000000000000000000", []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1a, 0x30}},
			{"0.0000000000000000001", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1a, 0x30}},
			{"0.01", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}},
			{"0.1", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}},
			{"1", []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{"0.9999999999999999999", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1a, 0x30}},
			{"9999999999999999.999", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3a, 0x30}},
			{"99999999999999999.99", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0x30}},
			{"999999999999999999.9", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3e, 0x30}},
			{"9999999999999999999", []byte{0xff, 0xff, 0xe7, 0x89, 0x4, 0x23, 0xc7, 0x8a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},

			// Exported constants
			{NegOne.String(), []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0xb0}},
			{Zero.String(), []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{One.String(), []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{Two.String(), []byte{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{Ten.String(), []byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{Hundred.String(), []byte{0x64, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{Thousand.String(), []byte{0xe8, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x30}},
			{E.String(), []byte{0x73, 0x61, 0xb3, 0xc0, 0xeb, 0x46, 0xb9, 0x25, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1c, 0x30}},
			{Pi.String(), []byte{0xd6, 0x49, 0x32, 0xa2, 0xdf, 0x2d, 0x99, 0x2b, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1c, 0x30}},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got := d.ieeeDecimal128()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("%q.ieeeDecimal128() = [% x], want [% x]", d, got, tt.want)
			}

		}
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
			{false, 1, 19, "0.0000000000000000001"},
			{false, 1, 2, "0.01"},
			{false, 1, 1, "0.1"},
			{false, 1, 0, "1"},
			{false, maxCoef, 19, "0.9999999999999999999"},
			{false, maxCoef, 3, "9999999999999999.999"},
			{false, maxCoef, 2, "99999999999999999.99"},
			{false, maxCoef, 1, "999999999999999999.9"},
			{false, maxCoef, 0, "9999999999999999999"},

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

func TestDecimal_Float64(t *testing.T) {
	tests := []struct {
		d         string
		wantFloat float64
		wantOk    bool
	}{
		{"-9999999999999999999", -9999999999999999999, true},
		{"-1000000000000000000", -1000000000000000000, true},
		{"-1", -1, true},
		{"-0.9999999999999999999", -0.9999999999999999999, true},
		{"-0.0000000000000000001", -0.0000000000000000001, true},
		{"0.0000000000000000001", 0.0000000000000000001, true},
		{"0.9999999999999999999", 0.9999999999999999999, true},
		{"1", 1, true},
		{"1000000000000000000", 1000000000000000000, true},
		{"9999999999999999999", 9999999999999999999, true},
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
		{"0.00", 2, 0, 0, true},
		{"0.0", 1, 0, 0, true},
		{"0", 0, 0, 0, true},

		// Trailing zeros
		{"0.1000", 4, 0, 1000, true},
		{"0.100", 4, 0, 1000, true},
		{"0.10", 4, 0, 1000, true},
		{"0.1", 4, 0, 1000, true},

		{"0.1000", 4, 0, 1000, true},
		{"0.100", 3, 0, 100, true},
		{"0.10", 2, 0, 10, true},
		{"0.1", 1, 0, 1, true},

		// Powers of ten
		{"0.0001", 4, 0, 1, true},
		{"0.0001", 4, 0, 1, true},
		{"0.001", 4, 0, 10, true},
		{"0.001", 3, 0, 1, true},
		{"0.01", 4, 0, 100, true},
		{"0.01", 2, 0, 1, true},
		{"0.1", 4, 0, 1000, true},
		{"0.1", 1, 0, 1, true},
		{"1", 4, 1, 0, true},
		{"1", 0, 1, 0, true},
		{"10", 4, 10, 0, true},
		{"10", 0, 10, 0, true},
		{"100", 4, 100, 0, true},
		{"100", 0, 100, 0, true},
		{"1000", 4, 1000, 0, true},
		{"1000", 0, 1000, 0, true},

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
		{"-9223372036854775808", 0, -9223372036854775808, 0, true},
		{"-922337203685477580.9", 1, -922337203685477580, -9, true},
		{"-9.223372036854775809", 18, -9, -223372036854775809, true},
		{"-0.9223372036854775808", 19, 0, -9223372036854775808, true},
		{"0.9223372036854775807", 19, 0, 9223372036854775807, true},
		{"9.223372036854775808", 18, 9, 223372036854775808, true},
		{"922337203685477580.8", 1, 922337203685477580, 8, true},
		{"9223372036854775807", 0, 9223372036854775807, 0, true},

		// Failures
		{"-9999999999999999999", 0, 0, 0, false},
		{"-9223372036854775809", 0, 0, 0, false},
		{"-0.9999999999999999999", 19, 0, 0, false},
		{"-0.9223372036854775809", 19, 0, 0, false},
		{"0.1", -1, 0, 0, false},
		{"0.1", 20, 0, 0, false},
		{"0.9223372036854775808", 19, 0, 0, false},
		{"0.9999999999999999999", 19, 0, 0, false},
		{"9223372036854775808", 0, 0, 0, false},
		{"9999999999999999999", 0, 0, 0, false},
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
			{math.SmallestNonzeroFloat64, "0.0000000000000000000"},
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
				t.Errorf("Scan(%v) failed: %v", tt.f, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Scan(%v) = %v, want %v", tt.f, got, want)
			}
		}
	})

	t.Run("float32", func(t *testing.T) {
		tests := []struct {
			f    float32
			want string
		}{
			{math.SmallestNonzeroFloat32, "0.0000000000000000000"},
			{1e-20, "0.0000000000000000000"},
			{1e-19, "0.0000000000000000001"},
			{1e-5, "0.0000099999997473788"},
			{1e-4, "0.0000999999974737875"},
			{1e-3, "0.0010000000474974513"},
			{1e-2, "0.009999999776482582"},
			{1e-1, "0.10000000149011612"},
			{1e0, "1"},
			{1e1, "10"},
			{1e2, "100"},
			{1e3, "1000"},
			{1e4, "10000"},
			{1e5, "100000"},
			{1e18, "999999984306749400"},
		}
		for _, tt := range tests {
			got := Decimal{}
			err := got.Scan(tt.f)
			if err != nil {
				t.Errorf("Scan(%v) failed: %v", tt.f, err)
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

	t.Run("uint64", func(t *testing.T) {
		tests := []struct {
			i    uint64
			want string
		}{
			{0, "0"},
			{9999999999999999999, "9999999999999999999"},
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
		{"9999999999999999999", "%k", "%!k(PANIC=Format method: formatting percent: computing [9999999999999999999 * 100]: decimal overflow: the integer part of a decimal.Decimal can have at most 19 digits, but it has 21 digits)"},
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

func TestSum(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d    []string
			want string
		}{
			{[]string{"0"}, "0"},
			{[]string{"1"}, "1"},
			{[]string{"1", "1"}, "2"},
			{[]string{"2", "3"}, "5"},
			{[]string{"5.75", "3.3"}, "9.05"},
			{[]string{"5", "-3"}, "2"},
			{[]string{"-5", "-3"}, "-8"},
			{[]string{"-7", "2.5"}, "-4.5"},
			{[]string{"0.7", "0.3"}, "1.0"},
			{[]string{"1.25", "1.25"}, "2.50"},
			{[]string{"1.1", "0.11"}, "1.21"},
			{[]string{"1.234567890", "1.000000000"}, "2.234567890"},
			{[]string{"1.234567890", "1.000000110"}, "2.234568000"},

			{[]string{"0.9998", "0.0000"}, "0.9998"},
			{[]string{"0.9998", "0.0001"}, "0.9999"},
			{[]string{"0.9998", "0.0002"}, "1.0000"},
			{[]string{"0.9998", "0.0003"}, "1.0001"},

			{[]string{"999999999999999999", "1"}, "1000000000000000000"},
			{[]string{"99999999999999999", "1"}, "100000000000000000"},
			{[]string{"9999999999999999", "1"}, "10000000000000000"},
			{[]string{"999999999999999", "1"}, "1000000000000000"},
			{[]string{"99999999999999", "1"}, "100000000000000"},
			{[]string{"9999999999999", "1"}, "10000000000000"},
			{[]string{"999999999999", "1"}, "1000000000000"},
			{[]string{"99999999999", "1"}, "100000000000"},
			{[]string{"9999999999", "1"}, "10000000000"},
			{[]string{"999999999", "1"}, "1000000000"},
			{[]string{"99999999", "1"}, "100000000"},
			{[]string{"9999999", "1"}, "10000000"},
			{[]string{"999999", "1"}, "1000000"},
			{[]string{"99999", "1"}, "100000"},
			{[]string{"9999", "1"}, "10000"},
			{[]string{"999", "1"}, "1000"},
			{[]string{"99", "1"}, "100"},
			{[]string{"9", "1"}, "10"},

			{[]string{"100000000000", "0.00000000"}, "100000000000.0000000"},
			{[]string{"100000000000", "0.00000001"}, "100000000000.0000000"},

			{[]string{"0.0", "0"}, "0.0"},
			{[]string{"0.00", "0"}, "0.00"},
			{[]string{"0.000", "0"}, "0.000"},
			{[]string{"0.0000000", "0"}, "0.0000000"},
			{[]string{"0", "0.0"}, "0.0"},
			{[]string{"0", "0.00"}, "0.00"},
			{[]string{"0", "0.000"}, "0.000"},
			{[]string{"0", "0.0000000"}, "0.0000000"},

			{[]string{"9999999999999999999", "0.4"}, "9999999999999999999"},
			{[]string{"-9999999999999999999", "-0.4"}, "-9999999999999999999"},
			{[]string{"1", "-9999999999999999999"}, "-9999999999999999998"},
			{[]string{"9999999999999999999", "-1"}, "9999999999999999998"},
		}
		for _, tt := range tests {
			d := make([]Decimal, len(tt.d))
			for i, s := range tt.d {
				d[i] = MustParse(s)
			}
			got, err := Sum(d...)
			if err != nil {
				t.Errorf("Sum(%v) failed: %v", d, err)
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Sum(%v) = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]string{
			"no arguments": {},
			"overflow 1":   {"9999999999999999999", "1"},
			"overflow 2":   {"9999999999999999999", "0.6"},
			"overflow 3":   {"-9999999999999999999", "-1"},
			"overflow 4":   {"-9999999999999999999", "-0.6"},
		}
		for name, ss := range tests {
			t.Run(name, func(t *testing.T) {
				d := make([]Decimal, len(ss))
				for i, s := range ss {
					d[i] = MustParse(s)
				}
				_, err := Sum(d...)
				if err == nil {
					t.Errorf("Sum(%v) did not fail", d)
				}
			})
		}
	})
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

func TestProd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d    []string
			want string
		}{
			{[]string{"0"}, "0"},
			{[]string{"1"}, "1"},
			{[]string{"2", "2"}, "4"},
			{[]string{"2", "3"}, "6"},
			{[]string{"5", "1"}, "5"},
			{[]string{"5", "2"}, "10"},
			{[]string{"1.20", "2"}, "2.40"},
			{[]string{"1.20", "0"}, "0.00"},
			{[]string{"1.20", "-2"}, "-2.40"},
			{[]string{"-1.20", "2"}, "-2.40"},
			{[]string{"-1.20", "0"}, "0.00"},
			{[]string{"-1.20", "-2"}, "2.40"},
			{[]string{"5.09", "7.1"}, "36.139"},
			{[]string{"2.5", "4"}, "10.0"},
			{[]string{"2.50", "4"}, "10.00"},
			{[]string{"0.70", "1.05"}, "0.7350"},
			{[]string{"1.000000000", "1"}, "1.000000000"},
			{[]string{"1.23456789", "1.00000000"}, "1.2345678900000000"},
			{[]string{"1.000000000000000000", "1.000000000000000000"}, "1.000000000000000000"},
			{[]string{"1.000000000000000001", "1.000000000000000001"}, "1.000000000000000002"},
			{[]string{"9.999999999999999999", "9.999999999999999999"}, "99.99999999999999998"},
			{[]string{"0.0000000000000000001", "0.0000000000000000001"}, "0.0000000000000000000"},
			{[]string{"0.0000000000000000001", "0.9999999999999999999"}, "0.0000000000000000001"},
			{[]string{"0.0000000000000000003", "0.9999999999999999999"}, "0.0000000000000000003"},
			{[]string{"0.9999999999999999999", "0.9999999999999999999"}, "0.9999999999999999998"},
			{[]string{"0.9999999999999999999", "0.9999999999999999999", "0.9999999999999999999"}, "0.9999999999999999997"},
			{[]string{"6963.788300835654596", "0.001436"}, "10.00000000000000000"},

			// Captured during fuzzing
			{[]string{"92233720368547757.26", "0.0000000000000000002"}, "0.0184467440737095515"},
			{[]string{"9223372036854775.807", "-0.0000000000000000013"}, "-0.0119903836479112085"},
		}
		for _, tt := range tests {
			d := make([]Decimal, len(tt.d))
			for i, s := range tt.d {
				d[i] = MustParse(s)
			}
			got, err := Prod(d...)
			if err != nil {
				t.Errorf("Prod(%v) failed: %v", d, err)
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Prod(%v) = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]string{
			"no arguments": {},
			"overflow 1":   {"10000000000", "1000000000"},
			"overflow 2":   {"1000000000000000000", "10"},
			"overflow 3":   {"4999999999999999995", "-2.000000000000000002"},
			"overflow 4": {
				"9999999999999999999", "9999999999999999999",
				"9999999999999999999", "9999999999999999999",
				"9999999999999999999", "9999999999999999999",
				"9999999999999999999", "9999999999999999999",
			},
		}
		for name, ss := range tests {
			t.Run(name, func(t *testing.T) {
				d := make([]Decimal, len(ss))
				for i, s := range ss {
					d[i] = MustParse(s)
				}
				_, err := Prod(d...)
				if err == nil {
					t.Errorf("Prod(%v) did not fail", d)
				}
			})
		}
	})
}

func TestMean(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d    []string
			want string
		}{
			{[]string{"1"}, "1"},
			{[]string{"1", "1"}, "1"},
			{[]string{"2", "3"}, "2.5"},
			{[]string{"5.75", "3.3"}, "4.525"},
			{[]string{"5", "-3"}, "1"},
			{[]string{"-5", "-3"}, "-4"},
			{[]string{"-7", "2.5"}, "-2.25"},
			{[]string{"0.7", "0.3"}, "0.5"},
			{[]string{"1.25", "1.25"}, "1.25"},
			{[]string{"1.1", "0.11"}, "0.605"},
			{[]string{"1.234567890", "1.000000000"}, "1.117283945"},
			{[]string{"1.234567890", "1.000000110"}, "1.117284000"},

			{[]string{"0.9998", "0.0000"}, "0.4999"},
			{[]string{"0.9998", "0.0001"}, "0.49995"},
			{[]string{"0.9998", "0.0002"}, "0.5000"},
			{[]string{"0.9998", "0.0003"}, "0.50005"},

			{[]string{"999999999999999999", "1"}, "500000000000000000"},
			{[]string{"99999999999999999", "1"}, "50000000000000000"},
			{[]string{"9999999999999999", "1"}, "5000000000000000"},
			{[]string{"999999999999999", "1"}, "500000000000000"},
			{[]string{"99999999999999", "1"}, "50000000000000"},
			{[]string{"9999999999999", "1"}, "5000000000000"},
			{[]string{"999999999999", "1"}, "500000000000"},
			{[]string{"99999999999", "1"}, "50000000000"},
			{[]string{"9999999999", "1"}, "5000000000"},
			{[]string{"999999999", "1"}, "500000000"},
			{[]string{"99999999", "1"}, "50000000"},
			{[]string{"9999999", "1"}, "5000000"},
			{[]string{"999999", "1"}, "500000"},
			{[]string{"99999", "1"}, "50000"},
			{[]string{"9999", "1"}, "5000"},
			{[]string{"999", "1"}, "500"},
			{[]string{"99", "1"}, "50"},
			{[]string{"9", "1"}, "5"},

			{[]string{"100000000000", "0.00000000"}, "50000000000.00000000"},
			{[]string{"100000000000", "0.00000001"}, "50000000000.00000000"},

			{[]string{"0.0", "0"}, "0.0"},
			{[]string{"0.00", "0"}, "0.00"},
			{[]string{"0.000", "0"}, "0.000"},
			{[]string{"0.0000000", "0"}, "0.0000000"},
			{[]string{"0", "0.0"}, "0.0"},
			{[]string{"0", "0.00"}, "0.00"},
			{[]string{"0", "0.000"}, "0.000"},
			{[]string{"0", "0.0000000"}, "0.0000000"},

			{[]string{"9999999999999999999", "0.4"}, "5000000000000000000"},
			{[]string{"-9999999999999999999", "-0.4"}, "-5000000000000000000"},
			{[]string{"1", "-9999999999999999999"}, "-4999999999999999999"},
			{[]string{"9999999999999999999", "-1"}, "4999999999999999999"},

			// Smallest and largest numbers
			{[]string{"-0.0000000000000000001", "-0.0000000000000000001"}, "-0.0000000000000000001"},
			{[]string{"0.0000000000000000001", "0.0000000000000000001"}, "0.0000000000000000001"},
			{[]string{"-9999999999999999999", "-9999999999999999999"}, "-9999999999999999999"},
			{[]string{"9999999999999999999", "9999999999999999999"}, "9999999999999999999"},

			// Captured during fuzzing
			{[]string{"9223372036854775807", "9223372036854775807", "922337203685477580.7"}, "6456360425798343065"},
			{[]string{"922.3372036854775807", "2", "-3000000000"}, "-999999691.8875987715"},
			{[]string{"0.5", "0.3", "0.2"}, "0.3333333333333333333"},
		}
		for _, tt := range tests {
			d := make([]Decimal, len(tt.d))
			for i, s := range tt.d {
				d[i] = MustParse(s)
			}
			got, err := Mean(d...)
			if err != nil {
				t.Errorf("Mean(%v) failed: %v", d, err)
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("Mean(%v) = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string][]string{
			"no arguments": {},
		}
		for name, ss := range tests {
			t.Run(name, func(t *testing.T) {
				d := make([]Decimal, len(ss))
				for i, s := range ss {
					d[i] = MustParse(s)
				}
				_, err := Mean(d...)
				if err == nil {
					t.Errorf("Mean(%v) did not fail", d)
				}
			})
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

			// Captured during fuzzing
			{"92233720368547757.26", "0.0000000000000000002", "0.0184467440737095515"},
			{"9223372036854775.807", "-0.0000000000000000013", "-0.0119903836479112085"},
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

func TestDecimal_AddMul(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, f, want string
		}{
			// Signs
			{"4", "2", "3", "10"},
			{"-4", "2", "3", "2"},
			{"4", "2", "-3", "-2"},
			{"-4", "2", "-3", "-10"},
			{"4", "-2", "3", "-2"},
			{"-4", "-2", "3", "-10"},
			{"4", "-2", "-3", "10"},
			{"-4", "-2", "-3", "2"},

			// Addition tests
			{"1", "1", "1", "2"},
			{"3", "1", "2", "5"},
			{"3.3", "1", "5.75", "9.05"},
			{"-3", "1", "5", "2"},
			{"-3", "1", "-5", "-8"},
			{"2.5", "1", "-7", "-4.5"},
			{"0.3", "1", "0.7", "1.0"},
			{"1.25", "1", "1.25", "2.50"},
			{"0.11", "1", "1.1", "1.21"},
			{"1.000000000", "1", "1.234567890", "2.234567890"},
			{"1.000000110", "1", "1.234567890", "2.234568000"},
			{"0.0000", "1", "0.9998", "0.9998"},
			{"0.0001", "1", "0.9998", "0.9999"},
			{"0.0002", "1", "0.9998", "1.0000"},
			{"0.0003", "1", "0.9998", "1.0001"},
			{"1", "1", "999999999999999999", "1000000000000000000"},
			{"1", "1", "99999999999999999", "100000000000000000"},
			{"1", "1", "9999999999999999", "10000000000000000"},
			{"1", "1", "999999999999999", "1000000000000000"},
			{"1", "1", "99999999999999", "100000000000000"},
			{"1", "1", "9999999999999", "10000000000000"},
			{"1", "1", "999999999999", "1000000000000"},
			{"1", "1", "99999999999", "100000000000"},
			{"1", "1", "9999999999", "10000000000"},
			{"1", "1", "999999999", "1000000000"},
			{"1", "1", "99999999", "100000000"},
			{"1", "1", "9999999", "10000000"},
			{"1", "1", "999999", "1000000"},
			{"1", "1", "99999", "100000"},
			{"1", "1", "9999", "10000"},
			{"1", "1", "999", "1000"},
			{"1", "1", "99", "100"},
			{"1", "1", "9", "10"},
			{"0.00000000", "1", "100000000000", "100000000000.0000000"},
			{"0.00000001", "1", "100000000000", "100000000000.0000000"},
			{"0", "1", "0.0", "0.0"},
			{"0", "1", "0.00", "0.00"},
			{"0", "1", "0.000", "0.000"},
			{"0", "1", "0.0000000", "0.0000000"},
			{"0.0", "1", "0", "0.0"},
			{"0.00", "1", "0", "0.00"},
			{"0.000", "1", "0", "0.000"},
			{"0.0000000", "1", "0", "0.0000000"},
			{"0.4", "1", "9999999999999999999", "9999999999999999999"},
			{"-0.4", "1", "-9999999999999999999", "-9999999999999999999"},
			{"-9999999999999999999", "1", "1", "-9999999999999999998"},
			{"-1", "1", "9999999999999999999", "9999999999999999998"},

			// Multiplication tests
			{"0", "2", "2", "4"},
			{"0", "2", "3", "6"},
			{"0", "5", "1", "5"},
			{"0", "5", "2", "10"},
			{"0", "1.20", "2", "2.40"},
			{"0", "1.20", "0", "0.00"},
			{"0", "1.20", "-2", "-2.40"},
			{"0", "-1.20", "2", "-2.40"},
			{"0", "-1.20", "0", "0.00"},
			{"0", "-1.20", "-2", "2.40"},
			{"0", "5.09", "7.1", "36.139"},
			{"0", "2.5", "4", "10.0"},
			{"0", "2.50", "4", "10.00"},
			{"0", "0.70", "1.05", "0.7350"},
			{"0", "1.000000000", "1", "1.000000000"},
			{"0", "1.23456789", "1.00000000", "1.2345678900000000"},
			{"0", "1.000000000000000000", "1.000000000000000000", "1.000000000000000000"},
			{"0", "1.000000000000000001", "1.000000000000000001", "1.000000000000000002"},
			{"0", "9.999999999999999999", "9.999999999999999999", "99.99999999999999998"},
			{"0", "0.0000000000000000001", "0.0000000000000000001", "0.0000000000000000000"},
			{"0", "0.0000000000000000001", "0.9999999999999999999", "0.0000000000000000001"},
			{"0", "0.0000000000000000003", "0.9999999999999999999", "0.0000000000000000003"},
			{"0", "0.9999999999999999999", "0.9999999999999999999", "0.9999999999999999998"},
			{"0", "6963.788300835654596", "0.001436", "10.00000000000000000"},

			// Captured during fuzzing
			{"0.0000000000000000121", "0.0000000000000000127", "12.5", "0.0000000000000001708"},
			{"-9.3", "0.0000000203", "-0.0000000116", "-9.300000000000000235"},
			{"5.8", "-0.0000000231", "0.0000000166", "5.799999999999999617"},

			// Tests from GDA
			{"2593183.42371", "27583489.6645", "2582471078.04", "71233564292579696.34"},
			{"2032.013252", "24280.355566", "939577.397653", "22813275328.80506589"},
			{"137903.517909", "7848976432", "-2586831.2281", "-20303977342780612.62"},
			{"339337.123410", "56890.388731", "35872030.4255", "2040774094814.077745"},
			{"5073392.31638", "7533543.57445", "360317763928", "2714469575205049785"},
			{"894450638.442", "437484.00601", "598906432790", "262011986336578659.5"},
			{"153127.446727", "203258304486", "-8628278.8066", "-1753769320861850379"},
			{"178277.96377", "42560533.1774", "-3643605282.86", "-155073783526334663.6"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			f := MustParse(tt.f)
			got, err := d.AddMul(e, f)
			if err != nil {
				t.Errorf("%q.AddMul(%q, %q) failed: %v", d, e, f, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.AddMul(%q, %q) = %q, want %q", d, e, f, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, f, e string
			scale   int
		}{
			"overflow 1": {"1", "1", "9999999999999999999", 0},
			"overflow 2": {"0.6", "1", "9999999999999999999", 0},
			"overflow 3": {"-1", "1", "-9999999999999999999", 0},
			"overflow 4": {"-0.6", "1", "-9999999999999999999", 0},
			"overflow 5": {"0", "10000000000", "1000000000", 0},
			"overflow 6": {"0", "1000000000000000000", "10", 0},
			"scale 1":    {"1", "1", "1", MaxScale},
			"scale 2":    {"0", "0", "0", MaxScale + 1},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			f := MustParse(tt.f)
			_, err := d.AddMulExact(e, f, tt.scale)
			if err == nil {
				t.Errorf("%q.AddMulExact(%q, %q, %v) did not fail", d, e, f, tt.scale)
			}
		}
	})
}

func TestDecimal_AddQuo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, f, want string
		}{
			// Signs
			{"3", "4", "2", "5"},
			{"3", "-4", "2", "1"},
			{"-3", "4", "2", "-1"},
			{"-3", "-4", "2", "-5"},
			{"3", "4", "-2", "1"},
			{"3", "-4", "-2", "5"},
			{"-3", "4", "-2", "-5"},
			{"-3", "-4", "-2", "-1"},

			// Addition tests
			{"1", "1", "1", "2"},
			{"3", "2", "1", "5"},
			{"3.3", "5.75", "1", "9.05"},
			{"-3", "5", "1", "2"},
			{"-3", "-5", "1", "-8"},
			{"2.5", "-7", "1", "-4.5"},
			{"0.3", "0.7", "1", "1.0"},
			{"1.25", "1.25", "1", "2.50"},
			{"0.11", "1.1", "1", "1.21"},
			{"1.000000000", "1.234567890", "1", "2.234567890"},
			{"1.000000110", "1.234567890", "1", "2.234568000"},
			{"0.0000", "0.9998", "1", "0.9998"},
			{"0.0001", "0.9998", "1", "0.9999"},
			{"0.0002", "0.9998", "1", "1.0000"},
			{"0.0003", "0.9998", "1", "1.0001"},
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
			{"0.00000000", "100000000000", "1", "100000000000.0000000"},
			{"0.00000001", "100000000000", "1", "100000000000.0000000"},
			{"0", "0.0", "1", "0.0"},
			{"0", "0.00", "1", "0.00"},
			{"0", "0.000", "1", "0.000"},
			{"0", "0.0000000", "1", "0.0000000"},
			{"0.0", "0", "1", "0.0"},
			{"0.00", "0", "1", "0.00"},
			{"0.000", "0", "1", "0.000"},
			{"0.0000000", "0", "1", "0.0000000"},
			{"0.4", "9999999999999999999", "1", "9999999999999999999"},
			{"-0.4", "-9999999999999999999", "1", "-9999999999999999999"},
			{"-9999999999999999999", "1", "1", "-9999999999999999998"},
			{"-1", "9999999999999999999", "1", "9999999999999999998"},

			// Division tests
			{"0", "9223372036854775807", "-9223372036854775808", "-0.9999999999999999999"},
			{"0", "0.000000000000000001", "20", "0.000000000000000000"},
			{"0", "105", "0.999999999999999990", "105.0000000000000011"},
			{"0", "0.05", "999999999999999954", "0.0000000000000000001"},
			{"0", "9.99999999999999998", "185", "0.0540540540540540539"},
			{"0", "7", "2.000000000000000002", "3.499999999999999997"},
			{"0", "0.000000009", "999999999999999999", "0.000000000"},
			{"0", "0.0000000000000000001", "9999999999999999999", "0.0000000000000000000"},
			{"0", "9999999999999999999", "2", "5000000000000000000"},
			{"0", "9999999999999999999", "5000000000000000000", "2"},

			// Captured during fuzzing
			{"47", "-126", "110", "45.85454545454545455"},
			{"-92", "94", "76", "-90.76315789473684211"},
			{"5", "-40", "139", "4.712230215827338129"},
			{"-3", "3", "0.9999999999999999999", "0.0000000000000000003"},
			{"-0.0000000000000000001", "1", "0.9999999999999999999", "1.000000000000000000"},
			{"0.00000000053", "4.3", "0.00000000071", "6056338028.169014085"},
			{"8.9", "0.0000000000082", "-0.000000110", "8.899925454545454545"},
			{"0.000000000000000", "0.9999999999999999940", "1", "0.9999999999999999940"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			f := MustParse(tt.f)
			got, err := d.AddQuo(e, f)
			if err != nil {
				t.Errorf("%q.AddQuo(%q, %q) failed: %v", d, e, f, err)
				continue
			}
			want := MustParse(tt.want)
			if got.CmpTotal(want) != 0 {
				t.Errorf("%q.AddQuo(%q, %q) = %q, want %q", d, e, f, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e, f string
			scale   int
		}{
			"overflow 1": {"9999999999999999999", "1", "1", 0},
			"overflow 2": {"9999999999999999999", "0.6", "1", 0},
			"overflow 3": {"-9999999999999999999", "-1", "1", 0},
			"overflow 4": {"-9999999999999999999", "-0.6", "1", 0},
			"overflow 5": {"0", "10000000000", "0.000000001", 0},
			"overflow 6": {"0", "1000000000000000000", "0.1", 0},
			"zero 1":     {"1", "1", "0", 0},
			"scale 1":    {"1", "1", "1", MaxScale},
			"scale 2":    {"0", "0", "1", MaxScale + 1},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			f := MustParse(tt.f)
			_, err := d.AddQuoExact(e, f, tt.scale)
			if err == nil {
				t.Errorf("%q.AddQuoExact(%q, %q, %v) did not fail", d, e, f, tt.scale)
			}
		}
	})
}

func TestDecimal_Pow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, e, want string
		}{
			//////////////////////
			// Fractional Powers
			//////////////////////

			{"0.0", "0.0", "1"},
			{"0.0", "0.5", "0"},

			{"4.0", "-0.5", "0.5000000000000000000"},
			{"4.0", "0.0", "1"},
			{"4.0", "0.5", "2.000000000000000000"},

			{"0.0000001", "0.0000001", "0.9999983881917338685"},
			{"0.003", "0.0000001", "0.9999994190858696993"},
			{"0.7", "0.0000001", "0.9999999643325062422"},
			{"1.2", "0.0000001", "1.000000018232155846"},
			{"71", "0.0000001", "1.000000426268078556"},
			{"9000000000", "0.0000001", "1.000002292051668175"},

			{"0.0000001", "0.003", "0.9527961640236518859"},
			{"0.003", "0.003", "0.9827235503366796915"},
			{"0.7", "0.003", "0.9989305474406207158"},
			{"1.2", "0.003", "1.000547114282833519"},
			{"71", "0.003", "1.012870156273545212"},
			{"9000000000", "0.003", "1.071180671278787089"},

			{"0.0000001", "0.7", "0.0000125892541179417"},
			{"0.003", "0.7", "0.0171389763028103005"},
			{"0.7", "0.7", "0.7790559126704490940"},
			{"1.2", "0.7", "1.136126977198888691"},
			{"71", "0.7", "19.76427300093869501"},
			{"9000000000", "0.7", "9289016.976853710315"},

			{"0.0000001", "1.2", "0.0000000039810717055"},
			{"0.003", "1.2", "0.0009387403933595694"},
			{"0.7", "1.2", "0.6518049405663863819"},
			{"1.2", "1.2", "1.244564747203977722"},
			{"71", "1.2", "166.5367244638552138"},
			{"9000000000", "1.2", "881233526124.8791107"},

			// Natural numbers
			{"10", "0.0000000000000000000", "1"},
			{"10", "0.3010299956639811952", "2.000000000000000000"},
			{"10", "0.4771212547196624373", "3.000000000000000000"},
			{"10", "0.6020599913279623904", "4.000000000000000000"},
			{"10", "0.6989700043360188048", "5.000000000000000000"},
			{"10", "0.7781512503836436325", "6.000000000000000000"},
			{"10", "0.8450980400142568307", "7.000000000000000000"},
			{"10", "0.9030899869919435856", "7.999999999999999999"},
			{"10", "0.9542425094393248746", "9.000000000000000000"},
			{"10", "1", "10"},
			{"10", "1.041392685158225041", "11.00000000000000001"},
			{"10", "1.079181246047624828", "12.00000000000000001"},
			{"10", "1.113943352306836769", "12.99999999999999999"},
			{"10", "1.146128035678238026", "14.00000000000000000"},
			{"10", "1.176091259055681242", "15.00000000000000000"},
			{"10", "1.204119982655924781", "16.00000000000000001"},
			{"10", "1.230448921378273929", "17.00000000000000002"},
			{"10", "1.255272505103306070", "18.00000000000000001"},
			{"10", "1.278753600952828962", "19.00000000000000002"},
			{"10", "1.301029995663981195", "19.99999999999999999"},

			// e^e
			{E.String(), E.String(), "15.15426224147926418"},

			// Closer and closer to e
			{"10", "0.4342944819032518275", "2.718281828459045234"},
			{"10", "0.4342944819032518276", "2.718281828459045235"},
			{"10", "0.4342944819032518277", "2.718281828459045236"},
			{"10", "0.4342944819032518278", "2.718281828459045236"},
			{"10", "0.4342944819032518279", "2.718281828459045237"},

			// Closer and closer to 1000
			{"10", "2.999999999999999998", "999.9999999999999954"},
			{"10", "2.999999999999999999", "999.9999999999999977"},
			{"10", "3", "1000"},
			{"10", "3.000000000000000001", "1000.000000000000002"},
			{"10", "3.000000000000000002", "1000.000000000000005"},

			// Captured during fuzzing
			{"1.000000000000000001", "9223372036854775807", "10131.16947077036074"},
			{"1.000000000000000001", "922337203685477580.7", "2.515161971551883079"},
			{"1.000000000000000001", "92233720368547758.07", "1.096621094818009214"},
			{"999999999.999999999", "-999999999.999999999", "0.0000000000000000000"},
			{"0.060", "0.999999999999999944", "0.0600000000000000095"},
			{"0.0000000091", "0.23", "0.0141442368531557249"},

			/////////////////
			// Square Roots
			/////////////////

			// Zeros
			{"0.00000000", "0.5", "0"},
			{"0.0000000", "0.5", "0"},
			{"0.000000", "0.5", "0"},
			{"0.00000", "0.5", "0"},
			{"0.0000", "0.5", "0"},
			{"0.000", "0.5", "0"},
			{"0.00", "0.5", "0"},
			{"0.0", "0.5", "0"},
			{"0", "0.5", "0"},

			// Trailing zeros
			{"0.010000000", "0.5", "0.1000000000000000000"},
			{"0.01000000", "0.5", "0.1000000000000000000"},
			{"0.0100000", "0.5", "0.1000000000000000000"},
			{"0.010000", "0.5", "0.1000000000000000000"},
			{"0.01000", "0.5", "0.1000000000000000000"},
			{"0.0100", "0.5", "0.1000000000000000000"},
			{"0.010", "0.5", "0.1000000000000000000"},
			{"0.01", "0.5", "0.1000000000000000000"},

			// Powers of ten
			{"0.00000001", "0.5", "0.0001000000000000000"},
			{"0.0000001", "0.5", "0.0003162277660168379"},
			{"0.000001", "0.5", "0.0010000000000000000"},
			{"0.00001", "0.5", "0.0031622776601683793"},
			{"0.0001", "0.5", "0.0100000000000000000"},
			{"0.001", "0.5", "0.0316227766016837933"},
			{"0.01", "0.5", "0.1000000000000000000"},
			{"0.1", "0.5", "0.3162277660168379332"},
			{"1", "0.5", "1.00000000000000000000"},
			{"10", "0.5", "3.162277660168379332"},
			{"100", "0.5", "10.00000000000000000"},
			{"1000", "0.5", "31.62277660168379332"},
			{"10000", "0.5", "100.0000000000000000"},
			{"100000", "0.5", "316.2277660168379332"},
			{"1000000", "0.5", "1000.000000000000000"},
			{"10000000", "0.5", "3162.277660168379332"},
			{"100000000", "0.5", "10000.00000000000000"},

			// Natural numbers
			{"0", "0.5", "0"},
			{"1", "0.5", "1.000000000000000000"},
			{"2", "0.5", "1.414213562373095049"},
			{"3", "0.5", "1.732050807568877294"},
			{"4", "0.5", "2.000000000000000000"},
			{"5", "0.5", "2.236067977499789696"},
			{"6", "0.5", "2.449489742783178098"},
			{"7", "0.5", "2.645751311064590591"},
			{"8", "0.5", "2.828427124746190098"},
			{"9", "0.5", "3.000000000000000000"},
			{"10", "0.5", "3.162277660168379332"},
			{"11", "0.5", "3.316624790355399849"},
			{"12", "0.5", "3.464101615137754587"},
			{"13", "0.5", "3.605551275463989293"},
			{"14", "0.5", "3.741657386773941386"},
			{"15", "0.5", "3.872983346207416885"},
			{"16", "0.5", "4.000000000000000000"},
			{"17", "0.5", "4.123105625617660550"},
			{"18", "0.5", "4.242640687119285146"},
			{"19", "0.5", "4.358898943540673552"},
			{"20", "0.5", "4.472135954999579393"},
			{"21", "0.5", "4.582575694955840007"},
			{"22", "0.5", "4.690415759823429555"},
			{"23", "0.5", "4.795831523312719542"},
			{"24", "0.5", "4.898979485566356196"},
			{"25", "0.5", "5.000000000000000000"},

			// Well-known squares
			{"1", "0.5", "1.000000000000000000"},
			{"4", "0.5", "2.000000000000000000"},
			{"9", "0.5", "3.000000000000000000"},
			{"16", "0.5", "4.000000000000000000"},
			{"25", "0.5", "5.000000000000000000"},
			{"36", "0.5", "6.000000000000000000"},
			{"49", "0.5", "7.000000000000000000"},
			{"64", "0.5", "8.000000000000000000"},
			{"81", "0.5", "9.000000000000000000"},
			{"100", "0.5", "10.00000000000000000"},
			{"121", "0.5", "11.00000000000000000"},
			{"144", "0.5", "12.00000000000000000"},
			{"169", "0.5", "13.00000000000000000"},
			{"256", "0.5", "16.00000000000000000"},
			{"1024", "0.5", "32.00000000000000000"},
			{"4096", "0.5", "64.00000000000000000"},

			{"0.01", "0.5", "0.1000000000000000000"},
			{"0.04", "0.5", "0.2000000000000000000"},
			{"0.09", "0.5", "0.3000000000000000000"},
			{"0.16", "0.5", "0.4000000000000000000"},
			{"0.25", "0.5", "0.5000000000000000000"},
			{"0.36", "0.5", "0.6000000000000000000"},
			{"0.49", "0.5", "0.7000000000000000000"},
			{"0.64", "0.5", "0.8000000000000000000"},
			{"0.81", "0.5", "0.9000000000000000000"},
			{"1.00", "0.5", "1.000000000000000000"},
			{"1.21", "0.5", "1.100000000000000000"},
			{"1.44", "0.5", "1.200000000000000000"},
			{"1.69", "0.5", "1.300000000000000000"},
			{"2.56", "0.5", "1.600000000000000000"},
			{"10.24", "0.5", "3.200000000000000000"},
			{"40.96", "0.5", "6.400000000000000000"},

			// Smallest and largest numbers
			{"0.0000000000000000001", "0.5", "0.0000000003162277660"},
			{"9999999999999999999", "0.5", "3162277660.168379332"},

			// Captured during fuzzing
			{"1.000000000000000063", "0.5", "1.000000000000000031"},
			{"0.000000272", "0.5", "0.0005215361924162119"},
			{"0.9999999999999999999", "0.5", "0.9999999999999999999"},

			///////////////////
			// Integer Powers
			///////////////////

			// Zeros
			{"0", "0", "1"},
			{"0", "1", "0"},
			{"0", "2", "0"},

			// Ones
			{"-1", "-2", "1"},
			{"-1", "-1", "-1"},
			{"-1", "0", "1"},
			{"-1", "1", "-1"},
			{"-1", "2", "1"},

			// One tenths
			{"0.1", "-18", "1000000000000000000"},
			{"0.1", "-10", "10000000000"},
			{"0.1", "-9", "1000000000"},
			{"0.1", "-8", "100000000"},
			{"0.1", "-7", "10000000"},
			{"0.1", "-6", "1000000"},
			{"0.1", "-5", "100000"},
			{"0.1", "-4", "10000"},
			{"0.1", "-3", "1000"},
			{"0.1", "-2", "100"},
			{"0.1", "-1", "10"},
			{"0.1", "0", "1"},
			{"0.1", "1", "0.1"},
			{"0.1", "2", "0.01"},
			{"0.1", "3", "0.001"},
			{"0.1", "4", "0.0001"},
			{"0.1", "5", "0.00001"},
			{"0.1", "6", "0.000001"},
			{"0.1", "7", "0.0000001"},
			{"0.1", "8", "0.00000001"},
			{"0.1", "9", "0.000000001"},
			{"0.1", "10", "0.0000000001"},
			{"0.1", "18", "0.000000000000000001"},
			{"0.1", "19", "0.0000000000000000001"},
			{"0.1", "20", "0.0000000000000000000"},
			{"0.1", "40", "0.0000000000000000000"},

			// Negative one tenths
			{"-0.1", "-18", "1000000000000000000"},
			{"-0.1", "-10", "10000000000"},
			{"-0.1", "-9", "-1000000000"},
			{"-0.1", "-8", "100000000"},
			{"-0.1", "-7", "-10000000"},
			{"-0.1", "-6", "1000000"},
			{"-0.1", "-5", "-100000"},
			{"-0.1", "-4", "10000"},
			{"-0.1", "-3", "-1000"},
			{"-0.1", "-2", "100"},
			{"-0.1", "-1", "-10"},
			{"-0.1", "0", "1"},
			{"-0.1", "1", "-0.1"},
			{"-0.1", "2", "0.01"},
			{"-0.1", "3", "-0.001"},
			{"-0.1", "4", "0.0001"},
			{"-0.1", "5", "-0.00001"},
			{"-0.1", "6", "0.000001"},
			{"-0.1", "7", "-0.0000001"},
			{"-0.1", "8", "0.00000001"},
			{"-0.1", "9", "-0.000000001"},
			{"-0.1", "10", "0.0000000001"},
			{"-0.1", "18", "0.000000000000000001"},
			{"-0.1", "19", "-0.0000000000000000001"},
			{"-0.1", "20", "0.0000000000000000000"},
			{"-0.1", "40", "0.0000000000000000000"},

			// Twos
			{"2", "-64", "0.0000000000000000001"},
			{"2", "-63", "0.0000000000000000001"},
			{"2", "-32", "0.0000000002328306437"},
			{"2", "-16", "0.0000152587890625"},
			{"2", "-9", "0.001953125"},
			{"2", "-8", "0.00390625"},
			{"2", "-7", "0.0078125"},
			{"2", "-6", "0.015625"},
			{"2", "-5", "0.03125"},
			{"2", "-4", "0.0625"},
			{"2", "-3", "0.125"},
			{"2", "-2", "0.25"},
			{"2", "-1", "0.5"},
			{"2", "0", "1"},
			{"2", "1", "2"},
			{"2", "2", "4"},
			{"2", "3", "8"},
			{"2", "4", "16"},
			{"2", "5", "32"},
			{"2", "6", "64"},
			{"2", "7", "128"},
			{"2", "8", "256"},
			{"2", "9", "512"},
			{"2", "16", "65536"},
			{"2", "32", "4294967296"},
			{"2", "63", "9223372036854775808"},

			// Negative twos
			{"-2", "-64", "0.0000000000000000001"},
			{"-2", "-63", "-0.0000000000000000001"},
			{"-2", "-32", "0.0000000002328306437"},
			{"-2", "-16", "0.0000152587890625"},
			{"-2", "-9", "-0.001953125"},
			{"-2", "-8", "0.00390625"},
			{"-2", "-7", "-0.0078125"},
			{"-2", "-6", "0.015625"},
			{"-2", "-5", "-0.03125"},
			{"-2", "-4", "0.0625"},
			{"-2", "-3", "-0.125"},
			{"-2", "-2", "0.25"},
			{"-2", "-1", "-0.5"},
			{"-2", "0", "1"},
			{"-2", "1", "-2"},
			{"-2", "2", "4"},
			{"-2", "3", "-8"},
			{"-2", "4", "16"},
			{"-2", "5", "-32"},
			{"-2", "6", "64"},
			{"-2", "7", "-128"},
			{"-2", "8", "256"},
			{"-2", "9", "-512"},
			{"-2", "16", "65536"},
			{"-2", "32", "4294967296"},
			{"-2", "63", "-9223372036854775808"},

			// Squares
			{"-3", "2", "9"},
			{"-2", "2", "4"},
			{"-1", "2", "1"},
			{"0", "2", "0"},
			{"1", "2", "1"},
			{"2", "2", "4"},
			{"3", "2", "9"},
			{"4", "2", "16"},
			{"5", "2", "25"},
			{"6", "2", "36"},
			{"7", "2", "49"},
			{"8", "2", "64"},
			{"9", "2", "81"},
			{"10", "2", "100"},
			{"11", "2", "121"},
			{"12", "2", "144"},
			{"13", "2", "169"},
			{"14", "2", "196"},

			{"-0.3", "2", "0.09"},
			{"-0.2", "2", "0.04"},
			{"-0.1", "2", "0.01"},
			{"0.0", "2", "0.00"},
			{"0.1", "2", "0.01"},
			{"0.2", "2", "0.04"},
			{"0.3", "2", "0.09"},
			{"0.4", "2", "0.16"},
			{"0.5", "2", "0.25"},
			{"0.6", "2", "0.36"},
			{"0.7", "2", "0.49"},
			{"0.8", "2", "0.64"},
			{"0.9", "2", "0.81"},
			{"1.0", "2", "1.00"},
			{"1.1", "2", "1.21"},
			{"1.2", "2", "1.44"},
			{"1.3", "2", "1.69"},
			{"1.4", "2", "1.96"},

			{"0.000000000316227766", "2", "0.0000000000000000001"},
			{"3162277660.168379331", "2", "9999999999999999994"},

			// Cubes
			{"-3", "3", "-27"},
			{"-2", "3", "-8"},
			{"-1", "3", "-1"},
			{"0", "3", "0"},
			{"1", "3", "1"},
			{"2", "3", "8"},
			{"3", "3", "27"},
			{"4", "3", "64"},
			{"5", "3", "125"},
			{"6", "3", "216"},
			{"7", "3", "343"},
			{"8", "3", "512"},
			{"9", "3", "729"},
			{"10", "3", "1000"},
			{"11", "3", "1331"},
			{"12", "3", "1728"},
			{"13", "3", "2197"},
			{"14", "3", "2744"},

			{"-0.3", "3", "-0.027"},
			{"-0.2", "3", "-0.008"},
			{"-0.1", "3", "-0.001"},
			{"0.0", "3", "0.000"},
			{"0.1", "3", "0.001"},
			{"0.2", "3", "0.008"},
			{"0.3", "3", "0.027"},
			{"0.4", "3", "0.064"},
			{"0.5", "3", "0.125"},
			{"0.6", "3", "0.216"},
			{"0.7", "3", "0.343"},
			{"0.8", "3", "0.512"},
			{"0.9", "3", "0.729"},
			{"1.0", "3", "1.000"},
			{"1.1", "3", "1.331"},
			{"1.2", "3", "1.728"},
			{"1.3", "3", "2.197"},
			{"1.4", "3", "2.744"},

			{"0.000000464158883361", "3", "0.0000000000000000001"},
			{"2154434.690031883721", "3", "9999999999999999989"},

			// Interest accrual
			{"1.1", "60", "304.4816395414180996"},         // no error
			{"1.01", "600", "391.5833969993197743"},       // no error
			{"1.001", "6000", "402.2211245663552923"},     // no error
			{"1.0001", "60000", "403.3077910727185433"},   // no error
			{"1.00001", "600000", "403.4166908911542153"}, // no error

			// Captured during fuzzing
			{"0.85", "-267", "7000786514887173012"},
			{"0.066", "-16", "7714309010612096020"},
			{"-0.9223372036854775808", "-128", "31197.15320234751783"},
			{"999999999.999999999", "-9223372036854775808", "0"},
			{"-0.9223372036854775807", "-541", "-9877744411719625497"},
			{"0.9223372036854775702", "-540", "9110611159425388150"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			e := MustParse(tt.e)
			got, err := d.Pow(e)
			if err != nil {
				t.Errorf("%q.Pow(%q) failed: %v", d, e, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Pow(%q) = %q, want %q", d, e, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d, e string
		}{
			"overflow 1": {"2", "64"},
			"overflow 2": {"0.5", "-64"},
			"overflow 3": {"10", "19"},
			"overflow 4": {"0.1", "-19"},
			"overflow 5": {"0.0000000000000000001", "-3"},
			"overflow 6": {"999999999.999999999", "999999999.999999999"},
			"zero 1":     {"0", "-1"},
			"negative 1": {"-1", "0.1"},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(tt.d)
				e := MustParse(tt.e)
				_, err := d.Pow(e)
				if err == nil {
					t.Errorf("%q.Pow(%d) did not fail", d, e)
				}
			})
		}
	})
}

func TestDecimal_PowInt(t *testing.T) {
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

			{"-0.3", 2, "0.09"},
			{"-0.2", 2, "0.04"},
			{"-0.1", 2, "0.01"},
			{"0.0", 2, "0.00"},
			{"0.1", 2, "0.01"},
			{"0.2", 2, "0.04"},
			{"0.3", 2, "0.09"},
			{"0.4", 2, "0.16"},
			{"0.5", 2, "0.25"},
			{"0.6", 2, "0.36"},
			{"0.7", 2, "0.49"},
			{"0.8", 2, "0.64"},
			{"0.9", 2, "0.81"},
			{"1.0", 2, "1.00"},
			{"1.1", 2, "1.21"},
			{"1.2", 2, "1.44"},
			{"1.3", 2, "1.69"},
			{"1.4", 2, "1.96"},

			{"0.000000000316227766", 2, "0.0000000000000000001"},
			{"3162277660.168379331", 2, "9999999999999999994"},

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

			{"-0.3", 3, "-0.027"},
			{"-0.2", 3, "-0.008"},
			{"-0.1", 3, "-0.001"},
			{"0.0", 3, "0.000"},
			{"0.1", 3, "0.001"},
			{"0.2", 3, "0.008"},
			{"0.3", 3, "0.027"},
			{"0.4", 3, "0.064"},
			{"0.5", 3, "0.125"},
			{"0.6", 3, "0.216"},
			{"0.7", 3, "0.343"},
			{"0.8", 3, "0.512"},
			{"0.9", 3, "0.729"},
			{"1.0", 3, "1.000"},
			{"1.1", 3, "1.331"},
			{"1.2", 3, "1.728"},
			{"1.3", 3, "2.197"},
			{"1.4", 3, "2.744"},

			{"0.000000464158883361", 3, "0.0000000000000000001"},
			{"2154434.690031883721", 3, "9999999999999999989"},

			// Interest accrual
			{"1.1", 60, "304.4816395414180996"},         // no error
			{"1.01", 600, "391.5833969993197743"},       // no error
			{"1.001", 6000, "402.2211245663552923"},     // no error
			{"1.0001", 60000, "403.3077910727185433"},   // no error
			{"1.00001", 600000, "403.4166908911542153"}, // no error

			// Captured during fuzzing
			{"-0.9223372036854775807", -541, "-9877744411719625497"},
			{"999999999.999999999", math.MinInt, "0"},
			{"-0.9223372036854775808", -128, "31197.15320234751783"},
			{"0.85", -267, "7000786514887173012"},
			{"0.066", -16, "7714309010612096020"},
			{"0.9223372036854775702", -540, "9110611159425388150"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.PowInt(tt.power)
			if err != nil {
				t.Errorf("%q.PowInt(%d) failed: %v", d, tt.power, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.PowInt(%d) = %q, want %q", d, tt.power, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]struct {
			d     string
			power int
		}{
			"overflow 1": {"2", 64},
			"overflow 2": {"0.5", -64},
			"overflow 3": {"10", 19},
			"overflow 4": {"0.1", -19},
			"overflow 5": {"0.0000000000000000001", -3},
			"overflow 6": {"999999999.999999999", math.MaxInt},
			"zero 1":     {"0", -1},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(tt.d)
				_, err := d.PowInt(tt.power)
				if err == nil {
					t.Errorf("%q.PowInt(%d) did not fail", d, tt.power)
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
			{"0.00000000", "0.0000"},
			{"0.0000000", "0.000"},
			{"0.000000", "0.000"},
			{"0.00000", "0.00"},
			{"0.0000", "0.00"},
			{"0.000", "0.0"},
			{"0.00", "0.0"},
			{"0.0", "0"},
			{"0", "0"},

			// Trailing zeros
			{"0.010000000", "0.1000"},
			{"0.01000000", "0.1000"},
			{"0.0100000", "0.100"},
			{"0.010000", "0.100"},
			{"0.01000", "0.10"},
			{"0.0100", "0.10"},
			{"0.010", "0.1"},
			{"0.01", "0.1"},

			// Powers of ten
			{"0.00000001", "0.0001"},
			{"0.0000001", "0.0003162277660168379"},
			{"0.000001", "0.001"},
			{"0.00001", "0.0031622776601683793"},
			{"0.0001", "0.01"},
			{"0.001", "0.0316227766016837933"},
			{"0.01", "0.1"},
			{"0.1", "0.3162277660168379332"},
			{"1", "1"},
			{"10", "3.162277660168379332"},
			{"100", "10"},
			{"1000", "31.62277660168379332"},
			{"10000", "100"},
			{"100000", "316.2277660168379332"},
			{"1000000", "1000"},
			{"10000000", "3162.277660168379332"},
			{"100000000", "10000"},

			// Natural numbers
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

			// Well-known squares
			{"1", "1"},
			{"4", "2"},
			{"9", "3"},
			{"16", "4"},
			{"25", "5"},
			{"36", "6"},
			{"49", "7"},
			{"64", "8"},
			{"81", "9"},
			{"100", "10"},
			{"121", "11"},
			{"144", "12"},
			{"169", "13"},
			{"256", "16"},
			{"1024", "32"},
			{"4096", "64"},

			{"0.01", "0.1"},
			{"0.04", "0.2"},
			{"0.09", "0.3"},
			{"0.16", "0.4"},
			{"0.25", "0.5"},
			{"0.36", "0.6"},
			{"0.49", "0.7"},
			{"0.64", "0.8"},
			{"0.81", "0.9"},
			{"1.00", "1.0"},
			{"1.21", "1.1"},
			{"1.44", "1.2"},
			{"1.69", "1.3"},
			{"2.56", "1.6"},
			{"10.24", "3.2"},
			{"40.96", "6.4"},

			// Smallest and largest numbers
			{"0.0000000000000000001", "0.000000000316227766"},
			{"9999999999999999999", "3162277660.168379332"},

			// Captured during fuzzing
			{"1.000000000000000063", "1.000000000000000031"},
			{"0.000000272", "0.0005215361924162119"},
			{"0.9999999999999999999", "0.9999999999999999999"},
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

func TestDecimal_Exp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Zeros
			{"0", "1"},
			{"0.0", "1"},
			{"0.00", "1"},
			{"0.000", "1"},
			{"0.0000", "1"},
			{"0.00000", "1"},

			// Ones
			{"1", E.String()},
			{"1.0", E.String()},
			{"1.00", E.String()},
			{"1.000", E.String()},
			{"1.0000", E.String()},
			{"1.00000", E.String()},

			// Closer and closer to negative one
			{"-0.9", "0.4065696597405991119"},
			{"-0.99", "0.3715766910220456905"},
			{"-0.999", "0.3682475046136629212"},
			{"-0.9999", "0.3679162309550179865"},
			{"-0.99999", "0.3678831199842480694"},
			{"-0.999999", "0.3678798090510674328"},
			{"-0.9999999", "0.3678794779593882781"},
			{"-0.99999999", "0.3678794448502367517"},
			{"-0.999999999", "0.3678794415393217630"},
			{"-0.9999999999", "0.3678794412082302657"},
			{"-0.99999999999", "0.3678794411751211160"},
			{"-0.999999999999", "0.3678794411718102010"},
			{"-0.9999999999999", "0.3678794411714791095"},
			{"-0.99999999999999", "0.3678794411714460004"},
			{"-0.999999999999999", "0.3678794411714426895"},
			{"-0.9999999999999999", "0.3678794411714423584"},
			{"-0.99999999999999999", "0.3678794411714423253"},
			{"-0.999999999999999999", "0.3678794411714423220"},
			{"-1", "0.3678794411714423216"},
			{"-1.000000000000000001", "0.3678794411714423212"},
			{"-1.00000000000000001", "0.3678794411714423179"},
			{"-1.0000000000000001", "0.3678794411714422848"},
			{"-1.000000000000001", "0.3678794411714419537"},
			{"-1.00000000000001", "0.3678794411714386428"},
			{"-1.0000000000001", "0.3678794411714055337"},
			{"-1.000000000001", "0.3678794411710744422"},
			{"-1.00000000001", "0.3678794411677635272"},
			{"-1.0000000001", "0.3678794411346543775"},
			{"-1.000000001", "0.3678794408035628806"},
			{"-1.00000001", "0.3678794374926479283"},
			{"-1.0000001", "0.3678794043835000438"},
			{"-1.000001", "0.3678790732921850898"},
			{"-1.00001", "0.3678757623954245179"},
			{"-1.0001", "0.3678426550666610715"},
			{"-1.001", "0.3675117456086935500"},
			{"-1.01", "0.3642189795715233198"},
			{"-1.1", "0.3328710836980795533"},

			// Closer and closer to zero
			{"-0.1", "0.9048374180359595732"},
			{"-0.01", "0.9900498337491680536"},
			{"-0.001", "0.9990004998333749917"},
			{"-0.0001", "0.9999000049998333375"},
			{"-0.00001", "0.9999900000499998333"},
			{"-0.000001", "0.9999990000004999998"},
			{"-0.0000001", "0.9999999000000050000"},
			{"-0.00000001", "0.9999999900000000500"},
			{"-0.000000001", "0.9999999990000000005"},
			{"-0.0000000001", "0.9999999999000000000"},
			{"-0.00000000001", "0.9999999999900000000"},
			{"-0.000000000001", "0.9999999999990000000"},
			{"-0.0000000000001", "0.9999999999999000000"},
			{"-0.00000000000001", "0.9999999999999900000"},
			{"-0.000000000000001", "0.9999999999999990000"},
			{"-0.0000000000000001", "0.9999999999999999000"},
			{"-0.00000000000000001", "0.9999999999999999900"},
			{"-0.000000000000000001", "0.9999999999999999990"},
			{"-0.0000000000000000001", "0.9999999999999999999"},
			{"0", "1"},
			{"0.0000000000000000001", "1.000000000000000000"},
			{"0.000000000000000001", "1.000000000000000001"},
			{"0.00000000000000001", "1.000000000000000010"},
			{"0.0000000000000001", "1.000000000000000100"},
			{"0.000000000000001", "1.000000000000001000"},
			{"0.00000000000001", "1.000000000000010000"},
			{"0.0000000000001", "1.000000000000100000"},
			{"0.000000000001", "1.000000000001000000"},
			{"0.00000000001", "1.000000000010000000"},
			{"0.0000000001", "1.000000000100000000"},
			{"0.000000001", "1.000000001000000001"},
			{"0.00000001", "1.000000010000000050"},
			{"0.0000001", "1.000000100000005000"},
			{"0.000001", "1.000001000000500000"},
			{"0.00001", "1.000010000050000167"},
			{"0.0001", "1.000100005000166671"},
			{"0.001", "1.001000500166708342"},
			{"0.01", "1.010050167084168058"},
			{"0.1", "1.105170918075647625"},

			// Closer and closer to one
			{"0.9", "2.459603111156949664"},
			{"0.99", "2.691234472349262289"},
			{"0.999", "2.715564905318566687"},
			{"0.9999", "2.718010013867155437"},
			{"0.99999", "2.718254645776674283"},
			{"0.999999", "2.718279110178575917"},
			{"0.9999999", "2.718281556630875981"},
			{"0.99999999", "2.718281801276227087"},
			{"0.999999999", "2.718281825740763408"},
			{"0.9999999999", "2.718281828187217053"},
			{"0.99999999999", "2.718281828431862417"},
			{"0.999999999999", "2.718281828456326954"},
			{"0.9999999999999", "2.718281828458773407"},
			{"0.99999999999999", "2.718281828459018053"},
			{"0.999999999999999", "2.718281828459042517"},
			{"0.9999999999999999", "2.718281828459044964"},
			{"0.99999999999999999", "2.718281828459045208"},
			{"0.999999999999999999", "2.718281828459045233"},
			{"0.9999999999999999999", "2.718281828459045235"},
			{"1", E.String()},
			{"1.000000000000000001", "2.718281828459045238"},
			{"1.00000000000000001", "2.718281828459045263"},
			{"1.0000000000000001", "2.718281828459045507"},
			{"1.000000000000001", "2.718281828459047954"},
			{"1.00000000000001", "2.718281828459072418"},
			{"1.0000000000001", "2.718281828459317064"},
			{"1.000000000001", "2.718281828461763517"},
			{"1.00000000001", "2.718281828486228054"},
			{"1.0000000001", "2.718281828730873418"},
			{"1.000000001", "2.718281831177327065"},
			{"1.00000001", "2.718281855641863656"},
			{"1.0000001", "2.718282100287241673"},
			{"1.000001", "2.718284546742232836"},
			{"1.00001", "2.718309011413244370"},
			{"1.0001", "2.718553670233753340"},
			{"1.001", "2.721001469881578766"},
			{"1.01", "2.745601015016916494"},
			{"1.1", "3.004166023946433112"},

			// Negated powers of ten
			{"-10000", "0.0000000000000000000"},
			{"-1000", "0.0000000000000000000"},
			{"-100", "0.0000000000000000000"},
			{"-10", "0.00004539992976248489"},
			{"-1", "0.3678794411714423216"},
			{"-0.1", "0.9048374180359595732"},
			{"-0.01", "0.9900498337491680536"},
			{"-0.001", "0.9990004998333749917"},
			{"-0.0001", "0.9999000049998333375"},
			{"-0.00001", "0.9999900000499998333"},
			{"-0.000001", "0.9999990000004999998"},
			{"-0.0000001", "0.9999999000000050000"},
			{"-0.00000001", "0.9999999900000000500"},
			{"-0.000000001", "0.9999999990000000005"},
			{"-0.0000000001", "0.9999999999000000000"},
			{"-0.00000000001", "0.9999999999900000000"},
			{"-0.000000000001", "0.9999999999990000000"},
			{"-0.0000000000001", "0.9999999999999000000"},
			{"-0.00000000000001", "0.9999999999999900000"},
			{"-0.000000000000001", "0.9999999999999990000"},
			{"-0.0000000000000001", "0.9999999999999999000"},
			{"-0.00000000000000001", "0.9999999999999999900"},
			{"-0.000000000000000001", "0.9999999999999999990"},
			{"-0.0000000000000000001", "0.9999999999999999999"},

			// Powers of ten
			{"0.0000000000000000001", "1.000000000000000000"},
			{"0.000000000000000001", "1.000000000000000001"},
			{"0.00000000000000001", "1.000000000000000010"},
			{"0.0000000000000001", "1.000000000000000100"},
			{"0.000000000000001", "1.000000000000001000"},
			{"0.00000000000001", "1.000000000000010000"},
			{"0.0000000000001", "1.000000000000100000"},
			{"0.000000000001", "1.000000000001000000"},
			{"0.00000000001", "1.000000000010000000"},
			{"0.0000000001", "1.000000000100000000"},
			{"0.000000001", "1.000000001000000001"},
			{"0.00000001", "1.000000010000000050"},
			{"0.0000001", "1.000000100000005000"},
			{"0.000001", "1.000001000000500000"},
			{"0.00001", "1.000010000050000167"},
			{"0.0001", "1.000100005000166671"},
			{"0.001", "1.001000500166708342"},
			{"0.01", "1.010050167084168058"},
			{"0.1", "1.105170918075647625"},
			{"1", E.String()},
			{"10", "22026.46579480671652"},

			// Logarithms of powers of ten
			{"-50.65687204586900505", "0.0000000000000000000"},
			{"-48.35428695287495936", "0.0000000000000000000"},
			{"-46.05170185988091368", "0.0000000000000000000"},
			{"-43.74911676688686799", "0.0000000000000000001"},
			{"-41.44653167389282231", "0.0000000000000000010"},
			{"-39.14394658089877663", "0.0000000000000000100"},
			{"-36.84136148790473094", "0.0000000000000001000"},
			{"-34.53877639491068526", "0.0000000000000010000"},
			{"-32.23619130191663958", "0.0000000000000100000"},
			{"-29.93360620892259389", "0.0000000000001000000"},
			{"-27.63102111592854821", "0.0000000000010000000"},
			{"-25.32843602293450252", "0.0000000000100000000"},
			{"-23.02585092994045684", "0.0000000001000000000"},
			{"-20.72326583694641116", "0.0000000010000000000"},
			{"-18.42068074395236547", "0.0000000100000000000"},
			{"-16.11809565095831979", "0.0000001000000000000"},
			{"-13.81551055796427410", "0.0000010000000000000"},
			{"-11.51292546497022842", "0.0000100000000000000"},
			{"-9.210340371976182736", "0.0001000000000000000"},
			{"-6.907755278982137052", "0.0010000000000000000"},
			{"-4.605170185988091368", "0.0100000000000000000"},
			{"-2.302585092994045684", "0.1000000000000000000"},
			{"0", "1"},
			{"2.302585092994045684", "10.00000000000000000"},
			{"4.605170185988091368", "100.0000000000000000"},
			{"6.907755278982137052", "999.9999999999999999"},
			{"9.210340371976182736", "9999.999999999999999"},
			{"11.51292546497022842", "99999.99999999999999"},
			{"13.81551055796427410", "999999.9999999999959"},
			{"16.11809565095831979", "10000000.00000000002"},
			{"18.42068074395236547", "99999999.99999999979"},
			{"20.72326583694641116", "1000000000.000000004"},
			{"23.02585092994045684", "9999999999.999999998"},
			{"25.32843602293450252", "99999999999.99999958"},
			{"27.63102111592854821", "1000000000000.000002"},
			{"29.93360620892259389", "9999999999999.999978"},
			{"32.23619130191663958", "100000000000000.0004"},
			{"34.53877639491068526", "999999999999999.9997"},
			{"36.84136148790473094", "9999999999999999.957"},
			{"39.14394658089877663", "100000000000000000.2"},
			{"41.44653167389282231", "999999999999999997.7"},
			{"43.74911676688686799", "9999999999999999937"},

			// Negative numbers
			{"-101", "0.0000000000000000000"},
			{"-100", "0.0000000000000000000"},
			{"-99", "0.0000000000000000000"},
			{"-50", "0.0000000000000000000"},
			{"-45", "0.0000000000000000000"},
			{"-44", "0.0000000000000000001"},
			{"-43", "0.0000000000000000002"},

			// Natural numbers
			{"1", E.String()},
			{"2", "7.389056098930650227"},
			{"3", "20.08553692318766774"},
			{"4", "54.59815003314423908"},
			{"5", "148.4131591025766034"},
			{"6", "403.4287934927351226"},
			{"7", "1096.633158428458599"},
			{"8", "2980.957987041728275"},
			{"9", "8103.083927575384008"},
			{"10", "22026.46579480671652"},
			{"11", "59874.14171519781846"},
			{"12", "162754.7914190039208"},
			{"13", "442413.3920089205033"},
			{"14", "1202604.284164776778"},
			{"15", "3269017.372472110639"},
			{"16", "8886110.520507872637"},
			{"17", "24154952.75357529821"},
			{"18", "65659969.13733051114"},
			{"19", "178482300.9631872608"},
			{"20", "485165195.4097902780"},
			{"21", "1318815734.483214697"},
			{"22", "3584912846.131591562"},
			{"23", "9744803446.248902600"},
			{"24", "26489122129.84347229"},
			{"25", "72004899337.38587252"},
			{"26", "195729609428.8387643"},
			{"27", "532048240601.7986167"},
			{"28", "1446257064291.475174"},
			{"29", "3931334297144.042074"},
			{"30", "10686474581524.46215"},
			{"31", "29048849665247.42523"},
			{"32", "78962960182680.69516"},
			{"33", "214643579785916.0646"},
			{"34", "583461742527454.8814"},
			{"35", "1586013452313430.728"},
			{"36", "4311231547115195.227"},
			{"37", "11719142372802611.31"},
			{"38", "31855931757113756.22"},
			{"39", "86593400423993746.95"},
			{"40", "235385266837019985.4"},
			{"41", "639843493530054949.2"},
			{"42", "1739274941520501047"},
			{"43", "4727839468229346561"},

			// Captured during fuzzing
			{"-2.999999999999999852", "0.0497870683678639503"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Exp()
			if err != nil {
				t.Errorf("%q.Exp() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Exp() = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]string{
			"overflow 1": "49",
			"overflow 2": "50",
		}
		for name, d := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(d)
				_, err := d.Exp()
				if err == nil {
					t.Errorf("%q.Exp() did not fail", d)
				}
			})
		}
	})
}

func TestDecimal_Expm1(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Zeros
			{"0", "0"},
			{"0.0", "0"},
			{"0.00", "0"},
			{"0.000", "0"},
			{"0.0000", "0"},
			{"0.00000", "0"},

			// Ones
			{"1", "1.718281828459045235"},
			{"1.0", "1.718281828459045235"},
			{"1.00", "1.718281828459045235"},
			{"1.000", "1.718281828459045235"},
			{"1.0000", "1.718281828459045235"},
			{"1.00000", "1.718281828459045235"},

			// Closer and closer to negative one
			{"-0.9", "-0.5934303402594008881"},
			{"-0.99", "-0.6284233089779543095"},
			{"-0.999", "-0.6317524953863370788"},
			{"-0.9999", "-0.6320837690449820135"},
			{"-0.99999", "-0.6321168800157519306"},
			{"-0.999999", "-0.6321201909489325672"},
			{"-0.9999999", "-0.6321205220406117219"},
			{"-0.99999999", "-0.6321205551497632483"},
			{"-0.999999999", "-0.6321205584606782370"},
			{"-0.9999999999", "-0.6321205587917697343"},
			{"-0.99999999999", "-0.6321205588248788840"},
			{"-0.999999999999", "-0.6321205588281897990"},
			{"-0.9999999999999", "-0.6321205588285208905"},
			{"-0.99999999999999", "-0.6321205588285539996"},
			{"-0.999999999999999", "-0.6321205588285573105"},
			{"-0.9999999999999999", "-0.6321205588285576416"},
			{"-0.99999999999999999", "-0.6321205588285576747"},
			{"-0.999999999999999999", "-0.6321205588285576780"},
			{"-1", "-0.6321205588285576784"},
			{"-1.000000000000000001", "-0.6321205588285576788"},
			{"-1.00000000000000001", "-0.6321205588285576821"},
			{"-1.0000000000000001", "-0.6321205588285577152"},
			{"-1.000000000000001", "-0.6321205588285580463"},
			{"-1.00000000000001", "-0.6321205588285613572"},
			{"-1.0000000000001", "-0.6321205588285944663"},
			{"-1.000000000001", "-0.6321205588289255578"},
			{"-1.00000000001", "-0.6321205588322364728"},
			{"-1.0000000001", "-0.6321205588653456225"},
			{"-1.000000001", "-0.6321205591964371194"},
			{"-1.00000001", "-0.6321205625073520717"},
			{"-1.0000001", "-0.6321205956164999562"},
			{"-1.000001", "-0.6321209267078149102"},
			{"-1.00001", "-0.6321242376045754821"},
			{"-1.0001", "-0.6321573449333389285"},
			{"-1.001", "-0.6324882543913064500"},
			{"-1.01", "-0.6357810204284766802"},
			{"-1.1", "-0.6671289163019204467"},

			// Closer and closer to zero
			{"-0.1", "-0.0951625819640404268"},
			{"-0.01", "-0.0099501662508319464"},
			{"-0.001", "-0.0009995001666250083"},
			{"-0.0001", "-0.0000999950001666625"},
			{"-0.00001", "-0.0000099999500001667"},
			{"-0.000001", "-0.0000009999995000002"},
			{"-0.0000001", "-0.0000000999999950000"},
			{"-0.00000001", "-0.0000000099999999500"},
			{"-0.000000001", "-0.0000000009999999995"},
			{"-0.0000000001", "-0.0000000001000000000"},
			{"-0.00000000001", "-0.0000000000100000000"},
			{"-0.000000000001", "-0.0000000000010000000"},
			{"-0.0000000000001", "-0.0000000000001000000"},
			{"-0.00000000000001", "-0.0000000000000100000"},
			{"-0.000000000000001", "-0.0000000000000010000"},
			{"-0.0000000000000001", "-0.0000000000000001000"},
			{"-0.00000000000000001", "-0.0000000000000000100"},
			{"-0.000000000000000001", "-0.0000000000000000010"},
			{"-0.0000000000000000001", "-0.0000000000000000001"},
			{"0", "0"},
			{"0.0000000000000000001", "0.0000000000000000001"},
			{"0.000000000000000001", "0.0000000000000000010"},
			{"0.00000000000000001", "0.0000000000000000100"},
			{"0.0000000000000001", "0.0000000000000001000"},
			{"0.000000000000001", "0.0000000000000010000"},
			{"0.00000000000001", "0.0000000000000100000"},
			{"0.0000000000001", "0.0000000000001000000"},
			{"0.000000000001", "0.0000000000010000000"},
			{"0.00000000001", "0.0000000000100000000"},
			{"0.0000000001", "0.0000000001000000000"},
			{"0.000000001", "0.0000000010000000005"},
			{"0.00000001", "0.0000000100000000500"},
			{"0.0000001", "0.0000001000000050000"},
			{"0.000001", "0.0000010000005000002"},
			{"0.00001", "0.0000100000500001667"},
			{"0.0001", "0.0001000050001666708"},
			{"0.001", "0.0010005001667083417"},
			{"0.01", "0.0100501670841680575"},
			{"0.1", "0.1051709180756476248"},

			// Closer and closer to one
			{"0.9", "1.459603111156949664"},
			{"0.99", "1.691234472349262289"},
			{"0.999", "1.715564905318566687"},
			{"0.9999", "1.718010013867155437"},
			{"0.99999", "1.718254645776674283"},
			{"0.999999", "1.718279110178575917"},
			{"0.9999999", "1.718281556630875981"},
			{"0.99999999", "1.718281801276227087"},
			{"0.999999999", "1.718281825740763408"},
			{"0.9999999999", "1.718281828187217053"},
			{"0.99999999999", "1.718281828431862417"},
			{"0.999999999999", "1.718281828456326954"},
			{"0.9999999999999", "1.718281828458773407"},
			{"0.99999999999999", "1.718281828459018053"},
			{"0.999999999999999", "1.718281828459042517"},
			{"0.9999999999999999", "1.718281828459044964"},
			{"0.99999999999999999", "1.718281828459045208"},
			{"0.999999999999999999", "1.718281828459045233"},
			{"0.9999999999999999999", "1.718281828459045235"},
			{"1", "1.718281828459045235"},
			{"1.000000000000000001", "1.718281828459045238"},
			{"1.00000000000000001", "1.718281828459045263"},
			{"1.0000000000000001", "1.718281828459045507"},
			{"1.000000000000001", "1.718281828459047954"},
			{"1.00000000000001", "1.718281828459072418"},
			{"1.0000000000001", "1.718281828459317064"},
			{"1.000000000001", "1.718281828461763517"},
			{"1.00000000001", "1.718281828486228054"},
			{"1.0000000001", "1.718281828730873418"},
			{"1.000000001", "1.718281831177327065"},
			{"1.00000001", "1.718281855641863656"},
			{"1.0000001", "1.718282100287241673"},
			{"1.000001", "1.718284546742232836"},
			{"1.00001", "1.718309011413244370"},
			{"1.0001", "1.718553670233753340"},
			{"1.001", "1.721001469881578766"},
			{"1.01", "1.745601015016916494"},
			{"1.1", "2.004166023946433112"},

			// Negated powers of ten
			{"-10000", "-1.000000000000000000"},
			{"-1000", "-1.000000000000000000"},
			{"-100", "-1.000000000000000000"},
			{"-10", "-0.9999546000702375151"},
			{"-1", "-0.6321205588285576784"},
			{"-0.1", "-0.0951625819640404268"},
			{"-0.01", "-0.0099501662508319464"},
			{"-0.001", "-0.0009995001666250083"},
			{"-0.0001", "-0.0000999950001666625"},
			{"-0.00001", "-0.0000099999500001667"},
			{"-0.000001", "-0.0000009999995000002"},
			{"-0.0000001", "-0.0000000999999950000"},
			{"-0.00000001", "-0.0000000099999999500"},
			{"-0.000000001", "-0.0000000009999999995"},
			{"-0.0000000001", "-0.0000000001000000000"},
			{"-0.00000000001", "-0.0000000000100000000"},
			{"-0.000000000001", "-0.0000000000010000000"},
			{"-0.0000000000001", "-0.0000000000001000000"},
			{"-0.00000000000001", "-0.0000000000000100000"},
			{"-0.000000000000001", "-0.0000000000000010000"},
			{"-0.0000000000000001", "-0.0000000000000001000"},
			{"-0.00000000000000001", "-0.0000000000000000100"},
			{"-0.000000000000000001", "-0.0000000000000000010"},
			{"-0.0000000000000000001", "-0.0000000000000000001"},

			// Powers of ten
			{"0.0000000000000000001", "0.0000000000000000001"},
			{"0.000000000000000001", "0.0000000000000000010"},
			{"0.00000000000000001", "0.0000000000000000100"},
			{"0.0000000000000001", "0.0000000000000001000"},
			{"0.000000000000001", "0.0000000000000010000"},
			{"0.00000000000001", "0.0000000000000100000"},
			{"0.0000000000001", "0.0000000000001000000"},
			{"0.000000000001", "0.0000000000010000000"},
			{"0.00000000001", "0.0000000000100000000"},
			{"0.0000000001", "0.0000000001000000000"},
			{"0.000000001", "0.0000000010000000005"},
			{"0.00000001", "0.0000000100000000500"},
			{"0.0000001", "0.0000001000000050000"},
			{"0.000001", "0.0000010000005000002"},
			{"0.00001", "0.0000100000500001667"},
			{"0.0001", "0.0001000050001666708"},
			{"0.001", "0.0010005001667083417"},
			{"0.01", "0.0100501670841680575"},
			{"0.1", "0.1051709180756476248"},
			{"1", "1.718281828459045235"},
			{"10", "22025.46579480671652"},

			// Negative numbers
			{"-101", "-1.000000000000000000"},
			{"-100", "-1.000000000000000000"},
			{"-99", "-1.000000000000000000"},
			{"-50", "-1.000000000000000000"},
			{"-45", "-1.000000000000000000"},
			{"-44", "-0.9999999999999999999"},
			{"-43", "-0.9999999999999999998"},

			// Natural numbers
			{"1", "1.718281828459045235"},
			{"2", "6.389056098930650227"},
			{"3", "19.08553692318766774"},
			{"4", "53.59815003314423908"},
			{"5", "147.4131591025766034"},
			{"6", "402.4287934927351226"},
			{"7", "1095.633158428458599"},
			{"8", "2979.957987041728275"},
			{"9", "8102.083927575384008"},
			{"10", "22025.46579480671652"},
			{"11", "59873.14171519781846"},
			{"12", "162753.7914190039208"},
			{"13", "442412.3920089205033"},
			{"14", "1202603.284164776778"},
			{"15", "3269016.372472110639"},
			{"16", "8886109.520507872637"},
			{"17", "24154951.75357529821"},
			{"18", "65659968.13733051114"},
			{"19", "178482299.9631872608"},
			{"20", "485165194.4097902780"},
			{"21", "1318815733.483214697"},
			{"22", "3584912845.131591562"},
			{"23", "9744803445.248902600"},
			{"24", "26489122128.84347229"},
			{"25", "72004899336.38587252"},
			{"26", "195729609427.8387643"},
			{"27", "532048240600.7986167"},
			{"28", "1446257064290.475174"},
			{"29", "3931334297143.042074"},
			{"30", "10686474581523.46215"},
			{"31", "29048849665246.42523"},
			{"32", "78962960182679.69516"},
			{"33", "214643579785915.0646"},
			{"34", "583461742527453.8814"},
			{"35", "1586013452313429.728"},
			{"36", "4311231547115194.227"},
			{"37", "11719142372802610.31"},
			{"38", "31855931757113755.22"},
			{"39", "86593400423993745.95"},
			{"40", "235385266837019984.4"},
			{"41", "639843493530054948.2"},
			{"42", "1739274941520501046"},
			{"43", "4727839468229346560"},
		}
		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Expm1()
			if err != nil {
				t.Errorf("%q.Expm1() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Expm1() = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]string{
			"overflow 1": "49",
			"overflow 2": "50",
		}
		for name, d := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(d)
				_, err := d.Expm1()
				if err == nil {
					t.Errorf("%q.Expm1() did not fail", d)
				}
			})
		}
	})
}

func TestDecimal_Log(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Ones
			{"1", "0"},
			{"1.0", "0"},
			{"1.00", "0"},
			{"1.000", "0"},

			// Powers of Euler's number
			{"0.0000000000000000002", "-43.05596958632692269"},
			{"0.0000000000000000006", "-41.95735729765881300"},
			{"0.0000000000000000016", "-40.97652804464708676"},
			{"0.0000000000000000042", "-40.01144714860349969"},
			{"0.0000000000000000115", "-39.00418463852361793"},
			{"0.0000000000000000314", "-37.99972378097861463"},
			{"0.0000000000000000853", "-37.00035721939518888"},
			{"0.0000000000000002320", "-35.99979430222651236"},
			{"0.0000000000000006305", "-35.00001851848784806"},
			{"0.0000000000000017139", "-34.00000491949429577"},
			{"0.0000000000000046589", "-32.99999702613981757"},
			{"0.0000000000000126642", "-31.99999727965819527"},
			{"0.0000000000000344248", "-30.99999916004414320"},
			{"0.0000000000000935762", "-30.00000031726440095"},
			{"0.0000000000002543666", "-28.99999986137208992"},
			{"0.0000000000006914400", "-28.00000001546630253"},
			{"0.0000000000018795288", "-27.00000000879959021"},
			{"0.0000000000051090890", "-26.00000000549282360"},
			{"0.0000000000138879439", "-24.99999999747723783"},
			{"0.0000000000377513454", "-24.00000000113349543"},
			{"0.0000000001026187963", "-23.00000000016584587"},
			{"0.0000000002789468093", "-21.99999999995301069"},
			{"0.0000000007582560428", "-20.99999999998838212"},
			{"0.0000000020611536224", "-20.00000000001870692"},
			{"0.0000000056027964375", "-19.00000000000665160"},
			{"0.0000000152299797447", "-18.00000000000082918"},
			{"0.0000000413993771879", "-16.99999999999883251"},
			{"0.0000001125351747193", "-15.99999999999963669"},
			{"0.0000003059023205018", "-15.00000000000008430"},
			{"0.0000008315287191036", "-13.99999999999996138"},
			{"0.0000022603294069811", "-12.99999999999997979"},
			{"0.0000061442123533282", "-12.00000000000000159"},
			{"0.0000167017007902457", "-10.99999999999999756"},
			{"0.0000453999297624849", "-9.999999999999998933"},
			{"0.0001234098040866795", "-9.000000000000000401"},
			{"0.0003354626279025118", "-8.000000000000000116"},
			{"0.0009118819655545162", "-7.000000000000000009"},
			{"0.0024787521766663584", "-6.000000000000000009"},
			{"0.0067379469990854671", "-5.000000000000000000"},
			{"0.0183156388887341803", "-4.000000000000000000"},
			{"0.0497870683678639430", "-3.000000000000000000"},
			{"0.1353352832366126919", "-2.000000000000000000"},
			{"0.3678794411714423216", "-1.000000000000000000"},
			{"1", "0"},
			{E.String(), "0.9999999999999999999"},
			{"7.389056098930650227", "2.000000000000000000"},
			{"20.08553692318766774", "3.000000000000000000"},
			{"54.59815003314423908", "4.000000000000000000"},
			{"148.4131591025766034", "5.000000000000000000"},
			{"403.4287934927351226", "6.000000000000000000"},
			{"1096.633158428458599", "7.000000000000000000"},
			{"2980.957987041728275", "8.000000000000000000"},
			{"8103.083927575384008", "9.000000000000000000"},
			{"22026.46579480671652", "10.00000000000000000"},
			{"59874.14171519781846", "11.00000000000000000"},
			{"162754.7914190039208", "12.00000000000000000"},
			{"442413.3920089205033", "13.00000000000000000"},
			{"1202604.284164776778", "14.00000000000000000"},
			{"3269017.372472110639", "15.00000000000000000"},
			{"8886110.520507872637", "16.00000000000000000"},
			{"24154952.75357529821", "17.00000000000000000"},
			{"65659969.13733051114", "18.00000000000000000"},
			{"178482300.9631872608", "19.00000000000000000"},
			{"485165195.4097902780", "20.00000000000000000"},
			{"1318815734.483214697", "21.00000000000000000"},
			{"3584912846.131591562", "22.00000000000000000"},
			{"9744803446.248902600", "23.00000000000000000"},
			{"26489122129.84347229", "24.00000000000000000"},
			{"72004899337.38587252", "25.00000000000000000"},
			{"195729609428.8387643", "26.00000000000000000"},
			{"532048240601.7986167", "27.00000000000000000"},
			{"1446257064291.475174", "28.00000000000000000"},
			{"3931334297144.042074", "29.00000000000000000"},
			{"10686474581524.46215", "30.00000000000000000"},
			{"29048849665247.42523", "31.00000000000000000"},
			{"78962960182680.69516", "32.00000000000000000"},
			{"214643579785916.0646", "33.00000000000000000"},
			{"583461742527454.8814", "34.00000000000000000"},
			{"1586013452313430.728", "35.00000000000000000"},
			{"4311231547115195.227", "36.00000000000000000"},
			{"11719142372802611.31", "37.00000000000000000"},
			{"31855931757113756.22", "38.00000000000000000"},
			{"86593400423993746.95", "39.00000000000000000"},
			{"235385266837019985.4", "40.00000000000000000"},
			{"639843493530054949.2", "41.00000000000000000"},
			{"1739274941520501047", "42.00000000000000000"},
			{"4727839468229346561", "43.00000000000000000"},

			// Closer and closer to Euler's number
			{"2.7", "0.9932517730102833902"},
			{"2.71", "0.9969486348916095321"},
			{"2.718", "0.9998963157289519689"},
			{"2.7182", "0.9999698965391098865"},
			{"2.71828", "0.9999993273472820032"},
			{"2.718281", "0.9999996952269029621"},
			{"2.7182818", "0.9999999895305022877"},
			{"2.71828182", "0.9999999968880911611"},
			{"2.718281828", "0.9999999998311266953"},
			{"2.7182818284", "0.9999999999782784718"},
			{"2.71828182845", "0.9999999999966724439"},
			{"2.718281828459", "0.9999999999999833588"},
			{"2.7182818284590", "0.9999999999999833588"},
			{"2.71828182845904", "0.9999999999999980740"},
			{"2.718281828459045", "0.9999999999999999134"},
			{"2.7182818284590452", "0.9999999999999999870"},
			{"2.71828182845904523", "0.9999999999999999980"},
			{"2.718281828459045234", "0.9999999999999999995"},
			{E.String(), "0.9999999999999999999"},
			{"2.718281828459045236", "1.000000000000000000"},
			{"2.71828182845904524", "1.000000000000000002"},
			{"2.7182818284590453", "1.000000000000000024"},
			{"2.718281828459046", "1.000000000000000281"},
			{"2.71828182845905", "1.000000000000001753"},
			{"2.7182818284591", "1.000000000000020147"},
			{"2.718281828460", "1.000000000000351238"},
			{"2.71828182846", "1.000000000000351238"},
			{"2.7182818285", "1.000000000015066416"},
			{"2.718281829", "1.000000000199006136"},
			{"2.71828183", "1.000000000566885578"},
			{"2.7182819", "1.000000026318446113"},
			{"2.718282", "1.000000063106388586"},
			{"2.71829", "1.000003006137401513"},
			{"2.7183", "1.000006684913987575"},
			{"2.719", "1.000264165650333661"},
			{"2.72", "1.000631880307905950"},
			{"2.8", "1.029619417181158240"},

			// Closer and closer to one
			{"0.9", "-0.1053605156578263012"},
			{"0.99", "-0.0100503358535014412"},
			{"0.999", "-0.0010005003335835335"},
			{"0.9999", "-0.0001000050003333583"},
			{"0.99999", "-0.0000100000500003333"},
			{"0.999999", "-0.0000010000005000003"},
			{"0.9999999", "-0.0000001000000050000"},
			{"0.99999999", "-0.0000000100000000500"},
			{"0.999999999", "-0.0000000010000000005"},
			{"0.9999999999", "-0.0000000001000000000"},
			{"0.99999999999", "-0.0000000000100000000"},
			{"0.999999999999", "-0.0000000000010000000"},
			{"0.9999999999999", "-0.0000000000001000000"},
			{"0.99999999999999", "-0.0000000000000100000"},
			{"0.999999999999999", "-0.0000000000000010000"},
			{"0.9999999999999999", "-0.0000000000000001000"},
			{"0.99999999999999999", "-0.0000000000000000100"},
			{"0.999999999999999999", "-0.0000000000000000010"},
			{"0.9999999999999999999", "-0.0000000000000000001"},
			{"1", "0"},
			{"1.000000000000000001", "0.0000000000000000010"},
			{"1.00000000000000001", "0.0000000000000000100"},
			{"1.0000000000000001", "0.0000000000000001000"},
			{"1.000000000000001", "0.0000000000000010000"},
			{"1.00000000000001", "0.0000000000000100000"},
			{"1.0000000000001", "0.0000000000001000000"},
			{"1.000000000001", "0.0000000000010000000"},
			{"1.00000000001", "0.0000000000100000000"},
			{"1.0000000001", "0.0000000001000000000"},
			{"1.000000001", "0.0000000009999999995"},
			{"1.00000001", "0.0000000099999999500"},
			{"1.0000001", "0.0000000999999950000"},
			{"1.000001", "0.0000009999995000003"},
			{"1.00001", "0.0000099999500003333"},
			{"1.0001", "0.0000999950003333083"},
			{"1.001", "0.0009995003330835332"},
			{"1.01", "0.0099503308531680828"},
			{"1.1", "0.0953101798043248600"},

			// Closer and closer to zero
			{"0.0000000000000000001", "-43.74911676688686800"},
			{"0.000000000000000001", "-41.44653167389282231"},
			{"0.00000000000000001", "-39.14394658089877663"},
			{"0.0000000000000001", "-36.84136148790473094"},
			{"0.000000000000001", "-34.53877639491068526"},
			{"0.00000000000001", "-32.23619130191663958"},
			{"0.0000000000001", "-29.93360620892259389"},
			{"0.000000000001", "-27.63102111592854821"},
			{"0.00000000001", "-25.32843602293450252"},
			{"0.0000000001", "-23.02585092994045684"},
			{"0.000000001", "-20.72326583694641116"},
			{"0.00000001", "-18.42068074395236547"},
			{"0.0000001", "-16.11809565095831979"},
			{"0.000001", "-13.81551055796427410"},
			{"0.00001", "-11.51292546497022842"},
			{"0.0001", "-9.210340371976182736"},
			{"0.001", "-6.907755278982137052"},
			{"0.01", "-4.605170185988091368"},
			{"0.1", "-2.302585092994045684"},

			// Natural numbers
			{"1", "0"},
			{"2", "0.6931471805599453094"},
			{"3", "1.098612288668109691"},
			{"4", "1.386294361119890619"},
			{"5", "1.609437912434100375"},
			{"6", "1.791759469228055001"},
			{"7", "1.945910149055313305"},
			{"8", "2.079441541679835928"},
			{"9", "2.197224577336219383"},
			{"10", "2.302585092994045684"},
			{"11", "2.397895272798370544"},
			{"12", "2.484906649788000310"},
			{"13", "2.564949357461536736"},
			{"14", "2.639057329615258615"},
			{"15", "2.708050201102210066"},
			{"16", "2.772588722239781238"},
			{"17", "2.833213344056216080"},
			{"18", "2.890371757896164692"},
			{"19", "2.944438979166440460"},
			{"20", "2.995732273553990993"},

			// Smallest and largest numbers
			{"0.0000000000000000001", "-43.74911676688686800"},
			{"9999999999999999999", "43.74911676688686800"},

			// Captured during fuzzing
			{"0.0000000000000097", "-32.26665050940134812"},
			{"0.00000000000018", "-29.34581954402047488"},
			{"0.00444", "-5.417100902538003665"},
			{"562", "6.331501849893691075"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Log()
			if err != nil {
				t.Errorf("%q.Log() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Log() = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]string{
			"negative": "-1",
			"zero":     "0",
		}
		for name, d := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(d)
				_, err := d.Log()
				if err == nil {
					t.Errorf("%q.Log() did not fail", d)
				}
			})
		}
	})
}

func TestDecimal_Log1p(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Zeros
			{"0", "0"},
			{"0.0", "0"},
			{"0.00", "0"},
			{"0.000", "0"},

			// Ones
			{"1", "0.6931471805599453094"},
			{"1.0", "0.6931471805599453094"},
			{"1.00", "0.6931471805599453094"},
			{"1.000", "0.6931471805599453094"},

			// Closer and closer to one
			{"0.9", "0.6418538861723947760"},
			{"0.99", "0.6881346387364010274"},
			{"0.999", "0.6926470555182630115"},
			{"0.9999", "0.6930971793099036412"},
			{"0.99999", "0.6931421805474452678"},
			{"0.999999", "0.6931466805598203094"},
			{"0.9999999", "0.6931471305599440594"},
			{"0.99999999", "0.6931471755599452969"},
			{"0.999999999", "0.6931471800599453093"},
			{"0.9999999999", "0.6931471805099453094"},
			{"0.99999999999", "0.6931471805549453094"},
			{"0.999999999999", "0.6931471805594453094"},
			{"0.9999999999999", "0.6931471805598953094"},
			{"0.99999999999999", "0.6931471805599403094"},
			{"0.999999999999999", "0.6931471805599448094"},
			{"0.9999999999999999", "0.6931471805599452594"},
			{"0.99999999999999999", "0.6931471805599453044"},
			{"0.999999999999999999", "0.6931471805599453089"},
			{"0.9999999999999999999", "0.6931471805599453094"},
			{"1", "0.6931471805599453094"},
			{"1.000000000000000001", "0.6931471805599453099"},
			{"1.00000000000000001", "0.6931471805599453144"},
			{"1.0000000000000001", "0.6931471805599453594"},
			{"1.000000000000001", "0.6931471805599458094"},
			{"1.00000000000001", "0.6931471805599503094"},
			{"1.0000000000001", "0.6931471805599953094"},
			{"1.000000000001", "0.6931471805604453094"},
			{"1.00000000001", "0.6931471805649453094"},
			{"1.0000000001", "0.6931471806099453094"},
			{"1.000000001", "0.6931471810599453093"},
			{"1.00000001", "0.6931471855599452969"},
			{"1.0000001", "0.6931472305599440594"},
			{"1.000001", "0.6931476805598203095"},
			{"1.00001", "0.6931521805474453511"},
			{"1.0001", "0.6931971793099869745"},
			{"1.001", "0.6936470556015963573"},
			{"1.01", "0.6981347220709843830"},
			{"1.1", "0.7419373447293773125"},

			// Closer and closer to zero
			{"-0.1", "-0.1053605156578263012"},
			{"-0.01", "-0.0100503358535014412"},
			{"-0.001", "-0.0010005003335835335"},
			{"-0.0001", "-0.0001000050003333583"},
			{"-0.00001", "-0.0000100000500003333"},
			{"-0.000001", "-0.0000010000005000003"},
			{"-0.0000001", "-0.0000001000000050000"},
			{"-0.00000001", "-0.0000000100000000500"},
			{"-0.000000001", "-0.0000000010000000005"},
			{"-0.0000000001", "-0.0000000001000000000"},
			{"-0.00000000001", "-0.0000000000100000000"},
			{"-0.000000000001", "-0.0000000000010000000"},
			{"-0.0000000000001", "-0.0000000000001000000"},
			{"-0.00000000000001", "-0.0000000000000100000"},
			{"-0.000000000000001", "-0.0000000000000010000"},
			{"-0.0000000000000001", "-0.0000000000000001000"},
			{"-0.00000000000000001", "-0.0000000000000000100"},
			{"-0.000000000000000001", "-0.0000000000000000010"},
			{"-0.0000000000000000001", "-0.0000000000000000001"},
			{"0", "0"},
			{"0.0000000000000000001", "0.0000000000000000001"},
			{"0.000000000000000001", "0.0000000000000000010"},
			{"0.00000000000000001", "0.0000000000000000100"},
			{"0.0000000000000001", "0.0000000000000001000"},
			{"0.000000000000001", "0.0000000000000010000"},
			{"0.00000000000001", "0.0000000000000100000"},
			{"0.0000000000001", "0.0000000000001000000"},
			{"0.000000000001", "0.0000000000010000000"},
			{"0.00000000001", "0.0000000000100000000"},
			{"0.0000000001", "0.0000000001000000000"},
			{"0.000000001", "0.0000000009999999995"},
			{"0.00000001", "0.0000000099999999500"},
			{"0.0000001", "0.0000000999999950000"},
			{"0.000001", "0.0000009999995000003"},
			{"0.00001", "0.0000099999500003333"},
			{"0.0001", "0.0000999950003333083"},
			{"0.001", "0.0009995003330835332"},
			{"0.01", "0.0099503308531680828"},
			{"0.1", "0.0953101798043248600"},

			// Closer and closer to negative one
			{"-0.9999999999999999999", "-43.74911676688686800"},
			{"-0.999999999999999999", "-41.44653167389282231"},
			{"-0.99999999999999999", "-39.14394658089877663"},
			{"-0.9999999999999999", "-36.84136148790473094"},
			{"-0.999999999999999", "-34.53877639491068526"},
			{"-0.99999999999999", "-32.23619130191663958"},
			{"-0.9999999999999", "-29.93360620892259389"},
			{"-0.999999999999", "-27.63102111592854821"},
			{"-0.99999999999", "-25.32843602293450252"},
			{"-0.9999999999", "-23.02585092994045684"},
			{"-0.999999999", "-20.72326583694641116"},
			{"-0.99999999", "-18.42068074395236547"},
			{"-0.9999999", "-16.11809565095831979"},
			{"-0.999999", "-13.81551055796427410"},
			{"-0.99999", "-11.51292546497022842"},
			{"-0.9999", "-9.210340371976182736"},
			{"-0.999", "-6.907755278982137052"},
			{"-0.99", "-4.605170185988091368"},
			{"-0.9", "-2.302585092994045684"},

			// Natural numbers
			{"0", "0"},
			{"1", "0.6931471805599453094"},
			{"2", "1.098612288668109691"},
			{"3", "1.386294361119890619"},
			{"4", "1.609437912434100375"},
			{"5", "1.791759469228055001"},
			{"6", "1.945910149055313305"},
			{"7", "2.079441541679835928"},
			{"8", "2.197224577336219383"},
			{"9", "2.302585092994045684"},
			{"10", "2.397895272798370544"},
			{"11", "2.484906649788000310"},
			{"12", "2.564949357461536736"},
			{"13", "2.639057329615258615"},
			{"14", "2.708050201102210066"},
			{"15", "2.772588722239781238"},
			{"16", "2.833213344056216080"},
			{"17", "2.890371757896164692"},
			{"18", "2.944438979166440460"},
			{"19", "2.995732273553990993"},
			{"20", "3.044522437723422997"},

			// Smallest and largest numbers
			{"0.0000000000000000001", "0.0000000000000000001"},
			{"9999999999999999999", "43.74911676688686800"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Log1p()
			if err != nil {
				t.Errorf("%q.Log1p() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Log1p() = %q, want %q", d, got, want)
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
				_, err := d.Log1p()
				if err == nil {
					t.Errorf("%q.Log1p() did not fail", d)
				}
			})
		}
	})
}

func TestDecimal_Log2(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Ones
			{"1", "0"},
			{"1.0", "0"},
			{"1.00", "0"},
			{"1.000", "0"},

			// Powers of two
			{"0.0000000000000000001", "-63.11663380285988461"},
			{"0.0000000000000000002", "-62.11663380285988461"},
			{"0.0000000000000000004", "-61.11663380285988461"},
			{"0.0000000000000000009", "-59.94670880141757225"},
			{"0.0000000000000000017", "-59.02917096160954520"},
			{"0.0000000000000000035", "-57.98735078591491815"},
			{"0.0000000000000000069", "-57.00810934608171556"},
			{"0.0000000000000000139", "-55.99769273013637718"},
			{"0.0000000000000000278", "-54.99769273013637718"},
			{"0.0000000000000000555", "-54.00028984162241630"},
			{"0.0000000000000001110", "-53.00028984162241630"},
			{"0.0000000000000002220", "-52.00028984162241630"},
			{"0.0000000000000004441", "-50.99996494689277185"},
			{"0.0000000000000008882", "-49.99996494689277185"},
			{"0.0000000000000017764", "-48.99996494689277185"},
			{"0.0000000000000035527", "-48.00000555473292341"},
			{"0.0000000000000071054", "-47.00000555473292341"},
			{"0.0000000000000142109", "-45.99999540266572901"},
			{"0.0000000000000284217", "-45.00000047869039635"},
			{"0.0000000000000568434", "-44.00000047869039635"},
			{"0.0000000000001136868", "-43.00000047869039635"},
			{"0.0000000000002273737", "-41.99999984418633623"},
			{"0.0000000000004547474", "-40.99999984418633623"},
			{"0.0000000000009094947", "-40.00000000281232510"},
			{"0.0000000000018189894", "-39.00000000281232510"},
			{"0.0000000000036379788", "-38.00000000281232510"},
			{"0.0000000000072759576", "-37.00000000281232510"},
			{"0.0000000000145519152", "-36.00000000281232510"},
			{"0.0000000000291038305", "-34.99999999785526269"},
			{"0.0000000000582076609", "-34.00000000033379389"},
			{"0.0000000001164153218", "-33.00000000033379389"},
			{"0.0000000002328306437", "-31.99999999971416109"},
			{"0.0000000004656612873", "-31.00000000002397749"},
			{"0.0000000009313225746", "-30.00000000002397749"},
			{"0.0000000018626451492", "-29.00000000002397749"},
			{"0.0000000037252902985", "-27.99999999998525044"},
			{"0.0000000074505805969", "-27.00000000000461396"},
			{"0.0000000149011611938", "-26.00000000000461396"},
			{"0.0000000298023223877", "-24.99999999999977308"},
			{"0.0000000596046447754", "-23.99999999999977308"},
			{"0.0000001192092895508", "-22.99999999999977308"},
			{"0.0000002384185791016", "-21.99999999999977308"},
			{"0.0000004768371582031", "-21.00000000000007564"},
			{"0.0000009536743164062", "-20.00000000000007564"},
			{"0.0000019073486328125", "-19"},
			{"0.000003814697265625", "-18"},
			{"0.00000762939453125", "-17"},
			{"0.0000152587890625", "-16"},
			{"0.000030517578125", "-15"},
			{"0.00006103515625", "-14"},
			{"0.0001220703125", "-13"},
			{"0.000244140625", "-12"},
			{"0.00048828125", "-11"},
			{"0.0009765625", "-10"},
			{"0.001953125", "-9"},
			{"0.00390625", "-8"},
			{"0.0078125", "-7"},
			{"0.015625", "-6"},
			{"0.03125", "-5"},
			{"0.0625", "-4"},
			{"0.125", "-3"},
			{"0.25", "-2"},
			{"0.5", "-1"},
			{"1", "0"},
			{"2", "1"},
			{"4", "2"},
			{"8", "3"},
			{"16", "4"},
			{"32", "5"},
			{"64", "6"},
			{"128", "7"},
			{"256", "8"},
			{"512", "9"},
			{"1024", "10"},
			{"2048", "11"},
			{"4096", "12"},
			{"8192", "13"},
			{"16384", "14"},
			{"32768", "15"},
			{"65536", "16"},
			{"131072", "17"},
			{"262144", "18"},
			{"524288", "19"},
			{"1048576", "20"},
			{"2097152", "21"},
			{"4194304", "22"},
			{"8388608", "23"},
			{"16777216", "24"},
			{"33554432", "25"},
			{"67108864", "26"},
			{"134217728", "27"},
			{"268435456", "28"},
			{"536870912", "29"},
			{"1073741824", "30"},
			{"2147483648", "31"},
			{"4294967296", "32"},
			{"8589934592", "33"},
			{"17179869184", "34"},
			{"34359738368", "35"},
			{"68719476736", "36"},
			{"137438953472", "37"},
			{"274877906944", "38"},
			{"549755813888", "39"},
			{"1099511627776", "40"},
			{"2199023255552", "41"},
			{"4398046511104", "42"},
			{"8796093022208", "43"},
			{"17592186044416", "44"},
			{"35184372088832", "45"},
			{"70368744177664", "46"},
			{"140737488355328", "47"},
			{"281474976710656", "48"},
			{"562949953421312", "49"},
			{"1125899906842624", "50"},
			{"2251799813685248", "51"},
			{"4503599627370496", "52"},
			{"9007199254740992", "53"},
			{"18014398509481984", "54"},
			{"36028797018963968", "55"},
			{"72057594037927936", "56"},
			{"144115188075855872", "57"},
			{"288230376151711744", "58"},
			{"576460752303423488", "59"},
			{"1152921504606846976", "60"},
			{"2305843009213693952", "61"},
			{"4611686018427387904", "62"},
			{"9223372036854775808", "63"},

			// Closer and closer to two
			{"1.9", "0.9259994185562231459"},
			{"1.99", "0.9927684307689241428"},
			{"1.999", "0.9992784720825405627"},
			{"1.9999", "0.9999278634445266362"},
			{"1.99999", "0.9999927865067618071"},
			{"1.999999", "0.9999992786522992186"},
			{"1.9999999", "0.9999999278652461522"},
			{"1.99999999", "0.9999999927865247775"},
			{"1.999999999", "0.9999999992786524794"},
			{"1.9999999999", "0.9999999999278652480"},
			{"1.99999999999", "0.9999999999927865248"},
			{"1.999999999999", "0.9999999999992786525"},
			{"1.9999999999999", "0.9999999999999278652"},
			{"1.99999999999999", "0.9999999999999927865"},
			{"1.999999999999999", "0.9999999999999992787"},
			{"1.9999999999999999", "0.9999999999999999279"},
			{"1.99999999999999999", "0.9999999999999999928"},
			{"2", "1"},
			{"2.00000000000000001", "1.000000000000000007"},
			{"2.0000000000000001", "1.000000000000000072"},
			{"2.000000000000001", "1.000000000000000721"},
			{"2.00000000000001", "1.000000000000007213"},
			{"2.0000000000001", "1.000000000000072135"},
			{"2.000000000001", "1.000000000000721348"},
			{"2.00000000001", "1.000000000007213475"},
			{"2.0000000001", "1.000000000072134752"},
			{"2.000000001", "1.000000000721347520"},
			{"2.00000001", "1.000000007213475186"},
			{"2.0000001", "1.000000072134750241"},
			{"2.000001", "1.000000721347340108"},
			{"2.00001", "1.000007213457170817"},
			{"2.0001", "1.000072132948735757"},
			{"2.001", "1.000721167243654131"},
			{"2.1", "1.070389327891397941"},

			// Closer and closer to one
			{"0.9", "-0.1520030934450499850"},
			{"0.99", "-0.0144995696951150766"},
			{"0.999", "-0.0014434168696687174"},
			{"0.9999", "-0.0001442767180450352"},
			{"0.99999", "-0.0000144270225441226"},
			{"0.999999", "-0.0000014426957622370"},
			{"0.9999999", "-0.0000001442695113024"},
			{"0.99999999", "-0.0000000144269504810"},
			{"0.999999999", "-0.0000000014426950416"},
			{"0.9999999999", "-0.0000000001442695041"},
			{"0.99999999999", "-0.0000000000144269504"},
			{"0.999999999999", "-0.0000000000014426950"},
			{"0.9999999999999", "-0.0000000000001442695"},
			{"0.99999999999999", "-0.0000000000000144270"},
			{"0.999999999999999", "-0.0000000000000014427"},
			{"0.9999999999999999", "-0.0000000000000001443"},
			{"0.99999999999999999", "-0.0000000000000000144"},
			{"0.999999999999999999", "-0.0000000000000000014"},
			{"1", "0"},
			{"1.000000000000000001", "0.0000000000000000014"},
			{"1.00000000000000001", "0.0000000000000000144"},
			{"1.0000000000000001", "0.0000000000000001443"},
			{"1.000000000000001", "0.0000000000000014427"},
			{"1.00000000000001", "0.0000000000000144270"},
			{"1.0000000000001", "0.0000000000001442695"},
			{"1.000000000001", "0.0000000000014426950"},
			{"1.00000000001", "0.0000000000144269504"},
			{"1.0000000001", "0.0000000001442695041"},
			{"1.000000001", "0.0000000014426950402"},
			{"1.00000001", "0.0000000144269503368"},
			{"1.0000001", "0.0000001442694968754"},
			{"1.000001", "0.0000014426943195419"},
			{"1.00001", "0.0000144268782746185"},
			{"1.0001", "0.0001442622910945542"},
			{"1.001", "0.0014419741739064804"},
			{"1.01", "0.0143552929770700414"},
			{"1.1", "0.1375035237499349083"},

			// Natural numbers
			{"1", "0"},
			{"2", "1"},
			{"3", "1.584962500721156181"},
			{"4", "2"},
			{"5", "2.321928094887362348"},
			{"6", "2.584962500721156181"},
			{"7", "2.807354922057604107"},
			{"8", "3"},
			{"9", "3.169925001442312363"},
			{"10", "3.321928094887362348"},
			{"11", "3.459431618637297256"},
			{"12", "3.584962500721156181"},
			{"13", "3.700439718141092160"},
			{"14", "3.807354922057604107"},
			{"15", "3.906890595608518529"},
			{"16", "4"},
			{"17", "4.087462841250339408"},
			{"18", "4.169925001442312363"},
			{"19", "4.247927513443585494"},
			{"20", "4.321928094887362348"},

			// Smallest and largest numbers
			{"0.0000000000000000001", "-63.11663380285988461"},
			{"9999999999999999999", "63.11663380285988461"},

			// Captured during fuzzing
			{"0.00375", "-8.058893689053568514"},
			{"9223372036854.775784", "43.06843143067582591"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Log2()
			if err != nil {
				t.Errorf("%q.Log2() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q,%q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]string{
			"negative": "-1",
			"zero":     "0",
		}
		for name, d := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(d)
				_, err := d.Log2()
				if err == nil {
					t.Errorf("%q.Log2() did not fail", d)
				}
			})
		}
	})
}

func TestDecimal_Log10(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			d, want string
		}{
			// Ones
			{"1", "0"},
			{"1.0", "0"},
			{"1.00", "0"},
			{"1.000", "0"},

			// Powers of ten
			{"0.0000000000000000001", "-19"},
			{"0.000000000000000001", "-18"},
			{"0.00000000000000001", "-17"},
			{"0.0000000000000001", "-16"},
			{"0.000000000000001", "-15"},
			{"0.00000000000001", "-14"},
			{"0.0000000000001", "-13"},
			{"0.000000000001", "-12"},
			{"0.00000000001", "-11"},
			{"0.0000000001", "-10"},
			{"0.000000001", "-9"},
			{"0.00000001", "-8"},
			{"0.0000001", "-7"},
			{"0.000001", "-6"},
			{"0.00001", "-5"},
			{"0.0001", "-4"},
			{"0.001", "-3"},
			{"0.01", "-2"},
			{"0.1", "-1"},
			{"1", "0"},
			{"10", "1"},
			{"100", "2"},
			{"1000", "3"},
			{"10000", "4"},
			{"100000", "5"},
			{"1000000", "6"},
			{"10000000", "7"},
			{"100000000", "8"},
			{"1000000000", "9"},
			{"10000000000", "10"},
			{"100000000000", "11"},
			{"1000000000000", "12"},
			{"10000000000000", "13"},
			{"100000000000000", "14"},
			{"1000000000000000", "15"},
			{"10000000000000000", "16"},
			{"100000000000000000", "17"},
			{"1000000000000000000", "18"},

			// Closer and closer to ten
			{"9.9", "0.9956351945975499153"},
			{"9.99", "0.9995654882259823087"},
			{"9.999", "0.9999565683801924896"},
			{"9.9999", "0.9999956570334660986"},
			{"9.99999", "0.9999995657053009494"},
			{"9.999999", "0.9999999565705496382"},
			{"9.9999999", "0.9999999956570551593"},
			{"9.99999999", "0.9999999995657055179"},
			{"9.999999999", "0.9999999999565705518"},
			{"9.9999999999", "0.9999999999956570552"},
			{"9.99999999999", "0.9999999999995657055"},
			{"9.999999999999", "0.9999999999999565706"},
			{"9.9999999999999", "0.9999999999999956571"},
			{"9.99999999999999", "0.9999999999999995657"},
			{"9.999999999999999", "0.9999999999999999566"},
			{"9.9999999999999999", "0.9999999999999999957"},
			{"9.99999999999999999", "0.9999999999999999996"},
			{"9.999999999999999999", "1"},
			{"10", "1"},
			{"10.00000000000000001", "1"},
			{"10.0000000000000001", "1.000000000000000004"},
			{"10.000000000000001", "1.000000000000000043"},
			{"10.00000000000001", "1.000000000000000434"},
			{"10.0000000000001", "1.000000000000004343"},
			{"10.000000000001", "1.000000000000043429"},
			{"10.00000000001", "1.000000000000434294"},
			{"10.0000000001", "1.000000000004342945"},
			{"10.000000001", "1.000000000043429448"},
			{"10.00000001", "1.000000000434294482"},
			{"10.0000001", "1.000000004342944797"},
			{"10.000001", "1.000000043429446019"},
			{"10.00001", "1.000000434294264756"},
			{"10.0001", "1.000004342923104453"},
			{"10.001", "1.000043427276862670"},
			{"10.01", "1.000434077479318641"},
			{"10.1", "1.004321373782642574"},

			// Closer and closer to one
			{"0.9", "-0.0457574905606751254"},
			{"0.99", "-0.0043648054024500847"},
			{"0.999", "-0.0004345117740176913"},
			{"0.9999", "-0.0000434316198075104"},
			{"0.99999", "-0.0000043429665339014"},
			{"0.999999", "-0.0000004342946990506"},
			{"0.9999999", "-0.0000000434294503618"},
			{"0.99999999", "-0.0000000043429448407"},
			{"0.999999999", "-0.0000000004342944821"},
			{"0.9999999999", "-0.0000000000434294482"},
			{"0.99999999999", "-0.0000000000043429448"},
			{"0.999999999999", "-0.0000000000004342945"},
			{"0.9999999999999", "-0.0000000000000434294"},
			{"0.99999999999999", "-0.0000000000000043429"},
			{"0.999999999999999", "-0.0000000000000004343"},
			{"0.9999999999999999", "-0.0000000000000000434"},
			{"0.99999999999999999", "-0.0000000000000000043"},
			{"0.999999999999999999", "-0.0000000000000000004"},
			{"0.9999999999999999999", "0"},
			{"1", "0"},
			{"1.000000000000000001", "0.0000000000000000004"},
			{"1.00000000000000001", "0.0000000000000000043"},
			{"1.0000000000000001", "0.0000000000000000434"},
			{"1.000000000000001", "0.0000000000000004343"},
			{"1.00000000000001", "0.0000000000000043429"},
			{"1.0000000000001", "0.0000000000000434294"},
			{"1.000000000001", "0.0000000000004342945"},
			{"1.00000000001", "0.0000000000043429448"},
			{"1.0000000001", "0.0000000000434294482"},
			{"1.000000001", "0.0000000004342944817"},
			{"1.00000001", "0.0000000043429447973"},
			{"1.0000001", "0.0000000434294460189"},
			{"1.000001", "0.0000004342942647562"},
			{"1.00001", "0.0000043429231044532"},
			{"1.0001", "0.0000434272768626696"},
			{"1.001", "0.0004340774793186407"},
			{"1.01", "0.0043213737826425743"},
			{"1.1", "0.0413926851582250408"},

			// Natural numbers
			{"1", "0"},
			{"2", "0.3010299956639811952"},
			{"3", "0.4771212547196624373"},
			{"4", "0.6020599913279623904"},
			{"5", "0.6989700043360188048"},
			{"6", "0.7781512503836436325"},
			{"7", "0.8450980400142568307"},
			{"8", "0.9030899869919435856"},
			{"9", "0.9542425094393248746"},
			{"10", "1"},
			{"11", "1.041392685158225041"},
			{"12", "1.079181246047624828"},
			{"13", "1.113943352306836769"},
			{"14", "1.146128035678238026"},
			{"15", "1.176091259055681242"},
			{"16", "1.204119982655924781"},
			{"17", "1.230448921378273929"},
			{"18", "1.255272505103306070"},
			{"19", "1.278753600952828962"},
			{"20", "1.301029995663981195"},

			// Smallest and largest numbers
			{"0.0000000000000000001", "-19"},
			{"9999999999999999999", "19"},

			// Captured during fuzzing
			{"0.00000000373", "-8.428291168191312394"},
			{"1.048", "0.0203612826477078465"},
		}

		for _, tt := range tests {
			d := MustParse(tt.d)
			got, err := d.Log10()
			if err != nil {
				t.Errorf("%q.Log10() failed: %v", d, err)
				continue
			}
			want := MustParse(tt.want)
			if got != want {
				t.Errorf("%q.Log10() = %q, want %q", d, got, want)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]string{
			"negative": "-1",
			"zero":     "0",
		}
		for name, d := range tests {
			t.Run(name, func(t *testing.T) {
				d := MustParse(d)
				_, err := d.Log10()
				if err == nil {
					t.Errorf("%q.Log10() did not fail", d)
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
			{"1.000000000000000049", "-99.9999999999999924", "-0.0100000000000000013"},
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

			// 1 divided by natural numbers
			{"1", "1", "1", "0"},
			{"1", "2", "0", "1"},
			{"1", "3", "0", "1"},
			{"1", "4", "0", "1"},
			{"1", "5", "0", "1"},
			{"1", "6", "0", "1"},
			{"1", "7", "0", "1"},
			{"1", "8", "0", "1"},
			{"1", "9", "0", "1"},

			// 2 divided by natural numbers
			{"2", "1", "2", "0"},
			{"2", "2", "1", "0"},
			{"2", "3", "0", "2"},
			{"2", "4", "0", "2"},
			{"2", "5", "0", "2"},
			{"2", "6", "0", "2"},
			{"2", "7", "0", "2"},
			{"2", "8", "0", "2"},
			{"2", "9", "0", "2"},

			// Closer and closer to five
			{"12345", "4.9", "2519", "1.9"},
			{"12345", "4.99", "2473", "4.73"},
			{"12345", "4.999", "2469", "2.469"},
			{"12345", "4.9999", "2469", "0.2469"},
			{"12345", "4.99999", "2469", "0.02469"},
			{"12345", "4.999999", "2469", "0.002469"},
			{"12345", "4.9999999", "2469", "0.0002469"},
			{"12345", "4.99999999", "2469", "0.00002469"},
			{"12345", "4.999999999", "2469", "0.000002469"},
			{"12345", "4.9999999999", "2469", "0.0000002469"},
			{"12345", "4.99999999999", "2469", "0.00000002469"},
			{"12345", "4.999999999999", "2469", "0.000000002469"},
			{"12345", "4.9999999999999", "2469", "0.0000000002469"},
			{"12345", "4.99999999999999", "2469", "0.00000000002469"},
			{"12345", "4.999999999999999", "2469", "0.000000000002469"},
			{"12345", "4.9999999999999999", "2469", "0.0000000000002469"},
			{"12345", "4.99999999999999999", "2469", "0.00000000000002469"},
			{"12345", "4.999999999999999999", "2469", "0.000000000000002469"},
			{"12345", "5", "2469", "0"},
			{"12345", "5.000000000000000001", "2468", "4.999999999999997532"},
			{"12345", "5.00000000000000001", "2468", "4.99999999999997532"},
			{"12345", "5.0000000000000001", "2468", "4.9999999999997532"},
			{"12345", "5.000000000000001", "2468", "4.999999999997532"},
			{"12345", "5.00000000000001", "2468", "4.99999999997532"},
			{"12345", "5.0000000000001", "2468", "4.9999999997532"},
			{"12345", "5.000000000001", "2468", "4.999999997532"},
			{"12345", "5.00000000001", "2468", "4.99999997532"},
			{"12345", "5.0000000001", "2468", "4.9999997532"},
			{"12345", "5.000000001", "2468", "4.999997532"},
			{"12345", "5.00000001", "2468", "4.99997532"},
			{"12345", "5.0000001", "2468", "4.9997532"},
			{"12345", "5.000001", "2468", "4.997532"},
			{"12345", "5.00001", "2468", "4.97532"},
			{"12345", "5.0001", "2468", "4.7532"},
			{"12345", "5.001", "2468", "2.532"},
			{"12345", "5.01", "2464", "0.36"},
			{"12345", "5.1", "2420", "3.0"},

			// Other tests
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

//nolint:revive
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
		for s := range MaxScale + 1 {
			d, err := newSafe(c.neg, fint(c.coef), c.scale)
			if err != nil {
				continue
			}
			f.Add(d.bytes(), s)
		}
	}

	f.Fuzz(
		func(t *testing.T, text []byte, scale int) {
			got, err := parseFint(text, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := parseBint(text, scale)
			if err != nil {
				t.Errorf("parseBint(%q) failed: %v", text, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("parseBint(%q) = %q, whereas parseFint(%q) = %q", text, want, text, got)
			}
		},
	)
}

func FuzzBSON(f *testing.F) {
	for _, c := range corpus {
		d := newUnsafe(c.neg, fint(c.coef), c.scale)
		f.Add(byte(19), d.ieeeDecimal128())
	}
	f.Fuzz(
		func(_ *testing.T, typ byte, data []byte) {
			var d Decimal
			_ = d.UnmarshalBSONValue(typ, data)
		},
	)
}

func FuzzDecimal_String_Parse(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			s := d.String()
			got, err := Parse(s)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", s, err)
				return
			}

			want := d

			if got.CmpTotal(want) != 0 {
				t.Errorf("Parse(%q) = %v, want %v", s, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_IEEE_ParseIEEE(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			b := d.ieeeDecimal128()
			got, err := parseIEEEDecimal128(b)
			if err != nil {
				t.Logf("%q.ieeeDecimal128() = % x", d, b)
				t.Errorf("parseIEEEDecimal128(% x) failed: %v", b, err)
				return
			}

			want := d

			if got.CmpTotal(want) != 0 {
				t.Logf("%q.ieeeDecimal128() = % x", d, b)
				t.Errorf("parseIEEEDecimal128(% x) = %v, want %v", b, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Binary_Text(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			b, err := d.MarshalBinary()
			if err != nil {
				t.Errorf("%q.MarshalBinary() failed: %v", d, err)
				return
			}

			var got Decimal
			err = got.UnmarshalText(b)
			if err != nil {
				t.Logf("%q.MarshalBinary() = % x", d, b)
				t.Errorf("UnmarshalText(% x) failed: %v", b, err)
				return
			}

			want := d

			if got.CmpTotal(want) != 0 {
				t.Logf("%q.MarshalBinary() = % x", d, b)
				t.Errorf("UnmarshalText(% x) = %v, want %v", b, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Text_Binary(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			b, err := d.MarshalText()
			if err != nil {
				t.Errorf("%q.MarshalText() failed: %v", d, err)
				return
			}

			var got Decimal
			err = got.UnmarshalBinary(b)
			if err != nil {
				t.Logf("%q.MarshalText() = % x", d, b)
				t.Errorf("UnmarshalBinary(% x) failed: %v", b, err)
				return
			}

			want := d

			if got.CmpTotal(want) != 0 {
				t.Logf("%q.MarshalText() = % x", d, b)
				t.Errorf("UnmarshalBinary(% x) = %v, want %v", b, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Int64_NewFromInt64(f *testing.F) {
	for _, d := range corpus {
		for s := range MaxScale + 1 {
			f.Add(d.neg, d.scale, d.coef, s)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, scale int) {
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}

			w, f, ok := d.Int64(scale)
			if !ok {
				t.Skip()
				return
			}
			got, err := NewFromInt64(w, f, scale)
			if err != nil {
				t.Logf("%q.Int64(%v) = (%v, %v)", d, scale, w, f)
				t.Errorf("NewFromInt64(%v, %v, %v) failed: %v", w, f, scale, err)
				return
			}

			want := d.Round(scale)

			if got.Cmp(want) != 0 {
				t.Logf("%q.Int64(%v) = (%v, %v)", d, scale, w, f)
				t.Errorf("NewFromInt64(%v, %v, %v) = %v, want %v", w, f, scale, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Float64_NewFromFloat64(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64) {
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil || d.Prec() > 17 {
				t.Skip()
				return
			}

			f, ok := d.Float64()
			if !ok {
				t.Errorf("%q.Float64() failed", d)
				return
			}
			got, err := NewFromFloat64(f)
			if err != nil {
				t.Logf("%q.Float64() = %v", d, f)
				t.Errorf("NewFromFloat64(%v) failed: %v", f, err)
				return
			}

			want := d

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
			for s := range MaxScale + 1 {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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
				t.Skip()
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

func FuzzDecimal_Mul_Prod(f *testing.F) {
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

			got, err := d.Mul(e)
			if err != nil {
				t.Skip()
				return
			}

			want, err := Prod(d, e)
			if err != nil {
				t.Errorf("Prod(%q, %q) failed: %v", d, e, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("Prod(%q, %q) = %q, whereas Mul(%q, %q) = %q", d, e, want, d, e, got)
			}
		},
	)
}

func FuzzDecimal_AddMul(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for _, g := range corpus {
				for s := range MaxScale + 1 {
					f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, g.neg, g.scale, g.coef, s)
				}
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, gneg bool, gscale int, gcoef uint64, scale int) {
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

			got, err := d.addMulFint(e, g, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.addMulBint(e, g, scale)
			if err != nil {
				t.Errorf("addMulBint(%q, %q, %q, %v) failed: %v", d, e, g, scale, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("addMulBint(%q, %q, %q, %v) = %q, whereas addMulFint(%q, %q, %q, %v) = %q", d, e, g, scale, want, d, e, g, scale, got)
			}
		},
	)
}

func FuzzDecimal_Add_AddMul(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := range MaxScale + 1 {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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

			got, err := d.AddExact(e, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.AddMulExact(e, One, scale)
			if err != nil {
				t.Errorf("AddMulExact(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("AddMulExact(%q, %q, %q, %v) = %q, whereas AddExact(%q, %q, %v) = %q", d, e, One, scale, want, d, e, scale, got)
				return
			}
		},
	)
}

func FuzzDecimal_Mul_AddMul(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := range MaxScale + 1 {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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

			got, err := d.MulExact(e, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := Zero.AddMulExact(d, e, scale)
			if err != nil {
				t.Errorf("AddMulExact(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("AddMulExact(%q, %q, %q, %v) = %q, whereas MulExact(%q, %q, %v) = %q", Zero, d, e, scale, want, d, e, scale, got)
			}
		},
	)
}

func FuzzDecimal_AddQuo(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for _, g := range corpus {
				for s := range MaxScale + 1 {
					f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, g.neg, g.scale, g.coef, s)
				}
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, gneg bool, gscale int, gcoef uint64, scale int) {
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

			got, err := d.addQuoFint(e, g, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.addQuoBint(e, g, scale)
			if err != nil {
				t.Errorf("addQuoBint(%q, %q, %q, %v) failed: %v", d, e, g, scale, err)
				return
			}

			if got.Cmp(want) != 0 {
				t.Errorf("addQuoBint(%q, %q, %q, %v) = %q, whereas addQuoFint(%q, %q, %q, %v) = %q", d, e, g, scale, want, d, e, g, scale, got)
			}
		},
	)
}

func FuzzDecimal_Add_AddQuo(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := range MaxScale + 1 {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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

			got, err := d.AddExact(e, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.AddQuoExact(e, One, scale)
			if err != nil {
				t.Errorf("AddQuoExact(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("AddQuoExact(%q, %q, %q, %v) = %q, whereas AddExact(%q, %q, %v) = %q", d, e, One, scale, want, d, e, scale, got)
				return
			}
		},
	)
}

func FuzzDecimal_Quo_AddQuo(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := 0; s <= MaxScale; s++ {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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

			got, err := d.QuoExact(e, scale)
			if err != nil {
				t.Skip()
				return
			}

			want, err := Zero.AddQuoExact(d, e, scale)
			if err != nil {
				t.Errorf("AddQuoExact(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("AddQuoExact(%q, %q, %q, %v) = %q, whereas QuoExact(%q, %q, %v) = %q", Zero, d, e, scale, want, d, e, scale, got)
			}
		},
	)
}

func FuzzDecimal_Add(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := range MaxScale + 1 {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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
				t.Skip()
				return
			}

			want, err := d.addBint(e, scale)
			if err != nil {
				t.Errorf("addBint(%q, %q, %v) failed: %v", d, e, scale, err)
				return
			}

			if got.Cmp(want) != 0 {
				t.Errorf("addBint(%q, %q, %v) = %q, whereas addFint(%q, %q, %v) = %q", d, e, scale, want, d, e, scale, got)
			}
		},
	)
}

func FuzzDecimal_Add_Sum(f *testing.F) {
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

			got, err := d.Add(e)
			if err != nil {
				t.Skip()
				return
			}

			want, err := Sum(d, e)
			if err != nil {
				t.Errorf("Sum(%q, %q) failed: %v", d, e, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("Sum(%q, %q) = %q, whereas Add(%q, %q) = %q", d, e, want, d, e, got)
			}
		},
	)
}

func FuzzDecimal_Quo(f *testing.F) {
	for _, d := range corpus {
		for _, e := range corpus {
			for s := range MaxScale + 1 {
				f.Add(d.neg, d.scale, d.coef, e.neg, e.scale, e.coef, s)
			}
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, eneg bool, escale int, ecoef uint64, scale int) {
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
				t.Skip()
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

			gotQuo, gotRem, err := d.quoRemFint(e)
			if err != nil {
				t.Skip()
				return
			}
			if gotQuo.Scale() != 0 {
				t.Errorf("quoRemFint(%q, %q) = (%q, _), expected integer quotient", d, e, gotQuo)
			}
			if gotRem.Scale() != max(d.Scale(), e.Scale()) {
				t.Errorf("quoRemFint(%q, %q) = (_, %q), expected remainder with scale %d", d, e, gotRem, max(d.Scale(), e.Scale()))
			}
			if !gotRem.IsZero() && gotRem.Sign() != d.Sign() {
				t.Errorf("quoRemFint(%q, %q) = (_, %q), expected remainder with the same sign as the dividend", d, e, gotRem)
			}

			wantQuo, wantRem, err := d.quoRemBint(e)
			if err != nil {
				t.Errorf("quoRemBint(%q, %q) failed: %v", d, e, err)
				return
			}

			if gotQuo.CmpTotal(wantQuo) != 0 || gotRem.CmpTotal(wantRem) != 0 {
				t.Errorf("quoRemBint(%q, %q) = (%q, %q), whereas quoRemFint(%q, %q) = (%q, %q)", d, e, wantQuo, wantRem, d, e, gotQuo, gotRem)
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
				t.Skip()
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

func FuzzDecimal_Sqrt_PowInt(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.Sqrt()
			if err != nil {
				t.Skip()
				return
			}
			got, err = got.PowInt(2)
			if err != nil {
				t.Skip()
				return
			}

			want := d

			if cmp, err := cmpULP(got, want, 3); err != nil {
				t.Errorf("cmpULP(%q, %q) failed: %v", got, want, err)
			} else if cmp != 0 {
				t.Errorf("%q.Sqrt().PowInt(2) = %q, want %q", want, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Pow_Sqrt(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			half := MustNew(5, 1)
			got, err := d.Pow(half)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.Sqrt()
			if err != nil {
				t.Errorf("%q.Sqrt() failed: %v", d, err)
				return
			}

			if got.Cmp(want) != 0 {
				t.Errorf("%q.Pow(%v) = %q, whereas %q.Sqrt() = %q", d, half, got, d, want)
				return
			}
		},
	)
}

func FuzzDecimal_Pow_PowInt(f *testing.F) {
	for _, d := range corpus {
		for _, e := range []int{-10, -5, -1, 1, 5, 10} {
			f.Add(d.neg, d.scale, d.coef, e)
		}
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64, power int) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}
			e, err := New(int64(power), 0)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.Pow(e)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.PowInt(power)
			if err != nil {
				t.Errorf("%q.PowInt(%v) failed: %v", d, power, err)
				return
			}

			if got.CmpTotal(want) != 0 {
				t.Errorf("%q.Pow(%v) = %q, whereas %q.PowInt(%v) = %q", d, power, got, d, power, want)
				return
			}
		},
	)
}

func FuzzDecimal_Pow_Exp(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := E.Pow(d)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.Exp()
			if err != nil {
				t.Errorf("%q.Exp() failed: %v", d, err)
				return
			}

			if cmp, err := cmpULP(got, want, 55); err != nil {
				t.Errorf("cmpULP(%q, %q) failed: %v", got, want, err)
			} else if cmp != 0 {
				t.Errorf("%v.Pow(%q) = %q, whereas %q.Exp() = %q", E, d, got, d, want)
				return
			}
		},
	)
}

func FuzzDecimal_Log_Exp(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.Log()
			if err != nil {
				t.Skip()
				return
			}
			got, err = got.Exp()
			if err != nil {
				t.Skip()
				return
			}

			want := d

			if cmp, err := cmpULP(got, want, 70); err != nil {
				t.Errorf("cmpULP(%q, %q) failed: %v", got, want, err)
				return
			} else if cmp != 0 {
				t.Errorf("%q.Log().Exp() = %q, want %q", want, got, want)
				return
			}
		},
	)
}

func FuzzDecimal_Expm1_Exp(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.Exp()
			if err != nil {
				t.Skip()
				return
			}
			got, err = got.Sub(One)
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.Expm1()
			if err != nil {
				t.Skip()
				return
			}

			if cmp, err := cmpULP(got, want, 5); err != nil {
				t.Errorf("cmpULP(%q, %q) failed: %v", got, want, err)
				return
			} else if cmp != 0 {
				t.Errorf("%q.Exp().Sub(1) = %q, whereas %q.Expm1() = %q", d, got, d, want)
				return
			}
		},
	)
}

func FuzzDecimal_Log1p_Log(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			d, err := newSafe(neg, fint(coef), scale)
			if err != nil {
				t.Skip()
				return
			}

			got, err := d.Add(One)
			if err != nil {
				t.Skip()
				return
			}
			got, err = got.Log()
			if err != nil {
				t.Skip()
				return
			}

			want, err := d.Log1p()
			if err != nil {
				t.Skip()
				return
			}

			if cmp, err := cmpULP(got, want, 5); err != nil {
				t.Errorf("cmpULP(%q, %q) failed: %v", got, want, err)
				return
			} else if cmp != 0 {
				t.Errorf("%q.Add(1).Log() = %q, whereas %q.Log1p() = %q", d, got, d, want)
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

func FuzzDecimal_Sub_Cmp(f *testing.F) {
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

			f, err := d.Sub(e)
			if err != nil {
				t.Skip()
				return
			}
			got := f.Sign()

			want := d.Cmp(e)

			if want != got {
				t.Errorf("%q.Cmp(%q) = %v, whereas %q.Sub(%q).Sign() = %v", d, e, want, d, e, got)
				return
			}
		},
	)
}

func FuzzDecimal_New(f *testing.F) {
	for _, d := range corpus {
		f.Add(d.neg, d.scale, d.coef)
	}

	toBint := func(coef uint64) *bint {
		b := new(big.Int)
		b.SetUint64(coef)
		return (*bint)(b)
	}

	f.Fuzz(
		func(t *testing.T, neg bool, scale int, coef uint64) {
			got, err := newFromFint(neg, fint(coef), scale, 0)
			if err != nil {
				t.Skip()
				return
			}

			want, err := newFromBint(neg, toBint(coef), scale, 0)
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

func FuzzDecimal_Pad(f *testing.F) {
	for _, d := range corpus {
		for s := range MaxScale + 1 {
			f.Add(d.neg, d.scale, d.coef, s)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, scale int) {
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}

			got := d.Pad(scale)

			want := d

			if got.Cmp(want) != 0 {
				t.Errorf("%q.Pad(%v) = %q", d, scale, got)
				return
			}
			if got.Scale() > MaxScale {
				t.Errorf("%q.Pad(%v).Scale() = %v", d, scale, got.Scale())
				return
			}
		},
	)
}

func FuzzDecimal_Trim(f *testing.F) {
	for _, d := range corpus {
		for s := range MaxScale + 1 {
			f.Add(d.neg, d.scale, d.coef, s)
		}
	}

	f.Fuzz(
		func(t *testing.T, dneg bool, dscale int, dcoef uint64, scale int) {
			d, err := newSafe(dneg, fint(dcoef), dscale)
			if err != nil {
				t.Skip()
				return
			}

			got := d.Trim(scale)

			want := d

			if got.Cmp(want) != 0 {
				t.Errorf("%q.Trim(%v) = %q", d, scale, got)
				return
			}
			if got.Scale() < MinScale {
				t.Errorf("%q.Trim(%v).Scale() = %v", d, scale, got.Scale())
				return
			}
		},
	)
}
