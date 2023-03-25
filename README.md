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

To create a new Decimal value, you can use one of the provided constructors,
such as `New`, `Parse` or `MustParse`.

```go
x := decimal.New(12345, 2) // x = 123,45
y := decimal.MustParse("123.45")
```

Once you have a Decimal value, you can perform arithmetic operations such as
addition, subtraction, multiplication, division, and exponentiation, as well
as rounding operations such as ceiling, floor, truncation, and rounding.

```go
sum := x.Add(y)
difference := x.Sub(y)
product := x.Mul(y)
quotient := x.Quo(y)
power := x.Pow(5)
ceil := x.Ceil(0)
floor := x.Floor(0)
trunc := x.Trunc(0)
round := x.Round(0)
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

| Test Case           | Expression | govalues | [cockroachdb] v3.1.2 | cockroachdb vs govalues | [shopspring] v1.3.1 | shopspring vs govalues |
| ------------------- | ---------- | -------: | -------------------: | ----------------------: | ------------------: | ---------------------: |
| Decimal_Add         | 2 + 3      |   16.32n |               46.76n |                +186.49% |             138.75n |               +750.18% |
| Decimal_Mul         | 2 * 3      |   15.81n |               51.41n |                +225.21% |             135.90n |               +759.58% |
| Decimal_QuoFinite   | 2 / 4      |   78.36n |              379.50n |                +384.30% |             641.60n |               +718.79% |
| Decimal_QuoInfinite | 2 / 3      |   573.7n |               948.4n |                 +65.31% |             2828.5n |               +393.03% |
| Decimal_Pow         | 1.1^60     |   1.029µ |               3.078µ |                +199.08% |             19.949µ |              +1838.68% |
| Parse               | -          |   106.6n |               252.1n |                +136.60% |              497.5n |               +366.92% |
| Decimal_String      | -          |   138.2n |               194.8n |                 +40.99% |              329.3n |               +138.28% |

The benchmark results shown in the table are provided for informational purposes only and may vary depending on your specific use case.

## Contributing to the project

The Decimal package is hosted on [GitHub](https://github.com/govalues/decimal).
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
