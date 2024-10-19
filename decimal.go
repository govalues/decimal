package decimal

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// Decimal represents a finite floating-point decimal number.
// Its zero value corresponds to the numeric value of 0.
// Decimal is designed to be safe for concurrent use by multiple goroutines.
type Decimal struct {
	neg   bool // indicates whether the decimal is negative
	scale int8 // position of the floating decimal point
	coef  fint // numeric value without decimal point
}

const (
	MaxPrec  = 19      // MaxPrec is a maximum length of the coefficient in decimal digits.
	MinScale = 0       // MinScale is a minimum number of digits after the decimal point.
	MaxScale = 19      // MaxScale is a maximum number of digits after the decimal point.
	maxCoef  = maxFint // maxCoef is a maximum absolute value of the coefficient, which is equal to (10^MaxPrec - 1).
)

var (
	NegOne              = MustNew(-1, 0)                         // NegOne represents the decimal value of -1.
	Zero                = MustNew(0, 0)                          // Zero represents the decimal value of 0. For comparison purposes, use IsZero method.
	One                 = MustNew(1, 0)                          // One represents the decimal value of 1.
	Two                 = MustNew(2, 0)                          // Two represents the decimal value of 2.
	Ten                 = MustNew(10, 0)                         // Ten represents the decimal value of 10.
	Hundred             = MustNew(100, 0)                        // Hundred represents the decimal value of 100.
	Thousand            = MustNew(1_000, 0)                      // Thousand represents the decimal value of 1,000.
	E                   = MustNew(2_718_281_828_459_045_235, 18) // E represents Euler’s number rounded to 18 digits.
	Pi                  = MustNew(3_141_592_653_589_793_238, 18) // Pi represents the value of π rounded to 18 digits.
	errDecimalOverflow  = errors.New("decimal overflow")
	errInvalidDecimal   = errors.New("invalid decimal")
	errScaleRange       = errors.New("scale out of range")
	errInvalidOperation = errors.New("invalid operation")
	errInexactDivision  = errors.New("inexact division")
	errDivisionByZero   = errors.New("division by zero")
)

// newUnsafe creates a new decimal without checking scale and coefficient.
// Use it only if you are absolutely sure that the arguments are valid.
func newUnsafe(neg bool, coef fint, scale int) Decimal {
	if coef == 0 {
		neg = false
	}
	//nolint:gosec
	return Decimal{neg: neg, coef: coef, scale: int8(scale)}
}

// newSafe creates a new decimal and checks scale and coefficient.
func newSafe(neg bool, coef fint, scale int) (Decimal, error) {
	switch {
	case scale < MinScale || scale > MaxScale:
		return Decimal{}, errScaleRange
	case coef > maxCoef:
		return Decimal{}, errDecimalOverflow
	}
	return newUnsafe(neg, coef, scale), nil
}

// newFromFint creates a new decimal from uint64 coefficient.
// This method does not use overflowError to return descriptive errors,
// as it must be as fast as possible.
func newFromFint(neg bool, coef fint, scale, minScale int) (Decimal, error) {
	var ok bool
	// Scale normalization
	switch {
	case scale < minScale:
		coef, ok = coef.lsh(minScale - scale)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		scale = minScale
	case scale > MaxScale:
		coef = coef.rshHalfEven(scale - MaxScale)
		scale = MaxScale
	}
	return newSafe(neg, coef, scale)
}

// newFromBint creates a new decimal from *big.Int coefficient.
// This method uses overflowError to return descriptive errors.
func newFromBint(neg bool, coef *bint, scale, minScale int) (Decimal, error) {
	// Overflow validation
	prec := coef.prec()
	if prec-scale > MaxPrec-minScale {
		return Decimal{}, overflowError(prec, scale, minScale)
	}
	// Scale normalization
	switch {
	case scale < minScale:
		coef.lsh(coef, minScale-scale)
		scale = minScale
	case scale >= prec && scale > MaxScale: // no integer part
		coef.rshHalfEven(coef, scale-MaxScale)
		scale = MaxScale
	case prec > scale && prec > MaxPrec: // there is an integer part
		coef.rshHalfEven(coef, prec-MaxPrec)
		scale = MaxPrec - prec + scale
	}
	// Handling the rare case when rshHalfEven rounded
	// a 19-digit coefficient to a 20-digit coefficient.
	if coef.hasPrec(MaxPrec + 1) {
		return newFromBint(neg, coef, scale, minScale)
	}
	return newSafe(neg, coef.fint(), scale)
}

func overflowError(gotPrec, gotScale, wantScale int) error {
	maxDigits := MaxPrec - wantScale
	gotDigits := gotPrec - gotScale
	switch wantScale {
	case 0:
		return fmt.Errorf("%w: the integer part of a %T can have at most %v digits, but it has %v digits", errDecimalOverflow, Decimal{}, maxDigits, gotDigits)
	default:
		return fmt.Errorf("%w: with %v significant digits after the decimal point, the integer part of a %T can have at most %v digits, but it has %v digits", errDecimalOverflow, wantScale, Decimal{}, maxDigits, gotDigits)
	}
}

func unknownOverflowError(wantScale int) error {
	maxDigits := MaxPrec - wantScale
	switch wantScale {
	case 0:
		return fmt.Errorf("%w: the integer part of a %T can have at most %v digits, but it has significantly more digits", errDecimalOverflow, Decimal{}, maxDigits)
	default:
		return fmt.Errorf("%w: with %v significant digits after the decimal point, the integer part of a %T can have at most %v digits, but it has significantly more digits", errDecimalOverflow, wantScale, Decimal{}, maxDigits)
	}
}

// New returns a decimal equal to coef / 10^scale.
// New keeps trailing zeros in the fractional part to preserve scale.
//
// New returns an error if scale is negative or greater than [MaxScale].
func New(coef int64, scale int) (Decimal, error) {
	var neg bool
	if coef < 0 {
		neg = true
		coef = -coef
	}
	// nolint:gosec
	return newSafe(neg, fint(coef), scale)
}

// MustNew is like [New] but panics if the decimal cannot be constructed.
// It simplifies safe initialization of global variables holding decimals.
func MustNew(coef int64, scale int) Decimal {
	d, err := New(coef, scale)
	if err != nil {
		panic(fmt.Sprintf("New(%v, %v) failed: %v", coef, scale, err))
	}
	return d
}

// NewFromInt64 converts a pair of integers, representing the whole and
// fractional parts, to a (possibly rounded) decimal equal to whole + frac / 10^scale.
// NewFromInt64 removes all trailing zeros from the fractional part.
// This method is useful for converting amounts from [protobuf] format.
// See also method [Decimal.Int64].
//
// NewFromInt64 returns an error if:
//   - the whole and fractional parts have different signs;
//   - the scale is negative or greater than [MaxScale];
//   - frac / 10^scale is not within the range (-1, 1).
//
// [protobuf]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func NewFromInt64(whole, frac int64, scale int) (Decimal, error) {
	// Whole
	d, err := New(whole, 0)
	if err != nil {
		return Decimal{}, fmt.Errorf("converting integers: %w", err)
	}
	// Fraction
	f, err := New(frac, scale)
	if err != nil {
		return Decimal{}, fmt.Errorf("converting integers: %w", err)
	}
	if !f.IsZero() {
		if !d.IsZero() && d.Sign() != f.Sign() {
			return Decimal{}, fmt.Errorf("converting integers: inconsistent signs")
		}
		if !f.WithinOne() {
			return Decimal{}, fmt.Errorf("converting integers: inconsistent fraction")
		}
		f = f.Trim(0)
		d, err = d.Add(f)
		if err != nil {
			return Decimal{}, fmt.Errorf("converting integers: %w", err)
		}
	}
	return d, nil
}

// NewFromFloat64 converts a float to a (possibly rounded) decimal.
// See also method [Decimal.Float64].
//
// NewFromFloat64 returns an error if:
//   - the float is a special value (NaN or Inf);
//   - the integer part of the result has more than [MaxPrec] digits.
func NewFromFloat64(f float64) (Decimal, error) {
	// Float
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return Decimal{}, fmt.Errorf("converting float: special value %v", f)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	// Decimal
	d, err := Parse(s)
	if err != nil {
		return Decimal{}, fmt.Errorf("converting float: %w", err)
	}
	return d, nil
}

// Zero returns a decimal with a value of 0, having the same scale as decimal d.
// See also methods [Decimal.One], [Decimal.ULP].
func (d Decimal) Zero() Decimal {
	return newUnsafe(false, 0, d.Scale())
}

// One returns a decimal with a value of 1, having the same scale as decimal d.
// See also methods [Decimal.Zero], [Decimal.ULP].
func (d Decimal) One() Decimal {
	return newUnsafe(false, pow10[d.Scale()], d.Scale())
}

// ULP (Unit in the Last Place) returns the smallest representable positive
// difference between two decimals with the same scale as decimal d.
// It can be useful for implementing rounding and comparison algorithms.
// See also methods [Decimal.Zero], [Decimal.One].
func (d Decimal) ULP() Decimal {
	return newUnsafe(false, 1, d.Scale())
}

