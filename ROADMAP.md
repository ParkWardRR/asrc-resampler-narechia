# ASRC (Asynchronous Sample Rate Converter) Roadmap

## Completed (Prior History)
- [x] Implemented core polyphase FIR filtering for asynchronous resampling.
- [x] Achieved comprehensive test coverage across extreme ratio boundary conditions.
- [x] Wrote `go test -bench` routines covering the 44.1kHz -> 48kHz conversions.
- [x] Deployed GitHub Actions CI and strict static analysis rules.

## Short-term Goals
- [ ] Implement dynamic ratio tracking (smooth interpolation of ratios over time).
- [ ] Stabilize the real-time jitter buffering APIs.
- [ ] Improve documentation on latency configuration and phase linearity.

## Mid-term Goals
- [ ] Offer SIMD (ARM64 NEON / AVX2) optimized paths for the convolution loops.
- [ ] Add Minimum-Phase and Linear-Phase filter toggle configurations.
- [ ] Reduce allocations to absolute zero for audio-thread safety.

## Long-term Vision
- [ ] Become a drop-in component for high-fidelity AirPlay 2 / Cast audio receivers.
- [ ] Hardware DSP integration for extremely low-power resampling contexts.
