package decimal_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/govalues/decimal"
)

func evaluate(input string) (decimal.Decimal, error) {
	tokens, err := parseTokens(input)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("parsing tokens: %w", err)
	}
	stack, err := processTokens(tokens)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("processing tokens: %w", err)
	}
	if len(stack) != 1 {
		return decimal.Decimal{}, fmt.Errorf("post-processed stack contains %v, expected exactly one item", stack)
	}
	return stack[0], nil
}

func parseTokens(input string) ([]string, error) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens")
	}
	return tokens, nil
}

func processTokens(tokens []string) ([]decimal.Decimal, error) {
	stack := make([]decimal.Decimal, 0, len(tokens))
	var err error
	for i := len(tokens) - 1; i >= 0; i-- {
		token := tokens[i]
		switch token {
		case "+", "-", "*", "/":
			stack, err = processOperator(stack, token)
		default:
			stack, err = processOperand(stack, token)
		}
		if err != nil {
			return nil, fmt.Errorf("processing token %q: %w", token, err)
		}
	}
	return stack, nil
}

func processOperator(stack []decimal.Decimal, token string) ([]decimal.Decimal, error) {
	if len(stack) < 2 {
		return nil, fmt.Errorf("not enough operands")
	}
	right := stack[len(stack)-2]
	left := stack[len(stack)-1]
	stack = stack[:len(stack)-2]
	var result decimal.Decimal
	var err error
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
	if err != nil {
		return nil, fmt.Errorf("evaluating \"%s %s %s\": %w", left, token, right, err)
	}
	return append(stack, result), nil
}

func processOperand(stack []decimal.Decimal, token string) ([]decimal.Decimal, error) {
	d, err := decimal.Parse(token)
	if err != nil {
		return nil, err
	}
	return append(stack, d), nil
}

// This example implements a simple calculator that evaluates mathematical
// expressions written in postfix (or reverse Polish) notation.
// The calculator can handle basic arithmetic operations such as addition,
// subtraction, multiplication, and division.
func Example_postfixCalculator() {
	d, err := evaluate("* 10 + 1.23 4.56")
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
	// Output:
	// 57.90
}

func approximate(terms int) (decimal.Decimal, error) {
	pi := decimal.Zero
	sign := decimal.One
	denominator := decimal.One
	increment := decimal.Two
	multiplier := decimal.MustParse("4")

	for i := 0; i < terms; i++ {
		term, err := multiplier.Quo(denominator)
		if err != nil {
			return decimal.Decimal{}, err
		}
		term = term.CopySign(sign)
		pi, err = pi.Add(term)
		if err != nil {
			return decimal.Decimal{}, err
		}
		denominator, err = denominator.Add(increment)
		if err != nil {
			return decimal.Decimal{}, err
		}
		sign = sign.Neg()
	}
	return pi, nil
}

// This example calculates an approximate value of pi using the Leibniz formula for pi.
// The Leibniz formula is an infinite series that converges to pi/4, and is
// given by the equation: 1 - 1/3 + 1/5 - 1/7 + 1/9 - 1/11 + ... = pi/4.
// This example computes the series up to the 50,000th term using decimal arithmetic
// and returns the approximate value of pi.
func Example_piApproximation() {
	pi, err := approximate(50000)
	if err != nil {
		panic(err)
	}
	fmt.Println(pi)
	fmt.Println(decimal.Pi)
	// Output:
	// 3.141572653589795330
	// 3.141592653589793238
}

func ExampleMustNew() {
	fmt.Println(decimal.MustNew(-123, 3))
	fmt.Println(decimal.MustNew(-123, 2))
	fmt.Println(decimal.MustNew(-123, 1))
	fmt.Println(decimal.MustNew(-123, 0))
	// Output:
	// -0.123
	// -1.23
	// -12.3
	// -123
}

func ExampleNew() {
	fmt.Println(decimal.New(-123, 3))
	fmt.Println(decimal.New(-123, 2))
	fmt.Println(decimal.New(-123, 1))
	fmt.Println(decimal.New(-123, 0))
	// Output:
	// -0.123 <nil>
	// -1.23 <nil>
	// -12.3 <nil>
	// -123 <nil>
}

