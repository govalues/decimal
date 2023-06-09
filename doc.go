/*
Package decimal implements immutable decimal floating-point numbers.

This packages is designed specifically for use in transactional financial systems.
The amounts involved in financial transactions typically do not exceed
99,999,999,999,999,999.99, so uint64 is used to store decimal coefficients,
which reduces heap allocations, lowers memory consumption, and improves performance.

# Features

  - Decimal values are immutable, making them safe to use in multiple goroutines.
  - Methods are panic-free and pure, returning errors in cases such as uint64
    overflow or division by zero.
  - [Decimal.String] produces simple and straightforward representation without
    scientific or engineering notation.
  - Arithmetic operations use half-even rounding, also known as "banker's rounding".
  - Special values such as NaN, Infinity, or signed zeros are not supported,
    ensuring that arithmetic operations always produce well-defined results.

# Supported Ranges

The range of a decimal value depends on the size of its coefficient.
Since the coefficient is stored as an uint64, a [Decimal] can have a maximum of
19 digits.
Additionally, the range of the [Decimal] depends on its scale, which determines
the number of decimal places.
Here are some examples of ranges supported for frequently used scales:

	| Scale | Minimum                              | Maximum                             | Example                    |
	| ----- | ------------------------------------ | ----------------------------------- | -------------------------- |
	|     0 | -9,999,999,999,999,999,999           | 9,999,999,999,999,999,999           | Japanese Yen               |
	|     2 |    -99,999,999,999,999,999.99        |    99,999,999,999,999,999.99        | US Dollar                  |
	|     3 |     -9,999,999,999,999,999.999       |     9,999,999,999,999,999.999       | Omani Rial                 |
	|     8 |            -99,999,999,999.99999999  |            99,999,999,999.99999999  | Bitcoin                    |
	|     9 |             -9,999,999,999.999999999 |             9,999,999,999.999999999 | US Dollar (high-precision) |
	|       |                                      |                                     | or Etherium                |

# Operations

Arithmetic operations in this package are based on [General Decimal Arithmetic]
and usually involve two steps:

 1. The operation is first performed using only uint64 variables.
    If no overflow occurs, the result is returned.
    If an overflow occurs, the operation proceeds to step 2.

 2. The operation is performed again using [big.Int] variables.
    The result is rounded to fit into 19 digits.
    If no significant digits are lost during rounding, the result is returned.
    If significant digits are lost, an error is returned.

The purpose of the first step is to optimize the performance of arithmetic
operations and reduce memory consumption.
Since the coefficient is stored as an uint64, arithmetic operations using only
uint64 variables can be performed quickly and efficiently.
It is expected that most of the arithmetic operations will be successfully
completed during the first step.

The following rules are used to determine the significance of digits:

  - [Decimal.Add], [Decimal.Sub], [Decimal.Mul], [Decimal.FMA], [Decimal.Pow],
    [Decimal.Quo], [Decimal.QuoRem]:
    All digits in the integer part are significant, while the digits in the
    fractional part are insignificant.
  - [Decimal.AddExact], [Decimal.SubExact], [Decimal.MulExact], [Decimal.FMAExact],
    [Decimal.PowExact], [Decimal.QuoExact]:
    All digits in the integer part are significant. The significance of digits
    in the fractional part is determined by the scale argument, which is typically
    equal to the scale of the currency.

# Rounding

To fit the results of arithmetic operations into 19 digits, the package
uses half-to-even rounding, which ensures that rounding errors are
evenly distributed between rounding up and rounding down.

In addition to implicit half-to-even rounding, the Decimal package provides
several methods for explicit rounding:

  - [Decimal.Ceil]: rounds towards positive infinity.
  - [Decimal.Floor]: rounds towards negative infinity.
  - [Decimal.Trunc]: rounds towards zero.
  - [Decimal.Round], [Decimal.Quantize], [Decimal.Rescale]: use half-to-even rounding.

# Errors

Arithmetic operations return errors in the following cases:

 1. Decimal overflow.
    This error occurs when significant digits are lost during rounding to fit
    19 digits.
    This typically happens when dealing with large numbers or when you requested
    large number of digits after the decimal point to be considered signigicant.
    Refer to the supported ranges section, if your application needs to handle
    numbers that are close to the minimum or maximum values, this package may
    not be suitable.
    Consider using packages that store coefficients using [big.Int] type,
    such as [ShopSpring Decimal] or [CockroachDB Decimal].

 2. Division by zero.
    Unlike the standard library, this package does not panic when dividing by zero.
    Instead, it returns an error.

[General Decimal Arithmetic]: https://speleotrove.com/decimal/daops.html
[ShopSpring Decimal]: https://pkg.go.dev/github.com/shopspring/decimal
[CockroachDB Decimal]: https://pkg.go.dev/github.com/cockroachdb/apd
[big.Int]: https://pkg.go.dev/math/big#Int
*/
package decimal
