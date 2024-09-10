package decimal

import (
	"fmt"
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

// quo calculates x / y and checks inexact division.
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

// quoRem calculates q = ⌊x / y⌋, r = x - y * q.
func (x fint) quoRem(y fint) (q, r fint, ok bool) {
	if y == 0 {
		return 0, 0, false
	}
	q = x / y
	r = x - q*y
	return q, r, true
}

// dist calculates |x - y|.
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

// rshHalfEven (Right Shift) calculates round(x / 10^shift) and rounds result
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

// rshUp (Right Shift) calculates ⌈x / 10^shift⌉ and rounds result away from zero.
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

// rshDown (Right Shift) calculates ⌊x / 10^shift⌋ and rounds result towards zero.
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
	mustParseBint("1"),
	mustParseBint("10"),
	mustParseBint("100"),
	mustParseBint("1000"),
	mustParseBint("10000"),
	mustParseBint("100000"),
	mustParseBint("1000000"),
	mustParseBint("10000000"),
	mustParseBint("100000000"),
	mustParseBint("1000000000"),
	mustParseBint("10000000000"),
	mustParseBint("100000000000"),
	mustParseBint("1000000000000"),
	mustParseBint("10000000000000"),
	mustParseBint("100000000000000"),
	mustParseBint("1000000000000000"),
	mustParseBint("10000000000000000"),
	mustParseBint("100000000000000000"),
	mustParseBint("1000000000000000000"),
	mustParseBint("10000000000000000000"),
	mustParseBint("100000000000000000000"),
	mustParseBint("1000000000000000000000"),
	mustParseBint("10000000000000000000000"),
	mustParseBint("100000000000000000000000"),
	mustParseBint("1000000000000000000000000"),
	mustParseBint("10000000000000000000000000"),
	mustParseBint("100000000000000000000000000"),
	mustParseBint("1000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
	mustParseBint("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
}

// bexp is a cache of powers of e, where bexp[x] = round(exp(x) * 10^38).
var bexp = [...]*bint{
	mustParseBint("100000000000000000000000000000000000000"),
	mustParseBint("271828182845904523536028747135266249776"),
	mustParseBint("738905609893065022723042746057500781318"),
	mustParseBint("2008553692318766774092852965458171789699"),
	mustParseBint("5459815003314423907811026120286087840279"),
	mustParseBint("14841315910257660342111558004055227962349"),
	mustParseBint("40342879349273512260838718054338827960590"),
	mustParseBint("109663315842845859926372023828812143244222"),
	mustParseBint("298095798704172827474359209945288867375597"),
	mustParseBint("810308392757538400770999668943275996501148"),
	mustParseBint("2202646579480671651695790064528424436635351"),
	mustParseBint("5987414171519781845532648579225778161426108"),
	mustParseBint("16275479141900392080800520489848678317020928"),
	mustParseBint("44241339200892050332610277594908828178439131"),
	mustParseBint("120260428416477677774923677076785944941248654"),
	mustParseBint("326901737247211063930185504609172131550573854"),
	mustParseBint("888611052050787263676302374078145035080271982"),
	mustParseBint("2415495275357529821477543518038582387986756735"),
	mustParseBint("6565996913733051113878650325906003356921635579"),
	mustParseBint("17848230096318726084491003378872270388361973317"),
	mustParseBint("48516519540979027796910683054154055868463898894"),
	mustParseBint("131881573448321469720999888374530278509144443738"),
	mustParseBint("358491284613159156168115994597842068922269306504"),
	mustParseBint("974480344624890260003463268482297527764938776404"),
	mustParseBint("2648912212984347229413916215281188234087019861925"),
	mustParseBint("7200489933738587252416135146612615791522353381340"),
	mustParseBint("19572960942883876426977639787609534279203610095070"),
	mustParseBint("53204824060179861668374730434117744165925580428369"),
	mustParseBint("144625706429147517367704742299692885690206232950992"),
	mustParseBint("393133429714404207438862058084352768579694233344390"),
	mustParseBint("1068647458152446214699046865074140165002449500547305"),
	mustParseBint("2904884966524742523108568211167982566676469509029698"),
	mustParseBint("7896296018268069516097802263510822421995619511535233"),
	mustParseBint("21464357978591606462429776153126088036922590605479790"),
	mustParseBint("58346174252745488140290273461039101900365923894110811"),
	mustParseBint("158601345231343072812964462577466012517620395013452615"),
	mustParseBint("431123154711519522711342229285692539078886361678034773"),
	mustParseBint("1171914237280261130877293979119019452167536369446182238"),
	mustParseBint("3185593175711375622032867170129864599954220990518100775"),
	mustParseBint("8659340042399374695360693271926493424970185470019598659"),
	mustParseBint("23538526683701998540789991074903480450887161725455546724"),
	mustParseBint("63984349353005494922266340351557081887933662139685527945"),
	mustParseBint("173927494152050104739468130361123522614798405772500840104"),
	mustParseBint("472783946822934656147445756274428037081975196238093817097"),
	mustParseBint("1285160011435930827580929963214309925780114322075882587192"),
	mustParseBint("3493427105748509534803479723340609953341165649751815426013"),
	mustParseBint("9496119420602448874513364911711832310181715892107998785044"),
	mustParseBint("25813128861900673962328580021527338043163708299304406081061"),
	mustParseBint("70167359120976317386547159988611740545593799872532198375455"),
	mustParseBint("190734657249509969052509984095384844738818973054378340247523"),
}

// bfact is a cache of factorials, where bfact[x] = x!.
var bfact = [...]*bint{
	mustParseBint("1"),
	mustParseBint("1"),
	mustParseBint("2"),
	mustParseBint("6"),
	mustParseBint("24"),
	mustParseBint("120"),
	mustParseBint("720"),
	mustParseBint("5040"),
	mustParseBint("40320"),
	mustParseBint("362880"),
	mustParseBint("3628800"),
	mustParseBint("39916800"),
	mustParseBint("479001600"),
	mustParseBint("6227020800"),
	mustParseBint("87178291200"),
	mustParseBint("1307674368000"),
	mustParseBint("20922789888000"),
	mustParseBint("355687428096000"),
	mustParseBint("6402373705728000"),
	mustParseBint("121645100408832000"),
	mustParseBint("2432902008176640000"),
	mustParseBint("51090942171709440000"),
	mustParseBint("1124000727777607680000"),
	mustParseBint("25852016738884976640000"),
	mustParseBint("620448401733239439360000"),
	mustParseBint("15511210043330985984000000"),
	mustParseBint("403291461126605635584000000"),
	mustParseBint("10888869450418352160768000000"),
	mustParseBint("304888344611713860501504000000"),
	mustParseBint("8841761993739701954543616000000"),
	mustParseBint("265252859812191058636308480000000"),
	mustParseBint("8222838654177922817725562880000000"),
	mustParseBint("263130836933693530167218012160000000"),
	mustParseBint("8683317618811886495518194401280000000"),
	mustParseBint("295232799039604140847618609643520000000"),
	mustParseBint("10333147966386144929666651337523200000000"),
	mustParseBint("371993326789901217467999448150835200000000"),
	mustParseBint("13763753091226345046315979581580902400000000"),
	mustParseBint("523022617466601111760007224100074291200000000"),
	mustParseBint("20397882081197443358640281739902897356800000000"),
	mustParseBint("815915283247897734345611269596115894272000000000"),
	mustParseBint("33452526613163807108170062053440751665152000000000"),
	mustParseBint("1405006117752879898543142606244511569936384000000000"),
	mustParseBint("60415263063373835637355132068513997507264512000000000"),
	mustParseBint("2658271574788448768043625811014615890319638528000000000"),
	mustParseBint("119622220865480194561963161495657715064383733760000000000"),
	mustParseBint("5502622159812088949850305428800254892961651752960000000000"),
	mustParseBint("258623241511168180642964355153611979969197632389120000000000"),
	mustParseBint("12413915592536072670862289047373375038521486354677760000000000"),
	mustParseBint("608281864034267560872252163321295376887552831379210240000000000"),
}

// mustParseBint converts a string to *big.Int, panicking on error.
// Use only for package variable initialization and test code!
func mustParseBint(s string) *bint {
	z, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(fmt.Errorf("mustParseBint(%q) failed: parsing error", s))
	}
	if z.Sign() < 0 {
		panic(fmt.Errorf("mustParseBint(%q) failed: negative number", s))
	}
	return (*bint)(z)
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

// dist calculates z = |x - y|.
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

// hlf (Half) calculates z = ⌊x / 2⌋.
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

// quo calculates z = ⌊x / y⌋.
func (z *bint) quo(x, y *bint) {
	// Passing r to prevent heap allocations.
	r := getBint()
	defer putBint(r)
	z.quoRem(x, y, r)
}

// quoRem calculates z = ⌊x / y⌋, r = x - y * z.
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

// rshDown (Right Shift) calculates z = ⌊x / 10^shift⌋ and rounds
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

// rshHalfEven (Right Shift) calculates z = round(x / 10^shift) and
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

// bpool is a cache of reusable *big.Int instances.
var bpool = sync.Pool{
	New: func() any {
		return (*bint)(new(big.Int))
	},
}

// getBint obtains a *big.Int from the pool.
func getBint() *bint {
	return bpool.Get().(*bint)
}

// putBint returns the *big.Int into the pool.
func putBint(b *bint) {
	bpool.Put(b)
}