// Parse converts a string to a (possibly rounded) decimal.
// The input string must be in one of the following formats:
//
//	1.234
//	-1234
//	+0.000001234
//	1.83e5
//	0.22e-9
//
// The formal EBNF grammar for the supported format is as follows:
//
//	sign           ::= '+' | '-'
//	digits         ::= { '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' }
//	significand    ::= digits '.' digits | '.' digits | digits '.' | digits
//	exponent       ::= ('e' | 'E') [sign] digits
//	numeric-string ::= [sign] significand [exponent]
//
// Parse removes leading zeros from the integer part of the input string,
// but tries to maintain trailing zeros in the fractional part to preserve scale.
//
// Parse returns an error if:
//   - the string contains any whitespaces;
//   - the string is longer than 330 bytes;
//   - the exponent is less than -330 or greater than 330;
//   - the string does not represent a valid decimal number;
//   - the integer part of the result has more than [MaxPrec] digits.
func Parse(s string) (Decimal, error) {
	return ParseExact(s, 0)
}

// ParseExact is similar to [Parse], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for parsing monetary amounts, where the scale should be
// equal to or greater than the currency's scale.
func ParseExact(s string, scale int) (Decimal, error) {
	if len(s) > 330 {
		return Decimal{}, fmt.Errorf("parsing decimal: %w", errInvalidDecimal)
	}
	if scale < MinScale || scale > MaxScale {
		return Decimal{}, fmt.Errorf("parsing decimal: %w", errScaleRange)
	}
	d, err := parseFint(s, scale)
	if err != nil {
		d, err = parseBint(s, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("parsing decimal: %w", err)
		}
	}
	return d, nil
}

// parseFint parses a decimal string using uint64 arithmetic.
// parseFint does not support exponential notation to make it as fast as possible.
//
//nolint:gocyclo
func parseFint(s string, minScale int) (Decimal, error) {
	var pos int
	width := len(s)

	// Sign
	var neg bool
	switch {
	case pos == width:
		// skip
	case s[pos] == '-':
		neg = true
		pos++
	case s[pos] == '+':
		pos++
	}

	// Coefficient
	var coef fint
	var scale int
	var hasCoef, ok bool

	// Integer
	for pos < width && s[pos] >= '0' && s[pos] <= '9' {
		coef, ok = coef.fsa(1, s[pos]-'0')
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		pos++
		hasCoef = true
	}

	// Fraction
	if pos < width && s[pos] == '.' {
		pos++
		for pos < width && s[pos] >= '0' && s[pos] <= '9' {
			coef, ok = coef.fsa(1, s[pos]-'0')
			if !ok {
				return Decimal{}, errDecimalOverflow
			}
			pos++
			scale++
			hasCoef = true
		}
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("%w: unexpected character %q", errInvalidDecimal, s[pos])
	}
	if !hasCoef {
		return Decimal{}, fmt.Errorf("%w: no coefficient", errInvalidDecimal)
	}
	return newFromFint(neg, coef, scale, minScale)
}

// parseBint parses a decimal string using *big.Int arithmetic.
// parseBint supports exponential notation.
//
//nolint:gocyclo
func parseBint(s string, minScale int) (Decimal, error) {
	var pos int
	width := len(s)

	// Sign
	var neg bool
	switch {
	case pos == width:
		// skip
	case s[pos] == '-':
		neg = true
		pos++
	case s[pos] == '+':
		pos++
	}

	// Coefficient
	bcoef := getBint()
	defer putBint(bcoef)
	bcoef.setFint(0)
	var fcoef fint
	var shift, scale int
	var hasCoef, ok bool

	// Algorithm:
	// 	1. Add as many digits as possible to the uint64 coefficient (fast).
	// 	2. Once the uint64 coefficient has reached its maximum value,
	//     add it to the *big.Int coefficient (slow).
	// 	3. Repeat until all digits are processed.

	// Integer
	for pos < width && s[pos] >= '0' && s[pos] <= '9' {
		fcoef, ok = fcoef.fsa(1, s[pos]-'0')
		if !ok {
			return Decimal{}, errDecimalOverflow // Should never happen
		}
		pos++
		shift++
		hasCoef = true
		if fcoef.hasPrec(MaxPrec) {
			bcoef.fsa(bcoef, shift, fcoef)
			fcoef, shift = 0, 0
		}
	}

	// Fraction
	if pos < width && s[pos] == '.' {
		pos++
		for pos < width && s[pos] >= '0' && s[pos] <= '9' {
			fcoef, ok = fcoef.fsa(1, s[pos]-'0')
			if !ok {
				return Decimal{}, errDecimalOverflow // Should never happen
			}
			pos++
			scale++
			shift++
			hasCoef = true
			if fcoef.hasPrec(MaxPrec) {
				bcoef.fsa(bcoef, shift, fcoef)
				fcoef, shift = 0, 0
			}
		}
	}
	if shift > 0 {
		bcoef.fsa(bcoef, shift, fcoef)
	}

	// Exponent
	var exp int
	var eneg, hasExp, hasE bool
	if pos < width && (s[pos] == 'e' || s[pos] == 'E') {
		pos++
		hasE = true
		// Sign
		switch {
		case pos == width:
			// skip
		case s[pos] == '-':
			eneg = true
			pos++
		case s[pos] == '+':
			pos++
		}
		// Integer
		for pos < width && s[pos] >= '0' && s[pos] <= '9' {
			exp = exp*10 + int(s[pos]-'0')
			if exp > 330 {
				return Decimal{}, errInvalidDecimal
			}
			pos++
			hasExp = true
		}
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("%w: unexpected character %q", errInvalidDecimal, s[pos])
	}
	if !hasCoef {
		return Decimal{}, fmt.Errorf("%w: no coefficient", errInvalidDecimal)
	}
	if hasE && !hasExp {
		return Decimal{}, fmt.Errorf("%w: no exponent", errInvalidDecimal)
	}

	if eneg {
		scale = scale + exp
	} else {
		scale = scale - exp
	}

	return newFromBint(neg, bcoef, scale, minScale)
}

// MustParse is like [Parse] but panics if the string cannot be parsed.
// It simplifies safe initialization of global variables holding decimals.
func MustParse(s string) Decimal {
	d, err := Parse(s)
	if err != nil {
		panic(fmt.Sprintf("Parse(%q) failed: %v", s, err))
	}
	return d
}

// String implements the [fmt.Stringer] interface and returns
// a string representation of the decimal.
// The returned string does not use scientific or engineering notation and is
// formatted according to the following formal EBNF grammar:
//
//	sign           ::= '-'
//	digits         ::= { '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' }
//	significand    ::= digits '.' digits | digits
//	numeric-string ::= [sign] significand
//
// See also method [Decimal.Format].
//
// [fmt.Stringer]: https://pkg.go.dev/fmt#Stringer
func (d Decimal) String() string {
	var buf [24]byte
	pos := len(buf) - 1
	coef := d.Coef()
	scale := d.Scale()

	// Coefficient
	for {
		buf[pos] = byte(coef%10) + '0'
		pos--
		coef /= 10
		if scale > 0 {
			scale--
			// Decimal point
			if scale == 0 {
				buf[pos] = '.'
				pos--
				// Leading 0
				if coef == 0 {
					buf[pos] = '0'
					pos--
				}
			}
		}
		if coef == 0 && scale == 0 {
			break
		}
	}

	// Sign
	if d.IsNeg() {
		buf[pos] = '-'
		pos--
	}

	return string(buf[pos+1:])
}

// parseBCD converts a [packed BCD] representation to a decimal.
//
// [packed BCD]: https://en.wikipedia.org/wiki/Binary-coded_decimal#Packed_BCD
func parseBCD(b []byte) (Decimal, error) {
	var pos int
	width := len(b)

	// Coefficient and sign
	var neg bool
	var coef fint
	var ok bool
	for pos < width {
		hi := b[pos] >> 4
		lo := b[pos] & 0x0f

		if hi > 9 {
			return Decimal{}, fmt.Errorf("%w: invalid high nibble \"%x\"", errInvalidDecimal, b[pos])
		}
		coef, ok = coef.fsa(1, hi)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}

		if lo > 9 {
			if lo == 0x0d {
				neg = true
			} else if lo != 0x0c {
				return Decimal{}, fmt.Errorf("%w: invalid low nibble \"%x\"", errInvalidDecimal, b[pos])
			}
			pos++
			break
		}
		coef, ok = coef.fsa(1, lo)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		pos++
	}

	// Scale
	var scale int
	var hasScale bool
	if pos < width {
		hi := b[pos] >> 4
		lo := b[pos] & 0x0f
		hasScale = true

		if hi > 1 {
			return Decimal{}, fmt.Errorf("%w: invalid high nibble \"%x\"", errInvalidDecimal, b[pos])
		}
		scale = int(hi) * 10

		if lo > 9 {
			return Decimal{}, fmt.Errorf("%w: invalid low nibble \"%x\"", errInvalidDecimal, b[pos])
		}
		scale += int(lo)

		pos++
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("%w: unexpected byte \"%x\"", errInvalidDecimal, b[pos])
	}
	if !hasScale {
		return Decimal{}, fmt.Errorf("%w: no scale", errInvalidDecimal)
	}

	return newSafe(neg, coef, scale)
}

