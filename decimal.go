package decimal

import (
	"errors"
	"fmt"
)

// Decimal type is a representation of a finite floating-point decimal number.
// The zero value is the numberic value of 0.
// It is designed to be safe for concurrent use by multiple goroutines.
//
// A decimal type is a struct with three parameters:
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
	coef  fint // the coefficient of the decimal
}

const (
	MaxPrec  = 19      // maximum length of the coefficient in decimal digits
	MaxScale = MaxPrec // maximum number of digits after the decimal point
	maxCoef  = maxFint // maximum absolute value of the coefficient, which is equal to (10^MaxPrec - 1)
)

var (
	errCoefficientOverflow = errors.New("coefficient overflow")
	errInvalidDecimal      = errors.New("invalid decimal")
	errScaleRange          = errors.New("scale out of range")
	errExponentRange       = errors.New("exponent out of range")
	errInexactDivision     = errors.New("inexact division")
	errDivisionByZero      = errors.New("division by zero")
)

func newDecimal(neg bool, coef fint, scale int) (Decimal, error) {
	switch {
	case scale < 0 || scale > MaxScale:
		return Decimal{}, errScaleRange
	case coef > maxCoef:
		return Decimal{}, errCoefficientOverflow
	}
	if coef == 0 {
		neg = false
	}
	return Decimal{neg: neg, coef: coef, scale: int8(scale)}, nil
}

func newDecimalFromRescaledFint(neg bool, coef fint, scale, minScale int) (Decimal, error) {
	var ok bool
	switch {
	case scale > MaxScale:
		coef, ok = coef.rshEven(scale - MaxScale)
		if !ok {
			return Decimal{}, errScaleRange
		}
		scale = MaxScale
	case scale < minScale:
		coef, ok = coef.lsh(minScale - scale)
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
		scale = minScale
	}
	return newDecimal(neg, coef, scale)
}

func newDecimalFromRescaledSint(neg bool, coef *sint, scale, minScale int) (Decimal, error) {
	prec := coef.prec()
	if prec-scale > MaxPrec-minScale {
		return Decimal{}, fmt.Errorf("with %v significant digit(s) after the decimal point, the integer part of a %T can have at most %v digit(s), but it has %v digit(s): %w", minScale, Decimal{}, MaxPrec-minScale, prec-scale, errCoefficientOverflow)
	}
	switch {
	case scale < minScale:
		coef.lsh(coef, minScale-scale)
		scale = minScale
	case scale >= prec && scale > MaxScale: // no integer part
		coef.rshEven(coef, scale-MaxScale)
		scale = MaxScale
	case prec > scale && prec > MaxPrec: // there is an integer part
		coef.rshEven(coef, prec-MaxPrec)
		scale = scale - (prec - MaxPrec)
	}
	// Handle the rare case when rshEven rounded a 19-digit coefficient
	// to a 20-digit coefficient.
	if coef.hasPrec(MaxPrec + 1) {
		return newDecimalFromRescaledSint(neg, coef, scale, minScale)
	}
	return newDecimal(neg, coef.fint(), scale)
}

// New returns a decimal equal to coef / 10^scale.
// New panics if scale is less than 0 or greater than [MaxScale].
func New(coef int64, scale int) Decimal {
	neg := false
	if coef < 0 {
		neg = true
		coef = -coef
	}
	d, err := newDecimal(neg, fint(coef), scale)
	if err != nil {
		panic(fmt.Sprintf("New(%v, %v) failed: %v", coef, scale, err))
	}
	return d
}

// Zero returns decimal with a value 0 but the same scale as d.
func (d Decimal) Zero() Decimal {
	return New(0, d.Scale())
}

// One returns decimal with a value 1 but the same scale as d.
func (d Decimal) One() Decimal {
	coef := int64(pow10[d.Scale()])
	return New(coef, d.Scale())
}

