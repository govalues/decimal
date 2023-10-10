# decimal

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]

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
- **Testing** - Fuzz testing is used to cross-validate arithmetic operations
  against both the [cockroachdb] and [shopspring] decimal packages.

## Getting Started

### Installation

To add the decimal package to your Go workspace:

```bash
go get github.com/govalues/decimal
```

### Usage

Create decimal values using constructors such as `MustNew` or `MustParse`.
After creating a decimal value, various arithmetic operations can be performed:

```go
package main

import (
    "fmt"
    "github.com/govalues/decimal"
)

func main() {
    d := decimal.MustNew(8, 0)     // d = 8
    e := decimal.MustParse("12.5") // e = 12.5
    fmt.Println(d.Add(e))          // 8 + 12.5
    fmt.Println(d.Sub(e))          // 8 - 12.5
    fmt.Println(d.Mul(e))          // 8 * 12.5
    fmt.Println(d.Quo(e))          // 8 / 12.5
    fmt.Println(d.QuoRem(e))       // 8 // 12.5 and 8 mod 12.5
    fmt.Println(d.FMA(e, e))       // 8 * 12.5 + 12.5
    fmt.Println(d.Pow(2))          // 8 ^ 2
    fmt.Println(d.Inv())           // 1 / 8
}
```

## Documentation

For detailed documentation and additional examples, visit the
[package documentation](https://pkg.go.dev/github.com/govalues/decimal#pkg-examples).
For examples related to financial calculations, see the
[money package documentation](https://pkg.go.dev/github.com/govalues/money#pkg-examples).

## Comparison

Comparison of decimal with other popular decimal packages:

| Feature          | govalues     | [cockroachdb] v3.2.0 | [shopspring] v1.3.1 |
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
pkg: github.com/govalues/benchmarks
cpu: AMD Ryzen 7 3700C  with Radeon Vega Mobile Gfx 
```

| Test Case   | Expression           | govalues | [cockroachdb] v3.2.0 | [shopspring] v1.3.1 | govalues vs cockroachdb | govalues vs shopspring |
| ----------- | -------------------- | -------: | -------------------: | ------------------: | ----------------------: | ---------------------: |
| Add         | 2 + 3                |   15.79n |               47.95n |             141.95n |                +203.64% |               +798.99% |
| Mul         | 2 * 3                |   16.61n |               54.66n |             144.95n |                +229.18% |               +772.93% |
| QuoFinite   | 2 / 4                |   64.74n |              381.15n |             645.35n |                +488.74% |               +896.83% |
| QuoInfinite | 2 / 3                |  595.30n |             1001.50n |            2810.50n |                 +68.23% |               +372.11% |
| Pow         | 1.1^60               |    1.31µ |                3.17µ |              20.50µ |                +142.42% |              +1469.53% |
| Pow         | 1.01^600             |    4.36µ |               13.86µ |              44.39µ |                +217.93% |               +918.44% |
| Pow         | 1.001^6000           |    7.39µ |               24.69µ |             656.84µ |                +234.34% |              +8793.66% |
| Parse       | 1                    |   17.27n |               78.25n |             128.80n |                +353.23% |               +646.02% |
| Parse       | 123.456              |   39.80n |              211.85n |             237.60n |                +432.22% |               +496.91% |
| Parse       | 123456789.1234567890 |  106.20n |              233.10n |             510.90n |                +119.59% |               +381.30% |
| String      | 1                    |    5.45n |               19.91n |             197.85n |                +265.49% |              +3531.94% |
| String      | 123.456              |   42.38n |               74.83n |             229.50n |                 +76.57% |               +441.53% |
| String      | 123456789.1234567890 |   77.90n |              210.40n |             328.90n |                +170.11% |               +322.24% |

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
[cockroachdb]: https://pkg.go.dev/github.com/cockroachdb/apd
[shopspring]: https://pkg.go.dev/github.com/shopspring/decimal
