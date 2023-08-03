package decimal

import (
	"errors"
	"fmt"
	"math"
)

// Decimal type represents a finite floating-point decimal number.
// Its zero value corresponds to the numeric value of 0.
// This type is designed to be safe for concurrent use by multiple goroutines.
//
// A decimal is a struct with three fields:
//
//   - Sign: a boolean indicating whether the decimal is negative.
//   - Scale: an integer indicating the position of the floating decimal point.
//   - Coefficient: an integer value of the decimal without the decimal point.
//
// The scale field determines the position of the decimal point in the coefficient.
// For example, a decimal value with a scale of 2 represents a value that has two
// digits after the decimal point.
// The coefficient field is the integer value of the decimal without the decimal point.
// For example, a decimal with a coefficient of 12345 and a scale of 2 represents
// the value 123.45.
// Such approach allows for multiple representations of the same numerical value.
// For example, 1, 1.0, and 1.00 all have the same value, but they
// have different scales and coefficients.
//
// One important aspect of the decimal is that it does not support
// special values such as NaN, Infinity, or signed zeros.
type Decimal struct {
	neg   bool // indicates whether the decimal is negative
	scale int8 // the position of the floating decimal point
	coef  fint // the value without decimal point
}

const (
	MaxPrec  = 19      // MaxPrec is a maximum length of the coefficient in decimal digits
	MaxScale = MaxPrec // MaxScale is a maximum number of digits after the decimal point
	maxCoef  = maxFint // maxCoef is a maximum absolute value of the coefficient, which is equal to (10^MaxPrec - 1)
)

var (
	Zero               = MustNew(0, 0)   // Zero is a decimal with a value of 0 and a scale of 0
	One                = MustNew(1, 0)   // One is a decimal with a value of 1 and a scale of 0
	Ten                = MustNew(10, 0)  // Ten is a decimal with a value of 10 and a scale of 0
	Hundred            = MustNew(100, 0) // Hundred is a decimal with a value of 100 and a scale of 0
	errDecimalOverflow = errors.New("decimal overflow")
	errInvalidDecimal  = errors.New("invalid decimal")
	errScaleRange      = errors.New("scale out of range")
	errExponentRange   = errors.New("exponent out of range")
	errInexactDivision = errors.New("inexact division")
	errDivisionByZero  = errors.New("division by zero")
)

func newDecimalUnsafe(neg bool, coef fint, scale int) Decimal {
	if coef == 0 {
		neg = false
	}
	return Decimal{neg: neg, coef: coef, scale: int8(scale)}
}

func newDecimalSafe(neg bool, coef fint, scale int) (Decimal, error) {
	switch {
	case scale < 0 || scale > MaxScale:
		return Decimal{}, errScaleRange
	case coef > maxCoef:
		return Decimal{}, errDecimalOverflow
	}
	return newDecimalUnsafe(neg, coef, scale), nil
}

// newDecimalFromFloat converts fint to decimal.
// This method does not use overflowError to return descriptive errors,
// as it must be as fast as possible.
func newDecimalFromFint(neg bool, coef fint, scale, minScale int) (Decimal, error) {
	var ok bool
	// Normaliztion
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
	return newDecimalSafe(neg, coef, scale)
}

func overflowError(gotPrec, gotScale, wantScale int) error {
	maxDigits := MaxPrec - wantScale
	gotDigits := gotPrec - gotScale
	switch wantScale {
	case 0:
		return fmt.Errorf("the integer part of a %T can have at most %v digits, but it has %v digits: %w", Decimal{}, maxDigits, gotDigits, errDecimalOverflow)
	case 1:
		return fmt.Errorf("with %v significant digit after the decimal point, the integer part of a %T can have at most %v digits, but it has %v digits: %w", wantScale, Decimal{}, maxDigits, gotDigits, errDecimalOverflow)
	default:
		return fmt.Errorf("with %v significant digits after the decimal point, the integer part of a %T can have at most %v digits, but it has %v digits: %w", wantScale, Decimal{}, maxDigits, gotDigits, errDecimalOverflow)
	}
}

