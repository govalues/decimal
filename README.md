# decimal

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]
[![awesomeb]][awesome]

Package decimal implements immutable decimal floating-point numbers for Go.
This package is designed specifically for use in transactional financial systems.

## Features

- **Optimized Performance** - Utilizes `uint64` for coefficients, reducing heap
  allocations and memory consumption.
- **Immutability** - Once a decimal is set, it remains unchanged.
  This immutability ensures safe concurrent access across goroutines.
- **Banker's Rounding** - Methods use half-to-even rounding, also known as "banker's rounding",
  which minimizes cumulative rounding errors commonly seen in financial calculations.
- **No Panics** - All methods are designed to be panic-free.
  Instead of potentially crashing your application, they return errors for issues
  such as overflow or division by zero.
- **Simple String Representation** - Decimals are represented without the complexities
  of scientific or engineering notation.
- **Correctness** - Fuzz testing is used to [cross-validate] arithmetic operations
  against the [cockroachdb/apd] and [shopspring/decimal] packages.

## Getting Started

### Installation

To add the decimal package to your Go workspace:

```bash
go get github.com/govalues/decimal
```

### Usage

Create decimal values using one of the constructors.
After creating a decimal value, various operations can be performed:

```go
package main

import (
    "fmt"
    "github.com/govalues/decimal"
)

func main() {
    // Constructors
    d, _ := decimal.New(8, 0)               // d = 8
    e, _ := decimal.Parse("12.5")           // e = 12.5
    f, _ := decimal.NewFromFloat64(2.567)   // f = 2.567
    g, _ := decimal.NewFromInt64(7, 896, 3) // g = 7.896

    // Operations
    fmt.Println(d.Add(e))          // 8 + 12.5
    fmt.Println(d.Sub(e))          // 8 - 12.5

    fmt.Println(d.Mul(e))          // 8 * 12.5
    fmt.Println(d.FMA(e, f))       // 8 * 12.5 + 2.567
    fmt.Println(d.Pow(2))          // 8 ^ 2

    fmt.Println(d.Quo(e))          // 8 ÷ 12.5
    fmt.Println(d.QuoRem(e))       // 8 div 12.5, 8 mod 12.5
    fmt.Println(d.Inv())           // 1 ÷ 8

    // Rounding to 2 decimal places
    fmt.Println(g.Round(2))        // 7.90
    fmt.Println(g.Ceil(2))         // 7.90
    fmt.Println(g.Floor(2))        // 7.89
    fmt.Println(g.Trunc(2))        // 7.89

    // Conversions
    fmt.Println(f.Int64(9))        // 2 567000000
    fmt.Println(f.Float64())       // 2.567
    fmt.Println(f.String())        // 2.567

    // Formatting
    fmt.Printf("%.2f\n", f)        // 2.57
    fmt.Printf("%.2k\n", f)        // 256.70%
}
```

## Documentation

For detailed documentation and additional examples, visit the package
[documentation](https://pkg.go.dev/github.com/govalues/decimal#section-documentation).
For examples related to financial calculations, see the `money` package
[documentation](https://pkg.go.dev/github.com/govalues/money#section-documentation).

## Comparison

Comparison with other popular packages:

| Feature          | govalues     | [cockroachdb/apd] v3.2.1 | [shopspring/decimal] v1.3.1 |
| ---------------- | ------------ | ------------------------ | --------------------------- |
| Speed            | High         | Medium                   | Low[^reason]                |
| Mutability       | Immutable    | Mutable[^reason]         | Immutable                   |
| Memory Footprint | Low          | Medium                   | High                        |
| Panic Free       | Yes          | Yes                      | No[^divzero]                |
| Precision        | 19 digits    | Arbitrary                | Arbitrary                   |
| Default Rounding | Half to even | Half up                  | Half away from 0            |
| Context          | Implicit     | Explicit                 | Implicit                    |

[^reason]: decimal package was created simply because [shopspring/decimal] was
too slow and [cockroachdb/apd] was mutable.

[^divzero]: [shopspring/decimal] panics on division by zero.

### Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/govalues/decimal-tests
cpu: AMD Ryzen 7 3700C  with Radeon Vega Mobile Gfx 
```

| Test Case   | Expression           | govalues | [cockroachdb/apd] v3.2.1 | [shopspring/decimal] v1.3.1 | govalues vs cockroachdb | govalues vs shopspring |
| ----------- | -------------------- | -------: | -----------------------: | --------------------------: | ----------------------: | ---------------------: |
| Add         | 5 + 6                |   16.89n |                   80.96n |                     140.50n |                +379.48% |               +732.10% |
| Mul         | 2 * 3                |   16.85n |                   58.14n |                     145.30n |                +245.15% |               +762.57% |
| QuoExact    | 2 ÷ 4                |   66.00n |                  193.25n |                     619.15n |                +192.78% |               +838.03% |
| QuoInfinite | 2 ÷ 3                |  453.30n |                  961.00n |                    2767.00n |                +112.01% |               +510.41% |
| Pow         | 1.1^60               |    1.04µ |                    3.42µ |                      15.76µ |                +227.72% |              +1408.43% |
| Pow         | 1.01^600             |    3.57µ |                   10.70µ |                      35.70µ |                +200.11% |               +901.23% |
| Pow         | 1.001^6000           |    6.19µ |                   20.72µ |                     634.41µ |                +234.65% |             +10148.95% |
| Parse       | 1                    |   18.10n |                   85.66n |                     136.75n |                +373.23% |               +655.52% |
| Parse       | 123.456              |   54.16n |                  197.25n |                     238.45n |                +264.20% |               +340.27% |
| Parse       | 123456789.1234567890 |  111.00n |                  238.20n |                     498.00n |                +114.59% |               +348.65% |
| String      | 1                    |    5.70n |                   20.89n |                     203.25n |                +266.24% |              +3464.23% |
| String      | 123.456              |   42.74n |                   75.71n |                     235.65n |                 +77.14% |               +451.36% |
| String      | 123456789.1234567890 |   72.34n |                  215.90n |                     331.20n |                +198.47% |               +357.87% |
| Telco       | see [specification]  |  148.00n |                 1075.00n |                    4010.50n |                +626.35% |              +2609.80% |

The benchmark results shown in the table are provided for informational purposes only and may vary depending on your specific use case.

## Contributing

Interested in contributing? Here's how to get started:

1. Fork and clone the repository.
1. Implement your changes.
1. Write tests to cover your changes.
1. Ensure all tests pass with `go test`.
1. Commit and push to your fork.
1. Open a pull request detailing your changes.

**Note**: If you're considering significant changes, please open an issue first to
discuss with the maintainers.
This ensures alignment with the project's objectives and roadmap.

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
[awesome]: https://github.com/avelino/awesome-go#financial
[awesomeb]: https://awesome.re/mentioned-badge.svg
[cockroachdb/apd]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring/decimal]: https://pkg.go.dev/github.com/shopspring/decimal
[specification]: https://speleotrove.com/decimal/telcoSpec.html
[cross-validate]: https://github.com/govalues/decimal-tests
