package decimal

import (
	"math/big"
)

// fint (Fast Integer) is a wrapper around uint64.
type fint uint64

// maxFint is a maximum value of fint.
const maxFint = fint(9_999_999_999_999_999_999)

// pow10 is a cache of powers of 10.
var pow10 = [...]fint{
	1,                          // 10^0
	10,                         // 10^1
	100,                        // 10^2
	1_000,                      // 10^3
	10_000,                     // 10^4
	100_000,                    // 10^5
	1_000_000,                  // 10^6
	10_000_000,                 // 10^7
	100_000_000,                // 10^8
	1_000_000_000,              // 10^9
	10_000_000_000,             // 10^10
	100_000_000_000,            // 10^11
	1_000_000_000_000,          // 10^12
	10_000_000_000_000,         // 10^13
	100_000_000_000_000,        // 10^14
	1_000_000_000_000_000,      // 10^15
	10_000_000_000_000_000,     // 10^16
	100_000_000_000_000_000,    // 10^17
	1_000_000_000_000_000_000,  // 10^18
	10_000_000_000_000_000_000, // 10^19
}

// add calculates x + y and checks overflow.
func (x fint) add(y fint) (z fint, ok bool) {
	if maxFint-x < y {
		return 0, false
	}
	z = x + y
	return z, true
}

// mul calculates x * y and checks overflow.
func (x fint) mul(y fint) (z fint, ok bool) {
	if x == 0 || y == 0 {
		return 0, true
	}
	z = x * y
	if z/y != x {
		return 0, false
	}
	if z > maxFint {
		return 0, false
	}
	return z, true
}

// quo calculates x / y and checks overflow or inexact division.
func (x fint) quo(y fint) (z fint, ok bool) {
	if y == 0 {
		return 0, false
	}
	z = x / y
	if y*z != x {
		return 0, false
	}
	return z, true
}

// dist calculates abs(x - y).
func (x fint) dist(y fint) fint {
	if x > y {
		return x - y
	}
	return y - x
}

