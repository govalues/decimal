package decimal_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/govalues/decimal"
)

func evaluate(input string) (decimal.Decimal, error) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return decimal.Decimal{}, fmt.Errorf("no tokens")
	}
	stack := make([]decimal.Decimal, 0, len(tokens))
	for i, token := range tokens {
		var err error
		var result decimal.Decimal
		if token == "+" || token == "-" || token == "*" || token == "/" {
			if len(stack) < 2 {
				return decimal.Decimal{}, fmt.Errorf("not enough operands")
			}
			left := stack[len(stack)-2]
			right := stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			switch token {
			case "+":
				result, err = left.Add(right)
			case "-":
				result, err = left.Sub(right)
			case "*":
				result, err = left.Mul(right)
			case "/":
				result, err = left.Quo(right)
			}
		} else {
			result, err = decimal.Parse(token)
		}
		if err != nil {
			return decimal.Decimal{}, fmt.Errorf("processing token %q at position %v: %w", token, i, err)
		}
		stack = append(stack, result)
	}
	if len(stack) != 1 {
		return decimal.Decimal{}, fmt.Errorf("stack contains %v, expected exactly one item", stack)
	}
	return stack[0], nil
}

// This example implements a simple calculator that evaluates mathematical
// expressions written in [postfix notation].
// The calculator can handle basic arithmetic operations such as addition,
// subtraction, multiplication, and division.
//
// [postfix notation]: https://en.wikipedia.org/wiki/Reverse_Polish_notation
func Example_postfixCalculator() {
	d, err := evaluate("1.23 4.56 + 10 *")
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
	// Output:
	// 57.90
}

func approximate(terms int) (decimal.Decimal, error) {
	pi := decimal.Zero
	denominator := decimal.One
	increment := decimal.Two
	multiplier := decimal.MustParse("4")

	for i := 0; i < terms; i++ {
		term, err := multiplier.Quo(denominator)
		if err != nil {
			return decimal.Decimal{}, err
		}
		pi, err = pi.Add(term)
		if err != nil {
			return decimal.Decimal{}, err
		}
		denominator, err = denominator.Add(increment)
		if err != nil {
			return decimal.Decimal{}, err
		}
		multiplier = multiplier.Neg()
	}
	return pi, nil
}

// This example calculates an approximate value of π using the [Leibniz formula].
// The Leibniz formula is an infinite series that converges to π/4, and is
// given by the equation: 1 - 1/3 + 1/5 - 1/7 + 1/9 - 1/11 + ... = π/4.
// This example computes the series up to the 500,000th term using decimal arithmetic
// and returns the approximate value of π.
//
// [Leibniz formula]: https://en.wikipedia.org/wiki/Leibniz_formula_for_%CF%80
func Example_piApproximation() {
	pi, err := approximate(500000)
	if err != nil {
		panic(err)
	}
	fmt.Println(pi)
	fmt.Println(decimal.Pi)
	// Output:
	// 3.141590653589793206
	// 3.141592653589793238
}

// This example demonstrates the advantage of decimals for financial calculations.
// It computes the sum 0.1 + 0.1 + 0.1 + 0.1 + 0.1 + 0.1 + 0.1 + 0.1 + 0.1 + 0.1.
// In decimal arithmetic, the result is exactly 1.0.
// In float64 arithmetic, the result slightly deviates from 1.0 due to binary
// floating-point representation.
func Example_floatInaccuracy() {
	d := decimal.MustParse("0.0")
	e := decimal.MustParse("0.1")
	for i := 0; i < 10; i++ {
		d, _ = d.Add(e)
	}
	fmt.Println(d)

	f := 0.0
	for i := 0; i < 10; i++ {
		f += 0.1
	}
	fmt.Println(f)
	// Output:
	// 1.0
	// 0.9999999999999999
}

func ExampleMustNew() {
	fmt.Println(decimal.MustNew(567, 0))
	fmt.Println(decimal.MustNew(567, 1))
	fmt.Println(decimal.MustNew(567, 2))
	fmt.Println(decimal.MustNew(567, 3))
	fmt.Println(decimal.MustNew(567, 4))
	// Output:
	// 567
	// 56.7
	// 5.67
	// 0.567
	// 0.0567
}

