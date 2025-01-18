package decimal_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"slices"
	"strings"

	"github.com/govalues/decimal"
)

// This example implements a simple calculator that evaluates mathematical
// expressions written in [postfix notation].
// The calculator can handle basic arithmetic operations such as addition,
// subtraction, multiplication, and division.
//
// [postfix notation]: https://en.wikipedia.org/wiki/Reverse_Polish_notation
func Example_postfixCalculator() {
	fmt.Println(evaluate("1.23 4.56 + 10 *"))
	// Output:
	// 57.90 <nil>
}

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

// This example calculates an approximate value of π using the [Leibniz formula].
// The Leibniz formula is an infinite series that converges to π/4, and is
// given by the equation: 1 - 1/3 + 1/5 - 1/7 + 1/9 - 1/11 + ... = π/4.
// This example computes the series up to the 500,000th term using decimal arithmetic
// and returns the approximate value of π.
//
// [Leibniz formula]: https://en.wikipedia.org/wiki/Leibniz_formula_for_%CF%80
func Example_piApproximation() {
	fmt.Println(approximate(500_000))
	fmt.Println(decimal.Pi)
	// Output:
	// 3.141590653589793192 <nil>
	// 3.141592653589793238
}

func approximate(terms int) (decimal.Decimal, error) {
	pi := decimal.Zero
	denominator := decimal.One
	increment := decimal.Two
	multiplier := decimal.MustParse("4")

	var err error
	for range terms {
		pi, err = pi.AddQuo(multiplier, denominator)
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

// This example demonstrates the advantage of decimals for financial calculations.
// It computes the sum 0.1 + 0.2 using both decimal and float64 arithmetic.
// In decimal arithmetic, the result is exactly 0.3, as expected.
// In float64 arithmetic, the result is 0.30000000000000004 due to floating-point inaccuracy.
func Example_floatInaccuracy() {
	a := decimal.MustParse("0.1")
	b := decimal.MustParse("0.2")
	fmt.Println(a.Add(b))

	x := 0.1
	y := 0.2
	fmt.Println(x + y)
	// Output:
	// 0.3 <nil>
	// 0.30000000000000004
}

func ExampleSum() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("-8")
	f := decimal.MustParse("23")
	fmt.Println(decimal.Sum(d, e, f))
	// Output: 20.67 <nil>
}

func ExampleMean() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("-8")
	f := decimal.MustParse("23")
	fmt.Println(decimal.Mean(d, e, f))
	// Output: 6.89 <nil>
}

func ExampleProd() {
	d := decimal.MustParse("5.67")
	e := decimal.MustParse("-8")
	f := decimal.MustParse("23")
	fmt.Println(decimal.Prod(d, e, f))
	// Output: -1043.28 <nil>
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

func ExampleDecimal_UnmarshalBinary_gob() {
	data := []byte{
		0x12, 0x7f, 0x06, 0x01,
		0x01, 0x07, 0x44, 0x65,
		0x63, 0x69, 0x6d, 0x61,
		0x6c, 0x01, 0xff, 0x80,
		0x00, 0x00, 0x00, 0x08,
		0xff, 0x80, 0x00, 0x04,
		0x35, 0x2e, 0x36, 0x37,
	}
	fmt.Println(unmarshalGOB(data))
	// Output:
	// 5.67 <nil>
}

func unmarshalGOB(data []byte) (decimal.Decimal, error) {
	var d decimal.Decimal
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&d)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return d, nil
}

func ExampleDecimal_AppendBinary() {
	d := decimal.MustParse("5.67")
	var data []byte
	data = append(data, 0x04)
	data, err := d.AppendBinary(data)
	data = append(data, 0x00)
	fmt.Printf("% x %v\n", data, err)
	// Output:
	// 04 35 2e 36 37 00 <nil>
}

func ExampleDecimal_MarshalBinary_gob() {
	data, err := marshalGOB("5.67")
	fmt.Printf("[% x] %v\n", data, err)
	// Output:
	// [12 7f 06 01 01 07 44 65 63 69 6d 61 6c 01 ff 80 00 00 00 08 ff 80 00 04 35 2e 36 37] <nil>
}

