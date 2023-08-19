# Decimal

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]

Package decimal implements immutable decimal floating-point numbers for Go.

## Getting started

To install the decimal package into your Go workspace, you can use the go get command:

```bash
go get github.com/govalues/decimal
```

To use the decimal package in your Go project, you can import it as follows:

```go
import "github.com/govalues/decimal"
```

## Using Decimal

To create a new decimal value, you can use one of the provided constructors,
such as `New`, `MustNew`, `Parse` or `MustParse`.

```go
d := decimal.MustNew(12345, 2) // d = 123.45
e := decimal.MustParse("123.45")
```

Once you have a decimal value, you can perform arithmetic operations such as
addition, subtraction, multiplication, division, and exponentiation, as well
as rounding operations such as ceiling, floor, truncation, and rounding.

```go
sum, _ := d.Add(e)
difference, _ := d.Sub(e)
product, _ := d.Mul(e)
quotient, _ := d.Quo(e)
power, _ := d.Pow(5)
ceil := d.Ceil(0)
floor := d.Floor(0)
trunc := d.Trunc(0)
round := d.Round(0)
```

For more details on these and other methods, see the package documentation
at [pkg.go.dev](https://pkg.go.dev/github.com/govalues/decimal).

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/govalues/benchmarks
cpu: AMD Ryzen 7 3700C  with Radeon Vega Mobile Gfx 
```

| Test Case           | Expression           | govalues | [cockroachdb] v3.2.0 | cockroachdb vs govalues | [shopspring] v1.3.1 | shopspring vs govalues |
| ------------------- | -------------------- | -------: | -------------------: | ----------------------: | ------------------: | ---------------------: |
| Decimal_Add         | 2 + 3                |   15.73n |               50.24n |                +219.39% |             141.85n |               +801.78% |
| Decimal_Mul         | 2 * 3                |   15.67n |               67.34n |                +329.57% |             139.55n |               +790.27% |
| Decimal_QuoFinite   | 2 / 4                |   61.08n |              371.35n |                +508.02% |             629.70n |               +931.03% |
| Decimal_QuoInfinite | 2 / 3                |  560.80n |              946.10n |                 +68.69% |            2736.50n |               +387.92% |
| Decimal_Pow         | 1.1^60               |    1.02µ |                2.89µ |                +182.58% |              20.08µ |              +1865.17% |
| Parse               | 123456789.1234567890 |   99.16n |              219.15n |                +121.01% |             480.10n |               +384.17% |
| Parse               | 123.456              |   37.41n |              203.80n |                +444.85% |             229.20n |               +512.75% |
| Parse               | 1                    |   16.37n |               76.27n |                +365.91% |             128.90n |               +687.42% |
| Decimal_String      | 123456789.1234567890 |   76.47n |              215.15n |                +181.35% |             329.35n |               +330.69% |
| Decimal_String      | 123.456              |   40.86n |               70.52n |                 +72.59% |             222.10n |               +443.56% |
| Decimal_String      | 1                    |    5.32n |               19.41n |                +264.88% |             192.75n |              +3523.46% |
| **Geometric Mean**  |                      |   52.70n |              170.10n |                +222.71% |             445.40n |               +745.22% |

The benchmark results shown in the table are provided for informational purposes only and may vary depending on your specific use case.

## Contributing to the project

The decimal package is hosted on [GitHub](https://github.com/govalues/decimal).
To contribute to the project, follow these steps:

 1. Fork the repository and clone it to your local machine.
 1. Make the desired changes to the code.
 1. Write tests for the changes you made.
 1. Ensure that all tests pass by running `go test`.
 1. Commit the changes and push them to your fork.
 1. Submit a pull request with a clear description of the changes you made.
 1. Wait for the maintainers to review and merge your changes.

Note: Before making any significant changes to the code, it is recommended to open an issue to discuss the proposed changes with the maintainers. This will help to ensure that the changes align with the project's goals and roadmap.

[codecov]: https://codecov.io/gh/govalues/decimal
[codecovb]: https://img.shields.io/codecov/c/github/govalues/decimal/main?color=brightcolor
[goreport]: https://goreportcard.com/report/github.com/govalues/decimal
[goreportb]: https://goreportcard.com/badge/github.com/govalues/decimal
[github]: https://github.com/govalues/decimal/actions/workflows/go.yml
[githubb]: https://img.shields.io/github/actions/workflow/status/govalues/decimal/go.yml
[godoc]: https://pkg.go.dev/github.com/govalues/decimal#section-documentation
[godocb]: https://img.shields.io/badge/go.dev-reference-blue
[version]: https://go.dev/dl
[versionb]: https://img.shields.io/github/go-mod/go-version/govalues/decimal?label=go
[license]: https://en.wikipedia.org/wiki/MIT_License
[licenseb]: https://img.shields.io/github/license/govalues/decimal?color=blue
[cockroachdb]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring]: https://pkg.go.dev/github.com/shopspring/decimal