// ULP (Unit in the Last Place) returns the smallest representable positive
// difference between d and the next larger decimal value with the same number
// of digits.
func (d Decimal) ULP() Decimal {
	return New(1, d.Scale())
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
// Parse returns error:
//   - if string does not represent a valid decimal number.
//   - if integer part of the result has more than [MaxPrec] digits.
//   - if exponent is less than -2 * [MaxScale] or greater than 2 * [MaxScale].
func Parse(dec string) (Decimal, error) {
	return ParseExact(dec, 0)
}

// ParseExact is similar to [Parse], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will return error.
// This method is useful for financial calculations, where the scale should be
// equal to or greater than the currency's scale.
func ParseExact(dec string, scale int) (Decimal, error) {
	if scale < 0 || MaxScale < scale {
		return Decimal{}, errScaleRange
	}
	d, err := parseFast(dec, scale)
	if err != nil {
		d, err = parseSlow(dec, scale)
		if err != nil {
			return Decimal{}, err
		}
	}
	return d, nil
}

func parseFast(dec string, minScale int) (Decimal, error) {
	var (
		pos     int
		width   int
		neg     bool
		coef    fint
		scale   int
		hascoef bool
		eneg    bool
		exp     int
		hasexp  bool
		hase    bool
		ok      bool
	)

	width = len(dec)

	// Sign
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
	for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
		hascoef = true
		coef, ok = coef.fsa(1, dec[pos]-'0')
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
		pos++
	}

	// Fraction
	if pos < width && dec[pos] == '.' {
		pos++
		for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
			hascoef = true
			if scale >= 2*MaxPrec {
				return Decimal{}, errCoefficientOverflow
			}
			coef, ok = coef.fsa(1, dec[pos]-'0')
			if !ok {
				return Decimal{}, errCoefficientOverflow
			}
			scale++
			pos++
		}
	}

	// Exponential part
	if pos < width && (dec[pos] == 'e' || dec[pos] == 'E') {
		hase = true
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
			if exp > 2*MaxScale {
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
	if hase && !hasexp {
		return Decimal{}, fmt.Errorf("no exponent: %w", errInvalidDecimal)
	}

	if eneg {
		scale = scale + exp
	} else {
		scale = scale - exp
	}
	if scale > 2*MaxPrec {
		return Decimal{}, errCoefficientOverflow
	}

	return newDecimalFromRescaledFint(neg, coef, scale, minScale)
}

func parseSlow(dec string, minScale int) (Decimal, error) {
	var (
		pos     int
		width   int
		neg     bool
		coef    *sint
		scale   int
		hascoef bool
		eneg    bool
		exp     int
		hasexp  bool
		hasesym bool
	)

	coef = new(sint)
	width = len(dec)

	// Sign
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
	for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
		hascoef = true
		if coef.hasPrec(2 * MaxPrec) {
			return Decimal{}, errCoefficientOverflow
		}
		coef.fsa(1, dec[pos]-'0')
		pos++
	}

	// Fraction
	if pos < width && dec[pos] == '.' {
		pos++
		for pos < width && dec[pos] >= '0' && dec[pos] <= '9' {
			hascoef = true
			if scale >= 2*MaxPrec {
				return Decimal{}, errCoefficientOverflow
			}
			coef.fsa(1, dec[pos]-'0')
			scale++
			pos++
		}
	}

	// Exponential part
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
			if exp > 2*MaxScale {
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

	return newDecimalFromRescaledSint(neg, coef, scale, minScale)
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

// String method implements the [fmt.Stringer] interface and returns
// a string representation of a decimal value.
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

	var (
		buf   [24]byte
		pos   int
		coef  uint64
		scale int
	)

	pos = len(buf) - 1
	coef = d.Coef()
	scale = d.Scale()

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

	// Convert bytes to string
	return string(buf[pos+1:])
}

// UnmarshalText implements [encoding.TextUnmarshaler] interface.
// Also see method [Parse].
//
// [encoding.TextUnmarshaler]: https://pkg.go.dev/encoding#TextUnmarshaler
func (d *Decimal) UnmarshalText(text []byte) error {
	var err error
	*d, err = Parse(string(text))
	return err
}

// MarshalText implements [encoding.TextMarshaler] interface.
// Also see method [Decimal.String].
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

	// Percentage
	if verb == 'k' || verb == 'K' {
		pfactor := New(100, 0)
		d = d.Mul(pfactor)
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
		switch {
		case scale < 0:
			scale = 0
		case scale > d.Scale():
			tzeroes = scale - d.Scale()
			scale = d.Scale()
		}
		d = d.Round(scale)
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

// Prec returns number of digits in the coefficient.
func (d Decimal) Prec() int {
	return d.coef.prec()
}

// Coef returns the coefficient of the decimal.
// Also see method [Decimal.Prec].
func (d Decimal) Coef() uint64 {
	return uint64(d.coef)
}

// Scale returns number of digits after the decimal point.
func (d Decimal) Scale() int {
	return int(d.scale)
}

// MinScale returns the smallest scale that d can be rescaled to without rounding.
// Also see method [Decimal.Reduce].
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

// IsInt returns true if fractional part of d is zero.
func (d Decimal) IsInt() bool {
	return d.coef%pow10[d.Scale()] == 0
}

// IsOne returns true if d == -1 or d == 1.
func (d Decimal) IsOne() bool {
	return d.coef == pow10[d.Scale()]
}

// WithinOne returns true if -1 < d < 1.
func (d Decimal) WithinOne() bool {
	return d.coef < pow10[d.Scale()]
}

// Round returns d that is rounded to the specified number of digits after
// the decimal point.
// If the scale of d is less than the specified scale, the result will be
// zero-padded to the right.
// For financial calculations, the scale should be equal to or greater than the scale
// of the currency.
//
// Round panics if:
//   - the integer part of the result has more than ([MaxPrec] - scale) digits;
//   - the scale is less than 0 or greater than [MaxScale].
func (d Decimal) Round(scale int) Decimal {

	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.Round(%v) failed: %v", d, scale, errScaleRange))
	}

	var (
		coef fint
		f    Decimal
		ok   bool
		err  error
	)

	coef = d.coef

	// Rounding
	switch {
	case scale == d.Scale():
		return d
	case scale < d.Scale():
		coef, ok = coef.rshEven(d.Scale() - scale)
		if !ok {
			panic(fmt.Sprintf("%q.Round(%v) failed: %v", d, scale, errScaleRange)) // unexpected by design
		}
	case d.Scale() < scale:
		coef, ok = coef.lsh(scale - d.Scale())
		if !ok {
			panic(fmt.Sprintf("%q.Round(%v) failed: the integer part of a %T can have at most %v digit(s), but it has %v digit(s): %v", d, scale, d, MaxPrec-scale, d.Prec()-d.Scale(), errCoefficientOverflow))
		}
	}

	// Result
	f, err = newDecimal(d.IsNeg(), coef, scale)
	if err != nil {
		panic(fmt.Sprintf("%q.Round(%v) failed: %v", d, scale, err)) // unexpected by design
	}
	return f
}

// Quantize returns d that is rounded to the same scale as e.
// The sign and coefficient of e are ignored.
// Also see method [Decimal.Round].
//
// Qunatize panics if the integer part of d has more than ([MaxPrec] - e.Scale()) digits.
func (d Decimal) Quantize(e Decimal) Decimal {
	return d.Round(e.Scale())
}

// Trunc returns d that is truncated to the specified number of digits after
// the decimal point.
// If the scale of d is less than the specified scale, the result will be
// zero-padded to the right.
// Also see method [Decimal.Reduce].
//
// Trunc panics if:
//   - the integer part of the result has more than ([MaxPrec] - scale) digits;
//   - the scale is less than 0 or greater than [MaxScale].
func (d Decimal) Trunc(scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.Trunc(%v) failed: %v", d, scale, errScaleRange))
	}

	var (
		coef fint
		f    Decimal
		ok   bool
		err  error
	)

	coef = d.coef

	// Truncating
	switch {
	case scale == d.Scale():
		return d
	case scale < d.Scale():
		coef, ok = coef.rshDown(d.Scale() - scale)
		if !ok {
			panic(fmt.Sprintf("%q.Trunc(%v) failed: %v", d, scale, errScaleRange)) // unexpected by design
		}
	case d.Scale() < scale:
		coef, ok = coef.lsh(scale - d.Scale())
		if !ok {
			panic(fmt.Sprintf("%q.Trunc(%v) failed: the integer part of a %T can have at most %v digit(s), but it has %v digit(s): %v", d, scale, d, MaxPrec-scale, d.Prec()-d.Scale(), errCoefficientOverflow))
		}
	}

	// Result
	f, err = newDecimal(d.IsNeg(), coef, scale)
	if err != nil {
		panic(fmt.Sprintf("%q.Trunc(%v) failed: %v", d, scale, err)) // unexpected by design
	}
	return f
}

