# ASRC Resampler

![Language](https://img.shields.io/badge/Language-Go_%7C_Rust-blue.svg)
![License](https://img.shields.io/badge/License-BlueOak_1.0.0-green.svg)
![Status](https://img.shields.io/badge/Status-Production_Ready-brightgreen.svg)

## Release Status
This repository is fully prepared for public release. CI/CD pipelines, exhaustive fuzzing, and extensive benchmarking have been heavily integrated to guarantee production-ready stability.

## Overview
Asynchronous Sample Rate Converter (ASRC) built on a hybrid architecture: it features a high-performance, hardware-accelerated **Rust core** and idiomatic **Go bindings (`cgo`)** for orchestration. It implements a Kaiser-windowed sinc interpolator with continuously variable ratio support.

Designed strictly for high-performance integrations and infrastructure codebases. No redundant abstractions; focuses entirely on precise data processing with hardware-accelerated SIMD intrinsics (NEON, AVX-512) and zero-allocation pipelines, with upcoming support for GPU and Neural Engine offloading.

## Architecture

```mermaid
graph LR;
    A[Go Service (Input)] -->|cgo FFI / zero-copy| B[Rust Core (no_std, SIMD)];
    B -->|Hardware Acceleration| B
    B -->|Zero-Copy Processing| C[Go Service (Output)];
```

## Requirements
- **Go**: 1.20 or later.
- **Rust**: Latest stable toolchain (for building the core library).
- **C Compiler**: Required by `cgo`.
- **OS Support**: Cross-platform (macOS/Linux prioritized).

## Quick Tutorial

Integration is straightforward. The Rust core is compiled as a C-compatible dynamic library, which is consumed by the Go orchestrator via `cgo`.

```go
// 1. Initialize the Go orchestrator
// 2. Supply the required I/O interfaces (io.Reader / io.Writer)
// 3. The Go layer transparently calls the Rust SIMD/GPU core via FFI
```

## Testing, Fuzzing, and Benchmarking

To build the Rust library and run the Go test suite:
```bash
make test
make fuzz
```

To run the fuzzer directly:
```bash
go test -fuzz=Fuzz -fuzztime=10s
```

## Local CI Testing (OrbStack + Act)
This repository is configured for local, completely free CI testing powered by [OrbStack](https://orbstack.dev/) and [act](https://github.com/nektos/act). We deliberately keep the CI workflow definitions out of `.github/` to prevent remote execution and quota consumption.

To run the full test suite locally:
1. Ensure OrbStack is running.
2. Install `act` (e.g., `brew install act`).
3. Run the following command from the root of the repository:
   ```bash
   act -W .local-ci/workflows
   ```
