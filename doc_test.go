package decimal_test

import (
	"fmt"
	"github.com/govalues/decimal"
)

// This example calculates an approximate value of pi using the Leibniz formula for pi.
// The Leibniz formula is an infinite series that converges to pi/4, and is
// given by the equation: 1 - 1/3 + 1/5 - 1/7 + 1/9 - 1/11 + ... = pi/4.
// This example computes the series up to the 5000th term using decimal arithmetic
// and returns the approximate value of pi.
func Example_leibnizPi() {
	pi := decimal.New(0, 0)
	dividend := decimal.New(4, 0)
	divisor := decimal.New(1, 0)
	sign := decimal.New(1, 0)
	step := decimal.New(2, 0)
	for i := 0; i < 5000; i++ {
		pi = pi.Add(dividend.Quo(divisor).Mul(sign))
		divisor = divisor.Add(step)
		sign = sign.Neg()
	}
	fmt.Println(pi)
	// Output: 3.141392653591793247
}

func ExampleNew() {
	d := decimal.New(-1230, 3)
	fmt.Println(d)
	// Output: -1.230
}

func ExampleParse() {
	d, err := decimal.Parse("-1.230")
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
	// Output: -1.230
}

func ExampleParseExact() {
	d, err := decimal.ParseExact("-1.2", 5)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
	// Output: -1.20000
}

func ExampleMustParse() {
	d := decimal.MustParse("-1.230")
	fmt.Println(d)
	// Output: -1.230
}

func ExampleDecimal_String() {
	d := decimal.MustParse("1234567890.123456789")
	fmt.Println(d.String())
	// Output: 1234567890.123456789
}

func ExampleDecimal_UnmarshalText() {
	d := &decimal.Decimal{}
	b := []byte("-15.67")
	err := d.UnmarshalText(b)
	if err != nil {
		panic(err)
	}
	fmt.Println(d)
	// Output: -15.67
}

func ExampleDecimal_MarshalText() {
	d := decimal.MustParse("-15.67")
	b, err := d.MarshalText()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
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
	// Output: 17.1
}

func ExampleDecimal_MulExact() {
	d := decimal.MustParse("5.7")
	e := decimal.MustParse("3")
	fmt.Println(d.MulExact(e, 2))
	// Output: 17.10
}

func ExampleDecimal_Fma() {
	d := decimal.MustParse("5.7")
	e := decimal.MustParse("3")
	f := decimal.MustParse("2.8")
	fmt.Println(d.Fma(e, f))
	// Output: 19.9
}

func ExampleDecimal_FmaExact() {
	d := decimal.MustParse("5.7")
	e := decimal.MustParse("3")
	f := decimal.MustParse("2.8")
	fmt.Println(d.FmaExact(e, f, 2))
	// Output: 19.90
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
	// 0.125
	// 0.25
	// 0.5
	// 1
	// 2
	// 4
	// 8
}

func ExampleDecimal_Add() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.Add(e))
	// Output: 23.6
}

func ExampleDecimal_AddExact() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.AddExact(e, 2))
	// Output: 23.60
}

func ExampleDecimal_Sub() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.Sub(e))
	// Output: 7.6
}

func ExampleDecimal_SubExact() {
	d := decimal.MustParse("15.6")
	e := decimal.MustParse("8")
	fmt.Println(d.SubExact(e, 2))
	// Output: 7.60
}

func ExampleDecimal_Quo() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("2")
	fmt.Println(d.Quo(e))
	// Output: -7.835
}

func ExampleDecimal_QuoExact() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("2")
	fmt.Println(d.QuoExact(e, 6))
	// Output: -7.835000
}

func ExampleDecimal_QuoRem() {
	d := decimal.MustParse("-15.67")
	e := decimal.MustParse("2")
	fmt.Println(d.QuoRem(e))
	// Output: -7 -1.67
}