// newDecimalFromSint converts *sint to decimal.
// This method uses overflowError to return descriptive errors.
func newDecimalFromSint(neg bool, coef *sint, scale, minScale int) (Decimal, error) {
	// Check for overflow
	prec := coef.prec()
	if prec-scale > MaxPrec-minScale {
		return Decimal{}, overflowError(prec, scale, minScale)
	}
	// Normaliztion
	switch {
	case scale < minScale:
		coef.lsh(coef, minScale-scale)
		scale = minScale
	case scale >= prec && scale > MaxScale: // no integer part
		coef.rshHalfEven(coef, scale-MaxScale)
		scale = MaxScale
	case prec > scale && prec > MaxPrec: // there is an integer part
		coef.rshHalfEven(coef, prec-MaxPrec)
		scale = scale - (prec - MaxPrec)
	}
	// Handle the rare case when rshHalfEven rounded
	// a 19-digit coefficient to a 20-digit coefficient.
	if coef.hasPrec(MaxPrec + 1) {
		return newDecimalFromSint(neg, coef, scale, minScale)
	}
	return newDecimalSafe(neg, coef.fint(), scale)
}

// New returns a (possibly rounded) decimal equal to coef / 10^scale.
//
// New returns an error if scale is negative or more than [MaxScale].
func New(coef int64, scale int) (Decimal, error) {
	var neg bool
	if coef < 0 {
		neg = true
		coef = -coef
	}
	return newDecimalSafe(neg, fint(coef), scale)
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

// Zero returns a decimal with a value of 0, having the same scale as decimal d.
func (d Decimal) Zero() Decimal {
	return newDecimalUnsafe(false, 0, d.Scale())
}

// One returns a decimal with a value of 1, having the same scale as decimal d.
func (d Decimal) One() Decimal {
	return newDecimalUnsafe(false, pow10[d.Scale()], d.Scale())
}

// ULP (Unit in the Last Place) returns the smallest representable positive
// difference between two decimals with the same scale as decimal d.
// It can be useful for implementing rounding and comparison algorithms.
// See also method [Decimal.One].
func (d Decimal) ULP() Decimal {
	return newDecimalUnsafe(false, 1, d.Scale())
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
//		sign           ::= '+' | '-'
//		digits         ::= { '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' }
//		significand    ::= digits '.' digits | '.' digits | digits '.' | digits
//	    exponent       ::= ('e' | 'E') [sign] digits
//		numeric-string ::= [sign] significand [exponent]
//
// Parse removes leading zeros from the integer part of the input string,
// but tries to maintain trailing zeros in the fractional part to preserve scale.
//
// Parse returns an error:
//   - if the integer part of the result has more than [MaxPrec] digits.
//   - if the string does not represent a valid decimal number.
//   - if the string is longer than 100 bytes.
//   - if the exponent is less than -100 or greater than 100.
func Parse(dec string) (Decimal, error) {
	return ParseExact(dec, 0)
}

// ParseExact is similar to [Parse], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for parsing monetary amounts, where the scale should be
// equal to or greater than the currency's scale.
func ParseExact(dec string, scale int) (Decimal, error) {
	if len(dec) > 100 {
		return Decimal{}, fmt.Errorf("parsing decimal: %w", errInvalidDecimal)
	}
	if scale < 0 || scale > MaxScale {
		return Decimal{}, fmt.Errorf("parsing decimal: %w", errScaleRange)
	}
	d, err := parseFint(dec, scale)
	if err != nil {
		d, err = parseSint(dec, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("parsing decimal: %w", err)
		}
	}
	return d, nil
}

// parseFint does not support exponential notation to make it as fast as possible.
func parseFint(dec string, minScale int) (Decimal, error) {
	pos := 0
	width := len(dec)

	// Sign
	var neg bool
	switch {
	case pos == width:
		// skip
	case dec[pos] == '-':
		neg = true
		pos++
	case dec[pos] == '+':
		pos++
	}

	// Integer
	coef := fint(0)
	hascoef, ok := false, false
	for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
		coef, ok = coef.fsa(1, dec[pos]-'0')
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		hascoef = true
		pos++
	}

	// Fraction
	scale := 0
	if pos < width && dec[pos] == '.' {
		pos++
		for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
			coef, ok = coef.fsa(1, dec[pos]-'0')
			if !ok {
				return Decimal{}, errDecimalOverflow
			}
			hascoef = true
			scale++
			pos++
		}
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("invalid character %q: %w", dec[pos], errInvalidDecimal)
	}
	if !hascoef {
		return Decimal{}, fmt.Errorf("no coefficient: %w", errInvalidDecimal)
	}
	return newDecimalFromFint(neg, coef, scale, minScale)
}

