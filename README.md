# Decimal

[![godocb]][godoc]
[![githubb]][github]
[![goreportb]][goreport]

Package decimal implements decimal floating-point numbers for Go.

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

| Case     | GoValues, v0.0.1 | [ShopSpring](https://pkg.go.dev/github.com/shopspring/decimal), v1.3.1 | [CockroachDB](https://pkg.go.dev/github.com/cockroachdb/apd), v1.1.0 |
| -------- | ---------------: | -----------------: | ------------------: |
| 2 + 3    |         10 ns/op |          156 ns/op |            88 ns/op |
| 2 * 3    |         12 ns/op |          160 ns/op |            84 ns/op |
| 2 / 4    |         31 ns/op |        2,650 ns/op |           854 ns/op |
| 2 / 3    |        917 ns/op |        3,347 ns/op |        14,095 ns/op |
| 1.1 ^ 60 |      1,050 ns/op |       11,889 ns/op |         4,044 ns/op |

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
[githubb]: https://github.com/govalues/decimal/actions/workflows/go.yml/badge.svg