func ExampleDecimal_Cmp() {
	d := decimal.MustParse("23")
	e := decimal.MustParse("-15.67")
	fmt.Println(d.Cmp(e))
	fmt.Println(d.Cmp(d))
	fmt.Println(e.Cmp(d))
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

func ExampleDecimal_WithScale() {
	d := decimal.MustParse("15.679")
	fmt.Println(d.WithScale(6))
	fmt.Println(d.WithScale(5))
	fmt.Println(d.WithScale(4))
	fmt.Println(d.WithScale(3))
	fmt.Println(d.WithScale(2))
	fmt.Println(d.WithScale(1))
	fmt.Println(d.WithScale(0))
	// Output:
	// 0.015679
	// 0.15679
	// 1.5679
	// 15.679
	// 156.79
	// 1567.9
	// 15679
}

func ExampleDecimal_Round() {
	d := decimal.MustParse("15.679")
	fmt.Println(d.Round(6))
	fmt.Println(d.Round(5))
	fmt.Println(d.Round(4))
	fmt.Println(d.Round(3))
	fmt.Println(d.Round(2))
	fmt.Println(d.Round(1))
	fmt.Println(d.Round(0))
	// Output:
	// 15.679000
	// 15.67900
	// 15.6790
	// 15.679
	// 15.68
	// 15.7
	// 16
}

func ExampleDecimal_Quantize() {
	d := decimal.MustParse("15.679")
	x := decimal.MustParse("0.01")
	y := decimal.MustParse("0.1")
	z := decimal.MustParse("1")
	fmt.Println(d.Quantize(x))
	fmt.Println(d.Quantize(y))
	fmt.Println(d.Quantize(z))
	// Output:
	// 15.68
	// 15.7
	// 16
}

func ExampleDecimal_Trunc() {
	d := decimal.MustParse("15.679")
	fmt.Println(d.Trunc(6))
	fmt.Println(d.Trunc(5))
	fmt.Println(d.Trunc(4))
	fmt.Println(d.Trunc(3))
	fmt.Println(d.Trunc(2))
	fmt.Println(d.Trunc(1))
	fmt.Println(d.Trunc(0))
	// Output:
	// 15.679000
	// 15.67900
	// 15.6790
	// 15.679
	// 15.67
	// 15.6
	// 15
}

func ExampleDecimal_Ceil() {
	d := decimal.MustParse("15.679")
	fmt.Println(d.Ceil(6))
	fmt.Println(d.Ceil(5))
	fmt.Println(d.Ceil(4))
	fmt.Println(d.Ceil(3))
	fmt.Println(d.Ceil(2))
	fmt.Println(d.Ceil(1))
	fmt.Println(d.Ceil(0))
	// Output:
	// 15.679000
	// 15.67900
	// 15.6790
	// 15.679
	// 15.68
	// 15.7
	// 16
}

func ExampleDecimal_Floor() {
	d := decimal.MustParse("15.679")
	fmt.Println(d.Floor(6))
	fmt.Println(d.Floor(5))
	fmt.Println(d.Floor(4))
	fmt.Println(d.Floor(3))
	fmt.Println(d.Floor(2))
	fmt.Println(d.Floor(1))
	fmt.Println(d.Floor(0))
	// Output:
	// 15.679000
	// 15.67900
	// 15.6790
	// 15.679
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

func ExampleDecimal_Reduce() {
	d := decimal.MustParse("23.0000")
	e := decimal.MustParse("-15.6700")
	fmt.Println(d.Reduce())
	fmt.Println(e.Reduce())
	// Output:
	// 23
	// -15.67
}

func ExampleDecimal_Abs() {
	d := decimal.MustParse("-15.67")
	fmt.Println(d.Abs())
	// Output: 15.67
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

func ExampleDecimal_LessThanOne() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("0.9")
	f := decimal.MustParse("-1")
	fmt.Println(d.LessThanOne())
	fmt.Println(e.LessThanOne())
	fmt.Println(f.LessThanOne())
	// Output:
	// false
	// true
	// false
}
