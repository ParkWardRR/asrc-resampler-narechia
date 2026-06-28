# ASRC Resampler

![Language](https://img.shields.io/badge/Language-Go-blue.svg)
![License](https://img.shields.io/badge/License-BlueOak_1.0.0-green.svg)
![Status](https://img.shields.io/badge/Status-Production_Ready-brightgreen.svg)
![CI](https://github.com/USER/asrc-resampler-narechia/actions/workflows/ci.yml/badge.svg)

## Overview
Asynchronous Sample Rate Converter in Go. Implements a Kaiser-windowed sinc interpolator with continuously variable ratio support.

Designed strictly for high-performance integrations and infrastructure codebases. No redundant abstractions; focuses entirely on precise data processing.

## Architecture

```mermaid
graph LR;
    A[44.1kHz] --> B[Polyphase Filter Bank];
    B --> C[48.0kHz];

```

## Requirements
- **Go**: Latest stable toolchain.
- **OS Support**: Cross-platform (macOS/Linux prioritized).
- **Dependencies**: Minimal to none (strictly constrained to standard library where mathematically possible).

## Quick Tutorial

Integration is straightforward. Consult the module source for exact API signatures.

```go
// 1. Initialize the primary component
// 2. Supply the required I/O interfaces or buffers
// 3. Execute the processing loop or listener
```
## Testing, Fuzzing, and Benchmarking

To run the test suite and benchmarks:
```bash
go test -v ./...
go test -bench .
```

To run the fuzzer:
```bash
go test -fuzz=Fuzz -fuzztime=10s
```
