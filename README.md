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

- **Immutability** - Once a decimal is set, it remains unchanged.
  This immutability ensures safe concurrent access across goroutines.
- **Banker's Rounding** - Methods use half-to-even rounding, also known as
  "banker's rounding", which minimizes cumulative rounding errors commonly seen
  in financial calculations.
- **No Panics** - All methods are designed to be panic-free.
  Instead of potentially crashing your application, they return errors for issues
  such as overflow or division by zero.
- **Zero Heap Allocation** - Methods are optimized to avoid heap allocations,
  reducing the impact on the garbage collector during arithmetic operations.
- **Simple String Representation** - Decimals are represented in a strightforward
  format avoiding the complexities of scientific or engineering notations.
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
    fmt.Println(d.Pow(2))          // 8²
    fmt.Println(d.Sqrt())          // √8

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

| Feature          | govalues     | [cockroachdb/apd] v3.2.1 | [shopspring/decimal] v1.4.0 |
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

| Test Case   | Expression           | govalues | [cockroachdb/apd] v3.2.1 | [shopspring/decimal] v1.4.0 | govalues vs cockroachdb | govalues vs shopspring |
| ----------- | -------------------- | -------: | -----------------------: | --------------------------: | ----------------------: | ---------------------: |
| Add         | 5 + 6                |   16.06n |                   74.88n |                     140.90n |                +366.22% |               +777.33% |
| Mul         | 2 * 3                |   16.93n |                   62.20n |                     146.00n |                +267.40% |               +762.37% |
| QuoExact    | 2 ÷ 4                |   59.52n |                  176.95n |                     657.40n |                +197.30% |              +1004.50% |
| QuoInfinite | 2 ÷ 3                |  391.60n |                  976.80n |                    2962.50n |                +149.39% |               +656.42% |
| Pow         | 1.1^60               |  950.90n |                 3302.50n |                    4599.50n |                +247.32% |               +383.73% |
| Pow         | 1.01^600             |    3.45µ |                   10.67µ |                      18.67µ |                +209.04% |               +440.89% |
| Pow         | 1.001^6000           |    5.94µ |                   20.50µ |                     722.22µ |                +244.88% |             +12052.44% |
| Sqrt        | √2                   |    3.40µ |                    4.96µ |                    2101.86µ |                 +46.00% |             +61755.71% |
| Parse       | 1                    |   16.52n |                   76.30n |                     136.55n |                +362.00% |               +726.82% |
| Parse       | 123.456              |   47.37n |                  176.90n |                     242.60n |                +273.44% |               +412.14% |
| Parse       | 123456789.1234567890 |   85.49n |                  224.15n |                     497.95n |                +162.19% |               +482.47% |
| String      | 1                    |    5.11n |                   19.57n |                     198.25n |                +283.21% |              +3783.07% |
| String      | 123.456              |   35.78n |                   77.12n |                     228.85n |                +115.52% |               +539.51% |
| String      | 123456789.1234567890 |   70.72n |                  239.10n |                     337.25n |                +238.12% |               +376.91% |
| Telco       | see [specification]  |  137.00n |                  969.40n |                    3981.00n |                +607.33% |              +2804.78% |

The benchmark results shown in the table are provided for informational purposes only and may vary depending on your specific use case.

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
