package decimal

import (
	"fmt"
	"math/big"
)

// fint (Fast Integer) is a wrapper around uint64.
type fint uint64

const (
	maxFint = fint(9999999999999999999)
)

var (
	pow10 = [...]fint{
		1,                    // 10^0
		10,                   // 10^1
		100,                  // 10^2
		1000,                 // 10^3
		10000,                // 10^4
		100000,               // 10^5
		1000000,              // 10^6
		10000000,             // 10^7
		100000000,            // 10^8
		1000000000,           // 10^9
		10000000000,          // 10^10
		100000000000,         // 10^11
		1000000000000,        // 10^12
		10000000000000,       // 10^13
		100000000000000,      // 10^14
		1000000000000000,     // 10^15
		10000000000000000,    // 10^16
		100000000000000000,   // 10^17
		1000000000000000000,  // 10^18
		10000000000000000000, // 10^19
	}
)

// add calculates x + y and checks overflow.
func (x fint) add(y fint) (fint, bool) {
	if maxFint-x < y {
		return 0, false
	}
	z := x + y
	return z, true
}

// mul calculates x * y and checks overflow.
func (x fint) mul(y fint) (fint, bool) {
	if x == 0 || y == 0 {
		return 0, true
	}
	z := x * y
	if z/y != x {
		return 0, false
	}
	if z > maxFint {
		return 0, false
	}
	return z, true
}

// quo calculates x / y and checks if it is an exact division.
func (x fint) quo(y fint) (fint, bool) {
	if y == 0 {
		return 0, false
	}
	z := x / y
	if y*z != x {
		return 0, false
	}
	return z, true
}

// dist calculates abs(x - y).
func (x fint) dist(y fint) fint {
	if x > y {
		return x - y
	} else {
		return y - x
	}
}

// lsh (Shift Left) calculates x * 10^shift and checks overflow.
func (x fint) lsh(shift int) (fint, bool) {
	// Special cases
	switch {
	case shift == 0:
		return x, true
	case shift == 1 && x < maxFint/10: // to speed up common case
		return x * 10, true
	case shift >= len(pow10):
		return 0, false
	}
	// General case
	y := pow10[shift]
	return x.mul(y)
}

// fsa (Fused Shift and Addition) calculates x * 10^shift + y and checks overflow.
func (x fint) fsa(shift int, y byte) (fint, bool) {
	z, ok := x.lsh(shift)
	if !ok {
		return 0, false
	}
	z, ok = z.add(fint(y))
	if !ok {
		return 0, false
	}
	return z, true
}

func (x fint) isOdd() bool {
	return x&1 != 0
}

// rshEven (Shift Right) calculates x / 10^shift and rounds result using "half to even" rule.
func (x fint) rshEven(shift int) (fint, bool) {
	switch {
	case shift == 0:
		return x, true
	case shift >= len(pow10):
		return 0, false
	}
	y := pow10[shift]
	z := x / y
	r := x - z*y                        // r = x % y
	y = y >> 1                          // y = y / 2, which is safe as y is a multiple of 10
	if y < r || (y == r && z.isOdd()) { // half-to-even
		z++
	}
	return z, true
}

// rshUp (Shift Right) calculates x / 10^shift and rounds result away from 0.
func (x fint) rshUp(shift int) (fint, bool) {
	switch {
	case shift == 0:
		return x, true
	case shift >= len(pow10):
		return 0, false
	}
	y := pow10[shift]
	z := x / y
	r := x - z*y // r = x % y
	if r > 0 {
		z++
	}
	return z, true
}

// rshDown (Shift Right) calculates x / 10^shift and rounds result towards 0.
func (x fint) rshDown(shift int) (fint, bool) {
	switch {
	case shift == 0:
		return x, true
	case shift >= len(pow10):
		return 0, false
	}
	y := pow10[shift]
	return x / y, true
}

