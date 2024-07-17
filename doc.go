/*
Package decimal implements immutable decimal floating-point numbers.
It is specifically designed for transactional financial systems
and adheres to the principles set by [ANSI X3.274-1996].

# Representation

Decimal is a struct with three fields:

  - Sign:
    A boolean indicating whether the decimal is negative.
  - Coefficient:
    An unsigned integer representing the numeric value of the decimal without
    the decimal point.
  - Scale:
    A non-negative integer indicating the position of the decimal point
    within the coefficient.
    For example, a decimal with a coefficient of 12345 and a scale of 2 represents
    the value 123.45.
    Conceptually, the scale can be understood as the inverse of the exponent in
    scientific notation.
    For example, a scale of 2 corresponds to an exponent of -2.
    The range of allowed values for the scale is from 0 to 19.

The numerical value of a decimal is calculated as follows:

  - -Coefficient / 10^Scale if Sign is true.
  - Coefficient / 10^Scale if Sign is false.

This approach allows the same numeric value to have multiple representations,
for example, 1, 1.0, and 1.00, which represent the same value but have different
scales and coefficients.

# Constraints

The range of a decimal is determined by its scale.
Here are the ranges for frequently used scales:

	| Example      | Scale | Minimum                              | Maximum                             |
	| ------------ | ----- | ------------------------------------ | ----------------------------------- |
	| Japanese Yen | 0     | -9,999,999,999,999,999,999           | 9,999,999,999,999,999,999           |
	| US Dollar    | 2     |    -99,999,999,999,999,999.99        |    99,999,999,999,999,999.99        |
	| Omani Rial   | 3     |     -9,999,999,999,999,999.999       |     9,999,999,999,999,999.999       |
	| Bitcoin      | 8     |            -99,999,999,999.99999999  |            99,999,999,999.99999999  |
	| Ethereum     | 9     |             -9,999,999,999.999999999 |             9,999,999,999.999999999 |

[Subnormal numbers] are not supported to ensure peak performance.
Consequently, decimals between -0.00000000000000000005 and 0.00000000000000000005
inclusive, are rounded to 0.

Special values such as [NaN], [Infinity], or [negative zeros] are not supported.
This ensures that arithmetic operations always produce either valid decimals
or errors.

# Operations

Each arithmetic operation, except for [Decimal.Sqrt] and [Decimal.SqrtExact],
occurs in two steps:

 1. The operation is initially performed using uint64 arithmetic.
    If no overflow occurs, the exact result is immediately returned.
    If overflow occurs, the operation proceeds to step 2.

 2. The operation is repeated with increased precision using [big.Int] arithmetic.
    The result is then rounded to 19 digits.
    If no significant digits are lost during rounding, the inexact result is returned.
    If any significant digit is lost, an overflow error is returned.

Step 1 improves performance by avoiding performance impact associated with [big.Int] arithmetic.
It is expected that, in transactional financial systems, most arithmetic operations
will compute an exact result during step 1.

The following rules determine the significance of digits during step 2:

  - [Decimal.Add], [Decimal.Sub], [Decimal.Mul], [Decimal.FMA], [Decimal.Pow],
    [Decimal.Quo], [Decimal.QuoRem], [Decimal.Inv], [Decimal.Sqrt]:
    All digits in the integer part are significant, while digits in the
    fractional part are considered insignificant.
  - [Decimal.AddExact], [Decimal.SubExact], [Decimal.MulExact], [Decimal.FMAExact],
    [Decimal.PowExact], [Decimal.QuoExact], [Decimal.SqrtExact]:
    All digits in the integer part are significant. The significance of digits
    in the fractional part is determined by the scale argument, which is typically
    equal to the scale of the currency.

# Context

Unlike many other decimal libraries, this package does not provide
an explicit context.
Instead, the [context] is implicit and can be approximately equated to
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

The equality of Etiny and Emin implies that this package does not support
subnormal numbers.

# Rounding

Implicit rounding is applied when a result exceeds 19 digits,
rounding it to 19 digits using half-to-even rounding.
This method ensures that rounding errors are evenly distributed between rounding up
and down.

For all arithmetic operations, except for [Decimal.Pow] and [Decimal.PowExact],
the result is the one that would be obtained by computing the exact mathematical
result with infinite precision and then rounding it to 19 digits.
[Decimal.Pow] and [Decimal.PowExact] may occasionally produce a result that is
off by 1 unit in the last place.

In addition to implicit rounding, the package provides several methods for
explicit rounding:

  - Half-to-even rounding:
    [Decimal.Round], [Decimal.Quantize], [Decimal.Rescale].
  - Rounding towards positive infinity:
    [Decimal.Ceil].
  - Rounding towards negative infinity:
    [Decimal.Floor].
  - Rounding towards zero:
    [Decimal.Trunc].

See the documentation for each method for more details.

# Errors

All methods are panic-free and pure.
Errors are returned in the following cases:

  - Division by Zero:
    Unlike the standard library, [Decimal.Quo], [Decimal.QuoRem], and [Decimal.Inv]
    do not panic when dividing by 0.
    Instead, they return an error.

  - Invalid Operation:
    [Decimal.Pow] and [Decimal.PowExact] return an error if 0 is raised to
    a negative power.
    [Decimal.Sqrt] and [Decimal.SqrtExact] return an error if the square root
    of a negative decimal is requested.

  - Overflow:
    Unlike standard integers, there is no "wrap around" for decimals at certain sizes.
    For out-of-range values, arithmetic operations return an error.

Errors are not returned in the following cases:

  - Underflow:
    Arithmetic operations do not return an error in case of decimal underflow.
    If the result is a decimal between -0.00000000000000000005 and
    0.00000000000000000005 inclusive, it will be rounded to 0.

# Conversions

A. JSON

The package integrates seamlessly with standard [encoding/json] through
the implementation of [encoding.TextMarshaller] and [encoding.TextUnmarshaler]
interfaces.
Below is an example structure:

	type Object struct {
	  Number decimal.Decimal `json:"some_number"`
	  // Other fields...
	}

This package marshals decimals as quoted strings, ensuring the preservation of
the exact numerical value.
Below is an example OpenAPI schema:

	Decimal:
	  type: string
	  format: decimal
	  pattern: '^(\-|\+)?((\d+(\.\d*)?)|(\.\d+))$'

B. XML

The package integrates with standard [encoding/xml] via the implementation of
[encoding.TextMarshaller] and [encoding.TextUnmarshaler] interfaces.
Below is an example structure:

	type Entity struct {
	  Number decimal.Decimal `xml:"SomeNumber"`
	  // Other fields...
	}

"xs:decimal" type can represent decimals in XML schema.
It is possible to impose restrictions on the length of the decimals
using the following type:

	<xs:simpleType name="Decimal">
	  <xs:restriction base="xs:decimal">
	    <xs:totalDigits value="19"/>
	  </xs:restriction>
	</xs:simpleType>

C. Protocol Buffers

Protocol Buffers can represent decimals as numerical strings, preserving trailing zeros.
To convert between numerical strings and decimals, use [Parse] and [Decimal.String].
Below is an example of a proto definition:

	message Decimal {
	  string value = 1;
	}

Alternatively, decimals can be represented as two integers:
one for the integer part and another for the fractional part.
However, this format does not preserve trailing zeros and rounds decimals
with more than nine digits in the fractional part.
For conversion between this format and decimals, use [NewFromInt64] and
[Decimal.Int64] with a scale argument of "9".
Below is an example of a proto definition:

	message Decimal {
	  int64 units = 1;
	  int32 nanos = 2;
	}

D. SQL

The package integrates with the standard [database/sql] via the implementation
of [sql.Scanner] and [driver.Valuer] interfaces.
To ensure accurate preservation of decimal scales, it is essential to choose
appropriate column types:

	| Database   | Type                          |
	| ---------- | ----------------------------- |
	| PostgreSQL | DECIMAL                       |
	| SQLite     | TEXT                          |
	| MySQL      | DECIMAL(19, d) or VARCHAR(22) |

Below are the reasons for these preferences:

  - PostgreSQL:
    Always use DECIMAL without precision or scale specifications, that is,
    avoid DECIMAL(p) or DECIMAL(p, s).
    DECIMAL accurately preserves the scale of decimals.

  - SQLite:
    Prefer TEXT, since DECIMAL is just an alias for binary floating-point numbers.
    TEXT accurately preserves the scale of decimals.

  - MySQL:
    Use DECIMAL(19, d), as DECIMAL is merely an alias for DECIMAL(10, 0).
    The downside of this format is that MySQL automatically rescales all decimals:
    it rounds values with more than d digits in the fractional part (using half
    away from zero) and pads with trailing zeros those with fewer than d digits
    in the fractional part.
    To prevent automatic rescaling, consider using VARCHAR(22), which accurately
    preserves the scale of decimals.

[Infinity]: https://en.wikipedia.org/wiki/Infinity#Computing
[Subnormal numbers]: https://en.wikipedia.org/wiki/Subnormal_number
[NaN]: https://en.wikipedia.org/wiki/NaN
[ANSI X3.274-1996]: https://speleotrove.com/decimal/dax3274.html
[big.Int]: https://pkg.go.dev/math/big#Int
[sql.Scanner]: https://pkg.go.dev/database/sql#Scanner
[negative zeros]: https://en.wikipedia.org/wiki/Signed_zero
[context]: https://speleotrove.com/decimal/damodel.html
*/
package decimal
