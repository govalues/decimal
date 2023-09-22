# decimal

[![githubb]][github]
[![codecovb]][codecov]
[![goreportb]][goreport]
[![godocb]][godoc]
[![licenseb]][license]
[![versionb]][version]

Package decimal implements immutable decimal floating-point numbers for Go.
This packages is designed specifically for use in transactional financial systems.

## Features

- **Optimized Performance**: Utilizes uint64 for coefficients, reducing heap allocations and memory consumption.
- **Immutability**: Once a decimal is set, it remains unchanged. This immutability ensures safe concurrent access across goroutines.
- **Banker's Rounding**: Methods use half even rounding, also known as "banker's rounding", which minimizes cumulative rounding errors commonly seen in financial calculations.
- **Error Handling**: All methods are designed to be panic-free. Instead of potentially crashing your application, they return errors for issues such as overflow or division by zero.
- **Simple String Representation**: Decimals are represented without the complexities of scientific or engineering notation.

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
    d := decimal.MustNew(156, 1) // d = 15.6
    e := decimal.MustParse("8")  // e = 8
    fmt.Println(d.Add(e))        // Sum
    fmt.Println(d.Sub(e))        // Difference
    fmt.Println(d.Mul(e))        // Product
    fmt.Println(d.Quo(e))        // Quotient
    fmt.Println(d.Pow(2))        // Square
    fmt.Println(d.Inv())         // Reciprocal
}
```

For detailed documentation and additional examples, visit the
[package documentation](https://pkg.go.dev/github.com/govalues/decimal#pkg-examples).
For examples related to financial calculations, see the
[money package documentation](https://pkg.go.dev/github.com/govalues/money#pkg-examples).

## Comparison

Comparison of decimal with other popular decimal packages:

| Feature          | govalues     | [cockroachdb] v3.2.0 | [shopspring] v1.3.1 |
| ---------------- | ------------ | -------------------- | ------------------- |
| Speed            | High         | Medium               | Low                 |
| Mutability       | Immutable    | Mutable              | Immutable           |
| Memory Footprint | Low          | Medium               | High                |
| Panic Free       | Yes          | Yes                  | No                  |
| Precision        | 19 digits    | Arbitrary            | Arbitrary           |
| Default Rounding | Half to even | Half up              | Half away from 0    |
| Context          | Implicit     | Explicit             | Implicit            |

decimal package was created simply because shopspring's decimal was too slow
and cockroachdb's decimal was mutable.

### Benchmarks

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