// prec returns length of x in decimal digits.
func (x fint) prec() int {
	left, right := 0, len(pow10)
	for left < right {
		mid := (left + right) / 2
		if x < pow10[mid] {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// tzeroes returns number of trailing zeros in x
func (x fint) tzeros() int {
	left, right := 1, x.prec()
	for left < right {
		mid := (right + left) / 2
		pow := pow10[mid]
		if x%pow == 0 {
			left = mid + 1
		} else {
			right = mid
		}
	}
	return left - 1
}

// hasPrec returns true if x has given number of digits or more.
func (x fint) hasPrec(prec int) bool {
	if prec < 1 {
		return true
	}
	return x >= pow10[prec-1]
}

// sint (Slow Integer) is a wrapper around big.Int.
type sint big.Int

var (
	sintOne   = newSint(1)
	sintPow10 = [...]*sint{
		mustParseSint("1"),                                                                                                     // 10^0
		mustParseSint("10"),                                                                                                    // 10^1
		mustParseSint("100"),                                                                                                   // 10^2
		mustParseSint("1000"),                                                                                                  // 10^3
		mustParseSint("10000"),                                                                                                 // 10^4
		mustParseSint("100000"),                                                                                                // 10^5
		mustParseSint("1000000"),                                                                                               // 10^6
		mustParseSint("10000000"),                                                                                              // 10^7
		mustParseSint("100000000"),                                                                                             // 10^8
		mustParseSint("1000000000"),                                                                                            // 10^9
		mustParseSint("10000000000"),                                                                                           // 10^10
		mustParseSint("100000000000"),                                                                                          // 10^11
		mustParseSint("1000000000000"),                                                                                         // 10^12
		mustParseSint("10000000000000"),                                                                                        // 10^13
		mustParseSint("100000000000000"),                                                                                       // 10^14
		mustParseSint("1000000000000000"),                                                                                      // 10^15
		mustParseSint("10000000000000000"),                                                                                     // 10^16
		mustParseSint("100000000000000000"),                                                                                    // 10^17
		mustParseSint("1000000000000000000"),                                                                                   // 10^18
		mustParseSint("10000000000000000000"),                                                                                  // 10^19
		mustParseSint("100000000000000000000"),                                                                                 // 10^20
		mustParseSint("1000000000000000000000"),                                                                                // 10^21
		mustParseSint("10000000000000000000000"),                                                                               // 10^22
		mustParseSint("100000000000000000000000"),                                                                              // 10^23
		mustParseSint("1000000000000000000000000"),                                                                             // 10^24
		mustParseSint("10000000000000000000000000"),                                                                            // 10^25
		mustParseSint("100000000000000000000000000"),                                                                           // 10^26
		mustParseSint("1000000000000000000000000000"),                                                                          // 10^27
		mustParseSint("10000000000000000000000000000"),                                                                         // 10^28
		mustParseSint("100000000000000000000000000000"),                                                                        // 10^29
		mustParseSint("1000000000000000000000000000000"),                                                                       // 10^30
		mustParseSint("10000000000000000000000000000000"),                                                                      // 10^31
		mustParseSint("100000000000000000000000000000000"),                                                                     // 10^32
		mustParseSint("1000000000000000000000000000000000"),                                                                    // 10^33
		mustParseSint("10000000000000000000000000000000000"),                                                                   // 10^34
		mustParseSint("100000000000000000000000000000000000"),                                                                  // 10^35
		mustParseSint("1000000000000000000000000000000000000"),                                                                 // 10^36
		mustParseSint("10000000000000000000000000000000000000"),                                                                // 10^37
		mustParseSint("100000000000000000000000000000000000000"),                                                               // 10^38
		mustParseSint("1000000000000000000000000000000000000000"),                                                              // 10^39
		mustParseSint("10000000000000000000000000000000000000000"),                                                             // 10^40
		mustParseSint("100000000000000000000000000000000000000000"),                                                            // 10^41
		mustParseSint("1000000000000000000000000000000000000000000"),                                                           // 10^42
		mustParseSint("10000000000000000000000000000000000000000000"),                                                          // 10^43
		mustParseSint("100000000000000000000000000000000000000000000"),                                                         // 10^44
		mustParseSint("1000000000000000000000000000000000000000000000"),                                                        // 10^45
		mustParseSint("10000000000000000000000000000000000000000000000"),                                                       // 10^46
		mustParseSint("100000000000000000000000000000000000000000000000"),                                                      // 10^47
		mustParseSint("1000000000000000000000000000000000000000000000000"),                                                     // 10^48
		mustParseSint("10000000000000000000000000000000000000000000000000"),                                                    // 10^49
		mustParseSint("100000000000000000000000000000000000000000000000000"),                                                   // 10^50
		mustParseSint("1000000000000000000000000000000000000000000000000000"),                                                  // 10^51
		mustParseSint("10000000000000000000000000000000000000000000000000000"),                                                 // 10^52
		mustParseSint("100000000000000000000000000000000000000000000000000000"),                                                // 10^53
		mustParseSint("1000000000000000000000000000000000000000000000000000000"),                                               // 10^54
		mustParseSint("10000000000000000000000000000000000000000000000000000000"),                                              // 10^55
		mustParseSint("100000000000000000000000000000000000000000000000000000000"),                                             // 10^56
		mustParseSint("1000000000000000000000000000000000000000000000000000000000"),                                            // 10^57
		mustParseSint("10000000000000000000000000000000000000000000000000000000000"),                                           // 10^58
		mustParseSint("100000000000000000000000000000000000000000000000000000000000"),                                          // 10^59
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000"),                                         // 10^60
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000"),                                        // 10^61
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000"),                                       // 10^62
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000"),                                      // 10^63
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000"),                                     // 10^64
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000"),                                    // 10^65
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000"),                                   // 10^66
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000"),                                  // 10^67
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000"),                                 // 10^68
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000"),                                // 10^69
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000"),                               // 10^70
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000"),                              // 10^71
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000"),                             // 10^72
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000"),                            // 10^73
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000"),                           // 10^74
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000"),                          // 10^75
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000"),                         // 10^76
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000"),                        // 10^77
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000"),                       // 10^78
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000"),                      // 10^79
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000"),                     // 10^80
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000"),                    // 10^81
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000"),                   // 10^82
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000"),                  // 10^83
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),                 // 10^84
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),                // 10^85
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),               // 10^86
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),              // 10^87
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),             // 10^88
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),            // 10^89
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),           // 10^90
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),          // 10^91
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),         // 10^92
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),        // 10^93
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),       // 10^94
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),      // 10^95
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),     // 10^96
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),    // 10^97
		mustParseSint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),   // 10^98
		mustParseSint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),  // 10^99
		mustParseSint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), // 10^100
	}
)

