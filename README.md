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

| Test Case      | Expression           | govalues | [cockroachdb] v3.2.0 | cockroachdb vs govalues | [shopspring] v1.3.1 | shopspring vs govalues |
| -------------- | -------------------- | -------: | -------------------: | ----------------------: | ------------------: | ---------------------: |
| Add            | 2 + 3                |   15.79n |               47.95n |                +203.64% |             141.95n |               +798.99% |
| Mul            | 2 * 3                |   16.61n |               54.66n |                +229.18% |             144.95n |               +772.93% |
| QuoFinite      | 2 / 4                |   64.74n |              381.15n |                +488.74% |             645.35n |               +896.83% |
| QuoInfinite    | 2 / 3                |  595.30n |             1001.50n |                 +68.23% |            2810.50n |               +372.11% |
| Pow            | 1.1^60               |    1.31µ |                3.17µ |                +142.42% |              20.50µ |              +1469.53% |
| Pow            | 1.01^600             |    4.36µ |               13.86µ |                +217.93% |              44.39µ |               +918.44% |
| Pow            | 1.001^6000           |    7.39µ |               24.69µ |                +234.34% |             656.84µ |              +8793.66% |
| Parse          | 1                    |   17.27n |               78.25n |                +353.23% |             128.80n |               +646.02% |
| Parse          | 123.456              |   39.80n |              211.85n |                +432.22% |             237.60n |               +496.91% |
| Parse          | 123456789.1234567890 |  106.20n |              233.10n |                +119.59% |             510.90n |               +381.30% |
| String         | 1                    |    5.45n |               19.91n |                +265.49% |             197.85n |              +3531.94% |
| String         | 123.456              |   42.38n |               74.83n |                 +76.57% |             229.50n |               +441.53% |
| String         | 123456789.1234567890 |   77.90n |              210.40n |                +170.11% |             328.90n |               +322.24% |
| NewFromFloat64 | 1                    |  155.50n |              361.00n |                +132.23% |             234.40n |                +50.76% |
| NewFromFloat64 | 123.456              |  237.10n |              588.80n |                +148.37% |             770.30n |               +224.95% |
| NewFromFloat64 | 123456789.1234567890 |  335.60n |              636.80n |                 +89.78% |             753.80n |               +124.65% |
| Float64        | 1                    |   28.92n |               51.64n |                 +78.59% |             456.70n |              +1479.46% |
| Float64        | 123.456              |   97.36n |              102.68n |                  +5.46% |             680.05n |               +598.49% |
| Float64        | 123456789.1234567890 |  206.10n |              304.60n |                 +47.79% |             792.60n |               +284.55% |
| **Geom. Mean** |                      |  121.40n |              313.60n |                +158.28% |             912.20n |               +651.24% |

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