// parseSint supports exponential notation.
func parseSint(dec string, minScale int) (Decimal, error) {
	pos := 0
	width := len(dec)

	// Sign
	var neg bool
	switch {
	case pos == width:
		// skip
	case dec[pos] == '-':
		neg = true
		pos++
	case dec[pos] == '+':
		pos++
	}

	// Integer
	coef := new(sint)
	hascoef := false
	for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
		coef.fsa(1, dec[pos]-'0')
		hascoef = true
		pos++
	}

	// Fraction
	scale := 0
	if pos < width && dec[pos] == '.' {
		pos++
		for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
			coef.fsa(1, dec[pos]-'0')
			hascoef = true
			scale++
			pos++
		}
	}

	// Exponential part
	exp := 0
	eneg, hasexp, hasesym := false, false, false
	if pos < width && (dec[pos] == 'e' || dec[pos] == 'E') {
		hasesym = true
		pos++
		// Sign
		switch {
		case pos == width:
			// skip
		case dec[pos] == '-':
			eneg = true
			pos++
		case dec[pos] == '+':
			pos++
		}
		// Integer
		for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
			exp = exp*10 + int(dec[pos]-'0')
			if exp > 100 {
				return Decimal{}, errExponentRange
			}
			hasexp = true
			pos++
		}
	}

	if pos != width {
		return Decimal{}, fmt.Errorf("invalid character %q: %w", dec[pos], errInvalidDecimal)
	}
	if !hascoef {
		return Decimal{}, fmt.Errorf("no coefficient: %w", errInvalidDecimal)
	}
	if hasesym && !hasexp {
		return Decimal{}, fmt.Errorf("no exponent: %w", errInvalidDecimal)
	}

	if eneg {
		scale = scale + exp
	} else {
		scale = scale - exp
	}

	return newDecimalFromSint(neg, coef, scale, minScale)
}

