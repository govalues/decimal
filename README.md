[![godocb]][godoc]
![versionb]
[![goreportb]][goreport]
[![githubb]][github]
[![codecovb]][codecov]
[![licenseb]][license]

# Decimal

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
rounded := x.Round(0)
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

| Expression                      | govalues | [cockroachdb] v3.1.2 | cockroachdb vs govalues | [shopspring] v1.3.1 | shopspring vs govalues |
| ------------------------------- | -------: | -------------------: | ----------------------: | ------------------: | ---------------------: |
| Parse("1234567890.123456789")   |   108.5n |               260.6n |                +140.30% |              556.8n |               +413.42% |
| "1234567890.123456789".String() |   140.9n |               221.7n |                 +57.37% |              373.6n |               +165.21% |
| 2 * 3                           |   18.77n |               77.20n |                +311.43% |             165.30n |               +780.90% |
| 2 + 3                           |   17.09n |               58.68n |                +243.46% |             158.05n |               +825.08% |
| 2 / 4                           |   40.62n |              366.90n |                +803.25% |             663.45n |              +1533.31% |
| 2 / 3                           |   747.6n |               970.5n |                 +29.81% |             2923.0n |               +290.96% |
| 1.1^60                          |   1.093µ |               3.078µ |                +181.61% |             14.949µ |              +1267.70% |

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

[godoc]: https://pkg.go.dev/github.com/govalues/decimal?tab=doc
[godocb]: https://img.shields.io/badge/go.dev-reference-blue
[goreport]: https://goreportcard.com/report/github.com/govalues/decimal
[goreportb]: https://goreportcard.com/badge/github.com/govalues/decimal
[github]: https://github.com/govalues/decimal/actions/workflows/go.yml
[githubb]: https://img.shields.io/github/actions/workflow/status/govalues/decimal/go.yml
[codecov]: https://codecov.io/gh/govalues/decimal
[codecovb]: https://img.shields.io/codecov/c/github/govalues/decimal/main
[license]: https://en.wikipedia.org/wiki/MIT_License
[licenseb]: https://img.shields.io/github/license/govalues/decimal
[versionb]: https://img.shields.io/github/go-mod/go-version/govalues/decimal?label=go
[cockroachdb]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring]: https://pkg.go.dev/github.com/shopspring/decimal