// bcd returns a [packed BCD] representation of a decimal.
//
// [packed BCD]: https://en.wikipedia.org/wiki/Binary-coded_decimal#Packed_BCD
func (d Decimal) bcd() []byte {
	var buf [11]byte
	pos := len(buf) - 1
	coef := d.Coef()
	scale := d.Scale()

	// Scale
	buf[pos] = byte(scale/10)<<4 | byte(scale%10)
	pos--

	// Sign and first digit
	if d.IsNeg() {
		buf[pos] = byte(coef%10)<<4 | 0x0d
	} else {
		buf[pos] = byte(coef%10)<<4 | 0x0c
	}
	pos--
	coef /= 10

	// Coefficient
	for coef > 0 {
		buf[pos] = byte(coef/10%10)<<4 | byte(coef%10)
		pos--
		coef /= 100
	}

	return buf[pos+1:]
}

// Float64 returns the nearest binary floating-point number rounded
// using [rounding half to even] (banker's rounding).
// See also constructor [NewFromFloat64].
//
// This conversion may lose data, as float64 has a smaller precision
// than the decimal type.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (d Decimal) Float64() (f float64, ok bool) {
	s := d.String()
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// Int64 returns a pair of integers representing the whole and
// (possibly rounded) fractional parts of the decimal.
// If given scale is greater than the scale of the decimal, then the fractional part
// is zero-padded to the right.
// If given scale is smaller than the scale of the decimal, then the fractional part
// is rounded using [rounding half to even] (banker's rounding).
// The relationship between the decimal and the returned values can be expressed
// as d = whole + frac / 10^scale.
// This method is useful for converting amounts to [protobuf] format.
// See also constructor [NewFromInt64].
//
// If the result cannot be represented as a pair of int64 values,
// then false is returned.
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
// [protobuf]: https://github.com/googleapis/googleapis/blob/master/google/type/money.proto
func (d Decimal) Int64(scale int) (whole, frac int64, ok bool) {
	if scale < MinScale || scale > MaxScale {
		return 0, 0, false
	}
	x := d.coef
	y := pow10[d.Scale()]
	if scale < d.Scale() {
		x = x.rshHalfEven(d.Scale() - scale)
		y = pow10[scale]
	}
	q, r, ok := x.quoRem(y)
	if !ok {
		return 0, 0, false // Should never happen
	}
	if scale > d.Scale() {
		r, ok = r.lsh(scale - d.Scale())
		if !ok {
			return 0, 0, false // Should never happen
		}
	}
	if d.IsNeg() {
		if q > -math.MinInt64 || r > -math.MinInt64 {
			return 0, 0, false
		}
		//nolint:gosec
		return -int64(q), -int64(r), true
	}
	if q > math.MaxInt64 || r > math.MaxInt64 {
		return 0, 0, false
	}
	//nolint:gosec
	return int64(q), int64(r), true
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
// See also constructor [Parse].
//
// [encoding.TextUnmarshaler]: https://pkg.go.dev/encoding#TextUnmarshaler
func (d *Decimal) UnmarshalText(text []byte) error {
	var err error
	*d, err = Parse(string(text))
	return err
}

// MarshalText implements the [encoding.TextMarshaler] interface.
// See also method [Decimal.String].
//
// [encoding.TextMarshaler]: https://pkg.go.dev/encoding#TextMarshaler
func (d Decimal) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
//
// [encoding.BinaryUnmarshaler]: https://pkg.go.dev/encoding#BinaryUnmarshaler
func (d *Decimal) UnmarshalBinary(data []byte) error {
	var err error
	*d, err = parseBCD(data)
	return err
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
//
// [encoding.BinaryMarshaler]: https://pkg.go.dev/encoding#BinaryMarshaler
func (d Decimal) MarshalBinary() ([]byte, error) {
	return d.bcd(), nil
}

// Scan implements the [sql.Scanner] interface.
// See also constructor [Parse].
//
// [sql.Scanner]: https://pkg.go.dev/database/sql#Scanner
func (d *Decimal) Scan(value any) error {
	var err error
	switch value := value.(type) {
	case string:
		*d, err = Parse(value)
	case []byte:
		*d, err = Parse(string(value))
	case int64:
		*d, err = New(value, 0)
	case float64:
		*d, err = NewFromFloat64(value)
	case nil:
		err = fmt.Errorf("converting to %T: nil is not supported", d)
	default:
		err = fmt.Errorf("converting from %T to %T: type %T is not supported", value, d, value)
	}
	return err
}

// Value implements the [driver.Valuer] interface.
// See also method [Decimal.String].
//
// [driver.Valuer]: https://pkg.go.dev/database/sql/driver#Valuer
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// Format implements the [fmt.Formatter] interface.
// The following [format verbs] are available:
//
//	| Verb       | Example | Description    |
//	| ---------- | ------- | -------------- |
//	| %f, %s, %v | 5.67    | Decimal        |
//	| %q         | "5.67"  | Quoted decimal |
//	| %k         | 567%    | Percentage     |
//
// The following format flags can be used with all verbs: '+', ' ', '0', '-'.
//
// Precision is only supported for %f and %k verbs.
// For %f verb, the default precision is equal to the actual scale of the decimal,
// whereas, for verb %k the default precision is the actual scale of the decimal minus 2.
//
// [format verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
//
//nolint:gocyclo
func (d Decimal) Format(state fmt.State, verb rune) {
	var err error

	// Percentage multiplier
	if verb == 'k' || verb == 'K' {
		d, err = d.Mul(Hundred)
		if err != nil {
			// This panic is handled inside the fmt package.
			panic(fmt.Errorf("formatting percent: %w", err))
		}
	}

	// Rescaling
	var tzeros int
	if verb == 'f' || verb == 'F' || verb == 'k' || verb == 'K' {
		var scale int
		switch p, ok := state.Precision(); {
		case ok:
			scale = p
		case verb == 'k' || verb == 'K':
			scale = d.Scale() - 2
		case verb == 'f' || verb == 'F':
			scale = d.Scale()
		}
		scale = max(scale, MinScale)
		switch {
		case scale < d.Scale():
			d = d.Round(scale)
		case scale > d.Scale():
			tzeros = scale - d.Scale()
		}
	}

	// Integer and fractional digits
	var intdigs int
	fracdigs := d.Scale()
	if dprec := d.Prec(); dprec > fracdigs {
		intdigs = dprec - fracdigs
	}
	if d.WithinOne() {
		intdigs++ // leading 0
	}

	// Decimal point
	var dpoint int
	if fracdigs > 0 || tzeros > 0 {
		dpoint = 1
	}

	// Arithmetic sign
	var rsign int
	if d.IsNeg() || state.Flag('+') || state.Flag(' ') {
		rsign = 1
	}

	// Percentage sign
	var psign int
	if verb == 'k' || verb == 'K' {
		psign = 1
	}

	// Openning and closing quotes
	var lquote, tquote int
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Calculating padding
	width := lquote + rsign + intdigs + dpoint + fracdigs + tzeros + psign + tquote
	var lspaces, tspaces, lzeros int
	if w, ok := state.Width(); ok && w > width {
		switch {
		case state.Flag('-'):
			tspaces = w - width
		case state.Flag('0'):
			lzeros = w - width
		default:
			lspaces = w - width
		}
		width = w
	}

	buf := make([]byte, width)
	pos := width - 1

	// Trailing spaces
	for range tspaces {
		buf[pos] = ' '
		pos--
	}

	// Closing quote
	for range tquote {
		buf[pos] = '"'
		pos--
	}

	// Percentage sign
	for range psign {
		buf[pos] = '%'
		pos--
	}

	// Trailing zeros
	for range tzeros {
		buf[pos] = '0'
		pos--
	}

	// Fractional digits
	dcoef := d.Coef()
	for range fracdigs {
		buf[pos] = byte(dcoef%10) + '0'
		pos--
		dcoef /= 10
	}

	// Decimal point
	for range dpoint {
		buf[pos] = '.'
		pos--
	}

	// Integer digits
	for range intdigs {
		buf[pos] = byte(dcoef%10) + '0'
		pos--
		dcoef /= 10
	}

	// Leading zeros
	for range lzeros {
		buf[pos] = '0'
		pos--
	}

	// Arithmetic sign
	for range rsign {
		if d.IsNeg() {
			buf[pos] = '-'
		} else if state.Flag(' ') {
			buf[pos] = ' '
		} else {
			buf[pos] = '+'
		}
		pos--
	}

	// Opening quote
	for range lquote {
		buf[pos] = '"'
		pos--
	}

	// Leading spaces
	for range lspaces {
		buf[pos] = ' '
		pos--
	}

	// Writing result
	//nolint:errcheck
	switch verb {
	case 'q', 'Q', 's', 'S', 'v', 'V', 'f', 'F', 'k', 'K':
		state.Write(buf)
	default:
		state.Write([]byte("%!"))
		state.Write([]byte{byte(verb)})
		state.Write([]byte("(decimal.Decimal="))
		state.Write(buf)
		state.Write([]byte(")"))
	}
}

// Prec returns the number of digits in the coefficient.
// See also method [Decimal.Coef].
func (d Decimal) Prec() int {
	return d.coef.prec()
}

// Coef returns the coefficient of the decimal.
// See also method [Decimal.Prec].
func (d Decimal) Coef() uint64 {
	return uint64(d.coef)
}

// Scale returns the number of digits after the decimal point.
// See also methods [Decimal.Prec], [Decimal.MinScale].
func (d Decimal) Scale() int {
	return int(d.scale)
}

// MinScale returns the smallest scale that the decimal can be rescaled to
// without rounding.
// See also method [Decimal.Trim].
func (d Decimal) MinScale() int {
	// Special case: zero
	if d.IsZero() {
		return MinScale
	}
	// General case
	dcoef := d.coef
	return max(MinScale, d.Scale()-dcoef.ntz())
}

// IsInt returns true if there are no significant digits after the decimal point.
func (d Decimal) IsInt() bool {
	return d.coef%pow10[d.Scale()] == 0
}

// IsOne returns:
//
//	true  if d = -1 or d = 1
//	false otherwise
func (d Decimal) IsOne() bool {
	return d.coef == pow10[d.Scale()]
}

// WithinOne returns:
//
//	true  if -1 < d < 1
//	false otherwise
func (d Decimal) WithinOne() bool {
	return d.coef < pow10[d.Scale()]
}

// Round returns a decimal rounded to the specified number of digits after
// the decimal point using [rounding half to even] (banker's rounding).
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also method [Decimal.Rescale].
//
// [rounding half to even]: https://en.wikipedia.org/wiki/Rounding#Rounding_half_to_even
func (d Decimal) Round(scale int) Decimal {
	scale = max(scale, MinScale)
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	coef = coef.rshHalfEven(d.Scale() - scale)
	return newUnsafe(d.IsNeg(), coef, scale)
}

// Pad returns a decimal zero-padded to the specified number of digits after
// the decimal point.
// The total number of digits in the result is limited by [MaxPrec].
// See also method [Decimal.Trim].
func (d Decimal) Pad(scale int) Decimal {
	scale = min(scale, MaxScale, MaxPrec-d.Prec()+d.Scale())
	if scale <= d.Scale() {
		return d
	}
	coef := d.coef
	coef, ok := coef.lsh(scale - d.Scale())
	if !ok {
		return d // Should never happen
	}
	return newUnsafe(d.IsNeg(), coef, scale)
}

// Rescale returns a decimal rounded or zero-padded to the given number of digits
// after the decimal point.
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also methods [Decimal.Round], [Decimal.Pad].
func (d Decimal) Rescale(scale int) Decimal {
	if scale > d.Scale() {
		return d.Pad(scale)
	}
	return d.Round(scale)
}

// Quantize returns a decimal rescaled to the same scale as decimal e.
// The sign and the coefficient of decimal e are ignored.
// See also methods [Decimal.SameScale] and [Decimal.Rescale].
func (d Decimal) Quantize(e Decimal) Decimal {
	return d.Rescale(e.Scale())
}

// SameScale returns true if decimals have the same scale.
// See also methods [Decimal.Scale], [Decimal.Quantize].
func (d Decimal) SameScale(e Decimal) bool {
	return d.Scale() == e.Scale()
}

// Trunc returns a decimal truncated to the specified number of digits
// after the decimal point using [rounding toward zero].
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
//
// [rounding toward zero]: https://en.wikipedia.org/wiki/Rounding#Rounding_toward_zero
func (d Decimal) Trunc(scale int) Decimal {
	scale = max(scale, MinScale)
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	coef = coef.rshDown(d.Scale() - scale)
	return newUnsafe(d.IsNeg(), coef, scale)
}

// Trim returns a decimal with trailing zeros removed up to the given number of
// digits after the decimal point.
// If the given scale is negative, it is redefined to zero.
// See also method [Decimal.Pad].
func (d Decimal) Trim(scale int) Decimal {
	if d.Scale() <= scale {
		return d
	}
	scale = max(scale, d.MinScale())
	return d.Trunc(scale)
}

// Ceil returns a decimal rounded up to the given number of digits
// after the decimal point using [rounding toward positive infinity].
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also method [Decimal.Floor].
//
// [rounding toward positive infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_up
func (d Decimal) Ceil(scale int) Decimal {
	scale = max(scale, MinScale)
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	if d.IsNeg() {
		coef = coef.rshDown(d.Scale() - scale)
	} else {
		coef = coef.rshUp(d.Scale() - scale)
	}
	return newUnsafe(d.IsNeg(), coef, scale)
}

// Floor returns a decimal rounded down to the specified number of digits
// after the decimal point using [rounding toward negative infinity].
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also method [Decimal.Ceil].
//
// [rounding toward negative infinity]: https://en.wikipedia.org/wiki/Rounding#Rounding_down
func (d Decimal) Floor(scale int) Decimal {
	scale = max(scale, MinScale)
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	if d.IsNeg() {
		coef = coef.rshUp(d.Scale() - scale)
	} else {
		coef = coef.rshDown(d.Scale() - scale)
	}
	return newUnsafe(d.IsNeg(), coef, scale)
}

// Neg returns a decimal with the opposite sign.
func (d Decimal) Neg() Decimal {
	return newUnsafe(!d.IsNeg(), d.coef, d.Scale())
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	return newUnsafe(false, d.coef, d.Scale())
}

// CopySign returns a decimal with the same sign as decimal e.
// CopySign treates 0 as positive.
// See also method [Decimal.Sign].
func (d Decimal) CopySign(e Decimal) Decimal {
	if d.IsNeg() == e.IsNeg() {
		return d
	}
	return d.Neg()
}

// Sign returns:
//
//	-1 if d < 0
//	 0 if d = 0
//	+1 if d > 0
//
// See also methods [Decimal.IsPos], [Decimal.IsNeg], [Decimal.IsZero].
func (d Decimal) Sign() int {
	switch {
	case d.neg:
		return -1
	case d.coef == 0:
		return 0
	}
	return 1
}

// IsPos returns:
//
//	true  if d > 0
//	false otherwise
func (d Decimal) IsPos() bool {
	return d.coef != 0 && !d.neg
}

// IsNeg returns:
//
//	true  if d < 0
//	false otherwise
func (d Decimal) IsNeg() bool {
	return d.neg
}

// IsZero returns:
//
//	true  if d = 0
//	false otherwise
func (d Decimal) IsZero() bool {
	return d.coef == 0
}

// Mul returns the (possibly rounded) product of decimals d and e.
//
// Mul returns an overflow error if the integer part of the result has
// more than [MaxPrec] digits.
func (d Decimal) Mul(e Decimal) (Decimal, error) {
	return d.MulExact(e, 0)
}

// MulExact is similar to [Decimal.Mul], but it allows you to specify the number
// of digits after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will
// return an overflow error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) MulExact(e Decimal, scale int) (Decimal, error) {
	if scale < MinScale || scale > MaxScale {
		return Decimal{}, fmt.Errorf("computing [%v * %v]: %w", d, e, errScaleRange)
	}

	// General case
	f, err := d.mulFint(e, scale)
	if err != nil {
		f, err = d.mulBint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v * %v]: %w", d, e, err)
		}
	}
	return f, nil
}