func ExampleNew() {
	fmt.Println(decimal.New(567, 0))
	fmt.Println(decimal.New(567, 1))
	fmt.Println(decimal.New(567, 2))
	fmt.Println(decimal.New(567, 3))
	fmt.Println(decimal.New(567, 4))
	// Output:
	// 567 <nil>
	// 56.7 <nil>
	// 5.67 <nil>
	// 0.567 <nil>
	// 0.0567 <nil>
}

func ExampleNewFromInt64() {
	fmt.Println(decimal.NewFromInt64(5, 6, 1))
	fmt.Println(decimal.NewFromInt64(5, 6, 2))
	fmt.Println(decimal.NewFromInt64(5, 6, 3))
	fmt.Println(decimal.NewFromInt64(5, 6, 4))
	fmt.Println(decimal.NewFromInt64(5, 6, 5))
	// Output:
	// 5.6 <nil>
	// 5.06 <nil>
	// 5.006 <nil>
	// 5.0006 <nil>
	// 5.00006 <nil>
}

func ExampleNewFromFloat64() {
	fmt.Println(decimal.NewFromFloat64(5.67e-2))
	fmt.Println(decimal.NewFromFloat64(5.67e-1))
	fmt.Println(decimal.NewFromFloat64(5.67e0))
	fmt.Println(decimal.NewFromFloat64(5.67e1))
	fmt.Println(decimal.NewFromFloat64(5.67e2))
	// Output:
	// 0.0567 <nil>
	// 0.567 <nil>
	// 5.67 <nil>
	// 56.7 <nil>
	// 567 <nil>
}

func ExampleDecimal_Zero() {
	d := decimal.MustParse("5")
	e := decimal.MustParse("5.6")
	f := decimal.MustParse("5.67")
	fmt.Println(d.Zero())
	fmt.Println(e.Zero())
	fmt.Println(f.Zero())
	// Output:
	// 0
	// 0.0
	// 0.00
}

func ExampleDecimal_One() {
	d := decimal.MustParse("5")
	e := decimal.MustParse("5.6")
	f := decimal.MustParse("5.67")
	fmt.Println(d.One())
	fmt.Println(e.One())
	fmt.Println(f.One())
	// Output:
	// 1
	// 1.0
	// 1.00
}

func ExampleDecimal_ULP() {
	d := decimal.MustParse("5")
	e := decimal.MustParse("5.6")
	f := decimal.MustParse("5.67")
	fmt.Println(d.ULP())
	fmt.Println(e.ULP())
	fmt.Println(f.ULP())
	// Output:
	// 1
	// 0.1
	// 0.01
}

func ExampleParse() {
	fmt.Println(decimal.Parse("5.67"))
	// Output: 5.67 <nil>
}

func ExampleParseExact() {
	fmt.Println(decimal.ParseExact("5.67", 0))
	fmt.Println(decimal.ParseExact("5.67", 1))
	fmt.Println(decimal.ParseExact("5.67", 2))
	fmt.Println(decimal.ParseExact("5.67", 3))
	fmt.Println(decimal.ParseExact("5.67", 4))
	// Output:
	// 5.67 <nil>
	// 5.67 <nil>
	// 5.67 <nil>
	// 5.670 <nil>
	// 5.6700 <nil>
}

func ExampleMustParse() {
	fmt.Println(decimal.MustParse("-1.23"))
	// Output: -1.23
}

func ExampleDecimal_String() {
	d := decimal.MustParse("1234567890.123456789")
	fmt.Println(d.String())
	// Output: 1234567890.123456789
}

func ExampleDecimal_Float64() {
	d := decimal.MustParse("0.1")
	e := decimal.MustParse("123.456")
	f := decimal.MustParse("1234567890.123456789")
	fmt.Println(d.Float64())
	fmt.Println(e.Float64())
	fmt.Println(f.Float64())
	// Output:
	// 0.1 true
	// 123.456 true
	// 1.2345678901234567e+09 true
}

