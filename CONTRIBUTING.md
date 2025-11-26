# Contributing to ML-DSA

Thank you for your interest in contributing to Trail of Bits' ML-DSA implementation! This document provides guidelines for contributing code, running tests, and ensuring your changes meet our security requirements.

## Prerequisites

- Go 1.24 or later
- `golangci-lint` ([installation instructions](https://golangci-lint.run/usage/install/))

## Development Workflow

### 1. Clone and Setup

```bash
git clone https://github.com/trailofbits/ml-dsa.git
cd ml-dsa
go mod download
```

### 2. Make Your Changes

Before writing code, please read [`.github/CLAUDE.md`](.github/CLAUDE.md) which contains critical security requirements for constant-time implementations.

### 3. Run Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run short tests only (faster)
go test -short ./...
```

### 4. Run the Linter

**All code must pass `golangci-lint` before being merged.**

```bash
golangci-lint run ./...
```

Common issues:
- **errcheck**: Handle or explicitly ignore error returns with `_ = func()`
- **unused**: Remove unused variables and imports
- **ineffassign**: Don't assign to variables that are never read

### 5. Run Constant-Time Verification

The `test_ct` tool analyzes compiled assembly to detect instructions that could leak timing information.

#### Basic Usage

```bash
# Analyze all code for the current architecture
go run ./test_ct

# Analyze for a specific target architecture
go run ./test_ct -arch=amd64 -os=linux
go run ./test_ct -arch=arm64 -os=linux
```

#### Analyzing Specific Functions

Use the `-func` flag with a regex pattern to analyze specific functions:

```bash
# Analyze all field arithmetic functions
go run ./test_ct -arch=amd64 -func='field\.'

# Analyze a specific function by name
go run ./test_ct -arch=amd64 -func='DivBarrett'

# Analyze multiple related functions
go run ./test_ct -arch=amd64 -func='DivBarrett|DivConstTime|Decompose'

# Analyze functions in a specific package
go run ./test_ct -arch=amd64 -func='internal/field'
```

#### Understanding the Output

```
Constant-Time Analysis Report
==============================

Architecture: x86-64 (AMD64) (amd64)
Functions analyzed: 6
Instructions analyzed: 273

No violations found.

Summary
-------
Errors:   0
Warnings: 0

PASSED
```

If violations are found:

```
Function: example.UnsafeDiv
  File: /path/to/file.go

  [ERROR] 0x46fb8b: DIVQ BX
         DIVQ has data-dependent timing; execution time varies based on operand values

Summary
-------
Errors:   1
Warnings: 0

FAILED
```

#### Additional Options

```bash
# Include warnings (e.g., conditional branches)
go run ./test_ct -arch=amd64 -warnings

# JSON output for CI integration
go run ./test_ct -arch=amd64 -json

# Verbose output showing build and disassembly steps
go run ./test_ct -arch=amd64 -verbose

# Keep the compiled binary for manual inspection
go run ./test_ct -arch=amd64 -keep-binary

# List all supported architectures
go run ./test_ct -list-arch
```

#### Supported Architectures

| GOARCH | Description |
|--------|-------------|
| `amd64` | x86-64 (Intel/AMD 64-bit) |
| `386` | x86 (Intel/AMD 32-bit) |
| `arm64` | AArch64 (Apple Silicon, ARM servers) |
| `arm` | ARM 32-bit |
| `ppc64le` | PowerPC 64-bit Little Endian |
| `riscv64` | RISC-V 64-bit |
| `s390x` | IBM z/Architecture |

### 6. Submit Your Changes

1. Ensure all tests pass: `go test ./...`
2. Ensure linter passes: `golangci-lint run ./...`
3. Run constant-time verification on critical functions
4. Create a pull request with a clear description of your changes

## Security Requirements

This is a cryptographic library. **All code must be constant-time with respect to secret data.**

### What This Means

- Execution time must not depend on secret values
- No branching (`if`/`switch`) based on secret data
- No division (`/`, `%`) on secret data (use Barrett reduction)
- No array indexing with secret indices
- No early returns based on secret comparisons

### Dangerous Instructions

The `test_ct` tool detects these variable-time instructions:

| Architecture | Dangerous Instructions |
|--------------|----------------------|
| x86/amd64 | `DIV`, `IDIV`, `DIVSS`, `DIVSD`, `SQRT*` |
| arm64 | `UDIV`, `SDIV`, `FDIV`, `FSQRT` |
| arm | `UDIV`, `SDIV`, `MUL*` (on older cores) |
| ppc64 | `divw`, `divd`, `fdiv`, `fsqrt` |
| riscv64 | `div`, `rem`, `fdiv`, `fsqrt` |

### Safe Alternatives

Instead of division, use:
- **Barrett reduction** for division by known constants (see `DivBarrett` in `internal/field/field.go`)
- **Bit-by-bit division** for general constant-time division (see `DivConstTime32`)

Instead of branching, use:
- **Bit masking** for conditional selection
- **`crypto/subtle`** functions for comparisons

See [`.github/CLAUDE.md`](.github/CLAUDE.md) for detailed examples and patterns.

## Code Style

- Follow standard Go conventions (`gofmt`, `goimports`)
- Use meaningful variable names
- Add comments for non-obvious security-critical code
- Document why something is constant-time if it's not obvious

## Questions?

- Open an issue for questions or discussions
- See [`.github/CLAUDE.md`](.github/CLAUDE.md) for security guidelines
- Reference the [FIPS 204 specification](https://csrc.nist.gov/pubs/fips/204/final) for algorithm details