func marshalGOB(s string) ([]byte, error) {
	d, err := decimal.Parse(s)
	if err != nil {
		return nil, err
	}
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err = enc.Encode(d)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
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

func ExampleDecimal_UnmarshalBSONValue_bson() {
	data := []byte{
		0x37, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x3c, 0x30,
	}

	var d decimal.Decimal
	err := d.UnmarshalBSONValue(19, data)
	fmt.Println(d, err)
	// Output:
	// 5.67 <nil>
}

func ExampleDecimal_MarshalBSONValue_bson() {
	d := decimal.MustParse("5.67")
	t, data, err := d.MarshalBSONValue()
	fmt.Printf("%v [% x] %v\n", t, data, err)
	// Output:
	// 19 [37 02 00 00 00 00 00 00 00 00 00 00 00 00 3c 30] <nil>
}

type Account struct {
	Balance decimal.Decimal `json:"balance"`
}

func ExampleDecimal_UnmarshalJSON_json() {
	fmt.Println(unmarshalJSON(`{"balance":"5.67"}`))
	fmt.Println(unmarshalJSON(`{"balance":"-5.67"}`))
	fmt.Println(unmarshalJSON(`{"balance":5.67e-5}`))
	fmt.Println(unmarshalJSON(`{"balance":5.67e5}`))
	// Output:
	// {5.67} <nil>
	// {-5.67} <nil>
	// {0.0000567} <nil>
	// {567000} <nil>
}

func unmarshalJSON(s string) (Account, error) {
	var a Account
	err := json.Unmarshal([]byte(s), &a)
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

func ExampleDecimal_MarshalJSON_json() {
	fmt.Println(marshalJSON("5.67"))
	fmt.Println(marshalJSON("-5.67"))
	// Output:
	// {"balance":"5.67"} <nil>
	// {"balance":"-5.67"} <nil>
}

func marshalJSON(s string) (string, error) {
	d, err := decimal.Parse(s)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(Account{Balance: d})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type Transaction struct {
	Amount decimal.Decimal `xml:"Amount"`
}

func ExampleDecimal_UnmarshalText_xml() {
	fmt.Println(unmarshalXML(`<Transaction><Amount>5.67</Amount></Transaction>`))
	fmt.Println(unmarshalXML(`<Transaction><Amount>-5.67</Amount></Transaction>`))
	fmt.Println(unmarshalXML(`<Transaction><Amount>5.67e-5</Amount></Transaction>`))
	fmt.Println(unmarshalXML(`<Transaction><Amount>5.67e5</Amount></Transaction>`))
	// Output:
	// {5.67} <nil>
	// {-5.67} <nil>
	// {0.0000567} <nil>
	// {567000} <nil>
}

func unmarshalXML(s string) (Transaction, error) {
	var t Transaction
	err := xml.Unmarshal([]byte(s), &t)
	return t, err
}

func ExampleDecimal_AppendText() {
	var text []byte
	d := decimal.MustParse("5.67")
	text = append(text, "<Decimal>"...)
	text, err := d.AppendText(text)
	text = append(text, "</Decimal>"...)
	fmt.Printf("%s %v\n", text, err)
	// Output:
	// <Decimal>5.67</Decimal> <nil>
}

func ExampleDecimal_MarshalText_xml() {
	fmt.Println(marshalXML("5.67"))
	fmt.Println(marshalXML("-5.67"))
	fmt.Println(marshalXML("5.67e-5"))
	fmt.Println(marshalXML("5.67e5"))
	// Output:
	// <Transaction><Amount>5.67</Amount></Transaction> <nil>
	// <Transaction><Amount>-5.67</Amount></Transaction> <nil>
	// <Transaction><Amount>0.0000567</Amount></Transaction> <nil>
	// <Transaction><Amount>567000</Amount></Transaction> <nil>
}

func marshalXML(s string) (string, error) {
	d, err := decimal.Parse(s)
	if err != nil {
		return "", err
	}
	data, err := xml.Marshal(Transaction{Amount: d})
	if err != nil {
		return "", err
	}
	return string(data), nil
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

func ExampleDecimal_SubMul() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.SubMul(e, f))
	// Output: -10 <nil>
}

func ExampleDecimal_SubMulExact() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.SubMulExact(e, f, 0))
	fmt.Println(d.SubMulExact(e, f, 1))
	fmt.Println(d.SubMulExact(e, f, 2))
	fmt.Println(d.SubMulExact(e, f, 3))
	fmt.Println(d.SubMulExact(e, f, 4))
	// Output:
	// -10 <nil>
	// -10.0 <nil>
	// -10.00 <nil>
	// -10.000 <nil>
	// -10.0000 <nil>
}