// Ceil returns d that is rounded up to the specified number of digits after
// the decimal point.
// If the scale of d is less than the specified scale, the result will be
// zero-padded to the right.
// Also see method [Decimal.Floor].
//
// Ceil panics if:
//   - the integer part of the result has more than ([MaxPrec] - scale) digits;
//   - the scale is less than 0 or greater than [MaxScale].
func (d Decimal) Ceil(scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.Ceil(%v) failed: %v", d, scale, errScaleRange))
	}

	var (
		coef fint
		f    Decimal
		ok   bool
		err  error
	)

	coef = d.coef

	// Rounding up
	switch {
	case scale == d.Scale():
		return d
	case scale < d.Scale():
		if d.IsNeg() {
			coef, ok = coef.rshDown(d.Scale() - scale)
		} else {
			coef, ok = coef.rshUp(d.Scale() - scale)
		}
		if !ok {
			panic(fmt.Sprintf("%q.Ceil(%v) failed: %v", d, scale, errScaleRange)) // unexpected by design
		}
	case d.Scale() < scale:
		coef, ok = coef.lsh(scale - d.Scale())
		if !ok {
			panic(fmt.Sprintf("%q.Ceil(%v) failed: the integer part of a %T can have at most %v digit(s), but it has %v digit(s): %v", d, scale, d, MaxPrec-scale, d.Prec()-d.Scale(), errCoefficientOverflow))
		}
	}

	// Result
	f, err = newDecimal(d.IsNeg(), coef, scale)
	if err != nil {
		panic(fmt.Sprintf("%q.Ceil(%v) failed: %v", d, scale, err)) // unexpected by design
	}
	return f
}

