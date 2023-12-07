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

- **Optimized Performance** - Utilizes uint64 for coefficients, reducing heap
  allocations and memory consumption.
- **Immutability** - Once a decimal is set, it remains unchanged.
  This immutability ensures safe concurrent access across goroutines.
- **Banker's Rounding** - Methods use half even rounding, also known as "banker's rounding",
  which minimizes cumulative rounding errors commonly seen in financial calculations.
- **No Panics** - All methods are designed to be panic-free.
  Instead of potentially crashing your application, they return errors for issues
  such as overflow or division by zero.
- **Simple String Representation** - Decimals are represented without the complexities
  of scientific or engineering notation.
- **Correctness** - Fuzz testing is used to [cross-validate] arithmetic operations
  against both the [cockroachdb] and [shopspring] decimal packages.

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

    fmt.Println(d.Quo(e))          // 8 / 12.5
    fmt.Println(d.QuoRem(e))       // 8 div 12.5, 8 mod 12.5
    fmt.Println(d.Inv())           // 1 / 8

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

| Feature          | govalues     | [cockroachdb] v3.2.1 | [shopspring] v1.3.1 |
| ---------------- | ------------ | -------------------- | ------------------- |
| Speed            | High         | Medium               | Low[^reason]        |
| Mutability       | Immutable    | Mutable[^reason]     | Immutable           |
| Memory Footprint | Low          | Medium               | High                |
| Panic Free       | Yes          | Yes                  | No                  |
| Precision        | 19 digits    | Arbitrary            | Arbitrary           |
| Default Rounding | Half to even | Half up              | Half away from 0    |
| Context          | Implicit     | Explicit             | Implicit            |

[^reason]: decimal package was created simply because shopspring's decimal was
too slow and cockroachdb's decimal was mutable.

### Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/govalues/decimal-tests
cpu: AMD Ryzen 7 3700C  with Radeon Vega Mobile Gfx 
```

| Test Case   | Expression           | govalues | [cockroachdb] v3.2.1 | [shopspring] v1.3.1 | govalues vs cockroachdb | govalues vs shopspring |
| ----------- | -------------------- | -------: | -------------------: | ------------------: | ----------------------: | ---------------------: |
| Add         | 2 + 3                |   15.53n |               46.68n |             142.30n |                +200.45% |               +816.00% |
| Mul         | 2 * 3                |   15.64n |               52.83n |             137.35n |                +237.76% |               +778.20% |
| QuoFinite   | 2 / 4                |   51.65n |              179.60n |             619.40n |                +247.76% |              +1099.34% |
| QuoInfinite | 2 / 3                |  568.80n |              935.20n |            2749.00n |                 +64.43% |               +383.30% |
| Pow         | 1.1^60               |    1.28µ |                3.28µ |              16.03µ |                +156.99% |              +1156.09% |
| Pow         | 1.01^600             |    4.31µ |               10.43µ |              37.00µ |                +142.15% |               +758.69% |
| Pow         | 1.001^6000           |    7.54µ |               20.39µ |             651.51µ |                +170.58% |              +8544.78% |
| Parse       | 1                    |   17.14n |               77.64n |             129.15n |                +353.00% |               +653.50% |
| Parse       | 123.456              |   36.15n |              201.85n |             235.25n |                +458.37% |               +550.76% |
| Parse       | 123456789.1234567890 |   98.90n |              210.95n |             475.05n |                +113.30% |               +380.33% |
| String      | 1                    |    5.18n |               21.43n |             208.00n |                +313.99% |              +3918.16% |
| String      | 123.456              |   42.31n |               67.55n |             226.55n |                 +59.66% |               +435.52% |
| String      | 123456789.1234567890 |   76.04n |              209.50n |             329.95n |                +175.49% |               +333.89% |
| Telco       | see [specification]  |  134.00n |              947.60n |            3945.50n |                +607.13% |              +2844.40% |

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
[cockroachdb]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring]: https://pkg.go.dev/github.com/shopspring/decimal
[specification]: https://speleotrove.com/decimal/telcoSpec.html
[cross-validate]: https://github.com/govalues/decimal-tests/blob/main/decimal_fuzz_test.go
