/*
Package decimal implements immutable decimal floating-point numbers.
It is designed specifically for use in transactional financial systems.
This package generally follows principles set by [ANSI X3.274-1996 (section 7.4)].

# Representation

A decimal value is a struct with three fields:

  - Sign: a boolean indicating whether the decimal is negative.
  - Coefficient: an unsigned integer that keeps numeric value of the decimal without
    the decimal point.
  - Scale: a non-negative integer indicating the position of the decimal point
    within the coefficient.
    For example, a decimal with a coefficient of 12345 and a scale of 2 represents
    the value 123.45.
    Conceptually the scale can be understood as the opposite of the exponent in
    scientific notation.
    For example, a scale of 2 (as in 0.01) corresponds to an exponent of -2.
    The range of allowed values for the scale is from 0 to 19.

The numerical value of a decimal is calculated as:

  - -Coefficient / 10^Scale, if Sign is true
  - Coefficient / 10^Scale, if Sign is false

In such approach, the same numeric value can have multiple representations.
For example, 1, 1.0, and 1.00 all have the same value, but they
have different scales and coefficients.

# Constraints

The range of a decimal value is determined by its scale.
Here are some examples of ranges for frequently used scales:

	| Currency     | Scale | Minimum                              | Maximum                             |
	| ------------ | ----- | ------------------------------------ | ----------------------------------- |
	| Japanese Yen | 0     | -9,999,999,999,999,999,999           | 9,999,999,999,999,999,999           |
	| US Dollar    | 2     |    -99,999,999,999,999,999.99        |    99,999,999,999,999,999.99        |
	| Omani Rial   | 3     |     -9,999,999,999,999,999.999       |     9,999,999,999,999,999.999       |
	| Bitcoin      | 8     |            -99,999,999,999.99999999  |            99,999,999,999.99999999  |
	| Etherium     | 9     |             -9,999,999,999.999999999 |             9,999,999,999.999999999 |

Subnormal numbers are not supported to ensure peak performance.
Consequently, decimals between -0.00000000000000000005 and 0.00000000000000000005
(inclusive) are rounded to 0.

Special values such as NaN, Infinity, or negative zeros are not supported.
This ensures that arithmetic operations always produce either valid decimal values
or errors.

# Conversions

The package provides methods for converting decimals:

  - [Parse], [Decimal.String]:
    from and to string.
  - [NewFromFloat64], [Decimal.Float64]:
    from and to float.
  - [New], [NewFromInt64], [Decimal.Int64]:
    from and to int.

See the documentation for each method for more details.

# Operations

Each operation is carried out in two steps:

 1. The operation is initially performed using uint64 arithmetic.
    If no overflow occurs, the exact result is immediately returned.
    If an overflow occurs, the operation proceeds to step 2.

 2. The operation is repeated with increased precision using [big.Int] arithmetic.
    The result is then rounded to 19 digits.
    If no significant digits are lost during rounding, the inexact result is returned.
    If any significant digit is lost, an overflow error is returned.

The step 1 was introduced to improve performance by avoiding heap allocation
of [big.Int] and the complexities associated with [big.Int] arithmetic.
It is expected that in transactional financial systems, majority of the arithmetic
operations will be successfully completed during the step 1.

The following rules are used to determine the significance of digits during step 2:

  - [Decimal.Add], [Decimal.Sub], [Decimal.Mul], [Decimal.FMA], [Decimal.Pow],
    [Decimal.Quo], [Decimal.QuoRem]:
    all digits in the integer part are significant, while the digits in the
    fractional part are insignificant.
  - [Decimal.AddExact], [Decimal.SubExact], [Decimal.MulExact], [Decimal.FMAExact],
    [Decimal.PowExact], [Decimal.QuoExact]:
    all digits in the integer part are significant. The significance of digits
    in the fractional part is determined by the scale argument, which is typically
    equal to the scale of the currency.

# Context

Unlike many other decimal libraries, this package does not provide
an explicit context.
Instead, the context is implicit and can be approximately equated to
the following settings:

	| Attribute               | Value                                           |
	| ----------------------- | ----------------------------------------------- |
	| Precision               | 19                                              |
	| Maximum Exponent (Emax) | 18                                              |
	| Minimum Exponent (Emin) | -19                                             |
	| Tiny Exponent (Etiny)   | -19                                             |
	| Rounding Method         | Half To Even                                    |
	| Enabled Traps           | Division by Zero, Invalid Operation, Overflow   |
	| Disabled Traps          | Inexact, Clamped, Rounded, Subnormal, Underflow |

Equality of Etiny and Emin implies that this package does not support subnormal numbers.

# Rounding

Implicit rounding is applied when a result coefficient has more than 19 digits,
in this case the coefficient is rounded to 19 digits using half-to-even rounding,
which ensures that rounding errors are evenly distributed between rounding up
and rounding down.

For all operations, excluding [Decimal.Pow] and [Decimal.PowExact], the result is
the one that would be obtained by computing the exact mathematical result with
infinite precision and then rounding it to 19 digits.

[Decimal.Pow] and [Decimal.PowExact] operations may occasionally produce a result that is
off by one unit in the last place.

In addition to implicit rounding, the package provides several
methods for explicit rounding:

  - [Decimal.Round], [Decimal.Quantize], [Decimal.Rescale]:
    round using half-to-even rounding.
  - [Decimal.Ceil]:
    rounds towards positive infinity.
  - [Decimal.Floor]:
    rounds towards negative infinity.
  - [Decimal.Trunc]:
    rounds towards zero.

# Errors

All methods are panic-free and pure, returning errors in cases such as
overflow, invalid operation or division by zero.

  - Division by Zero.
    Unlike the standard library, arithmetic operations do not panic when
    dividing by zero.  Instead, an error is returned.

  - Invalid Operation.
    [Decimal.Pow] may return invalid operation error if zero is raised to
    a negative power.

  - Overflow Error.
    Unlike standard integers, there is no "wrap around" for decimals at certain
    sizes, for out-of-range values operations return an overflow error.

  - Underflow Error.
    Operations do not return an error in case of decimal underflow.  If result
    of an operation is a decimal between -0.00000000000000000005 and
    0.00000000000000000005 inclusive it will be rounded to 0.

[ANSI X3.274-1996 (section 7.4)]: https://speleotrove.com/decimal/dax3274.html
[big.Int]: https://pkg.go.dev/math/big#Int
*/
package decimal
