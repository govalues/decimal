# Decimal

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![licenseb]][license]
[![godocb]][godoc]
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

| Test Case           | Expression           |   govalues | [cockroachdb] v3.1.2 | cockroachdb vs govalues | [shopspring] v1.3.1 | shopspring vs govalues |
| ------------------- | -------------------- | ---------: | -------------------: | ----------------------: | ------------------: | ---------------------: |
| Decimal_Add         | 2 + 3                |     15.54n |               47.39n |                +204.89% |             142.65n |               +817.66% |
| Decimal_Mul         | 2 * 3                |     15.87n |               61.18n |                +285.51% |             136.60n |               +760.74% |
| Decimal_QuoFinite   | 2 / 4                |     55.95n |              360.70n |                +544.74% |             654.45n |              +1069.81% |
| Decimal_QuoInfinite | 2 / 3                |    570.10n |              936.40n |                 +64.27% |            2858.00n |               +401.36% |
| Decimal_Pow         | 1.1^60               |      1.01µ |                3.06µ |                +202.92% |              20.09µ |              +1889.90% |
| Parse               | 1                    |     17.35n |               88.44n |                +409.74% |             132.10n |               +661.38% |
| Parse               | 123.456              |     36.75n |              220.15n |                +499.05% |             244.45n |               +565.17% |
| Parse               | 123456789.1234567890 |    105.80n |              240.40n |                +127.33% |             496.00n |               +369.08% |
| Decimal_String      | 1                    |      6.39n |               23.42n |                +266.32% |             237.95n |              +3622.62% |
| Decimal_String      | 123.456              |     37.65n |               38.27n |                  +1.66% |             230.40n |               +512.03% |
| Decimal_String      | 123456789.1234567890 |     74.95n |              189.70n |                +153.10% |             324.10n |               +332.42% |
| **Geometric Mean**  |                      | **53.20n** |          **164.50n** |            **+209.17%** |         **462.80n** |           **+769.86%** |

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
