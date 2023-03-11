/*
Package decimal implements decimal floating-point numbers.

This packages is designed specifically for use in transactional financial systems. 
The amounts involved in financial transactions typically do not exceed 
99,999,999,999,999,999.99, so uint64 is used to store decimal coefficients,
which reduces heap allocations, lowers memory consumption, and improves performance.

# Features

 - [Decimal] values are immutable, making them safe to use in multiple goroutines.
 - Supports simple, straightforward string representation without scientific or 
   engineering notation.
 - Uses half-even rounding for arithmetic operations, with the ability to panic 
   if significant digits are lost.
 - Does not support special values such as NaN, Infinity, or signed zeros, 
   ensuring that arithmetic operations always produce well-defined results.

# Representation

A [Decimal] value is represented as a struct with three parameters:

 1. Sign: a boolean indicating whether the decimal is negative.
 2. Coefficient: an uint64 value.
 3. Scale: an integer indicating the position of the floating decimal point.
 
The scale field determines the position of the decimal point in the coefficient. 
For example, a decimal value with a scale of 2 represents a value that has two 
digits after the decimal point. 
The coefficient field is the integer value of the decimal without the decimal point. 
For example, a decimal with a coefficient of 12345 and a scale of 2 represents 
the value 123.45.
Such approach allows for multiple representations of the same numerical value. 
For example, 1, 1.0, and 1.00 all represent the same value, but they 
have different scales and coefficients.

One important aspect of the [Decimal] is that it does not support 
special values such as NaN, Infinity, or signed zeros. 
This makes the representation simpler and more efficient, and it ensures that 
arithmetic operations always produce well-defined results.

# Supported Ranges

The range of a decimal value depends on its scale and the size of its coefficient. 
Since the coefficient is stored as an uint64, a [Decimal] can have a maximum of 19 digits.
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
and generally involve two steps:

   1. The operation is first performed using only uint64 variables. 
      If no overflow occurs, the result is returned. 
      If an overflow occurs, the operation proceeds to step 2.

   2. The operation is performed again using [big.Int] variables. 
      The result is rounded to fit into 19 digits. 
      If no significant digits are lost during rounding, the result is returned. 
      If significant digits are lost, a panic is raised.

The purpose of the first step is to optimize the performance of arithmetic
operations and reduce memory consumption. 
Since the coefficient is stored as an uint64, arithmetic operations using only 
uint64 variables can be performed quickly and efficiently.
It is expected that most of the arithmetic operations will be successfully
completed during the first step.

The following rules are used to determine the significance of digits:

   - [Decimal.Add], [Decimal.Sub], [Decimal.Mul], [Decimal.Pow], [Decimal.Quo], [Decimal.QuoRem]:
      All digits in the integer part are significant, while the digits in the
      fractional part are insignificant.
   - [Decimal.AddExact], [Decimal.SubExact], [Decimal.MulExact], [Decimal.QuoExact]:
      All digits in the integer part are significant. The significance of digits
      in the fractional part is determined by the scale argument, which is typically
      equal to the scale of the currency.

# Rounding

To fit the results of arithmetic operations into 19 digits, the package 
uses half-to-even rounding, which ensures that rounding errors are 
evenly distributed between rounding up and rounding down.

In addition to implicit half-to-even rounding, the Decimal package provides 
several methods for explicit rounding:

   - [Decimal.Ceil]: rounds towards positive infinity. For example, the ceil of 1.5 is 2.
   - [Decimal.Floor]: rounds towards negative infinity. For example, the floor of 1.5 is 1.
   - [Decimal.Trunc]: rounds towards zero. For example, the trunc of 1.5 is 1.
   - [Decimal.Round]: uses half-to-even rounding.

# Errors

Arithmetic operations panic in the following cases:

 1. Out-of-range scale.
    This error is expected to be very rare as the scale usually comes from
    a global constant or an established standard, such as ISO 4217.
    If this error occurs, it suggests a significant bug.

 2. Coefficient overflow.
    This error occurs when significant digits are lost during rounding to fit 19 digits.
    This typically happens when dealing with extremely large amounts or high scales.
    Refer to the supported ranges section.
    If your application needs to handle numbers that are close to the minimum or
    maximum values, this package may not be suitable.
    Consider using packages that store coefficients using [big.Int] type,
    such as [ShopSpring Decimal] or [CockroachDB Decimal].

 3. Division by zero.
    This package follows the convention of the standard library and will panic
    in case of division by zero.

[General Decimal Arithmetic]: https://speleotrove.com/decimal/daops.html
[ShopSpring Decimal]: https://pkg.go.dev/github.com/shopspring/decimal
[CockroachDB Decimal]: https://pkg.go.dev/github.com/cockroachdb/apd
[big.Int]: https://pkg.go.dev/math/big#Int
*/
package decimal