// MustParse is like [Parse] but panics if the string cannot be parsed.
// It simplifies safe initialization of global variables holding decimals.
func MustParse(dec string) Decimal {
	d, err := Parse(dec)
	if err != nil {
		panic(fmt.Sprintf("MustParse(%q) failed: %v", dec, err))
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

// Float64 returns a float64 representation of the decimal.
// This conversion may lose data, as float64 has a limited precision
// compared to the decimal type.
func (d Decimal) Float64() (f float64, ok bool) {
	x := float64(d.Coef())
	y := float64(pow10[d.Scale()])
	z := x / y
	if math.IsInf(z, 0) {
		return 0, false
	}
	if d.IsNeg() {
		z = -z
	}
	return z, true
}

// Int64 returns a pair of int64 values representing the integer part i and the
// fractional part f of the decimal.
// The relationship between the decimal and the returned values can be expressed
// as d = i + f / 10^scale.
// If the result cannot be accurately represented as a pair of int64 values,
// the method returns false.
func (d Decimal) Int64(scale int) (i, f int64, ok bool) {
	if scale < 0 {
		return 0, 0, false
	}
	x := d.coef
	y := pow10[d.Scale()]
	if scale < d.Scale() {
		x = x.rshHalfEven(d.Scale() - scale)
		y = pow10[scale]
	}
	p := x / y
	q := x - p*y
	if scale > d.Scale() {
		q, ok = q.lsh(scale - d.Scale())
		if !ok {
			return 0, 0, false
		}
	}
	if d.IsNeg() {
		if p > -math.MinInt64 || q > -math.MinInt64 {
			return 0, 0, false
		}
		return -int64(p), -int64(q), true
	} else {
		if p > math.MaxInt64 || q > math.MaxInt64 {
			return 0, 0, false
		}
		return int64(p), int64(q), true
	}
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
// See also method [Parse].
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

// Format implements [fmt.Formatter] interface.
// The following [verbs] are available:
//
//	%f, %s, %v: -123.456
//	%q:        "-123.456"
//	%k:         -12345.6%
//
// The following format flags can be used with all verbs: '+', ' ', '0', '-'.
//
// Precision is only supported for %f and %k verbs.
// For %f verb, the default precision is equal to the actual scale of the decimal,
// whereas, for verb %k the default precision is the actual scale of the decimal minus 2.
//
// [verbs]: https://pkg.go.dev/fmt#hdr-Printing
// [fmt.Formatter]: https://pkg.go.dev/fmt#Formatter
func (d Decimal) Format(state fmt.State, verb rune) {
	var err error

	// Percentage
	if verb == 'k' || verb == 'K' {
		d, err = d.Mul(Hundred)
		if err != nil {
			panic(fmt.Errorf("formatting percent: %w", err)) // this panic is handled by the fmt package
		}
	}

	// Rescaling
	tzeroes := 0
	if verb == 'f' || verb == 'F' || verb == 'k' || verb == 'K' {
		scale := 0
		switch p, ok := state.Precision(); {
		case ok:
			scale = p
		case verb == 'k' || verb == 'K':
			scale = d.Scale() - 2
		case verb == 'f' || verb == 'F':
			scale = d.Scale()
		}
		if scale < 0 {
			scale = 0
		}
		switch {
		case scale < d.Scale():
			d = d.Round(scale)
		case scale > d.Scale():
			tzeroes = scale - d.Scale()
		}
	}

	// Integer and fractional digits
	intdigs, fracdigs := 0, d.Scale()
	if dprec := d.Prec(); dprec > fracdigs {
		intdigs = dprec - fracdigs
	}
	if d.WithinOne() {
		intdigs++ // leading 0
	}

	// Decimal point
	dpoint := 0
	if fracdigs > 0 || tzeroes > 0 {
		dpoint = 1
	}

	// Arithmetic sign
	rsign := 0
	if d.IsNeg() || state.Flag('+') || state.Flag(' ') {
		rsign = 1
	}

	// Percentage sign
	psign := 0
	if verb == 'k' || verb == 'K' {
		psign = 1
	}

	// Quotes
	lquote, tquote := 0, 0
	if verb == 'q' || verb == 'Q' {
		lquote, tquote = 1, 1
	}

	// Padding
	width := lquote + rsign + intdigs + dpoint + fracdigs + tzeroes + psign + tquote
	lspaces, tspaces, lzeroes := 0, 0, 0
	if w, ok := state.Width(); ok && w > width {
		switch {
		case state.Flag('-'):
			tspaces = w - width
		case state.Flag('0'):
			lzeroes = w - width
		default:
			lspaces = w - width
		}
		width = w
	}

	// Writing buffer
	buf := make([]byte, width)
	pos := width - 1
	for i := 0; i < tspaces; i++ {
		buf[pos] = ' '
		pos--
	}
	if tquote > 0 {
		buf[pos] = '"'
		pos--
	}
	if psign > 0 {
		buf[pos] = '%'
		pos--
	}
	for i := 0; i < tzeroes; i++ {
		buf[pos] = '0'
		pos--
	}
	dcoef := d.Coef()
	for i := 0; i < fracdigs; i++ {
		buf[pos] = byte(dcoef%10) + '0'
		pos--
		dcoef /= 10
	}
	if dpoint > 0 {
		buf[pos] = '.'
		pos--
	}
	for i := 0; i < intdigs; i++ {
		buf[pos] = byte(dcoef%10) + '0'
		pos--
		dcoef /= 10
	}
	for i := 0; i < lzeroes; i++ {
		buf[pos] = '0'
		pos--
	}
	if rsign > 0 {
		if d.IsNeg() {
			buf[pos] = '-'
		} else if state.Flag(' ') {
			buf[pos] = ' '
		} else {
			buf[pos] = '+'
		}
		pos--
	}
	if lquote > 0 {
		buf[pos] = '"'
		pos--
	}
	for i := 0; i < lspaces; i++ {
		buf[pos] = ' '
		pos--
	}

	// Writing result
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
// See also method [Decimal.Prec].
func (d Decimal) Scale() int {
	return int(d.scale)
}

// MinScale returns the smallest scale that the decimal can be rescaled to without rounding.
// See also method [Decimal.Trim].
func (d Decimal) MinScale() int {
	// Special case: no scale
	if d.Scale() == 0 || d.IsZero() {
		return 0
	}
	// General case
	z := d.coef.tzeros()
	if z > d.Scale() {
		return 0
	}
	return d.Scale() - z
}

// IsInt returns true if the fractional part of the decimal is zero.
func (d Decimal) IsInt() bool {
	return d.coef%pow10[d.Scale()] == 0
}

// IsOne returns:
//
//	true  if d == -1 or d == 1
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
// the decimal point.
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also method [Decimal.Rescale].
func (d Decimal) Round(scale int) Decimal {
	if scale < 0 {
		scale = 0
	}
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	coef = coef.rshHalfEven(d.Scale() - scale)
	return newDecimalUnsafe(d.IsNeg(), coef, scale)
}

// Pad returns a decimal zero-padded to the specified number of digits after
// the decimal point.
// See also method [Decimal.Trim].
//
// Pad returns an error if the integer part of the result has more than
// ([MaxPrec] - scale) digits.
func (d Decimal) Pad(scale int) (Decimal, error) {
	if scale > MaxScale {
		return Decimal{}, fmt.Errorf("zero-padding %v: %w", d, errScaleRange)
	}
	if scale <= d.Scale() {
		return d, nil
	}
	coef := d.coef
	coef, ok := coef.lsh(scale - d.Scale())
	if !ok {
		return Decimal{}, fmt.Errorf("zero-padding %v with %v digits: %w", d, scale-d.Scale(), overflowError(d.Prec(), d.Scale(), scale))
	}
	return newDecimalSafe(d.IsNeg(), coef, scale)
}

// Rescale returns a decimal rounded or zero-padded to the given number of digits
// after the decimal point.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// Rescale returns an error if the integer part of the result has more
// than ([MaxPrec] - scale) digits.
func (d Decimal) Rescale(scale int) (Decimal, error) {
	if scale < 0 || scale > MaxScale {
		return Decimal{}, fmt.Errorf("rescaling %v: %w", d, errScaleRange)
	}
	switch {
	case scale < d.Scale():
		return d.Round(scale), nil
	case scale > d.Scale():
		return d.Pad(scale)
	}
	return d, nil
}

// Quantize returns a decimal rounded to the same scale as decimal e.
// The sign and coefficient of decimal e are ignored.
// See also method [Decimal.Rescale].
//
// Qunatize returns an error if the integer part of result has more than ([MaxPrec] - e.Scale()) digits.
func (d Decimal) Quantize(e Decimal) (Decimal, error) {
	return d.Rescale(e.Scale())
}

// Trunc returns a decimal truncated to the specified number of digits
// after the decimal point.
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
func (d Decimal) Trunc(scale int) Decimal {
	if scale < 0 {
		scale = 0
	}
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	coef = coef.rshDown(d.Scale() - scale)
	return newDecimalUnsafe(d.IsNeg(), coef, scale)
}

// Trim returns a decimal with trailing zeros removed
// up to the given number of digits after the decimal point.
// If the given scale is negative, it is redefined to zero.
// See also method [Decimal.Pad].
func (d Decimal) Trim(scale int) Decimal {
	m := d.MinScale()
	if m > scale {
		scale = m
	}
	return d.Trunc(scale)
}

// Ceil returns a decimal rounded up to the given number of digits
// after the decimal point.
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also method [Decimal.Floor].
func (d Decimal) Ceil(scale int) Decimal {
	if scale < 0 {
		scale = 0
	}
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	if d.IsNeg() {
		coef = coef.rshDown(d.Scale() - scale)
	} else {
		coef = coef.rshUp(d.Scale() - scale)
	}
	return newDecimalUnsafe(d.IsNeg(), coef, scale)
}

// Floor returns a decimal rounded down to the specified number of digits
// after the decimal point.
// If the given scale is negative, it is redefined to zero.
// For financial calculations, the scale should be equal to or greater than
// the scale of the currency.
// See also method [Decimal.Ceil].
func (d Decimal) Floor(scale int) Decimal {
	if scale < 0 {
		scale = 0
	}
	if scale >= d.Scale() {
		return d
	}
	coef := d.coef
	if d.IsNeg() {
		coef = coef.rshUp(d.Scale() - scale)
	} else {
		coef = coef.rshDown(d.Scale() - scale)
	}
	return newDecimalUnsafe(d.IsNeg(), coef, scale)
}

// Neg returns a decimal with the opposite sign.
func (d Decimal) Neg() Decimal {
	return newDecimalUnsafe(!d.IsNeg(), d.coef, d.Scale())
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	return newDecimalUnsafe(false, d.coef, d.Scale())
}

// CopySign returns a decimal with the same sign as decimal e.
// If decimal e is zero, the sign of the result remains unchanged.
func (d Decimal) CopySign(e Decimal) Decimal {
	switch {
	case e.IsZero():
		return d
	case d.IsNeg() != e.IsNeg():
		return d.Neg()
	}
	return d
}

// Sign returns:
//
//	-1 if d < 0
//	 0 if d == 0
//	+1 if d > 0
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
//	true  if d == 0
//	false otherwise
func (d Decimal) IsZero() bool {
	return d.coef == 0
}

// Mul returns the (possibly rounded) product of decimals d and e.
//
// Mul returns an error if the integer part of the product has more than [MaxPrec] digits.
func (d Decimal) Mul(e Decimal) (Decimal, error) {
	return d.MulExact(e, 0)
}

// MulExact is similar to [Decimal.Mul], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) MulExact(e Decimal, scale int) (Decimal, error) {
	if scale < 0 || scale > MaxScale {
		return Decimal{}, fmt.Errorf("%v * %v: %w", d, e, errScaleRange)
	}
	f, err := d.mulFint(e, scale)
	if err != nil {
		f, err = d.mulSint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("%v * %v: %w", d, e, err)
		}
	}
	return f, nil
}

func (d Decimal) mulFint(e Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	ecoef := e.coef

	// Coefficient
	dcoef, ok := dcoef.mul(ecoef)
	if !ok {
		return Decimal{}, errDecimalOverflow
	}

	// Sign
	neg := d.IsNeg() != e.IsNeg()

	// Scale
	scale := d.Scale() + e.Scale()

	return newDecimalFromFint(neg, dcoef, scale, minScale)
}

func (d Decimal) mulSint(e Decimal, minScale int) (Decimal, error) {
	dcoef := newSintFromFint(d.coef)
	ecoef := newSintFromFint(e.coef)

	// Coefficient
	dcoef.mul(dcoef, ecoef)

	// Sign
	neg := d.IsNeg() != e.IsNeg()

	// Scale
	scale := d.Scale() + e.Scale()

	return newDecimalFromSint(neg, dcoef, scale, minScale)
}

// Pow returns the (possibly rounded) decimal raised to the given power.
//
// Pow returns an error if the integer part of the power has more than [MaxPrec] digits.
func (d Decimal) Pow(exp int) (Decimal, error) {
	return d.PowExact(exp, 0)
}

// PowExact is similar to [Decimal.Pow], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) PowExact(exp, scale int) (Decimal, error) {
	e, err := d.powLoop(exp, scale)
	if err != nil {
		return Decimal{}, fmt.Errorf("%v^%v: %w", d, exp, err)
	}
	// Trailing zeros (Workaround)
	e = e.Trim(scale)
	return e, nil
}

func (d Decimal) powLoop(exp, scale int) (Decimal, error) {
	// Special case: power of 0
	if exp == 0 {
		return One, nil
	}

	// General case
	e, err := d.powLoop(exp/2, scale)
	if err != nil {
		return Decimal{}, err
	}
	e, err = e.MulExact(e, scale)
	if err != nil {
		return Decimal{}, err
	}
	if exp%2 == 0 {
		return e, nil
	}
	if exp > 0 {
		return e.MulExact(d, scale)
	}
	return e.QuoExact(d, scale)
}

// Add returns the (possibly rounded) sum of decimals d and e.
//
// Add returns an error if the integer part of the sum has more than [MaxPrec] digits.
func (d Decimal) Add(e Decimal) (Decimal, error) {
	return d.AddExact(e, 0)
}

// AddExact is similar to [Decimal.Add], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) AddExact(e Decimal, scale int) (Decimal, error) {
	if scale < 0 || scale > MaxScale {
		return Decimal{}, fmt.Errorf("%v + %v: %w", d, e, errScaleRange)
	}
	f, err := d.addFint(e, scale)
	if err != nil {
		f, err = d.addSint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("%v + %v: %w", d, e, err)
		}
	}
	return f, nil
}