func ExampleNewFromInt64() {
	fmt.Println(decimal.NewFromInt64(-1, -23, 2))
	fmt.Println(decimal.NewFromInt64(-1, -23, 3))
	fmt.Println(decimal.NewFromInt64(-1, -23, 4))
	fmt.Println(decimal.NewFromInt64(-1, -23, 5))
	// Output:
	// -1.23 <nil>
	// -1.023 <nil>
	// -1.0023 <nil>
	// -1.00023 <nil>
}

func ExampleNewFromFloat64() {
	fmt.Println(decimal.NewFromFloat64(1.23e-2))
	fmt.Println(decimal.NewFromFloat64(1.23e-1))
	fmt.Println(decimal.NewFromFloat64(1.23e0))
	fmt.Println(decimal.NewFromFloat64(1.23e1))
	fmt.Println(decimal.NewFromFloat64(1.23e2))
	// Output:
	// 0.0123 <nil>
	// 0.123 <nil>
	// 1.23 <nil>
	// 12.3 <nil>
	// 123 <nil>
}

func ExampleDecimal_Zero() {
	d := decimal.MustParse("-1.23")
	e := decimal.MustParse("0.4")
	f := decimal.MustParse("15")
	fmt.Println(d.Zero())
	fmt.Println(e.Zero())
	fmt.Println(f.Zero())
	// Output:
	// 0.00
	// 0.0
	// 0
}

func ExampleDecimal_One() {
	d := decimal.MustParse("-1.23")
	e := decimal.MustParse("0.4")
	f := decimal.MustParse("15")
	fmt.Println(d.One())
	fmt.Println(e.One())
	fmt.Println(f.One())
	// Output:
	// 1.00
	// 1.0
	// 1
}

func ExampleDecimal_ULP() {
	d := decimal.MustParse("-1.23")
	e := decimal.MustParse("0.4")
	f := decimal.MustParse("15")
	fmt.Println(d.ULP())
	fmt.Println(e.ULP())
	fmt.Println(f.ULP())
	// Output:
	// 0.01
	// 0.1
	// 1
}

func ExampleParse() {
	fmt.Println(decimal.Parse("-1.23"))
	// Output: -1.23 <nil>
}

func ExampleParseExact() {
	fmt.Println(decimal.ParseExact("-1.23", 0))
	fmt.Println(decimal.ParseExact("-1.23", 1))
	fmt.Println(decimal.ParseExact("-1.23", 2))
	fmt.Println(decimal.ParseExact("-1.23", 3))
	fmt.Println(decimal.ParseExact("-1.23", 4))
	fmt.Println(decimal.ParseExact("-1.23", 5))
	// Output:
	// -1.23 <nil>
	// -1.23 <nil>
	// -1.23 <nil>
	// -1.230 <nil>
	// -1.2300 <nil>
	// -1.23000 <nil>
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
	d := decimal.MustParse("123.567")
	fmt.Println(d.Int64(5))
	fmt.Println(d.Int64(4))
	fmt.Println(d.Int64(3))
	fmt.Println(d.Int64(2))
	fmt.Println(d.Int64(1))
	fmt.Println(d.Int64(0))
	// Output:
	// 123 56700 true
	// 123 5670 true
	// 123 567 true
	// 123 57 true
	// 123 6 true
	// 124 0 true
}

type Value struct {
	Number decimal.Decimal `json:"number"`
}

func ExampleDecimal_UnmarshalText() {
	b := []byte(`{"number": "-15.67"}`)
	var v Value
	err := json.Unmarshal(b, &v)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	// Output: {-15.67}
}

func ExampleDecimal_MarshalText() {
	d := decimal.MustParse("-15.67")
	v := Value{Number: d}
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	// Output: {"number":"-15.67"}
}