// Floor returns d that is rounded down to the specified number of digits after
// the decimal point.
// If the scale of d is less than the specified scale, the result will be
// zero-padded to the right.
// Also see method [Decimal.Ceil].
//
// Floor panics if:
//   - the integer part of the result has more than ([MaxPrec] - scale) digits;
//   - the scale is less than 0 or greater than [MaxScale].
func (d Decimal) Floor(scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.Floor(%v) failed: %v", d, scale, errScaleRange))
	}

	var (
		coef fint
		f    Decimal
		ok   bool
		err  error
	)

	coef = d.coef

	// Rounding down
	switch {
	case scale == d.Scale():
		return d
	case scale < d.Scale():
		if d.IsNeg() {
			coef, ok = coef.rshUp(d.Scale() - scale)
		} else {
			coef, ok = coef.rshDown(d.Scale() - scale)
		}
		if !ok {
			panic(fmt.Sprintf("%q.Floor(%v) failed: %v", d, scale, errScaleRange)) // unexpected by design
		}
	case d.Scale() < scale:
		coef, ok = coef.lsh(scale - d.Scale())
		if !ok {
			panic(fmt.Sprintf("%q.Floor(%v) failed: the integer part of %T can have at most %v digit(s), but it has %v digit(s): %v", d, scale, d, MaxPrec-scale, d.Prec()-d.Scale(), errCoefficientOverflow))
		}
	}

	// Result
	f, err = newDecimal(d.IsNeg(), coef, scale)
	if err != nil {
		panic(fmt.Sprintf("%q.Floor(%v) failed: %v", d, scale, err)) // unexpected by design
	}
	return f
}