func (d Decimal) addFint(e Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	ecoef := e.coef

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

	// Coefficient
	if d.IsNeg() != e.IsNeg() {
		dcoef = dcoef.dist(ecoef)
	} else {
		dcoef, ok = dcoef.add(ecoef)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	}

	return newDecimalFromFint(neg, dcoef, scale, minScale)
}

func (d Decimal) addSint(e Decimal, minScale int) (Decimal, error) {
	dcoef := newSintFromFint(d.coef)
	ecoef := newSintFromFint(e.coef)

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

	// Coefficient
	if d.IsNeg() != e.IsNeg() {
		dcoef.dist(dcoef, ecoef)
	} else {
		dcoef.add(dcoef, ecoef)
	}

	return newDecimalFromSint(neg, dcoef, scale, minScale)
}

// Sub returns the (possibly rounded) difference between decimals d and e.
//
// Sub returns an error if the integer part of the difference has more than [MaxPrec] digits.
func (d Decimal) Sub(e Decimal) (Decimal, error) {
	return d.SubExact(e, 0)
}

// SubExact is similar to [Decimal.Sub], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) SubExact(e Decimal, scale int) (Decimal, error) {
	return d.AddExact(e.Neg(), scale)
}

// FMA returns the (possibly rounded) [fused multiply-addition] of decimals d, e, and f.
// It computes d * e + f without any intermediate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of products, such as daily interest accrual.
//
// FMA returns an error if the integer part of the result has more than [MaxPrec] digits.
//
// [fused multiply-addition]: https://en.wikipedia.org/wiki/Multiply%E2%80%93accumulate_operation#Fused_multiply%E2%80%93add
func (d Decimal) FMA(e, f Decimal) (Decimal, error) {
	return d.FMAExact(e, f, 0)
}

