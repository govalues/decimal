package decimal

import (
	"math/big"
	"sync"
)

// fint (Fast INTeger) is a wrapper around uint64.
type fint uint64

// maxFint is a maximum value of fint.
const maxFint = 9_999_999_999_999_999_999

// pow10 is a cache of powers of 10, where pow10[x] = 10^x.
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
	if y == 0 {
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
	if z*y != x {
		return 0, false
	}
	return z, true
}

// quoRem calculates x div y and x mod y.
func (x fint) quoRem(y fint) (q, r fint, ok bool) {
	if y == 0 {
		return 0, 0, false
	}
	q = x / y
	r = x - q*y
	return q, r, true
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

// fsa (Fused Shift and Addition) calculates x * 10^shift + b and checks overflow.
func (x fint) fsa(shift int, b byte) (z fint, ok bool) {
	z, ok = x.lsh(shift)
	if !ok {
		return 0, false
	}
	z, ok = z.add(fint(b))
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
// prec assumes that 0 has no digits.
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

// ntz returns number of trailing zeros in x.
// ntz assumes that 0 has no trailing zeros.
func (x fint) ntz() int {
	left, right := 1, x.prec()
	for left < right {
		mid := (left + right) / 2
		if x%pow10[mid] == 0 {
			left = mid + 1
		} else {
			right = mid
		}
	}
	return left - 1
}

// hasPrec returns true if x has given number of digits or more.
// hasPrec assumes that 0 has no digits.
//
// x.hasPrec(p) is significantly faster than x.prec() >= p.
func (x fint) hasPrec(prec int) bool {
	// Special cases
	switch {
	case prec < 1:
		return true
	case prec > len(pow10):
		return false
	}
	// General case
	return x >= pow10[prec-1]
}

// bint (Big INTeger) is a wrapper around big.Int.
type bint big.Int

// bpow10 is a cache of powers of 10, where bpow10[x] = 10^x.
var bpow10 = [...]*bint{
	newBintFromPow10(0),
	newBintFromPow10(1),
	newBintFromPow10(2),
	newBintFromPow10(3),
	newBintFromPow10(4),
	newBintFromPow10(5),
	newBintFromPow10(6),
	newBintFromPow10(7),
	newBintFromPow10(8),
	newBintFromPow10(9),
	newBintFromPow10(10),
	newBintFromPow10(11),
	newBintFromPow10(12),
	newBintFromPow10(13),
	newBintFromPow10(14),
	newBintFromPow10(15),
	newBintFromPow10(16),
	newBintFromPow10(17),
	newBintFromPow10(18),
	newBintFromPow10(19),
	newBintFromPow10(20),
	newBintFromPow10(21),
	newBintFromPow10(22),
	newBintFromPow10(23),
	newBintFromPow10(24),
	newBintFromPow10(25),
	newBintFromPow10(26),
	newBintFromPow10(27),
	newBintFromPow10(28),
	newBintFromPow10(29),
	newBintFromPow10(30),
	newBintFromPow10(31),
	newBintFromPow10(32),
	newBintFromPow10(33),
	newBintFromPow10(34),
	newBintFromPow10(35),
	newBintFromPow10(36),
	newBintFromPow10(37),
	newBintFromPow10(38),
	newBintFromPow10(39),
	newBintFromPow10(40),
	newBintFromPow10(41),
	newBintFromPow10(42),
	newBintFromPow10(43),
	newBintFromPow10(44),
	newBintFromPow10(45),
	newBintFromPow10(46),
	newBintFromPow10(47),
	newBintFromPow10(48),
	newBintFromPow10(49),
	newBintFromPow10(50),
	newBintFromPow10(51),
	newBintFromPow10(52),
	newBintFromPow10(53),
	newBintFromPow10(54),
	newBintFromPow10(55),
	newBintFromPow10(56),
	newBintFromPow10(57),
	newBintFromPow10(58),
	newBintFromPow10(59),
	newBintFromPow10(60),
	newBintFromPow10(61),
	newBintFromPow10(62),
	newBintFromPow10(63),
	newBintFromPow10(64),
	newBintFromPow10(65),
	newBintFromPow10(66),
	newBintFromPow10(67),
	newBintFromPow10(68),
	newBintFromPow10(69),
	newBintFromPow10(70),
	newBintFromPow10(71),
	newBintFromPow10(72),
	newBintFromPow10(73),
	newBintFromPow10(74),
	newBintFromPow10(75),
	newBintFromPow10(76),
	newBintFromPow10(77),
	newBintFromPow10(78),
	newBintFromPow10(79),
	newBintFromPow10(80),
	newBintFromPow10(81),
	newBintFromPow10(82),
	newBintFromPow10(83),
	newBintFromPow10(84),
	newBintFromPow10(85),
	newBintFromPow10(86),
	newBintFromPow10(87),
	newBintFromPow10(88),
	newBintFromPow10(89),
	newBintFromPow10(90),
	newBintFromPow10(91),
	newBintFromPow10(92),
	newBintFromPow10(93),
	newBintFromPow10(94),
	newBintFromPow10(95),
	newBintFromPow10(96),
	newBintFromPow10(97),
	newBintFromPow10(98),
	newBintFromPow10(99),
}

// newBintFromPow10 creates a *big.Int equal to 10^power.
func newBintFromPow10(power int) *bint {
	z := (*bint)(new(big.Int))
	z.pow10(power)
	return z
}

func (z *bint) sign() int {
	return (*big.Int)(z).Sign()
}

func (z *bint) cmp(x *bint) int {
	return (*big.Int)(z).Cmp((*big.Int)(x))
}

func (z *bint) string() string {
	return (*big.Int)(z).String()
}

func (z *bint) setBint(x *bint) {
	(*big.Int)(z).Set((*big.Int)(x))
}

func (z *bint) setInt64(x int64) {
	(*big.Int)(z).SetInt64(x)
}

func (z *bint) setFint(x fint) {
	(*big.Int)(z).SetUint64(uint64(x))
}

// fint converts *big.Int to uint64.
// If z cannot be represented as uint64, the result is undefined.
func (z *bint) fint() fint {
	f := (*big.Int)(z).Uint64()
	return fint(f)
}

// add calculates z = x + y.
func (z *bint) add(x, y *bint) {
	(*big.Int)(z).Add((*big.Int)(x), (*big.Int)(y))
}

// inc calcualtes z = x + 1.
func (z *bint) inc(x *bint) {
	y := bpow10[0]
	z.add(x, y)
}

// sub calculates z = x - y.
func (z *bint) sub(x, y *bint) {
	(*big.Int)(z).Sub((*big.Int)(x), (*big.Int)(y))
}

// dist calculates z = abs(x - y).
func (z *bint) dist(x, y *bint) {
	switch x.cmp(y) {
	case 1:
		z.sub(x, y)
	default:
		z.sub(y, x)
	}
}

// dbl (Double) calculates z = x * 2.
func (z *bint) dbl(x *bint) {
	(*big.Int)(z).Lsh((*big.Int)(x), 1)
}

// hlf (Half) calculates z = x / 2.
func (z *bint) hlf(x *bint) {
	(*big.Int)(z).Rsh((*big.Int)(x), 1)
}

// mul calculates z = x * y.
func (z *bint) mul(x, y *bint) {
	// Copying x, y to prevent heap allocations.
	if z == x {
		b := getBint()
		defer putBint(b)
		b.setBint(x)
		x = b
	}
	if z == y {
		b := getBint()
		defer putBint(b)
		b.setBint(y)
		y = b
	}
	(*big.Int)(z).Mul((*big.Int)(x), (*big.Int)(y))
}

// exp calculates z = x^y.
// If y is negative, the result is unpredictable.
func (z *bint) exp(x, y *bint) {
	(*big.Int)(z).Exp((*big.Int)(x), (*big.Int)(y), nil)
}

// pow10 calculates z = 10^power.
// If power is negative, the result is unpredictable.
func (z *bint) pow10(power int) {
	x := getBint()
	defer putBint(x)
	x.setInt64(10)
	y := getBint()
	defer putBint(y)
	y.setInt64(int64(power))
	z.exp(x, y)
}

// quo calculates z = x / y.
func (z *bint) quo(x, y *bint) {
	r := getBint()
	defer putBint(r)
	// Passing r to prevent heap allocations.
	z.quoRem(x, y, r)
}

// quoRem calculates z and r such that x = z * y + r.
func (z *bint) quoRem(x, y, r *bint) {
	(*big.Int)(z).QuoRem((*big.Int)(x), (*big.Int)(y), (*big.Int)(r))
}

func (z *bint) isOdd() bool {
	return (*big.Int)(z).Bit(0) != 0
}

// lsh (Left Shift) calculates z = x * 10^shift.
func (z *bint) lsh(x *bint, shift int) {
	var y *bint
	if shift < len(bpow10) {
		y = bpow10[shift]
	} else {
		y = getBint()
		defer putBint(y)
		y.pow10(shift)
	}
	z.mul(x, y)
}

// fsa (Fused Shift and Addition) calculates z = x * 10^shift + f.
func (z *bint) fsa(x *bint, shift int, f fint) {
	y := getBint()
	defer putBint(y)
	y.setFint(f)
	z.lsh(x, shift)
	z.add(z, y)
}

// rshDown (Right Shift) calculates z = x / 10^shift and rounds
// result towards zero.
func (z *bint) rshDown(x *bint, shift int) {
	// Special cases
	switch {
	case x.sign() == 0:
		z.setFint(0)
		return
	case shift <= 0:
		z.setBint(x)
		return
	}
	// General case
	var y *bint
	if shift < len(bpow10) {
		y = bpow10[shift]
	} else {
		y = getBint()
		defer putBint(y)
		y.pow10(shift)
	}
	z.quo(x, y)
}

// rshHalfEven (Right Shift) calculates z = x / 10^shift and
// rounds result using "half to even" rule.
func (z *bint) rshHalfEven(x *bint, shift int) {
	// Special cases
	switch {
	case x.sign() == 0:
		z.setFint(0)
		return
	case shift <= 0:
		z.setBint(x)
		return
	}
	// General case
	var y, r *bint
	r = getBint()
	defer putBint(r)
	if shift < len(bpow10) {
		y = bpow10[shift]
	} else {
		y = getBint()
		defer putBint(y)
		y.pow10(shift)
	}
	z.quoRem(x, y, r)
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

// prec returns length of z in decimal digits.
// prec assumes that 0 has no digits.
// If z is negative, the result is unpredictable.
//
// z.prec() is significantly faster than len(z.string()),
// if z has less than len(bpow10) digits.
func (z *bint) prec() int {
	// Special case
	if z.cmp(bpow10[len(bpow10)-1]) > 0 {
		return len(z.string())
	}
	// General case
	left, right := 0, len(bpow10)
	for left < right {
		mid := (left + right) / 2
		if z.cmp(bpow10[mid]) < 0 {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// hasPrec checks if z has a given number of digits or more.
// hasPrec assumes that 0 has no digits.
// If z is negative, the result is unpredictable.
//
// z.hasPrec(p) is significantly faster than z.prec() >= p,
// if z has no more than len(bpow10) digits.
func (z *bint) hasPrec(prec int) bool {
	// Special cases
	switch {
	case prec < 1:
		return true
	case prec > len(bpow10):
		return len(z.string()) >= prec
	}
	// General case
	return z.cmp(bpow10[prec-1]) >= 0
}

// pool is a cache of reusable *big.Int instances.
var pool = sync.Pool{
	New: func() any {
		return (*bint)(new(big.Int))
	},
}

// getBint obtains a *big.Int from the pool.
func getBint() *bint {
	return pool.Get().(*bint)
}

// putBint returns the *big.Int into the pool.
func putBint(b *bint) {
	pool.Put(b)
}