// mulFint computes the product of two decimals using uint64 arithmetic.
func (d Decimal) mulFint(e Decimal, minScale int) (Decimal, error) {
	dcoef, ecoef := d.coef, e.coef

	// Compute d = d * e
	dcoef, ok := dcoef.mul(ecoef)
	if !ok {
		return Decimal{}, errDecimalOverflow
	}
	neg := d.IsNeg() != e.IsNeg()
	scale := d.Scale() + e.Scale()

	return newFromFint(neg, dcoef, scale, minScale)
}

// mulBint computes the product of two decimals using *big.Int arithmetic.
func (d Decimal) mulBint(e Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	// Compute d = d * e
	dcoef.mul(dcoef, ecoef)
	neg := d.IsNeg() != e.IsNeg()
	scale := d.Scale() + e.Scale()

	return newFromBint(neg, dcoef, scale, minScale)
}

// Deprecated: use [Decimal.PowInt] instead.
// This method will change its signature in the v1.0 release.
func (d Decimal) Pow(power int) (Decimal, error) {
	return d.PowInt(power)
}

// PowInt returns the (possibly rounded) decimal raised to the given integer power.
// If zero is raised to zero power then the result is one.
//
// PowInt returns an error if:
//   - the integer part of the result has more than [MaxPrec] digits;
//   - zero is raised to a negative power.
func (d Decimal) PowInt(power int) (Decimal, error) {
	// Special case: zero to a negative power
	if power < 0 && d.IsZero() {
		return Decimal{}, fmt.Errorf("computing [%v^%v]: %w", d, power, errInvalidOperation)
	}

	// General case
	e, err := d.powIntFint(power)
	if err != nil {
		e, err = d.powIntBint(power)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v^%v]: %w", d, power, err)
		}
	}

	// Preferred scale
	if power < 0 {
		e = e.Trim(0)
	}

	return e, nil
}