// FMAExact is similar to [Decimal.FMA], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will return an error.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) FMAExact(e, f Decimal, scale int) (Decimal, error) {
	if scale < 0 || scale > MaxScale {
		return Decimal{}, fmt.Errorf("%v * %v + %v: %w", d, e, f, errScaleRange)
	}
	g, err := d.fmaFint(e, f, scale)
	if err != nil {
		g, err = d.fmaSint(e, f, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("%v * %v + %v: %w", d, e, f, err)
		}
	}
	return g, nil
}

func (d Decimal) fmaFint(e, f Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	ecoef := e.coef
	fcoef := f.coef

	// Coefficient (Multiplication)
	var ok bool
	dcoef, ok = dcoef.mul(ecoef)
	if !ok {
		return Decimal{}, errDecimalOverflow
	}

	// Alignment and scale
	scale := d.Scale() + e.Scale()
	switch {
	case scale > f.Scale():
		fcoef, ok = fcoef.lsh(scale - f.Scale())
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	case scale < f.Scale():
		dcoef, ok = dcoef.lsh(f.Scale() - scale)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		scale = f.Scale()
	}

	// Sign
	var neg bool
	if dcoef > fcoef {
		neg = d.IsNeg() != e.IsNeg()
	} else {
		neg = f.IsNeg()
	}

	// Coefficient (Addition)
	if (d.IsNeg() != e.IsNeg()) != f.IsNeg() {
		dcoef = dcoef.dist(fcoef)
	} else {
		dcoef, ok = dcoef.add(fcoef)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
	}

	return newDecimalFromFint(neg, dcoef, scale, minScale)
}