func ExampleDecimal_Scan() {
	d := &decimal.Decimal{}
	s := "-15.67"
	err := d.Scan(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
	// Output: -15.67
}

func ExampleDecimal_Value() {
	d := decimal.MustParse("-15.67")
	s, err := d.Value()
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
	// Output: -15.67
}

func ExampleDecimal_Format() {
	d := decimal.MustParse("-15.679")
	fmt.Printf("%k\n", d)
	fmt.Printf("%f\n", d)
	fmt.Printf("%.2f\n", d)
	// Output:
	// -1567.9%
	// -15.679
	// -15.68
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
	fmt.Println(d.MulExact(e, 2))
	// Output: 17.10 <nil>
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
	fmt.Println(d.FMAExact(e, f, 2))
	// Output: 10.00 <nil>
}

func ExampleDecimal_Pow() {
	d := decimal.MustParse("2")
	fmt.Println(d.Pow(-3))
	fmt.Println(d.Pow(-2))
	fmt.Println(d.Pow(-1))
	fmt.Println(d.Pow(0))
	fmt.Println(d.Pow(1))
	fmt.Println(d.Pow(2))
	fmt.Println(d.Pow(3))
	// Output:
	// 0.125 <nil>
	// 0.25 <nil>
	// 0.5 <nil>
	// 1 <nil>
	// 2 <nil>
	// 4 <nil>
	// 8 <nil>
}

func ExampleDecimal_PowExact() {
	d := decimal.MustParse("2")
	fmt.Println(d.PowExact(3, 5))
	fmt.Println(d.PowExact(3, 4))
	fmt.Println(d.PowExact(3, 3))
	fmt.Println(d.PowExact(3, 2))
	fmt.Println(d.PowExact(3, 1))
	fmt.Println(d.PowExact(3, 0))
	// Output:
	// 8.00000 <nil>
	// 8.0000 <nil>
	// 8.000 <nil>
	// 8.00 <nil>
	// 8.0 <nil>
	// 8 <nil>
}

func ExampleDecimal_Add() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.Add(e))
	// Output: 23.6 <nil>
}

func ExampleDecimal_AddExact() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.AddExact(e, 2))
	// Output: 23.60 <nil>
}

func ExampleDecimal_Sub() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.Sub(e))
	// Output: 7.6 <nil>
}

func ExampleDecimal_SubAbs() {
	d := decimal.MustParse("-15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.SubAbs(e))
	// Output: 23.6 <nil>
}

func ExampleDecimal_SubExact() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.SubExact(e, 2))
	// Output: 7.60 <nil>
}

func ExampleDecimal_Quo() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("2")
	fmt.Println(d.Quo(e))
	// Output: -7.835 <nil>
}

func ExampleDecimal_QuoExact() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("2")
	fmt.Println(d.QuoExact(e, 6))
	// Output: -7.835000 <nil>
}

func ExampleDecimal_QuoRem() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("2")
	fmt.Println(d.QuoRem(e))
	// Output: -7 -1.67 <nil>
}

func ExampleDecimal_Inv() {
	d := decimal.MustParse("2")
	fmt.Println(d.Inv())
	// Output: 0.5 <nil>
}

func ExampleDecimal_Cmp() {
	d := decimal.MustParse("-23")
	e := decimal.MustParse("15.67")
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
	e := decimal.MustParse("15.67")
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
	e := decimal.MustParse("-15.67")
	fmt.Println(d.Max(e))
	// Output: 23
}

func ExampleDecimal_Min() {
	d := decimal.MustParse("23")
	e := decimal.MustParse("-15.67")
	fmt.Println(d.Min(e))
	// Output: -15.67
}

func ExampleDecimal_Clamp() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("0")
	f := decimal.MustParse("23")
	min := decimal.MustParse("-10")
	max := decimal.MustParse("10")
	fmt.Println(d.Clamp(min, max))
	fmt.Println(e.Clamp(min, max))
	fmt.Println(f.Clamp(min, max))
	// Output:
	// -10 <nil>
	// 0 <nil>
	// 10 <nil>
}

func ExampleDecimal_Rescale() {
	d := decimal.MustParse("15.679")
	fmt.Println(d.Rescale(5))
	fmt.Println(d.Rescale(4))
	fmt.Println(d.Rescale(3))
	fmt.Println(d.Rescale(2))
	fmt.Println(d.Rescale(1))
	fmt.Println(d.Rescale(0))
	// Output:
	// 15.67900 <nil>
	// 15.6790 <nil>
	// 15.679 <nil>
	// 15.68 <nil>
	// 15.7 <nil>
	// 16 <nil>
}

func ExampleDecimal_Quantize() {
	d := decimal.MustParse("15.6789")
	x := decimal.MustParse("0.01")
	y := decimal.MustParse("0.1")
	z := decimal.MustParse("1")
	fmt.Println(d.Quantize(x))
	fmt.Println(d.Quantize(y))
	fmt.Println(d.Quantize(z))
	// Output:
	// 15.68 <nil>
	// 15.7 <nil>
	// 16 <nil>
}