// lsh (Left Shift) calculates x * 10^shift and checks overflow.
func (x fint) lsh(shift int) (z fint, ok bool) {
	// Special cases
	switch {
	case x == 0:
		return 0, true
	case shift <= 0:
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
func (x fint) fsa(shift int, y byte) (z fint, ok bool) {
	z, ok = x.lsh(shift)
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

// rshHalfEven (Right Shift) calculates x / 10^shift and rounds result
// using "half to even" rule.
func (x fint) rshHalfEven(shift int) fint {
	// Special cases
	switch {
	case x == 0:
		return 0
	case shift <= 0:
		return x
	case shift >= len(pow10):
		return 0
	}
	// General case
	y := pow10[shift]
	z := x / y
	r := x - z*y                        // r = x % y
	y = y >> 1                          // y = y / 2, which is safe as y is a multiple of 10
	if y < r || (y == r && z.isOdd()) { // half-to-even
		z++
	}
	return z
}

// rshUp (Right Shift) calculates x / 10^shift and rounds result away from zero.
func (x fint) rshUp(shift int) fint {
	// Special cases
	switch {
	case x == 0:
		return 0
	case shift <= 0:
		return x
	case shift >= len(pow10):
		return 1
	}
	// General case
	y := pow10[shift]
	z := x / y
	r := x - z*y // r = x % y
	if r > 0 {
		z++
	}
	return z
}

// rshDown (Right Shift) calculates x / 10^shift and rounds result towards zero.
func (x fint) rshDown(shift int) fint {
	// Special cases
	switch {
	case x == 0:
		return 0
	case shift <= 0:
		return x
	case shift >= len(pow10):
		return 0
	}
	// General case
	y := pow10[shift]
	return x / y
}

// prec returns length of x in decimal digits.
// prec assumes that 0 has zero digits.
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

// tzeroes returns number of trailing zeros in x.
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
// hasPrec assumes that 0 has zero digits.
// x.hasPrec() is significantly faster than (x.prec() >= prec).
func (x fint) hasPrec(prec int) bool {
	switch {
	case prec < 1:
		return true
	case prec > len(pow10):
		return false
	}
	return x >= pow10[prec-1]
}

// sint (Slow Integer) is a wrapper around big.Int.
type sint big.Int

// spow10 is a cache of powers of 10.
var spow10 = [...]*sint{
	newSintFromPow10(0),
	newSintFromPow10(1),
	newSintFromPow10(2),
	newSintFromPow10(3),
	newSintFromPow10(4),
	newSintFromPow10(5),
	newSintFromPow10(6),
	newSintFromPow10(7),
	newSintFromPow10(8),
	newSintFromPow10(9),
	newSintFromPow10(10),
	newSintFromPow10(11),
	newSintFromPow10(12),
	newSintFromPow10(13),
	newSintFromPow10(14),
	newSintFromPow10(15),
	newSintFromPow10(16),
	newSintFromPow10(17),
	newSintFromPow10(18),
	newSintFromPow10(19),
	newSintFromPow10(20),
	newSintFromPow10(21),
	newSintFromPow10(22),
	newSintFromPow10(23),
	newSintFromPow10(24),
	newSintFromPow10(25),
	newSintFromPow10(26),
	newSintFromPow10(27),
	newSintFromPow10(28),
	newSintFromPow10(29),
	newSintFromPow10(30),
	newSintFromPow10(31),
	newSintFromPow10(32),
	newSintFromPow10(33),
	newSintFromPow10(34),
	newSintFromPow10(35),
	newSintFromPow10(36),
	newSintFromPow10(37),
	newSintFromPow10(38),
	newSintFromPow10(39),
	newSintFromPow10(40),
	newSintFromPow10(41),
	newSintFromPow10(42),
	newSintFromPow10(43),
	newSintFromPow10(44),
	newSintFromPow10(45),
	newSintFromPow10(46),
	newSintFromPow10(47),
	newSintFromPow10(48),
	newSintFromPow10(49),
	newSintFromPow10(50),
	newSintFromPow10(51),
	newSintFromPow10(52),
	newSintFromPow10(53),
	newSintFromPow10(54),
	newSintFromPow10(55),
	newSintFromPow10(56),
	newSintFromPow10(57),
	newSintFromPow10(58),
	newSintFromPow10(59),
	newSintFromPow10(60),
	newSintFromPow10(61),
	newSintFromPow10(62),
	newSintFromPow10(63),
	newSintFromPow10(64),
	newSintFromPow10(65),
	newSintFromPow10(66),
	newSintFromPow10(67),
	newSintFromPow10(68),
	newSintFromPow10(69),
	newSintFromPow10(70),
	newSintFromPow10(71),
	newSintFromPow10(72),
	newSintFromPow10(73),
	newSintFromPow10(74),
	newSintFromPow10(75),
	newSintFromPow10(76),
	newSintFromPow10(77),
	newSintFromPow10(78),
	newSintFromPow10(79),
	newSintFromPow10(80),
	newSintFromPow10(81),
	newSintFromPow10(82),
	newSintFromPow10(83),
	newSintFromPow10(84),
	newSintFromPow10(85),
	newSintFromPow10(86),
	newSintFromPow10(87),
	newSintFromPow10(88),
	newSintFromPow10(89),
	newSintFromPow10(90),
	newSintFromPow10(91),
	newSintFromPow10(92),
	newSintFromPow10(93),
	newSintFromPow10(94),
	newSintFromPow10(95),
	newSintFromPow10(96),
	newSintFromPow10(97),
	newSintFromPow10(98),
	newSintFromPow10(99),
}

// newSintFromFint converts fint to *sint.
func newSintFromFint(x fint) *sint {
	z := new(big.Int).SetUint64(uint64(x))
	return (*sint)(z)
}

// newSintFromPow10 returns 10^exp as *sint.
func newSintFromPow10(exp int) *sint {
	z := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exp)), nil)
	return (*sint)(z)
}