// Reduce returns d with all trailing zeros removed.
func (d Decimal) Reduce() Decimal {
	return d.Trunc(d.MinScale())
}

// Neg returns d with opposite sign.
func (d Decimal) Neg() Decimal {
	f, err := newDecimal(!d.IsNeg(), d.coef, d.Scale())
	if err != nil {
		panic(fmt.Sprintf("%q.Neg() failed: %v", d, err)) // unexpected by design
	}
	return f
}

// Abs returns absolute value of d.
func (d Decimal) Abs() Decimal {
	f, err := newDecimal(false, d.coef, d.Scale())
	if err != nil {
		panic(fmt.Sprintf("%q.Abs() failed: %v", d, err)) // unexpected by design
	}
	return f
}

// CopySign returns d with the same sign as e.
// If e is zero, sign of the result remains unchanged.
func (d Decimal) CopySign(e Decimal) Decimal {
	switch {
	case e.IsZero():
		return d
	case d.IsNeg() != e.IsNeg():
		return d.Neg()
	default:
		return d
	}
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

// IsPos returns true if d > 0.
func (d Decimal) IsPos() bool {
	return d.coef != 0 && !d.neg
}

// IsNeg returns true if d < 0.
func (d Decimal) IsNeg() bool {
	return d.neg
}

// IsZero returns true if d == 0.
func (d Decimal) IsZero() bool {
	return d.coef == 0
}

// Mul returns (possibly rounded) product of d and e.
//
// Mul panics if the integer part of the product has more than [MaxPrec] digits.
func (d Decimal) Mul(e Decimal) Decimal {
	return d.MulExact(e, 0)
}

// MulExact is similar to [Decimal.Mul], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will panic.
// This method is useful for financial calculations, where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) MulExact(e Decimal, scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.MulExact(%q, %v) failed: %v", d, e, scale, errScaleRange))
	}
	f, err := mulFast(d, e, scale)
	if err != nil {
		f, err = mulSlow(d, e, scale)
		if err != nil {
			panic(fmt.Sprintf("%q.MulExact(%q, %v) failed: %v", d, e, scale, err))
		}
	}
	return f
}

func mulFast(d, e Decimal, minScale int) (Decimal, error) {

	var (
		dcoef fint
		ecoef fint
		neg   bool
		scale int
		ok    bool
	)

	dcoef = d.coef
	ecoef = e.coef

	// Coefficient
	dcoef, ok = dcoef.mul(ecoef)
	if !ok {
		return Decimal{}, errCoefficientOverflow
	}

	// Sign
	neg = d.IsNeg() != e.IsNeg()

	// Scale
	scale = d.Scale() + e.Scale()

	return newDecimalFromRescaledFint(neg, dcoef, scale, minScale)
}

func mulSlow(d, e Decimal, minScale int) (Decimal, error) {

	var (
		dcoef *sint
		ecoef *sint
		neg   bool
		scale int
	)

	dcoef = new(sint)
	ecoef = new(sint)
	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)

	// Coefficient
	dcoef.mul(dcoef, ecoef)

	// Sign
	neg = d.IsNeg() != e.IsNeg()

	// Scale
	scale = d.Scale() + e.Scale()

	return newDecimalFromRescaledSint(neg, dcoef, scale, minScale)
}

// Pow returns (possibly rounded) d raised to the exp.
//
// Pow panics if the integer part of the power has more than [MaxPrec] digits.
func (d Decimal) Pow(exp int) Decimal {
	// Special case
	if exp == 0 {
		return New(1, 0)
	}
	// General case
	f := d.Pow(exp / 2)
	if exp%2 == 0 {
		return f.Mul(f)
	}
	if exp > 0 {
		return f.Mul(f).Mul(d)
	}
	return f.Mul(f).Quo(d)
}

// Add returns (possibly rounded) sum of d and e.
//
// Add panics if the integer part of the sum has more than [MaxPrec] digits.
func (d Decimal) Add(e Decimal) Decimal {
	return d.AddExact(e, 0)
}