func ExampleDecimal_Pad() {
	d := decimal.MustParse("15.67")
	fmt.Println(d.Pad(5))
	fmt.Println(d.Pad(4))
	fmt.Println(d.Pad(3))
	fmt.Println(d.Pad(2))
	fmt.Println(d.Pad(1))
	fmt.Println(d.Pad(0))
	// Output:
	// 15.67000 <nil>
	// 15.6700 <nil>
	// 15.670 <nil>
	// 15.67 <nil>
	// 15.67 <nil>
	// 15.67 <nil>
}

func ExampleDecimal_Round() {
	d := decimal.MustParse("15.6789")
	fmt.Println(d.Round(5))
	fmt.Println(d.Round(4))
	fmt.Println(d.Round(3))
	fmt.Println(d.Round(2))
	fmt.Println(d.Round(1))
	fmt.Println(d.Round(0))
	// Output:
	// 15.6789
	// 15.6789
	// 15.679
	// 15.68
	// 15.7
	// 16
}

func ExampleDecimal_Trunc() {
	d := decimal.MustParse("15.6789")
	fmt.Println(d.Trunc(5))
	fmt.Println(d.Trunc(4))
	fmt.Println(d.Trunc(3))
	fmt.Println(d.Trunc(2))
	fmt.Println(d.Trunc(1))
	fmt.Println(d.Trunc(0))
	// Output:
	// 15.6789
	// 15.6789
	// 15.678
	// 15.67
	// 15.6
	// 15
}

func ExampleDecimal_Ceil() {
	d := decimal.MustParse("15.6789")
	fmt.Println(d.Ceil(5))
	fmt.Println(d.Ceil(4))
	fmt.Println(d.Ceil(3))
	fmt.Println(d.Ceil(2))
	fmt.Println(d.Ceil(1))
	fmt.Println(d.Ceil(0))
	// Output:
	// 15.6789
	// 15.6789
	// 15.679
	// 15.68
	// 15.7
	// 16
}

func ExampleDecimal_Floor() {
	d := decimal.MustParse("15.6789")
	fmt.Println(d.Floor(5))
	fmt.Println(d.Floor(4))
	fmt.Println(d.Floor(3))
	fmt.Println(d.Floor(2))
	fmt.Println(d.Floor(1))
	fmt.Println(d.Floor(0))
	// Output:
	// 15.6789
	// 15.6789
	// 15.678
	// 15.67
	// 15.6
	// 15
}

func ExampleDecimal_Scale() {
	d := decimal.MustParse("23.0000")
	e := decimal.MustParse("-15.670")
	fmt.Println(d.Scale())
	fmt.Println(e.Scale())
	// Output:
	// 4
	// 3
}

func ExampleDecimal_MinScale() {
	d := decimal.MustParse("23.0000")
	e := decimal.MustParse("-15.6700")
	fmt.Println(d.MinScale())
	fmt.Println(e.MinScale())
	// Output:
	// 0
	// 2
}

func ExampleDecimal_Trim() {
	d := decimal.MustParse("23.4000")
	fmt.Println(d.Trim(5))
	fmt.Println(d.Trim(4))
	fmt.Println(d.Trim(3))
	fmt.Println(d.Trim(2))
	fmt.Println(d.Trim(1))
	fmt.Println(d.Trim(0))
	// Output:
	// 23.4000
	// 23.4000
	// 23.400
	// 23.40
	// 23.4
	// 23.4
}

func ExampleDecimal_Abs() {
	d := decimal.MustParse("-15.67")
	fmt.Println(d.Abs())
	// Output: 15.67
}

func ExampleDecimal_CopySign() {
	d := decimal.MustParse("23.00")
	e := decimal.MustParse("-15.67")
	fmt.Println(d.CopySign(e))
	fmt.Println(e.CopySign(d))
	// Output:
	// -23.00
	// 15.67
}

func ExampleDecimal_Neg() {
	d := decimal.MustParse("15.67")
	fmt.Println(d.Neg())
	// Output: -15.67
}

func ExampleDecimal_Sign() {
	d := decimal.MustParse("-15.67")
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
	d := decimal.MustParse("-15.67")
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
	d := decimal.MustParse("-15.67")
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
	d := decimal.MustParse("-15.67")
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