func (z *sint) sign() int {
	return (*big.Int)(z).Sign()
}

func (z *sint) cmp(x *sint) int {
	return (*big.Int)(z).Cmp((*big.Int)(x))
}

func (z *sint) string() string {
	return (*big.Int)(z).String()
}

func (z *sint) setSint(x *sint) {
	(*big.Int)(z).Set((*big.Int)(x))
}

func (z *sint) setFint(x fint) {
	(*big.Int)(z).SetUint64(uint64(x))
}

// fint converts *sint to fint.
// If z cannot be represented as fint, the result is undefined.
func (z *sint) fint() fint {
	i := (*big.Int)(z).Uint64()
	return fint(i)
}

// add calculates z = x + y.
func (z *sint) add(x, y *sint) {
	(*big.Int)(z).Add((*big.Int)(x), (*big.Int)(y))
}

// inc calcualtes z = x + 1.
func (z *sint) inc(x *sint) {
	y := spow10[0]
	z.add(x, y)
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

// quo calculates z = x / y.
func (z *sint) quo(x, y *sint) {
	(*big.Int)(z).Quo((*big.Int)(x), (*big.Int)(y))
}

// quoRem calculates z and r such that x = z * y + r.
func (z *sint) quoRem(x, y *sint) *sint {
	_, r := (*big.Int)(z).QuoRem((*big.Int)(x), (*big.Int)(y), new(big.Int))
	return (*sint)(r)
}

func (z *sint) isOdd() bool {
	return (*big.Int)(z).Bit(0) != 0
}

// lsh (Left Shift) calculates x * 10^shift.
func (z *sint) lsh(x *sint, shift int) {
	var y *sint
	if shift < len(spow10) {
		y = spow10[shift]
	} else {
		y = newSintFromPow10(shift)
	}
	z.mul(x, y)
}

// fsa (Fused Shift and Addition) calculates x * 10^shift + y.
func (z *sint) fsa(shift int, y byte) {
	z.lsh(z, shift)
	z.add(z, newSintFromFint(fint(y)))
}

// rshDown (Right Shift) calculates x / 10^shift and rounds result towards zero.
func (z *sint) rshDown(x *sint, shift int) {
	var y *sint
	if shift < len(spow10) {
		y = spow10[shift]
	} else {
		y = newSintFromPow10(shift)
	}
	z.quo(x, y)
}

// rshHalfEven (Right Shift) calculates x / 10^shift and
// rounds result using "half to even" rule.
func (z *sint) rshHalfEven(x *sint, shift int) {
	// Special cases
	switch {
	case x.sign() == 0:
		z.setFint(0)
		return
	case shift <= 0:
		z.setSint(x)
		return
	}
	// General case
	var y *sint
	if shift < len(spow10) {
		y = spow10[shift]
	} else {
		y = newSintFromPow10(shift)
	}
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

// prec returns length of *sint in decimal digits.
// It considers 0 to have zero digits.
// If *sint is negative, the result is unpredictable.
//
// z.prec() provides a more efficient approach than len(z.string())
// when dealing with decimals having less than len(spow10) digits.
func (z *sint) prec() int {
	// Special case
	if z.cmp(spow10[len(spow10)-1]) > 0 {
		return len(z.string())
	}
	// General case
	left, right := 0, len(spow10)
	for left < right {
		mid := (left + right) / 2
		if z.cmp(spow10[mid]) < 0 {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// hasPrec checks if *sint has a given number of digits or more.
// It considers 0 to have zero digits.
// If *sint is negative, the result is unpredictable.
//
// z.hasPrec() provides a more efficient approach than (z.prec() >= prec)
// when dealing with decimals having less than len(spow10) digits.
func (z *sint) hasPrec(prec int) bool {
	// Special cases
	switch {
	case prec < 1:
		return true
	case prec > len(spow10):
		return len(z.string()) >= prec
	}
	// General case
	return z.cmp(spow10[prec-1]) >= 0
}
