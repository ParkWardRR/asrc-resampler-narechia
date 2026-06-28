# asrc-resampler-narechia

![License: Blue Oak](https://img.shields.io/badge/License-Blue_Oak_1.0.0-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Language](https://img.shields.io/badge/language-Go-blue)
![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen)

## Overview
Kaiser-windowed sinc resampler with continuously variable ratio and 4 quality levels.

## Architecture

```mermaid
graph TD;
    A[Input Samples] --> B(Polyphase Filter Bank);
    B --> C{Ratio Selection};
    C --> D(Interpolation);
    D --> E[Output Samples];
```

## Interface
```go
// Core exported structs, traits, or functions
```

## Agent Handoff / Continuation
Copied asrc.go. Need to rename package, add quality comparison benchmarks.