// powIntFint computes the integer power of a decimal using uint64 arithmetic.
// powIntFint does not support negative powers.
func (d Decimal) powIntFint(power int) (Decimal, error) {
	dcoef := d.coef
	dneg := d.IsNeg()
	dscale := d.Scale()

	ecoef := One.coef
	eneg := One.IsNeg()
	escale := One.Scale()

	if power < 0 {
		return Decimal{}, errInvalidOperation
	}

	// Exponentiation by squaring
	var ok bool
	for power > 0 {
		if power%2 == 1 {
			power = power - 1

			// Compute e = e * d
			ecoef, ok = ecoef.mul(dcoef)
			if !ok {
				return Decimal{}, errDecimalOverflow
			}
			eneg = eneg != dneg
			escale = escale + dscale
		}
		if power > 0 {
			power = power / 2

			// Compute d = d * d
			dcoef, ok = dcoef.mul(dcoef)
			if !ok {
				return Decimal{}, errDecimalOverflow
			}
			dneg = false
			dscale = dscale * 2
		}
	}

	return newFromFint(eneg, ecoef, escale, 0)
}

// powIntBint computes the integer power of a decimal using *big.Int arithmetic.
// powIntBint supports negative powers.
func (d Decimal) powIntBint(power int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)
	dneg := d.IsNeg()
	dscale := d.Scale()

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(One.coef)
	eneg := One.IsNeg()
	escale := One.Scale()

	inv := false
	if power < 0 {
		power = -power
		inv = true
	}

	// Exponentiation by squaring
	for power > 0 {
		if power%2 == 1 {
			power = power - 1

			// Compute e = e * d
			ecoef.mul(ecoef, dcoef)
			eneg = eneg != dneg
			escale = escale + dscale

			// Intermediate truncation
			if escale > 3*MaxScale {
				shift := escale - 3*MaxScale
				ecoef.rshDown(ecoef, shift)
				escale = 3 * MaxScale
			}
		}
		if power > 0 {
			power = power / 2

			// Compute d = d * d
			dcoef.mul(dcoef, dcoef)
			dneg = false
			dscale = dscale * 2

			// Intermediate truncation
			if dscale > 3*MaxScale {
				shift := dscale - 3*MaxScale
				dcoef.rshDown(dcoef, shift)
				dscale = 3 * MaxScale
			}
		}
	}

	if inv {
		if ecoef.sign() == 0 {
			return Decimal{}, unknownOverflowError(0)
		}

		// Compute e = 1 / e
		ecoef.quo(bpow10[2*MaxScale+escale], ecoef)
		escale = 2 * MaxScale
	}

	return newFromBint(eneg, ecoef, escale, 0)
}

// Sqrt computes the square root of a decimal.
//
// Sqrt returns an error if the decimal is negative.
func (d Decimal) Sqrt() (Decimal, error) {
	// Special case: negative
	if d.IsNeg() {
		return Decimal{}, fmt.Errorf("computing sqrt(%v): %w", d, errInvalidOperation)
	}

	// Special case: zero
	if d.IsZero() {
		return newSafe(false, 0, d.Scale()/2)
	}

	// General case
	e, err := d.sqrtBint()
	if err != nil {
		return Decimal{}, fmt.Errorf("computing sqrt(%v): %w", d, err)
	}

	// Preferred scale
	e = e.Trim(d.Scale() / 2)

	return e, nil
}

// sqrtBint computes the square root of a decimal using *big.Int arithmetic.
func (d Decimal) sqrtBint() (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	escale := 2 * MaxScale

	fcoef := getBint()
	defer putBint(fcoef)
	fcoef.setFint(0)

	// Alignment
	dcoef.lsh(dcoef, 4*MaxScale-d.Scale())

	// Initial guess is calculated as 10^(n/2), where n is the position of
	// the most significant digit (n is negative if -1 < d < 1).
	n := dcoef.prec() - 4*MaxScale
	ecoef.setBint(bpow10[n/2+escale])

	// Newton's method
	for range 50 {
		if ecoef.cmp(fcoef) == 0 {
			break
		}
		fcoef.setBint(ecoef)
		ecoef.quo(dcoef, ecoef)
		ecoef.add(ecoef, fcoef)
		ecoef.hlf(ecoef)
	}

	return newFromBint(false, ecoef, escale, 0)
}

// Exp returns the (possibly rounded) exponential of a decimal.
//
// Exp returns an error if the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Exp() (Decimal, error) {
	// Special case: zero
	if d.IsZero() {
		return newSafe(false, 1, 0)
	}

	// General case
	e, err := d.expBint()
	if err != nil {
		return Decimal{}, fmt.Errorf("computing exp(%v): %w", d, err)
	}

	// Preferred scale
	e = e.Trim(0)

	return e, nil
}

// expBint computes exponential of a decimal using *big.Int arithmetic.
func (d Decimal) expBint() (Decimal, error) {
	dcoef := d.coef
	dscale := d.Scale()

	// Split |d| into integer part q and fractional part r
	q, r, ok := dcoef.quoRem(pow10[dscale])
	if !ok {
		return Decimal{}, errDecimalOverflow // Should never happen
	}

	// Check underflow and overflow
	if q >= fint(len(bexp)) {
		if d.IsNeg() {
			return newSafe(false, 0, 0)
		}
		return Decimal{}, unknownOverflowError(0)
	}

	// Retrieve e = exp(q) from precomputed cache
	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setBint(bexp[q])
	escale := 2 * MaxScale

	if r != 0 {
		// Compute f = exp(r) using Taylor series expansion
		fcoef := getBint()
		defer putBint(fcoef)
		fcoef.setFint(0)
		fscale := 2 * MaxScale

		rcoef := getBint()
		defer putBint(rcoef)
		rcoef.setFint(r)
		rscale := dscale

		gcoef := getBint()
		defer putBint(gcoef)
		gcoef.setBint(bpow10[2*MaxScale])
		gscale := 2 * MaxScale

		hcoef := getBint()
		defer putBint(hcoef)

		// Alignment
		if rscale < 2*MaxScale {
			rcoef.lsh(rcoef, 2*MaxScale-rscale)
			rscale = 2 * MaxScale
		}

		// Compute f = exp(r) = r^0 / 0! + r^1 / 1! + ... + r^n / n!
		for i := range len(bfact) {
			// Accumulate f = f + r^i / i!
			hcoef.quo(gcoef, bfact[i])
			if hcoef.sign() == 0 {
				break
			}
			fcoef.add(fcoef, hcoef)

			// Compute g = r^(i+1)
			gcoef.mul(gcoef, rcoef)
			gscale = gscale + rscale

			// Intermediate truncation
			if gscale > 2*MaxScale {
				shift := gscale - 2*MaxScale
				gcoef.rshDown(gcoef, shift)
				gscale = 2 * MaxScale
			}
		}

		// Compute exp(|d|) = exp(q) * exp(r)
		ecoef.mul(ecoef, fcoef)
		escale = escale + fscale

		// Intermediate truncation
		if escale > 2*MaxScale {
			shift := escale - 2*MaxScale
			ecoef.rshDown(ecoef, shift)
			escale = 2 * MaxScale
		}
	}

	if d.IsNeg() {
		if ecoef.sign() == 0 {
			return Decimal{}, unknownOverflowError(0)
		}

		// Compute exp(d) = 1 / exp(|d|)
		ecoef.quo(bpow10[2*MaxScale+escale], ecoef)
		escale = 2 * MaxScale
	}

	return newFromBint(false, ecoef, escale, 0)
}

// Log returns the (possibly rounded) natural logarithm of a decimal.
//
// Log returns an error if the decimal is zero or negative.
func (d Decimal) Log() (Decimal, error) {
	// Special case: zero or negative
	if !d.IsPos() {
		return Decimal{}, fmt.Errorf("computing log(%v): %w", d, errInvalidOperation)
	}

	// Special case: one
	if d.IsOne() {
		return newSafe(false, 0, 0)
	}

	// General case
	e, err := d.logBint()
	if err != nil {
		return Decimal{}, fmt.Errorf("computing log(%v): %w", d, err)
	}

	// Preferred scale
	e = e.Trim(0)

	return e, nil
}