func ExampleDecimal_AddMul() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.AddMul(e, f))
	// Output: 14 <nil>
}

func ExampleDecimal_AddMulExact() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.AddMulExact(e, f, 0))
	fmt.Println(d.AddMulExact(e, f, 1))
	fmt.Println(d.AddMulExact(e, f, 2))
	fmt.Println(d.AddMulExact(e, f, 3))
	fmt.Println(d.AddMulExact(e, f, 4))
	// Output:
	// 14 <nil>
	// 14.0 <nil>
	// 14.00 <nil>
	// 14.000 <nil>
	// 14.0000 <nil>
}

func ExampleDecimal_SubQuo() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.SubQuo(e, f))
	// Output: 1.25 <nil>
}

func ExampleDecimal_SubQuoExact() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.SubQuoExact(e, f, 0))
	fmt.Println(d.SubQuoExact(e, f, 1))
	fmt.Println(d.SubQuoExact(e, f, 2))
	fmt.Println(d.SubQuoExact(e, f, 3))
	fmt.Println(d.SubQuoExact(e, f, 4))
	// Output:
	// 1.25 <nil>
	// 1.25 <nil>
	// 1.25 <nil>
	// 1.250 <nil>
	// 1.2500 <nil>
}

func ExampleDecimal_AddQuo() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.AddQuo(e, f))
	// Output: 2.75 <nil>
}

func ExampleDecimal_AddQuoExact() {
	d := decimal.MustParse("2")
	e := decimal.MustParse("3")
	f := decimal.MustParse("4")
	fmt.Println(d.AddQuoExact(e, f, 0))
	fmt.Println(d.AddQuoExact(e, f, 1))
	fmt.Println(d.AddQuoExact(e, f, 2))
	fmt.Println(d.AddQuoExact(e, f, 3))
	fmt.Println(d.AddQuoExact(e, f, 4))
	// Output:
	// 2.75 <nil>
	// 2.75 <nil>
	// 2.75 <nil>
	// 2.750 <nil>
	// 2.7500 <nil>
}

func ExampleDecimal_Pow() {
	d := decimal.MustParse("4")
	e := decimal.MustParse("0.5")
	f := decimal.MustParse("-0.5")
	fmt.Println(d.Pow(e))
	fmt.Println(d.Pow(f))
	// Output:
	// 2.000000000000000000 <nil>
	// 0.5000000000000000000 <nil>
}

func ExampleDecimal_PowInt() {
	d := decimal.MustParse("2")
	fmt.Println(d.PowInt(-2))
	fmt.Println(d.PowInt(-1))
	fmt.Println(d.PowInt(0))
	fmt.Println(d.PowInt(1))
	fmt.Println(d.PowInt(2))
	// Output:
	// 0.25 <nil>
	// 0.5 <nil>
	// 1 <nil>
	// 2 <nil>
	// 4 <nil>
}

func ExampleDecimal_Sqrt() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("2")
	f := decimal.MustParse("3")
	g := decimal.MustParse("4")
	fmt.Println(d.Sqrt())
	fmt.Println(e.Sqrt())
	fmt.Println(f.Sqrt())
	fmt.Println(g.Sqrt())
	// Output:
	// 1 <nil>
	// 1.414213562373095049 <nil>
	// 1.732050807568877294 <nil>
	// 2 <nil>
}

