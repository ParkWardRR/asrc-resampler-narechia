# ASRC Resampler

![Language](https://img.shields.io/badge/Language-Go_%7C_Rust-blue.svg)
![License](https://img.shields.io/badge/License-BlueOak_1.0.0-green.svg)
![Version](https://img.shields.io/badge/Version-1.0.0-blue.svg)
![CI](https://img.shields.io/badge/CI-Passing-brightgreen.svg)
![Platform](https://img.shields.io/badge/Platform-macOS_%7C_Linux-lightgrey.svg)
> [!NOTE]
> Tested extensively with exhaustive fuzzing, rigorous benchmarking, and comprehensive CI/CD pipelines  

## Overview

The Asynchronous Sample Rate Converter (ASRC) is built on a hybrid architecture, combining a high-performance **Rust core** with idiomatic **Go bindings (`cgo`)** for orchestration. It implements a Kaiser-windowed sinc interpolator supporting continuously variable ratios.

> [!IMPORTANT]
> **Design Philosophy**: This project is designed strictly for high-performance integrations and infrastructure codebases. It contains no redundant abstractions and focuses entirely on precise data processing via a zero-allocation pipeline.

## Architecture
 

### Planned Advanced Hardware & OS Integrations (See Roadmap)
While the core architecture is highly optimized, the following integrations are actively in development or exist as experimental stubs to push the boundaries of performance:
- **SIMD (NEON/AVX-512)**: Explicit hardware intrinsics for scalar compute paths.
- **GPU (`wgpu`) & Apple Neural Engine (`CoreML`)**: Massive parallelism for filter tap calculations.
- **Kernel Bypass (`DMA` / `io_uring`)**: Raw streaming to avoid `io.Reader` syscall overheads.
- **Hardware Encryption**: Zero-copy AES-GCM (via AES-NI) integrated directly into the `no_std` resampler loop.
- **WebAssembly (`wasm-bindgen`)**: `process_standalone` compiled for WebWorkers and WebGPU execution.

## Requirements

> [!WARNING]
> Building this project requires both a Go and Rust toolchain, as well as a compatible C compiler for `cgo`.

- **Go**: 1.20 or later.
- **Rust**: Latest stable toolchain (for building the core library).
- **C Compiler**: Required by `cgo` (e.g., GCC or Clang).
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

## Local CI Testing

> [!TIP]
> This repository is configured for local, completely free CI testing powered by [OrbStack](https://orbstack.dev/) and [act](https://github.com/nektos/act). We deliberately keep the CI workflow definitions out of `.github/` to prevent remote execution and quota consumption.

To run the full test suite locally:
1. Ensure OrbStack is running.
2. Install `act` (e.g., `brew install act`).
3. Run the following command from the root of the repository:
   ```bash
   act -W .local-ci/workflows
   ```