// logBint computes the natural logarithm of a decimal using *big.Int arithmetic.
func (d Decimal) logBint() (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	escale := 2 * MaxScale

	fcoef := getBint()
	defer putBint(fcoef)
	fcoef.setFint(0)

	// Alignment and sign
	eneg := true
	if d.WithinOne() {
		dcoef.quo(bpow10[2*MaxScale+d.Scale()], dcoef)
	} else {
		dcoef.lsh(dcoef, 2*MaxScale-d.Scale())
		eneg = false
	}

	// The initial guess is calculated as n * ln(10),
	// where n is the position of the most significant digit.
	n := dcoef.prec() - 2*MaxScale
	ecoef.setBint(bnlog10[n])

	Ecoef := getBint()
	defer putBint(Ecoef)

	ncoef := getBint()
	defer putBint(ncoef)

	mcoef := getBint()
	defer putBint(mcoef)

	// Halley's method
	for range 50 {
		Ecoef.e(ecoef)

		ncoef.sub(Ecoef, dcoef)
		ncoef.dbl(ncoef)

		mcoef.add(Ecoef, dcoef)

		ncoef.lsh(ncoef, 2*MaxScale)
		ncoef.quo(ncoef, mcoef)

		fcoef.sub(ecoef, ncoef)

		if ecoef.cmp(fcoef) == 0 {
			break
		}

		ecoef.setBint(fcoef)
	}

	return newFromBint(eneg, ecoef, escale, 0)
}

// e computes the exponential of a decimal using *big.Int arithmetic.
// TODO: refactor to improve performance even more.
func (z *bint) e(x *bint) {
	qcoef := getBint()
	defer putBint(qcoef)

	rcoef := getBint()
	defer putBint(rcoef)
	rscale := 2 * MaxScale

	qcoef.quoRem(x, bpow10[rscale], rcoef)

	zcoef := getBint()
	defer putBint(zcoef)
	zcoef.setFint(0)

	gcoef := getBint()
	defer putBint(gcoef)
	gcoef.setBint(bpow10[2*MaxScale])
	gscale := 2 * MaxScale

	hcoef := getBint()
	defer putBint(hcoef)

	// Compute f = exp(r) = r^0 / 0! + r^1 / 1! + ... + r^n / n!
	for i := range len(bfact) {
		// Accumulate f = f + r^i / i!
		hcoef.quo(gcoef, bfact[i])
		if hcoef.sign() == 0 {
			break
		}
		zcoef.add(zcoef, hcoef)

		// Compute g = r^(i+1)
		gcoef.mul(gcoef, rcoef)
		gscale = gscale + rscale

		// Intermediate truncation
		if gscale > 2*MaxScale {
			shift := gscale - 2*MaxScale
			gcoef.rshDown(gcoef, shift)
			gscale = 2 * MaxScale
		}
	}

	// nolint:gosec
	zcoef.mul(zcoef, bexp[int(qcoef.fint())])
	zcoef.quo(zcoef, bpow10[2*MaxScale])

	z.setBint(zcoef)
}

// SubAbs returns the (possibly rounded) absolute difference between decimals d and e.
//
// SubAbs returns an error if the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) SubAbs(e Decimal) (Decimal, error) {
	f, err := d.Sub(e)
	if err != nil {
		return Decimal{}, fmt.Errorf("computing [abs(%v - %v)]: %w", d, e, err)
	}
	return f.Abs(), nil
}

// Sub returns the (possibly rounded) difference between decimals d and e.
//
// Sub returns an error if the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Sub(e Decimal) (Decimal, error) {
	return d.AddExact(e.Neg(), 0)
}

// SubExact is similar to [Decimal.Sub], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) SubExact(e Decimal, scale int) (Decimal, error) {
	return d.AddExact(e.Neg(), scale)
}

// Add returns the (possibly rounded) sum of decimals d and e.
//
// Add returns an error if the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Add(e Decimal) (Decimal, error) {
	return d.AddExact(e, 0)
}

// AddExact is similar to [Decimal.Add], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) AddExact(e Decimal, scale int) (Decimal, error) {
	if scale < MinScale || scale > MaxScale {
		return Decimal{}, fmt.Errorf("computing [%v + %v]: %w", d, e, errScaleRange)
	}

	// General case
	f, err := d.addFint(e, scale)
	if err != nil {
		f, err = d.addBint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v + %v]: %w", d, e, err)
		}
	}

	return f, nil
}

// addFint computes the sum of two decimals using uint64 arithmetic.
func (d Decimal) addFint(e Decimal, minScale int) (Decimal, error) {
	dcoef, ecoef := d.coef, e.coef

	// Alignment and scale
	var scale int
	var ok bool
	switch {
	case d.Scale() == e.Scale():
		scale = d.Scale()
	case d.Scale() > e.Scale():
		scale = d.Scale()
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	case d.Scale() < e.Scale():
		scale = e.Scale()
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	}

	// Sign
	var neg bool
	if ecoef < dcoef {
		neg = d.IsNeg()
	} else {
		neg = e.IsNeg()
	}

	// Compute d = d + e
	if d.IsNeg() != e.IsNeg() {
		dcoef = dcoef.subAbs(ecoef)
	} else {
		dcoef, ok = dcoef.add(ecoef)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	}

	return newFromFint(neg, dcoef, scale, minScale)
}

// addBint computes the sum of two decimals using *big.Int arithmetic.
func (d Decimal) addBint(e Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	// Alignment and scale
	var scale int
	switch {
	case d.Scale() == e.Scale():
		scale = d.Scale()
	case d.Scale() > e.Scale():
		scale = d.Scale()
		ecoef.lsh(ecoef, d.Scale()-e.Scale())
	case d.Scale() < e.Scale():
		scale = e.Scale()
		dcoef.lsh(dcoef, e.Scale()-d.Scale())
	}

	// Sign
	var neg bool
	if dcoef.cmp(ecoef) > 0 {
		neg = d.IsNeg()
	} else {
		neg = e.IsNeg()
	}

	// Compute d = d + e
	if d.IsNeg() != e.IsNeg() {
		dcoef.subAbs(dcoef, ecoef)
	} else {
		dcoef.add(dcoef, ecoef)
	}

	return newFromBint(neg, dcoef, scale, minScale)
}

// Deprecated: use [Decimal.AddMul] instead.
// Pay attention to the order of arguments, [Decimal.FMA] computes d * e + f,
// whereas [Decimal.AddMul] computes d + e * f.
// This method will be removed in the v1.0 release.
func (d Decimal) FMA(e, f Decimal) (Decimal, error) {
	return f.AddMulExact(d, e, 0)
}

// Deprecated: use [Decimal.AddMulExact] instead.
// Pay attention to the order of arguments, [Decimal.FMAExact] computes d * e + f,
// whereas [Decimal.AddMulExact] computes d + e * f.
// This method will be removed in the v1.0 release.
func (d Decimal) FMAExact(e, f Decimal, scale int) (Decimal, error) {
	return f.AddMulExact(d, e, scale)
}

// SubMul returns the (possibly rounded) [fused multiply-subtraction] of decimals d, e, and f.
// It computes d - e * f without any intermediate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of products, such as daily interest accrual.
//
// SubMul returns an error if the integer part of the result has more than [MaxPrec] digits.
//
// [fused multiply-subtraction]: https://en.wikipedia.org/wiki/Multiply%E2%80%93accumulate_operation#Fused_multiply%E2%80%93add
func (d Decimal) SubMul(e, f Decimal) (Decimal, error) {
	return d.AddMulExact(e.Neg(), f, 0)
}

// SubMulExact is similar to [Decimal.SubMul], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) SubMulExact(e, f Decimal, scale int) (Decimal, error) {
	return d.AddMulExact(e.Neg(), f, scale)
}

// AddMul returns the (possibly rounded) [fused multiply-addition] of decimals d, e, and f.
// It computes d + e * f without any intermediate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of products, such as daily interest accrual.
//
// AddMul returns an error if the integer part of the result has more than [MaxPrec] digits.
//
// [fused multiply-addition]: https://en.wikipedia.org/wiki/Multiply%E2%80%93accumulate_operation#Fused_multiply%E2%80%93add
func (d Decimal) AddMul(e, f Decimal) (Decimal, error) {
	return d.AddMulExact(e, f, 0)
}

// AddMulExact is similar to [Decimal.AddMul], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) AddMulExact(e, f Decimal, scale int) (Decimal, error) {
	if scale < MinScale || scale > MaxScale {
		return Decimal{}, fmt.Errorf("computing [%v + %v * %v]: %w", d, e, f, errScaleRange)
	}

	// General case
	g, err := d.addMulFint(e, f, scale)
	if err != nil {
		g, err = d.addMulBint(e, f, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v + %v * %v]: %w", d, e, f, err)
		}
	}

	return g, nil
}