func ExampleDecimal_Exp() {
	d := decimal.MustParse("-2.302585092994045684")
	e := decimal.MustParse("0")
	f := decimal.MustParse("2.302585092994045684")
	fmt.Println(d.Exp())
	fmt.Println(e.Exp())
	fmt.Println(f.Exp())
	// Output:
	// 0.1000000000000000000 <nil>
	// 1 <nil>
	// 10.00000000000000000 <nil>
}

func ExampleDecimal_Expm1() {
	d := decimal.MustParse("-2.302585092994045684")
	e := decimal.MustParse("0")
	f := decimal.MustParse("2.302585092994045684")
	fmt.Println(d.Expm1())
	fmt.Println(e.Expm1())
	fmt.Println(f.Expm1())
	// Output:
	// -0.9000000000000000000 <nil>
	// 0 <nil>
	// 9.000000000000000000 <nil>
}

func ExampleDecimal_Log() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("2")
	f := decimal.MustParse("2.718281828459045236")
	g := decimal.MustParse("10")
	fmt.Println(d.Log())
	fmt.Println(e.Log())
	fmt.Println(f.Log())
	fmt.Println(g.Log())
	// Output:
	// 0 <nil>
	// 0.6931471805599453094 <nil>
	// 1.000000000000000000 <nil>
	// 2.302585092994045684 <nil>
}

func ExampleDecimal_Log1p() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("2")
	f := decimal.MustParse("2.718281828459045236")
	g := decimal.MustParse("10")
	fmt.Println(d.Log1p())
	fmt.Println(e.Log1p())
	fmt.Println(f.Log1p())
	fmt.Println(g.Log1p())
	// Output:
	// 0.6931471805599453094 <nil>
	// 1.098612288668109691 <nil>
	// 1.313261687518222834 <nil>
	// 2.397895272798370544 <nil>
}

func ExampleDecimal_Log2() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("2")
	f := decimal.MustParse("2.718281828459045236")
	g := decimal.MustParse("10")
	fmt.Println(d.Log2())
	fmt.Println(e.Log2())
	fmt.Println(f.Log2())
	fmt.Println(g.Log2())
	// Output:
	// 0 <nil>
	// 1 <nil>
	// 1.442695040888963408 <nil>
	// 3.321928094887362348 <nil>
}

func ExampleDecimal_Log10() {
	d := decimal.MustParse("1")
	e := decimal.MustParse("2")
	f := decimal.MustParse("2.718281828459045236")
	g := decimal.MustParse("10")
	fmt.Println(d.Log10())
	fmt.Println(e.Log10())
	fmt.Println(f.Log10())
	fmt.Println(g.Log10())
	// Output:
	// 0 <nil>
	// 0.3010299956639811952 <nil>
	// 0.4342944819032518278 <nil>
	// 1 <nil>
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
	fmt.Println(e.Sub(d))
	// Output:
	// -13.67 <nil>
	// 13.67 <nil>
}

func ExampleDecimal_SubAbs() {
	d := decimal.MustParse("-5.67")
	e := decimal.MustParse("8")
	fmt.Println(d.SubAbs(e))
	fmt.Println(e.SubAbs(d))
	// Output:
	// 13.67 <nil>
	// 13.67 <nil>
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

func ExampleDecimal_Less() {
	d := decimal.MustParse("-23")
	e := decimal.MustParse("5.67")
	fmt.Println(d.Less(e))
	fmt.Println(e.Less(d))
	// Output:
	// true
	// false
}

func ExampleDecimal_Equal() {
	d := decimal.MustParse("-23")
	e := decimal.MustParse("5.67")
	fmt.Println(d.Equal(e))
	fmt.Println(d.Equal(d))
	// Output:
	// false
	// true
}

func ExampleDecimal_Equal_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("0"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.EqualFunc(s, s, decimal.Decimal.Equal))
	fmt.Println(slices.CompactFunc(s, decimal.Decimal.Equal))
	// Output:
	// true
	// [-5.67 0]
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

func ExampleDecimal_Cmp_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.CompareFunc(s, s, decimal.Decimal.Cmp))
	fmt.Println(slices.MaxFunc(s, decimal.Decimal.Cmp))
	fmt.Println(slices.MinFunc(s, decimal.Decimal.Cmp))
	fmt.Println(s, slices.IsSortedFunc(s, decimal.Decimal.Cmp))
	slices.SortFunc(s, decimal.Decimal.Cmp)
	fmt.Println(s, slices.IsSortedFunc(s, decimal.Decimal.Cmp))
	fmt.Println(slices.BinarySearchFunc(s, decimal.MustParse("1"), decimal.Decimal.Cmp))
	// Output:
	// 0
	// 23
	// -5.67
	// [-5.67 23 0] false
	// [-5.67 0 23] true
	// 2 false
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