func ExampleDecimal_Int64() {
	d := decimal.MustParse("5.67")
	fmt.Println(d.Int64(0))
	fmt.Println(d.Int64(1))
	fmt.Println(d.Int64(2))
	fmt.Println(d.Int64(3))
	fmt.Println(d.Int64(4))
	// Output:
	// 6 0 true
	// 5 7 true
	// 5 67 true
	// 5 670 true
	// 5 6700 true
}

type Value struct {
	Number decimal.Decimal `json:"number"`
}

func ExampleDecimal_UnmarshalText() {
	var v Value
	_ = json.Unmarshal([]byte(`{"number": "5.67"}`), &v)
	fmt.Println(v)
	// Output: {5.67}
}

func ExampleDecimal_MarshalText() {
	v := Value{
		Number: decimal.MustParse("5.67"),
	}
	b, _ := json.Marshal(v)
	fmt.Println(string(b))
	// Output: {"number":"5.67"}
}

func ExampleDecimal_Scan() {
	var d decimal.Decimal
	_ = d.Scan("5.67")
	fmt.Println(d)
	// Output: 5.67
}

func ExampleDecimal_Value() {
	d := decimal.MustParse("5.67")
	fmt.Println(d.Value())
	// Output: 5.67 <nil>
}

func ExampleDecimal_Format() {
	d := decimal.MustParse("5.67")
	fmt.Printf("%f\n", d)
	fmt.Printf("%k\n", d)
	// Output:
	// 5.67
	// 567%
}

func ExampleDecimal_Coef() {
	d := decimal.MustParse("-123")
	e := decimal.MustParse("5.7")
	f := decimal.MustParse("0.4")
	fmt.Println(d.Coef())
	fmt.Println(e.Coef())
	fmt.Println(f.Coef())
	// Output:
	// 123
	// 57
	// 4
}

func ExampleDecimal_Prec() {
	d := decimal.MustParse("-123")
	e := decimal.MustParse("5.7")
	f := decimal.MustParse("0.4")
	fmt.Println(d.Prec())
	fmt.Println(e.Prec())
	fmt.Println(f.Prec())
	// Output:
	// 3
	// 2
	// 1
}

func ExampleDecimal_Mul() {
	d := decimal.MustParse("5.7")
	e := decimal.MustParse("3")
	fmt.Println(d.Mul(e))
	// Output: 17.1 <nil>
}

func ExampleDecimal_MulExact() {
	d := decimal.MustParse("5.7")
	e := decimal.MustParse("3")
	fmt.Println(d.MulExact(e, 0))
	fmt.Println(d.MulExact(e, 1))
	fmt.Println(d.MulExact(e, 2))
	fmt.Println(d.MulExact(e, 3))
	fmt.Println(d.MulExact(e, 4))
	// Output:
	// 17.1 <nil>
	// 17.1 <nil>
	// 17.10 <nil>
	// 17.100 <nil>
	// 17.1000 <nil>
}

func ExampleDecimal_FMA() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.FMA(e, f))
	// Output: 10 <nil>
}

func ExampleDecimal_FMAExact() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.FMAExact(e, f, 0))
	fmt.Println(d.FMAExact(e, f, 1))
	fmt.Println(d.FMAExact(e, f, 2))
	fmt.Println(d.FMAExact(e, f, 3))
	fmt.Println(d.FMAExact(e, f, 4))
	// Output:
	// 10 <nil>
	// 10.0 <nil>
	// 10.00 <nil>
	// 10.000 <nil>
	// 10.0000 <nil>
}

func ExampleDecimal_Pow() {
	d := decimal.MustParse("2")
	fmt.Println(d.Pow(-2))
	fmt.Println(d.Pow(-1))
	fmt.Println(d.Pow(0))
	fmt.Println(d.Pow(1))
	fmt.Println(d.Pow(2))
	// Output:
	// 0.25 <nil>
	// 0.5 <nil>
	// 1 <nil>
	// 2 <nil>
	// 4 <nil>
}