func (d Decimal) fmaSint(e, f Decimal, minScale int) (Decimal, error) {
	dcoef := newSintFromFint(d.coef)
	ecoef := newSintFromFint(e.coef)
	fcoef := newSintFromFint(f.coef)

	// Coefficient (Multiplication)
	dcoef.mul(dcoef, ecoef)

	// Alignment and scale
	scale := d.Scale() + e.Scale()
	switch {
	case scale > f.Scale():
		fcoef.lsh(fcoef, scale-f.Scale())
	case scale < f.Scale():
		dcoef.lsh(dcoef, f.Scale()-scale)
		scale = f.Scale()
	}

	// Sign
	var neg bool
	if dcoef.cmp(fcoef) > 0 {
		neg = d.IsNeg() != e.IsNeg()
	} else {
		neg = f.IsNeg()
	}

	// Coefficient (Addition)
	if (d.IsNeg() != e.IsNeg()) != f.IsNeg() {
		dcoef.dist(dcoef, fcoef)
	} else {
		dcoef.add(dcoef, fcoef)
	}

	return newDecimalFromSint(neg, dcoef, scale, minScale)
}

// Quo returns the (possibly rounded) quotient of decimals d and e.
//
// Quo returns an error if:
//   - the integer part of the result has more than [MaxPrec] digits;
//   - divisor e is zero.
func (d Decimal) Quo(e Decimal) (Decimal, error) {
	return d.QuoExact(e, 0)
}