func ExampleDecimal_CmpAbs_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.CompareFunc(s, s, decimal.Decimal.CmpAbs))
	fmt.Println(slices.MaxFunc(s, decimal.Decimal.CmpAbs))
	fmt.Println(slices.MinFunc(s, decimal.Decimal.CmpAbs))
	fmt.Println(s, slices.IsSortedFunc(s, decimal.Decimal.CmpAbs))
	slices.SortFunc(s, decimal.Decimal.CmpAbs)
	fmt.Println(s, slices.IsSortedFunc(s, decimal.Decimal.CmpAbs))
	fmt.Println(slices.BinarySearchFunc(s, decimal.MustParse("1"), decimal.Decimal.CmpAbs))
	// Output:
	// 0
	// 23
	// 0
	// [-5.67 23 0] false
	// [0 -5.67 23] true
	// 1 false
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

func ExampleDecimal_CmpTotal_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.CompareFunc(s, s, decimal.Decimal.CmpTotal))
	fmt.Println(slices.MaxFunc(s, decimal.Decimal.CmpTotal))
	fmt.Println(slices.MinFunc(s, decimal.Decimal.CmpTotal))
	fmt.Println(s, slices.IsSortedFunc(s, decimal.Decimal.CmpTotal))
	slices.SortFunc(s, decimal.Decimal.CmpTotal)
	fmt.Println(s, slices.IsSortedFunc(s, decimal.Decimal.CmpTotal))
	fmt.Println(slices.BinarySearchFunc(s, decimal.MustParse("10"), decimal.Decimal.CmpTotal))
	// Output:
	// 0
	// 23
	// -5.67
	// [-5.67 23 0] false
	// [-5.67 0 23] true
	// 2 false
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

//nolint:revive
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
	// 6
	// 5.7
	// 5.68
	// 5.678
	// 5.6780
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
	// 6
	// 5.7
	// 5.68
}

func ExampleDecimal_Pad() {
	d := decimal.MustParse("5.67")
	fmt.Println(d.Pad(0))
	fmt.Println(d.Pad(1))
	fmt.Println(d.Pad(2))
	fmt.Println(d.Pad(3))
	fmt.Println(d.Pad(4))
	// Output:
	// 5.67
	// 5.67
	// 5.67
	// 5.670
	// 5.6700
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

func ExampleDecimal_IsNeg_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.ContainsFunc(s, decimal.Decimal.IsNeg))
	fmt.Println(slices.IndexFunc(s, decimal.Decimal.IsNeg))
	fmt.Println(slices.DeleteFunc(s, decimal.Decimal.IsNeg))
	// Output:
	// true
	// 0
	// [23 0]
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

func ExampleDecimal_IsPos_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.ContainsFunc(s, decimal.Decimal.IsPos))
	fmt.Println(slices.IndexFunc(s, decimal.Decimal.IsPos))
	fmt.Println(slices.DeleteFunc(s, decimal.Decimal.IsPos))
	// Output:
	// true
	// 1
	// [-5.67 0]
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

func ExampleDecimal_IsZero_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.ContainsFunc(s, decimal.Decimal.IsZero))
	fmt.Println(slices.IndexFunc(s, decimal.Decimal.IsZero))
	fmt.Println(slices.DeleteFunc(s, decimal.Decimal.IsZero))
	// Output:
	// true
	// 2
	// [-5.67 23]
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