func ExampleDecimal_PowExact() {
	d := decimal.MustParse("2")
	fmt.Println(d.PowExact(3, 0))
	fmt.Println(d.PowExact(3, 1))
	fmt.Println(d.PowExact(3, 2))
	fmt.Println(d.PowExact(3, 3))
	fmt.Println(d.PowExact(3, 4))
	// Output:
	// 8 <nil>
	// 8.0 <nil>
	// 8.00 <nil>
	// 8.000 <nil>
	// 8.0000 <nil>
}

func ExampleDecimal_Add() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("8")
	fmt.Println(d.Add(e))
	// Output: 13.67 <nil>
}

func ExampleDecimal_AddExact() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("8")
	fmt.Println(d.AddExact(e, 0))
	fmt.Println(d.AddExact(e, 1))
	fmt.Println(d.AddExact(e, 2))
	fmt.Println(d.AddExact(e, 3))
	fmt.Println(d.AddExact(e, 4))
	// Output:
	// 13.67 <nil>
	// 13.67 <nil>
	// 13.67 <nil>
	// 13.670 <nil>
	// 13.6700 <nil>
}

func ExampleDecimal_Sub() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("8")
	fmt.Println(d.Sub(e))
	// Output: -13.67 <nil>
}

func ExampleDecimal_SubAbs() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("8")
	fmt.Println(d.SubAbs(e))
	// Output: 13.67 <nil>
}

func ExampleDecimal_SubExact() {
	d := decimal.MustParse("8")
	e := decimal.MustParse("5.67")
	fmt.Println(d.SubExact(e, 0))
	fmt.Println(d.SubExact(e, 1))
	fmt.Println(d.SubExact(e, 2))
	fmt.Println(d.SubExact(e, 3))
	fmt.Println(d.SubExact(e, 4))
	// Output:
	// 2.33 <nil>
	// 2.33 <nil>
	// 2.33 <nil>
	// 2.330 <nil>
	// 2.3300 <nil>
}

func ExampleDecimal_Quo() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("2")
	fmt.Println(d.Quo(e))
	// Output: 2.835 <nil>
}

func ExampleDecimal_QuoExact() {
	d := decimal.MustParse("5.66")
	e := decimal.MustParse("2")
	fmt.Println(d.QuoExact(e, 0))
	fmt.Println(d.QuoExact(e, 1))
	fmt.Println(d.QuoExact(e, 2))
	fmt.Println(d.QuoExact(e, 3))
	fmt.Println(d.QuoExact(e, 4))
	// Output:
	// 2.83 <nil>
	// 2.83 <nil>
	// 2.83 <nil>
	// 2.830 <nil>
	// 2.8300 <nil>
}

func ExampleDecimal_QuoRem() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("2")
	fmt.Println(d.QuoRem(e))
	// Output: 2 1.67 <nil>
}

func ExampleDecimal_Inv() {
	d := decimal.MustParse("2")
	fmt.Println(d.Inv())
	// Output: 0.5 <nil>
}

func ExampleDecimal_Cmp() {
	d := decimal.MustParse("-23")
	e := decimal.MustParse("5.67")
	fmt.Println(d.Cmp(e))
	fmt.Println(d.Cmp(d))
	fmt.Println(e.Cmp(d))
	// Output:
	// -1
	// 0
	// 1
}

func ExampleDecimal_CmpAbs() {
	d := decimal.MustParse("-23")
	e := decimal.MustParse("5.67")
	fmt.Println(d.CmpAbs(e))
	fmt.Println(d.CmpAbs(d))
	fmt.Println(e.CmpAbs(d))
	// Output:
	// 1
	// 0
	// -1
}

func ExampleDecimal_CmpTotal() {
	d := decimal.MustParse("2.0")
	e := decimal.MustParse("2.00")
	fmt.Println(d.CmpTotal(e))
	fmt.Println(d.CmpTotal(d))
	fmt.Println(e.CmpTotal(d))
	// Output:
	// 1
	// 0
	// -1
}