// newSint converts uint to big.Int.
func newSint(i uint) *sint {
	z := new(big.Int).SetUint64(uint64(i))
	return (*sint)(z)
}

// mustParseSint converts string to big.Int.
func mustParseSint(str string) *sint {
	z, ok := new(big.Int).SetString(str, 10)
	if !ok {
		panic(fmt.Sprintf("mustParseSint(%q) failed: parsing error", str)) // unexpected by design
	}
	if z.Sign() < 0 {
		panic(fmt.Sprintf("mustParseSint(%q) failed: negative number", str)) // unexpected by design
	}
	return (*sint)(z)
}

func (x *sint) sign() int {
	return (*big.Int)(x).Sign()
}

func (x *sint) cmp(y *sint) int {
	return (*big.Int)(x).Cmp((*big.Int)(y))
}

// prec returns length of sint in decimal digits.
func (x *sint) prec() int {
	if x.sign() < 0 {
		panic("prec() failed: negative number") // unexpected by design
	}
	if maxSint := sintPow10[len(sintPow10)-1]; x.cmp(maxSint) > 0 {
		panic("prec() failed: number overflow") // unexpected by design
	}
	left, right := 0, len(sintPow10)
	for left < right {
		mid := (left + right) / 2
		if x.cmp(sintPow10[mid]) < 0 {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// hasPrec returns true if x has given number of digits or more.
func (x *sint) hasPrec(prec int) bool {
	if prec < 1 {
		return true
	}
	return x.cmp(sintPow10[prec-1]) >= 0
}

func (z *sint) setSint(x *sint) {
	(*big.Int)(z).Set((*big.Int)(x))
}

func (z *sint) setFint(x fint) {
	(*big.Int)(z).SetUint64(uint64(x))
}

// fint converts big.Int to uint64.
func (x *sint) fint() fint {
	return fint((*big.Int)(x).Uint64())
}

// inc calcualtes z = x + 1.
func (z *sint) inc(x *sint) {
	(*big.Int)(z).Add((*big.Int)(x), (*big.Int)(sintOne))
}

// add calculates z = x + y.
func (z *sint) add(x, y *sint) {
	(*big.Int)(z).Add((*big.Int)(x), (*big.Int)(y))
}

// sub calculates z = x - y.
func (z *sint) sub(x, y *sint) {
	(*big.Int)(z).Sub((*big.Int)(x), (*big.Int)(y))
}

// dist calculates z = abs(x - y).
func (z *sint) dist(x, y *sint) {
	switch x.cmp(y) {
	case 1:
		z.sub(x, y)
	default:
		z.sub(y, x)
	}
}

// dbl calculates z = 2 * x.
func (z *sint) dbl(x *sint) {
	(*big.Int)(z).Lsh((*big.Int)(x), 1)
}

// mul calculates z = x * y.
func (z *sint) mul(x, y *sint) {
	(*big.Int)(z).Mul((*big.Int)(x), (*big.Int)(y))
}

// quo calculates z = x / y
func (z *sint) quo(x, y *sint) {
	(*big.Int)(z).Quo((*big.Int)(x), (*big.Int)(y))
}

// quoRem calculates z and r such that x = z * y + r.
func (z *sint) quoRem(x, y *sint) *sint {
	_, r := (*big.Int)(z).QuoRem((*big.Int)(x), (*big.Int)(y), new(big.Int))
	return (*sint)(r)
}

func (x *sint) isOdd() bool {
	return (*big.Int)(x).Bit(0) != 0
}

// lsh (Shift Left) calculates x * 10^shift.
func (z *sint) lsh(x *sint, shift int) {
	y := sintPow10[shift]
	z.mul(x, y)
}

// fsa (Fused Shift and Addition) calculates x * 10^shift + y
func (z *sint) fsa(shift int, y byte) {
	z.lsh(z, shift)
	z.add(z, newSint(uint(y)))
}

// rshEven (Shift Right) calculates x / 10^shift and rounds result using "half to even" rule.
func (z *sint) rshEven(x *sint, shift int) {
	if shift == 0 {
		z.setSint(x)
		return
	}
	y := sintPow10[shift]
	r := z.quoRem(x, y)
	r.dbl(r) // r = r * 2
	switch y.cmp(r) {
	case -1:
		z.inc(z) // z = z + 1
	case 0:
		// half-to-even
		if z.isOdd() {
			z.inc(z) // z = z + 1
		}
	}
}