// QuoExact is similar to [Decimal.Quo], but it allows you to specify the number of digits
// after the decimal point that should be considered significant.
// If any of the significant digits are lost during rounding, the method will panic.
// This method is useful for financial calculations where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) QuoExact(e Decimal, scale int) (Decimal, error) {
	if scale < 0 || scale > MaxScale {
		return Decimal{}, fmt.Errorf("%v / %v: %w", d, e, errScaleRange)
	}

	// Special case: zero divisor
	if e.IsZero() {
		return Decimal{}, errDivisionByZero
	}

	// Special case: zero dividend
	if d.IsZero() {
		if t := d.Scale() - e.Scale(); scale < t {
			scale = t
		}
		return newDecimalSafe(false, 0, scale)
	}

	// General case
	f, err := d.quoFint(e, scale)
	if err != nil {
		f, err = d.quoSint(e, scale)
		if err != nil {
			return Decimal{}, fmt.Errorf("%v / %v: %w", d, e, err)
		}
	}

	// Trailing zeroes
	if t := d.Scale() - e.Scale(); scale < t {
		scale = t
	}
	f = f.Trim(scale)

	return f, nil
}

func (d Decimal) quoFint(e Decimal, minScale int) (Decimal, error) {
	dcoef := d.coef
	ecoef := e.coef

	// Scale
	scale := d.Scale() - e.Scale()

	// Dividend alignment
	var ok bool
	if p := MaxPrec - dcoef.prec(); p > 0 {
		dcoef, ok = dcoef.lsh(p)
		if !ok {
			return Decimal{}, errDecimalOverflow
		}
		scale = scale + p
	}

	// Divisor alignment
	if t := ecoef.tzeros(); t > 0 {
		ecoef = ecoef.rshDown(t)
		scale = scale + t
	}

	// Coefficient
	dcoef, ok = dcoef.quo(ecoef)
	if !ok {
		return Decimal{}, errInexactDivision
	}

	// Sign
	neg := d.IsNeg() != e.IsNeg()

	return newDecimalFromFint(neg, dcoef, scale, minScale)
}

func (d Decimal) quoSint(e Decimal, minScale int) (Decimal, error) {
	dcoef := newSintFromFint(d.coef)
	ecoef := newSintFromFint(e.coef)

	// Scale
	scale := 2 * MaxScale

	// Divident alignment
	dcoef.lsh(dcoef, scale+e.Scale()-d.Scale())

	// Coefficient
	dcoef.quo(dcoef, ecoef)

	// Sign
	neg := d.IsNeg() != e.IsNeg()

	return newDecimalFromSint(neg, dcoef, scale, minScale)
}

// QuoRem returns the quotient q and remainder r of decimals d and e
// such that d = e * q + r, where q is an integer.
//
// QuoRem returns an error if:
//   - the integer part of the quotient q has more than [MaxPrec] digits;
//   - the divisor e is zero.
func (d Decimal) QuoRem(e Decimal) (q, r Decimal, err error) {
	q, r, err = d.quoRem(e)
	if err != nil {
		return Decimal{}, Decimal{}, fmt.Errorf("⌊%v / %v⌋ and %v mod %v: %w", d, e, d, e, err)
	}
	return q, r, nil
}

func (d Decimal) quoRem(e Decimal) (q, r Decimal, err error) {
	// Quotient
	q, err = d.Quo(e)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	q = q.Trunc(0)

	// Reminder
	r, err = e.Mul(q)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}
	r, err = d.Sub(r)
	if err != nil {
		return Decimal{}, Decimal{}, err
	}

	return q, r, nil
}

// Cmp numerically compares decimals and returns:
//
//	-1 if d < e
//	 0 if d == e
//	+1 if d > e
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
		r = d.cmpSint(e)
	}
	return r
}

func (d Decimal) cmpFint(e Decimal) (int, error) {
	dcoef := d.coef
	ecoef := e.coef

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
	default:
		return 0, nil
	}
}

func (d Decimal) cmpSint(e Decimal) int {
	dcoef := newSintFromFint(d.coef)
	ecoef := newSintFromFint(e.coef)

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
	default:
		return 0
	}
}

// CmpTotal compares decimal representations and returns:
//
//	-1 if d < e
//	-1 if d == e && d.scale >  e.scale
//	 0 if d == e && d.scale == e.scale
//	+1 if d == e && d.scale <  e.scale
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