// AddExact is similar to [Decimal.Add], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will panic.
// This method is useful for financial calculations, where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) AddExact(e Decimal, scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.AddExact(%q, %v) failed: %v", d, e, scale, errScaleRange))
	}
	f, err := addFast(d, e, scale)
	if err != nil {
		f, err = addSlow(d, e, scale)
		if err != nil {
			panic(fmt.Sprintf("%q.AddExact(%q, %v) failed: %v", d, e, scale, err))
		}
	}
	return f
}

func addFast(d, e Decimal, minScale int) (Decimal, error) {

	var (
		dcoef fint
		ecoef fint
		neg   bool
		scale int
		ok    bool
	)

	dcoef = d.coef
	ecoef = e.coef

	// Alignment and scale
	switch {
	case d.Scale() == e.Scale():
		scale = d.Scale()
	case e.Scale() < d.Scale():
		scale = d.Scale()
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
	case d.Scale() < e.Scale():
		scale = e.Scale()
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
	}

	// Sign
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
			return Decimal{}, errCoefficientOverflow
		}
	}

	return newDecimalFromRescaledFint(neg, dcoef, scale, minScale)
}

func addSlow(d, e Decimal, minScale int) (Decimal, error) {

	var (
		dcoef *sint
		ecoef *sint
		neg   bool
		scale int
	)

	dcoef = new(sint)
	ecoef = new(sint)
	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)

	// Alignment and scale
	switch {
	case d.Scale() == e.Scale():
		scale = d.Scale()
	case e.Scale() < d.Scale():
		ecoef.lsh(ecoef, d.Scale()-e.Scale())
		scale = d.Scale()
	case d.Scale() < e.Scale():
		dcoef.lsh(dcoef, e.Scale()-d.Scale())
		scale = e.Scale()
	}

	// Sign
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

	return newDecimalFromRescaledSint(neg, dcoef, scale, minScale)
}

// Sub returns (possibly rounded) difference of d and e.
//
// Sub panics if the integer part of the difference has more than [MaxPrec] digits.
func (d Decimal) Sub(e Decimal) Decimal {
	return d.SubExact(e, 0)
}

// SubExact is similar to [Decimal.Sub], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will panic.
// This method is useful for financial calculations, where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) SubExact(e Decimal, scale int) Decimal {
	return d.AddExact(e.Neg(), scale)
}

// FMA returns (possibly rounded) [fused multiply-addition] of d, e, and f.
// It computes d * e + f without any intermeddiate rounding.
// This method is useful for improving the accuracy and performance of algorithms
// that involve the accumulation of products, such as daily interest accrual.
//
// [fused multiply-addition]: https://en.wikipedia.org/wiki/Multiply%E2%80%93accumulate_operation#Fused_multiply%E2%80%93add
func (d Decimal) FMA(e, f Decimal) Decimal {
	return d.FMAExact(e, f, 0)
}

// FMAExact is similar to [Decimal.FMA], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will panic.
// This method is useful for financial calculations, where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) FMAExact(e, f Decimal, scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.FMAExact(%q, %q, %v) failed: %v", d, e, f, scale, errScaleRange))
	}
	g, err := fmaFast(d, e, f, scale)
	if err != nil {
		g, err = fmaSlow(d, e, f, scale)
		if err != nil {
			panic(fmt.Sprintf("%q.FMAExact(%q, %q, %v) failed: %v", d, e, f, scale, err))
		}
	}
	return g
}

func fmaFast(d, e, f Decimal, minScale int) (Decimal, error) {

	var (
		dcoef fint
		ecoef fint
		fcoef fint
		neg   bool
		scale int
		ok    bool
	)

	dcoef = d.coef
	ecoef = e.coef
	fcoef = f.coef

	// Coefficient (Multiplication)
	dcoef, ok = dcoef.mul(ecoef)
	if !ok {
		return Decimal{}, errCoefficientOverflow
	}

	// Alignment and scale
	scale = d.Scale() + e.Scale()
	switch {
	case f.Scale() < scale:
		fcoef, ok = fcoef.lsh(scale - f.Scale())
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
	case scale < f.Scale():
		dcoef, ok = dcoef.lsh(f.Scale() - scale)
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
		scale = f.Scale()
	}

	// Sign
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
			return Decimal{}, errCoefficientOverflow
		}
	}

	return newDecimalFromRescaledFint(neg, dcoef, scale, minScale)
}