func ExampleDecimal_Max() {
	d := decimal.MustParse("23")
	e := decimal.MustParse("-5.67")
	fmt.Println(d.Max(e))
	// Output: 23
}

func ExampleDecimal_Min() {
	d := decimal.MustParse("23")
	e := decimal.MustParse("-5.67")
	fmt.Println(d.Min(e))
	// Output: -5.67
}

func ExampleDecimal_Clamp() {
	min := decimal.MustParse("-20")
	max := decimal.MustParse("20")
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("0")
	f := decimal.MustParse("23")
	fmt.Println(d.Clamp(min, max))
	fmt.Println(e.Clamp(min, max))
	fmt.Println(f.Clamp(min, max))
	// Output:
	// -5.67 <nil>
	// 0 <nil>
	// 20 <nil>
}

func ExampleDecimal_Rescale() {
	d := decimal.MustParse("5.678")
	fmt.Println(d.Rescale(0))
	fmt.Println(d.Rescale(1))
	fmt.Println(d.Rescale(2))
	fmt.Println(d.Rescale(3))
	fmt.Println(d.Rescale(4))
	// Output:
	// 6 <nil>
	// 5.7 <nil>
	// 5.68 <nil>
	// 5.678 <nil>
	// 5.6780 <nil>
}

func ExampleDecimal_Quantize() {
	d := decimal.MustParse("5.678")
	x := decimal.MustParse("1")
	y := decimal.MustParse("0.1")
	z := decimal.MustParse("0.01")
	fmt.Println(d.Quantize(x))
	fmt.Println(d.Quantize(y))
	fmt.Println(d.Quantize(z))
	// Output:
	// 6 <nil>
	// 5.7 <nil>
	// 5.68 <nil>
}

func ExampleDecimal_Pad() {
	d := decimal.MustParse("5.67")
	fmt.Println(d.Pad(0))
	fmt.Println(d.Pad(1))
	fmt.Println(d.Pad(2))
	fmt.Println(d.Pad(3))
	fmt.Println(d.Pad(4))
	// Output:
	// 5.67 <nil>
	// 5.67 <nil>
	// 5.67 <nil>
	// 5.670 <nil>
	// 5.6700 <nil>
}

func ExampleDecimal_Round() {
	d := decimal.MustParse("5.678")
	fmt.Println(d.Round(0))
	fmt.Println(d.Round(1))
	fmt.Println(d.Round(2))
	fmt.Println(d.Round(3))
	fmt.Println(d.Round(4))
	// Output:
	// 6
	// 5.7
	// 5.68
	// 5.678
	// 5.678
}

func ExampleDecimal_Trunc() {
	d := decimal.MustParse("5.678")
	fmt.Println(d.Trunc(0))
	fmt.Println(d.Trunc(1))
	fmt.Println(d.Trunc(2))
	fmt.Println(d.Trunc(3))
	fmt.Println(d.Trunc(4))
	// Output:
	// 5
	// 5.6
	// 5.67
	// 5.678
	// 5.678
}

func ExampleDecimal_Ceil() {
	d := decimal.MustParse("5.678")
	fmt.Println(d.Ceil(0))
	fmt.Println(d.Ceil(1))
	fmt.Println(d.Ceil(2))
	fmt.Println(d.Ceil(3))
	fmt.Println(d.Ceil(4))
	// Output:
	// 6
	// 5.7
	// 5.68
	// 5.678
	// 5.678
}

func ExampleDecimal_Floor() {
	d := decimal.MustParse("5.678")
	fmt.Println(d.Floor(0))
	fmt.Println(d.Floor(1))
	fmt.Println(d.Floor(2))
	fmt.Println(d.Floor(3))
	fmt.Println(d.Floor(4))
	// Output:
	// 5
	// 5.6
	// 5.67
	// 5.678
	// 5.678
}

func ExampleDecimal_Scale() {
	d := decimal.MustParse("23")
	e := decimal.MustParse("5.67")
	fmt.Println(d.Scale())
	fmt.Println(e.Scale())
	// Output:
	// 0
	// 2
}