// addMulFint computes the fused multiply-addition of three decimals using uint64 arithmetic.
func (d Decimal) addMulFint(e, f Decimal, minScale int) (Decimal, error) {
	dcoef, ecoef, fcoef := d.coef, e.coef, f.coef

	// Compute e = e * f
	var ok bool
	ecoef, ok = ecoef.mul(fcoef)
	if !ok {
		return Decimal{}, errDecimalOverflow
	}

	// Alignment and scale
	scale := e.Scale() + f.Scale()
	switch {
	case scale > d.Scale():
		dcoef, ok = dcoef.lsh(scale - d.Scale())
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	case scale < d.Scale():
		ecoef, ok = ecoef.lsh(d.Scale() - scale)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		scale = d.Scale()
	}

	// Sign
	var neg bool
	if ecoef > dcoef {
		neg = e.IsNeg() != f.IsNeg()
	} else {
		neg = d.IsNeg()
	}

	// Compute d = d + e
	if d.IsNeg() != (e.IsNeg() != f.IsNeg()) {
		dcoef = dcoef.subAbs(ecoef)
	} else {
		dcoef, ok = dcoef.add(ecoef)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	}

	return newFromFint(neg, dcoef, scale, minScale)
}

// addMulBint computes the fused multiply-addition of three decimals using *big.Int arithmetic.
func (d Decimal) addMulBint(e, f Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	fcoef := getBint()
	defer putBint(fcoef)
	fcoef.setFint(f.coef)

	// Compute e = e * f
	ecoef.mul(ecoef, fcoef)

	// Alignment and scale
	scale := e.Scale() + f.Scale()
	switch {
	case scale > d.Scale():
		dcoef.lsh(dcoef, scale-d.Scale())
	case scale < d.Scale():
		ecoef.lsh(ecoef, d.Scale()-scale)
		scale = d.Scale()
	}

	// Sign
	var neg bool
	if ecoef.cmp(dcoef) > 0 {
		neg = e.IsNeg() != f.IsNeg()
	} else {
		neg = d.IsNeg()
	}

	// Compute d = d + e
	if d.IsNeg() != (e.IsNeg() != f.IsNeg()) {
		dcoef.subAbs(dcoef, ecoef)
	} else {
		dcoef.add(dcoef, ecoef)
	}

	return newFromBint(neg, dcoef, scale, minScale)
}

// SubQuo returns the (possibly rounded) fused quotient-subtraction of decimals d, e, and f.
// It computes d - e / f with double precision during intermediate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of quotients, such as internal rate of return.
//
// AddQuo returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) SubQuo(e, f Decimal) (Decimal, error) {
	return d.AddQuoExact(e.Neg(), f, 0)
}

// SubQuoExact is similar to [Decimal.SubQuo], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) SubQuoExact(e, f Decimal, scale int) (Decimal, error) {
	return d.AddQuoExact(e.Neg(), f, scale)
}

// AddQuo returns the (possibly rounded) fused quotient-addition of decimals d, e, and f.
// It computes d + e / f with double precision during intermediate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of quotients, such as internal rate of return.
//
// AddQuo returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) AddQuo(e, f Decimal) (Decimal, error) {
	return d.AddQuoExact(e, f, 0)
}

// AddQuoExact is similar to [Decimal.AddQuo], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) AddQuoExact(e, f Decimal, scale int) (Decimal, error) {
	if scale < MinScale || scale > MaxScale {
		return Decimal{}, fmt.Errorf("computing [%v + %v / %v]: %w", d, e, f, errScaleRange)
	}

	// Special case: zero divisor
	if f.IsZero() {
		return Decimal{}, fmt.Errorf("computing [%v + %v / %v]: %w", d, e, f, errDivisionByZero)
	}

	// Special case: zero dividend
	if e.IsZero() {
		scale = max(scale, e.Scale()-f.Scale())
		return d.Pad(scale), nil
	}

	// General case
	g, err := d.addQuoFint(e, f, scale)
	if err != nil {
		g, err = d.addQuoBint(e, f, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v + %v / %v]: %w", d, e, f, err)
		}
	}

	// Preferred scale
	scale = max(scale, d.Scale(), e.Scale()-f.Scale())
	g = g.Trim(scale)

	return g, nil
}

// addQuoFint computes the fused quotient-addition of three decimals using uint64 arithmetic.
func (d Decimal) addQuoFint(e, f Decimal, minScale int) (Decimal, error) {
	dcoef, ecoef, fcoef := d.coef, e.coef, f.coef

	// Scale
	scale := e.Scale() - f.Scale()

	// Alignment
	var ok bool
	if shift := MaxPrec - ecoef.prec(); shift > 0 {
		ecoef, ok = ecoef.lsh(shift)
		if !ok {
			return Decimal{}, errDecimalOverflow // Should never happen
		}
		scale = scale + shift
	}
	if shift := fcoef.ntz(); shift > 0 {
		fcoef = fcoef.rshDown(shift)
		scale = scale + shift
	}

	// Compute e = e / f
	ecoef, ok = ecoef.quo(fcoef)
	if !ok {
		return Decimal{}, errInexactDivision
	}

	// Alignment and scale
	switch {
	case scale > d.Scale():
		if shift := min(scale-d.Scale(), scale-(e.Scale()-f.Scale()), ecoef.ntz()); shift > 0 {
			ecoef = ecoef.rshDown(shift)
			scale = scale - shift
		}
		dcoef, ok = dcoef.lsh(scale - d.Scale())
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	case scale < d.Scale():
		ecoef, ok = ecoef.lsh(d.Scale() - scale)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		scale = d.Scale()
	}

	// Sign
	var neg bool
	if ecoef > dcoef {
		neg = e.IsNeg() != f.IsNeg()
	} else {
		neg = d.IsNeg()
	}

	// Compute d = d + e
	if d.IsNeg() != (e.IsNeg() != f.IsNeg()) {
		dcoef = dcoef.subAbs(ecoef)
	} else {
		dcoef, ok = dcoef.add(ecoef)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	}

	return newFromFint(neg, dcoef, scale, minScale)
}

// addQuoBint computes the fused quotient-addition of three decimals using *big.Int arithmetic.
func (d Decimal) addQuoBint(e, f Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	fcoef := getBint()
	defer putBint(fcoef)
	fcoef.setFint(f.coef)

	// Compute e = ⌊e / f⌋
	scale := 2 * MaxScale
	ecoef.lsh(ecoef, scale+f.Scale()-e.Scale())
	ecoef.quo(ecoef, fcoef)

	// Alignment
	dcoef.lsh(dcoef, scale-d.Scale())

	// Sign
	var neg bool
	if ecoef.cmp(dcoef) > 0 {
		neg = e.IsNeg() != f.IsNeg()
	} else {
		neg = d.IsNeg()
	}

	// Compute d = d + e
	if d.IsNeg() != (e.IsNeg() != f.IsNeg()) {
		dcoef.subAbs(dcoef, ecoef)
	} else {
		dcoef.add(dcoef, ecoef)
	}

	return newFromBint(neg, dcoef, scale, minScale)
}

// Inv returns the (possibly rounded) inverse of the decimal.
//
// Inv returns an error if:
//   - the integer part of the result has more than [MaxPrec] digits;
//   - the decimal is 0.
func (d Decimal) Inv() (Decimal, error) {
	f, err := One.Quo(d)
	if err != nil {
		return Decimal{}, fmt.Errorf("inverting %v: %w", d, err)
	}
	return f, nil
}

// Quo returns the (possibly rounded) quotient of decimals d and e.
//
// Quo returns an error if:
//   - the divisor is 0;
//   - the integer part of the result has more than [MaxPrec] digits.
func (d Decimal) Quo(e Decimal) (Decimal, error) {
	return d.QuoExact(e, 0)
}

// QuoExact is similar to [Decimal.Quo], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) QuoExact(e Decimal, scale int) (Decimal, error) {
	if scale < MinScale || scale > MaxScale {
		return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, errScaleRange)
	}

	// Special case: zero divisor
	if e.IsZero() {
		return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, errDivisionByZero)
	}

	// Special case: zero dividend
	if d.IsZero() {
		scale = max(scale, d.Scale()-e.Scale())
		return newSafe(false, 0, scale)
	}

	// General case
	f, err := d.quoFint(e, scale)
	if err != nil {
		f, err = d.quoBint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("computing [%v / %v]: %w", d, e, err)
		}
	}

	// Preferred scale
	scale = max(scale, d.Scale()-e.Scale())
	f = f.Trim(scale)

	return f, nil
}

// quoFint computes the quotient of two decimals using uint64 arithmetic.
func (d Decimal) quoFint(e Decimal, minScale int) (Decimal, error) {
	dcoef, ecoef := d.coef, e.coef

	// Scale
	scale := d.Scale() - e.Scale()

	// Alignment
	var ok bool
	if shift := MaxPrec - dcoef.prec(); shift > 0 {
		dcoef, ok = dcoef.lsh(shift)
		if !ok {
			return Decimal{}, errDecimalOverflow // Should never happen
		}
		scale = scale + shift
	}

	if shift := ecoef.ntz(); shift > 0 {
		ecoef = ecoef.rshDown(shift)
		scale = scale + shift
	}

	// Compute d = d / e
	dcoef, ok = dcoef.quo(ecoef)
	if !ok {
		return Decimal{}, errInexactDivision
	}

	// Sign
	neg := d.IsNeg() != e.IsNeg()

	return newFromFint(neg, dcoef, scale, minScale)
}