func fmaSlow(d, e, f Decimal, minScale int) (Decimal, error) {

	var (
		dcoef *sint
		ecoef *sint
		fcoef *sint
		neg   bool
		scale int
	)

	dcoef = new(sint)
	ecoef = new(sint)
	fcoef = new(sint)
	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)
	fcoef.setFint(f.coef)

	// Coefficient (Multiplication)
	dcoef.mul(dcoef, ecoef)

	// Alignment and scale
	scale = d.Scale() + e.Scale()
	switch {
	case f.Scale() < scale:
		fcoef.lsh(fcoef, scale-f.Scale())
	case scale < f.Scale():
		dcoef.lsh(dcoef, f.Scale()-scale)
		scale = f.Scale()
	}

	// Sign
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

	return newDecimalFromRescaledSint(neg, dcoef, scale, minScale)
}

// Quo returns (possibly rounded) quotient of d and e.
//
// Quo panics if:
//   - the integer part of the result has more than [MaxPrec] digits;
//   - e is 0.
func (d Decimal) Quo(e Decimal) Decimal {
	return d.QuoExact(e, 0)
}

// QuoExact is similar to [Decimal.Quo], but it allows you to specify how many digits
// after the decimal point should be considered significant.
// If any of the significant digits are lost during rounding, the method will panic.
// This method is useful for financial calculations, where the scale should be
// equal to or greater than the currency's scale.
func (d Decimal) QuoExact(e Decimal, scale int) Decimal {
	if scale < 0 || MaxScale < scale {
		panic(fmt.Sprintf("%q.QuoExact(%q, %v) failed: %v", d, e, scale, errScaleRange))
	}

	// Special case: zero divisor
	if e.IsZero() {
		panic(fmt.Sprintf("%q.QuoExact(%q, %v) failed: %v", d, e, scale, errDivisionByZero))
	}

	// Special case: zero dividend
	if d.IsZero() {
		fscale := scale
		if t := d.Scale() - e.Scale(); fscale < t {
			fscale = t
		}
		f, err := newDecimal(d.IsNeg(), d.coef, fscale)
		if err != nil {
			panic(fmt.Sprintf("%q.QuoExact(%q) failed: zero dividend: %v", d, e, err)) // unexpected by design
		}
		return f
	}

	// General case
	f, err := quoFast(d, e, scale)
	if err != nil {
		f, err = quoSlow(d, e, scale)
		if err != nil {
			panic(fmt.Sprintf("%q.QuoExact(%q, %v) failed: %v", d, e, scale, err))
		}
	}

	// Trailing zeroes
	if t := d.Scale() - e.Scale(); scale < t {
		scale = t
	}
	if m := f.MinScale(); scale < m {
		scale = m
	}
	f = f.Trunc(scale)

	return f
}

func quoFast(d, e Decimal, minScale int) (Decimal, error) {

	var (
		dcoef fint
		ecoef fint
		neg   bool
		scale int
		ok    bool
	)

	dcoef = d.coef
	ecoef = e.coef

	// Scale
	scale = d.Scale() - e.Scale()

	// Dividend alignment
	if p := MaxPrec - dcoef.prec(); p > 0 {
		dcoef, ok = dcoef.lsh(p)
		if !ok {
			return Decimal{}, errCoefficientOverflow
		}
		scale = scale + p
	}

	// Divisor alignment
	if t := ecoef.tzeros(); t > 0 {
		ecoef, ok = ecoef.rshDown(t)
		if !ok {
			return Decimal{}, errScaleRange
		}
		scale = scale + t
	}

	// Coefficient
	dcoef, ok = dcoef.quo(ecoef)
	if !ok {
		return Decimal{}, errInexactDivision // inexact division
	}

	// Sign
	neg = d.IsNeg() != e.IsNeg()

	return newDecimalFromRescaledFint(neg, dcoef, scale, minScale)
}