func ExampleDecimal_IsInt_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0"),
	}
	fmt.Println(slices.ContainsFunc(s, decimal.Decimal.IsInt))
	fmt.Println(slices.IndexFunc(s, decimal.Decimal.IsInt))
	fmt.Println(slices.DeleteFunc(s, decimal.Decimal.IsInt))
	// Output:
	// true
	// 1
	// [-5.67]
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

func ExampleDecimal_IsOne_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("1"),
	}
	fmt.Println(slices.ContainsFunc(s, decimal.Decimal.IsOne))
	fmt.Println(slices.IndexFunc(s, decimal.Decimal.IsOne))
	fmt.Println(slices.DeleteFunc(s, decimal.Decimal.IsOne))
	// Output:
	// true
	// 2
	// [-5.67 23]
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

func ExampleDecimal_WithinOne_slices() {
	s := []decimal.Decimal{
		decimal.MustParse("-5.67"),
		decimal.MustParse("23"),
		decimal.MustParse("0.1"),
	}
	fmt.Println(slices.ContainsFunc(s, decimal.Decimal.WithinOne))
	fmt.Println(slices.IndexFunc(s, decimal.Decimal.WithinOne))
	fmt.Println(slices.DeleteFunc(s, decimal.Decimal.WithinOne))
	// Output:
	// true
	// 2
	// [-5.67 23]
}

func ExampleNullDecimal_Scan() {
	var n decimal.NullDecimal
	_ = n.Scan(nil)
	fmt.Println(n)

	var m decimal.NullDecimal
	_ = m.Scan("5.67")
	fmt.Println(m)
	// Output:
	// {0 false}
	// {5.67 true}
}

func ExampleNullDecimal_Value() {
	n := decimal.NullDecimal{
		Valid: false,
	}
	fmt.Println(n.Value())

	m := decimal.NullDecimal{
		Decimal: decimal.MustParse("5.67"),
		Valid:   true,
	}
	fmt.Println(m.Value())
	// Output:
	// <nil> <nil>
	// 5.67 <nil>
}

func ExampleNullDecimal_UnmarshalJSON_json() {
	var n decimal.NullDecimal
	_ = json.Unmarshal([]byte(`null`), &n)
	fmt.Println(n)

	var m decimal.NullDecimal
	_ = json.Unmarshal([]byte(`"5.67"`), &m)
	fmt.Println(m)
	// Output:
	// {0 false}
	// {5.67 true}
}

func ExampleNullDecimal_MarshalJSON_json() {
	n := decimal.NullDecimal{
		Valid: false,
	}
	data, _ := json.Marshal(n)
	fmt.Println(string(data))

	m := decimal.NullDecimal{
		Decimal: decimal.MustParse("5.67"),
		Valid:   true,
	}
	data, _ = json.Marshal(m)
	fmt.Println(string(data))
	// Output:
	// null
	// "5.67"
}

func ExampleNullDecimal_UnmarshalBSONValue_bson() {
	var n decimal.NullDecimal
	_ = n.UnmarshalBSONValue(10, nil)
	fmt.Println(n)

	data := []byte{
		0x37, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x3c, 0x30,
	}
	var m decimal.NullDecimal
	_ = m.UnmarshalBSONValue(19, data)
	fmt.Println(m)
	// Output:
	// {0 false}
	// {5.67 true}
}

func ExampleNullDecimal_MarshalBSONValue_bson() {
	n := decimal.NullDecimal{
		Valid: false,
	}
	t, data, _ := n.MarshalBSONValue()
	fmt.Printf("%v [% x]\n", t, data)

	m := decimal.NullDecimal{
		Decimal: decimal.MustParse("5.67"),
		Valid:   true,
	}
	t, data, _ = m.MarshalBSONValue()
	fmt.Printf("%v [% x]\n", t, data)
	// Output:
	// 10 []
	// 19 [37 02 00 00 00 00 00 00 00 00 00 00 00 00 3c 30]
}