// quoBint computes the quotient of two decimals using *big.Int arithmetic.
func (d Decimal) quoBint(e Decimal, minScale int) (Decimal, error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	// Scale
	scale := 2 * MaxScale

	// Alignment
	dcoef.lsh(dcoef, scale+e.Scale()-d.Scale())

	// Compute d = ⌊d / e⌋
	dcoef.quo(dcoef, ecoef)

	// Sign
	neg := d.IsNeg() != e.IsNeg()

	return newFromBint(neg, dcoef, scale, minScale)
}

// QuoRem returns the quotient q and remainder r of decimals d and e
// such that d = e * q + r, where q is an integer and the sign of the
// reminder r is the same as the sign of the dividend d.
//
// QuoRem returns an error if:
//   - the divisor is 0;
//   - the integer part of the quotient has more than [MaxPrec] digits.
func (d Decimal) QuoRem(e Decimal) (q, r Decimal, err error) {
	// Special case: zero divisor
	if e.IsZero() {
		return Decimal{}, Decimal{}, fmt.Errorf("computing [%v div %v] and [%v mod %v]: %w", d, e, d, e, errDivisionByZero)
	}

	// General case
	q, r, err = d.quoRemFint(e)
	if err != nil {
		q, r, err = d.quoRemBint(e)
		if err != nil {
			return Decimal{}, Decimal{}, fmt.Errorf("computing [%v div %v] and [%v mod %v]: %w", d, e, d, e, err)
		}
	}

	return q, r, nil
}

// quoRemFint computes the quotient and remainder of two decimals using uint64 arithmetic.
func (d Decimal) quoRemFint(e Decimal) (q, r Decimal, err error) {
	dcoef, ecoef := d.coef, e.coef

	// Alignment and rscale
	var rscale int
	var ok bool
	switch {
	case d.Scale() == e.Scale():
		rscale = d.Scale()
	case d.Scale() > e.Scale():
		rscale = d.Scale()
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return Decimal{}, Decimal{}, errDecimalOverflow
		}
	case d.Scale() < e.Scale():
		rscale = e.Scale()
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return Decimal{}, Decimal{}, errDecimalOverflow
		}
	}

	// Compute q = ⌊d / e⌋, r = d - e * q
	qcoef, rcoef, ok := dcoef.quoRem(ecoef)
	if !ok {
		return Decimal{}, Decimal{}, errDivisionByZero // Should never happen
	}

	// Signs
	qsign := d.IsNeg() != e.IsNeg()
	rsign := d.IsNeg()

	q, err = newFromFint(qsign, qcoef, 0, 0)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	r, err = newFromFint(rsign, rcoef, rscale, rscale)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	return q, r, nil
}

// quoRemBint computes the quotient and remainder of two decimals using *big.Int arithmetic.
func (d Decimal) quoRemBint(e Decimal) (q, r Decimal, err error) {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	qcoef := getBint()
	defer putBint(qcoef)

	rcoef := getBint()
	defer putBint(rcoef)

	// Alignment and scale
	var rscale int
	switch {
	case d.Scale() == e.Scale():
		rscale = d.Scale()
	case d.Scale() > e.Scale():
		rscale = d.Scale()
		ecoef.lsh(ecoef, d.Scale()-e.Scale())
	case d.Scale() < e.Scale():
		rscale = e.Scale()
		dcoef.lsh(dcoef, e.Scale()-d.Scale())
	}

	// Compute q = ⌊d / e⌋, r = d - e * q
	qcoef.quoRem(dcoef, ecoef, rcoef)

	// Signs
	qsign := d.IsNeg() != e.IsNeg()
	rsign := d.IsNeg()

	q, err = newFromBint(qsign, qcoef, 0, 0)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	r, err = newFromBint(rsign, rcoef, rscale, rscale)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	return q, r, nil
}

// Max returns the larger decimal.
// See also method [Decimal.CmpTotal].
func (d Decimal) Max(e Decimal) Decimal {
	if d.CmpTotal(e) >= 0 {
		return d
	}
	return e
}

// Min returns the smaller decimal.
// See also method [Decimal.CmpTotal].
func (d Decimal) Min(e Decimal) Decimal {
	if d.CmpTotal(e) <= 0 {
		return d
	}
	return e
}

// Clamp compares decimals and returns:
//
//	min if d < min
//	max if d > max
//	  d otherwise
//
// See also method [Decimal.CmpTotal].
//
// Clamp returns an error if min is greater than max numerically.
// nolint:predeclared
func (d Decimal) Clamp(min, max Decimal) (Decimal, error) {
	if min.Cmp(max) > 0 {
		return Decimal{}, fmt.Errorf("clamping %v: invalid range", d)
	}
	if min.CmpTotal(max) > 0 {
		// min and max are equal numerically but have different scales.
		// Swaping min and max to ensure total ordering.
		min, max = max, min
	}
	if d.CmpTotal(min) < 0 {
		return min, nil
	}
	if d.CmpTotal(max) > 0 {
		return max, nil
	}
	return d, nil
}

// CmpTotal compares decimal representations and returns:
//
//	-1 if d < e
//	-1 if d = e and d.scale > e.scale
//	 0 if d = e and d.scale = e.scale
//	+1 if d = e and d.scale < e.scale
//	+1 if d > e
//
// See also method [Decimal.Cmp].
func (d Decimal) CmpTotal(e Decimal) int {
	switch d.Cmp(e) {
	case -1:
		return -1
	case 1:
		return 1
	}
	switch {
	case d.Scale() > e.Scale():
		return -1
	case d.Scale() < e.Scale():
		return 1
	}
	return 0
}

// CmpAbs compares absolute values of decimals and returns:
//
//	-1 if |d| < |e|
//	 0 if |d| = |e|
//	+1 if |d| > |e|
//
// See also method [Decimal.Cmp].
func (d Decimal) CmpAbs(e Decimal) int {
	d, e = d.Abs(), e.Abs()
	return d.Cmp(e)
}

// Equal compares decimals and returns:
//
//	 true if d = e
//	false otherwise
//
// See also method [Decimal.Cmp].
func (d Decimal) Equal(e Decimal) bool {
	return d.Cmp(e) == 0
}

// Less compares decimals and returns:
//
//	 true if d < e
//	false otherwise
//
// See also method [Decimal.Cmp].
func (d Decimal) Less(e Decimal) bool {
	return d.Cmp(e) < 0
}

// Cmp compares decimals and returns:
//
//	-1 if d < e
//	 0 if d = e
//	+1 if d > e
//
// See also methods [Decimal.CmpAbs], [Decimal.CmpTotal].
func (d Decimal) Cmp(e Decimal) int {
	// Special case: different signs
	switch {
	case d.Sign() > e.Sign():
		return 1
	case d.Sign() < e.Sign():
		return -1
	}

	// General case
	r, err := d.cmpFint(e)
	if err != nil {
		r = d.cmpBint(e)
	}
	return r
}

// cmpFint compares decimals using uint64 arithmetic.
func (d Decimal) cmpFint(e Decimal) (int, error) {
	dcoef, ecoef := d.coef, e.coef

	// Alignment
	var ok bool
	switch {
	case d.Scale() > e.Scale():
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return 0, errDecimalOverflow
		}
	case d.Scale() < e.Scale():
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return 0, errDecimalOverflow
		}
	}

	// Comparison
	switch {
	case dcoef > ecoef:
		return d.Sign(), nil
	case ecoef > dcoef:
		return -e.Sign(), nil
	}
	return 0, nil
}

// cmpBint compares decimals using *big.Int arithmetic.
func (d Decimal) cmpBint(e Decimal) int {
	dcoef := getBint()
	defer putBint(dcoef)
	dcoef.setFint(d.coef)

	ecoef := getBint()
	defer putBint(ecoef)
	ecoef.setFint(e.coef)

	// Alignment
	switch {
	case d.Scale() > e.Scale():
		ecoef.lsh(ecoef, d.Scale()-e.Scale())
	case d.Scale() < e.Scale():
		dcoef.lsh(dcoef, e.Scale()-d.Scale())
	}

	// Comparison
	switch dcoef.cmp(ecoef) {
	case 1:
		return d.Sign()
	case -1:
		return -e.Sign()
	}
	return 0
}

// NullDecimal represents a decimal that can be null.
// Its zero value is null.
// NullDecimal is not thread-safe.
type NullDecimal struct {
	Decimal Decimal
	Valid   bool
}

// Scan implements the [sql.Scanner] interface.
// See also constructor [Parse].
//
// [sql.Scanner]: https://pkg.go.dev/database/sql#Scanner
func (n *NullDecimal) Scan(value any) error {
	if value == nil {
		n.Decimal = Decimal{}
		n.Valid = false
		return nil
	}
	err := n.Decimal.Scan(value)
	if err != nil {
		n.Decimal = Decimal{}
		n.Valid = false
		return err
	}
	n.Valid = true
	return nil
}

// Value implements the [driver.Valuer] interface.
// See also method [Decimal.String].
//
// [driver.Valuer]: https://pkg.go.dev/database/sql/driver#Valuer
func (n NullDecimal) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Decimal.Value()
}
