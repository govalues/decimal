# Changelog

## [0.1.29] - 2024-06-29

### Changed

- Improved `Decimal.Sqrt` and `Decimal.QuoRem` performance.

## [0.1.28] - 2024-06-22

### Added

- Implemented `Decimal.Sqrt`.

## [0.1.27] - 2024-05-19

### Changed

- `Decimal.Pad`, `Decimal.Rescale`, and `Descimal.Quantize` methods
  do not return errors anymore.

## [0.1.25] - 2024-05-17

### Added

- Implemented binary marshaling.

## [0.1.24] - 2024-05-05

### Changed

- Bumped go version to 1.21.
- Improved documentation.

## [0.1.23] - 2024-03-04

### Changed

- Improved documentation.

## [0.1.22] - 2024-01-11

### Changed

- Supported MySQL in `Decimal.Scan` method.
- Added examples for XML marshaling.

## [0.1.21] - 2024-01-05

### Changed

- Optimized parsing performance for long strings.
- Improved documentation.

## [0.1.20] - 2024-01-01

### Changed

- Eliminated heap allocations in big.Int arithmetic.
- Improved documentation.

## [0.1.19] - 2023-12-18

### Changed

- Improved table formatting in documentation.

## [0.1.18] - 2023-12-18

### Changed

- Improved examples and documentation.

## [0.1.17] - 2023-12-01

### Added

- Implemented `Decimal.SameScale` method.

### Changed

- Improved examples and documentation.

## [0.1.16] - 2023-11-21

### Changed

- Improved examples and documentation.
- Improved test coverage.

## [0.1.15] - 2023-10-31

### Changed

- Improved examples and documentation.

## [0.1.14] - 2023-10-13

### Changed

- Improved examples and documentation.

## [0.1.13] - 2023-10-10

### Added

- Implemented `NullDecimal` type.

## [0.1.12] - 2023-10-01

### Changed

- Improved accuracy of `Decimal.Pow` method for negative powers.
- Reviewed and improved documentation.

## [0.1.11] - 2023-09-21

### Added

- Implemented `Decimal.Clamp` method.

### Changed

- Reviewed and improved documentation.

## [0.1.10] - 2023-09-09

### Added

- Implemented `Decimal.SubAbs`, `Decimal.CmpAbs`, `Decimal.Inv`.
- Added `Decimal.Pi`, `Decimal.E`, `Decimal.NegOne`, `Decimal.Two`, `Decimal.Thousand`.

### Changed

- Reviewed descriptions of rounding methods.

## [0.1.9] - 2023-08-27

### Changed

- Reviewed error descriptions.

## [0.1.8] - 2023-08-23

### Changed

- Improved accuracy of `Decimal.Pow`.

## [0.1.7] - 2023-08-20

### Changed

- Enabled `gocyclo` linter.

## [0.1.6] - 2023-08-19

### Added

- Implemented `Decimal.Scan` and `Decimal.Value`.

### Changed

- `Decimal.CopySign` treats 0 as a positive.
- Enabled `gosec`, `godot`, and `stylecheck` linters.

## [0.1.5] - 2023-08-12

### Added

- Implemented `NewFromFloat64` method.
- Added fuzzing job to continuous integration.

### Changed

- `NewFromInt64` can round to nearest if coefficient is too large.

## [0.1.4] - 2023-08-04

### Changed

- Implemented `NewFromInt64` method.

## [0.1.3] - 2023-08-03

### Changed

- Implemented scale argument for `Decimal.Int64` method.

## [0.1.2] - 2023-06-17

### Changed

- `Rescale`, `ParseExact`, `MulExact`, `AddExact`, `FMAExact`, and `QuoExact` methods
  return error if scale is out of range.

## [0.1.1] - 2023-06-10

### Changed

- `New` method returns error if scale is out of range.

## [0.1.0] - 2023-06-03

### Changed

- All methods now return errors, instead of panicking.
- Implemented `Decimal.Pad` method.
- Implemented `Decimal.PowExact` method.
- Renamed `Decimal.Round` to `Decimal.Rescale`.
- Renamed `Decimal.Reduce` to `Decimal.Trim`.

## [0.0.13] - 2023-04-22

### Fixed

- Testing on Windows.

## [0.0.12] - 2023-04-21

### Changed

- Testing on Windows and macOS.
- Improved documentation.

## [0.0.11] - 2023-04-15

### Added

- Implemented `Decimal.Int64` method.
- Implemented `Decimal.Float64` method.

### Changed

- Reviewed and improved documentation.

## [0.0.10] - 2023-04-13

### Changed

- Reviewed and improved documentation.
- Improved continuous integration.

## [0.0.9] - 2023-04-05

### Added

- Implemented `Decimal.One` method.
- Implemented `Decimal.Zero` method.

### Changed

- Reduced memory consumption.
- Renamed `Decimal.LessThanOne` to `Decimal.WithinOne`.

### Deleted

- Removed `Decimal.WithScale`.

## [0.0.8] - 2023-03-25

### Changed

- Simplified `Decimal.Quo` method.

## [0.0.7] - 2023-03-22

### Added

- Implemented `Decimal.CopySign` method.

## [0.0.6] - 2023-03-21

### Added

- Implemented `Decimal.ULP` method.

## [0.0.5] - 2023-03-19

### Added

- Polish notation calculator example.
- Benchmarks statistics.

## [0.0.4] - 2023-03-19

### Fixed

- Fixed index out of range in `Parse`.
- Rounding error in `Decimal.Quo`.

## [0.0.3] - 2023-03-18

### Changed

- Removed errors from public API.
- Renamed `Decimal.Fma` to `Decimal.FMA`.

## [0.0.2] - 2023-03-13

### Added

- Implemented `Decimal.Fma`.

## [0.0.1] - 2023-02-28

### Added

- Initial version.