func ExampleDecimal_SameScale() {
	a := decimal.MustParse("23")
	b := decimal.MustParse("5.67")
	c := decimal.MustParse("1.23")
	fmt.Println(a.SameScale(b))
	fmt.Println(b.SameScale(c))
	// Output:
	// false
	// true
}

func ExampleDecimal_MinScale() {
	d := decimal.MustParse("23.0000")
	e := decimal.MustParse("-5.6700")
	fmt.Println(d.MinScale())
	fmt.Println(e.MinScale())
	// Output:
	// 0
	// 2
}

func ExampleDecimal_Trim() {
	d := decimal.MustParse("23.400")
	fmt.Println(d.Trim(0))
	fmt.Println(d.Trim(1))
	fmt.Println(d.Trim(2))
	fmt.Println(d.Trim(3))
	fmt.Println(d.Trim(4))
	// Output:
	// 23.4
	// 23.4
	// 23.40
	// 23.400
	// 23.400
}

func ExampleDecimal_Abs() {
	d := decimal.MustParse("-5.67")
	fmt.Println(d.Abs())
	// Output: 5.67
}

func ExampleDecimal_CopySign() {
	d := decimal.MustParse("23.00")
	e := decimal.MustParse("-5.67")
	fmt.Println(d.CopySign(e))
	fmt.Println(e.CopySign(d))
	// Output:
	// -23.00
	// 5.67
}

func ExampleDecimal_Neg() {
	d := decimal.MustParse("5.67")
	fmt.Println(d.Neg())
	// Output: -5.67
}

func ExampleDecimal_Sign() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("23")
	f := decimal.MustParse("0")
	fmt.Println(d.Sign())
	fmt.Println(e.Sign())
	fmt.Println(f.Sign())
	// Output:
	// -1
	// 1
	// 0
}

func ExampleDecimal_IsNeg() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("23")
	f := decimal.MustParse("0")
	fmt.Println(d.IsNeg())
	fmt.Println(e.IsNeg())
	fmt.Println(f.IsNeg())
	// Output:
	// true
	// false
	// false
}

func ExampleDecimal_IsPos() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("23")
	f := decimal.MustParse("0")
	fmt.Println(d.IsPos())
	fmt.Println(e.IsPos())
	fmt.Println(f.IsPos())
	// Output:
	// false
	// true
	// false
}

func ExampleDecimal_IsZero() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("23")
	f := decimal.MustParse("0")
	fmt.Println(d.IsZero())
	fmt.Println(e.IsZero())
	fmt.Println(f.IsZero())
	// Output:
	// false
	// false
	// true
}

func ExampleDecimal_IsInt() {
	d := decimal.MustParse("1.00")
	e := decimal.MustParse("1.01")
	fmt.Println(d.IsInt())
	fmt.Println(e.IsInt())
	// Output:
	// true
	// false
}

func ExampleDecimal_IsOne() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("2")
	fmt.Println(d.IsOne())
	fmt.Println(e.IsOne())
	// Output:
	// true
	// false
}

func ExampleDecimal_WithinOne() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("0.9")
	f := decimal.MustParse("-0.9")
	g := decimal.MustParse("-1")
	fmt.Println(d.WithinOne())
	fmt.Println(e.WithinOne())
	fmt.Println(f.WithinOne())
	fmt.Println(g.WithinOne())
	// Output:
	// false
	// true
	// true
	// false
}

func ExampleNullDecimal_Scan() {
	var n, m decimal.NullDecimal
	_ = n.Scan("5.67")
	_ = m.Scan(nil)
	fmt.Println(n)
	fmt.Println(m)
	// Output:
	// {5.67 true}
	// {0 false}
}

func ExampleNullDecimal_Value() {
	n := decimal.NullDecimal{
		Decimal: decimal.MustParse("5.67"),
		Valid:   true,
	}
	m := decimal.NullDecimal{
		Decimal: decimal.MustParse("0"),
		Valid:   false,
	}
	fmt.Println(n.Value())
	fmt.Println(m.Value())
	// Output:
	// 5.67 <nil>
	// <nil> <nil>
}
