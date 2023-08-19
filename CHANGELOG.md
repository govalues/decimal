# Changelog

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