func quoSlow(d, e Decimal, minScale int) (Decimal, error) {

	var (
		dcoef *sint
		ecoef *sint
		neg   bool
		scale int
	)

	dcoef = new(sint)
	ecoef = new(sint)
	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)

	// Scale
	scale = 2 * MaxScale

	// Divident alignment
	dcoef.lsh(dcoef, scale+e.Scale()-d.Scale())

	// Coefficient
	dcoef.quo(dcoef, ecoef)

	// Sign
	neg = d.IsNeg() != e.IsNeg()

	return newDecimalFromRescaledSint(neg, dcoef, scale, minScale)
}

// QuoRem returns the quotient and remainder of d and e such that d = q * e + r.
//
// QuoRem panics if:
//   - the integer part of the quotient has more than [MaxPrec] digits;
//   - the divisor is 0.
func (d Decimal) QuoRem(e Decimal) (Decimal, Decimal) {
	q := d.Quo(e).Trunc(0)
	r := d.Sub(e.Mul(q))
	return q, r
}

// Cmp compares d and e numerically and returns:
//
//	-1 if d < e
//	 0 if d == e
//	+1 if d > e
func (d Decimal) Cmp(e Decimal) int {

	// Special case: different signs
	switch {
	case e.Sign() < d.Sign():
		return 1
	case d.Sign() < e.Sign():
		return -1
	}

	// General case
	r, err := cmpFast(d, e)
	if err != nil {
		r = cmpSlow(d, e)
	}
	return r
}

func cmpFast(d, e Decimal) (int, error) {

	var (
		dcoef fint
		ecoef fint
		ok    bool
	)

	dcoef = d.coef
	ecoef = e.coef

	// Alignment
	switch {
	case e.Scale() < d.Scale():
		ecoef, ok = ecoef.lsh(d.Scale() - e.Scale())
		if !ok {
			return 0, errCoefficientOverflow
		}
	case d.Scale() < e.Scale():
		dcoef, ok = dcoef.lsh(e.Scale() - d.Scale())
		if !ok {
			return 0, errCoefficientOverflow
		}
	}

	// Comparison
	switch {
	case ecoef < dcoef:
		return d.Sign(), nil
	case dcoef < ecoef:
		return -e.Sign(), nil
	default:
		return 0, nil
	}
}

func cmpSlow(d, e Decimal) int {

	var (
		dcoef *sint
		ecoef *sint
	)

	dcoef = new(sint)
	ecoef = new(sint)
	dcoef.setFint(d.coef)
	ecoef.setFint(e.coef)

	// Alignment
	switch {
	case e.Scale() < d.Scale():
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

// CmpTotal compares representation of d and e and returns:
//
//	-1 if d < e
//	-1 if d == e && d.scale > e.scale
//	 0 if d == e && d.scale == e.scale
//	+1 if d == e && d.scale < e.scale
//	+1 if d > e
//
// Also see method [Decimal.Cmp].
func (d Decimal) CmpTotal(e Decimal) int {
	switch d.Cmp(e) {
	case -1:
		return -1
	case 1:
		return 1
	}
	switch {
	case e.Scale() < d.Scale():
		return -1
	case d.Scale() < e.Scale():
		return 1
	}
	return 0
}

// Max returns maximum of d and e.
// Also see method [Decimal.CmpTotal]
func (d Decimal) Max(e Decimal) Decimal {
	if d.CmpTotal(e) >= 0 {
		return d
	} else {
		return e
	}
}

// Min returns minimum of d and e.
// Also see method [Decimal.CmpTotal]
func (d Decimal) Min(e Decimal) Decimal {
	if d.CmpTotal(e) <= 0 {
		return d
	} else {
		return e
	}
}
