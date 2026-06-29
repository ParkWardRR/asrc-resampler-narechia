package asrc

/*
#cgo LDFLAGS: -L${SRCDIR}/asrc-rs/target/release -lasrc
#cgo osx LDFLAGS: -framework CoreFoundation -framework Security
#include <stdlib.h>
#include "asrc-rs/asrc.h"
*/
import "C"
import (
	"math"
	"runtime"
	"sync"
	"unsafe"
)

// ASRCQuality controls the resampler filter quality.
type ASRCQuality int

const (
	ASRCFast       ASRCQuality = 0 // 16 taps
	ASRCBalanced   ASRCQuality = 1 // 32 taps
	ASRCHigh       ASRCQuality = 2 // 48 taps
	ASRCAudiophile ASRCQuality = 3 // 64 taps
)

// ASRCResampler performs asynchronous sample rate conversion with variable ratio.
// Backed by a high-performance Rust CGO implementation.
type ASRCResampler struct {
	mu       sync.Mutex
	inner    *C.ASRCResampler
	channels int
	ratio    float64
}

// NewASRCResampler creates an ASRC resampler.
// channels is the number of interleaved channels (typically 2 for stereo).
func NewASRCResampler(quality ASRCQuality, channels int) *ASRCResampler {
	inner := C.asrc_create(C.int32_t(quality), C.int32_t(channels))
	if inner == nil {
		panic("Failed to create ASRCResampler (Rust OOM or panic)")
	}
	r := &ASRCResampler{
		inner:    inner,
		channels: channels,
		ratio:    1.0,
	}
	runtime.SetFinalizer(r, (*ASRCResampler).Close)
	return r
}

// Close explicitly destroys the underlying Rust resources.
// It is also called automatically by the garbage collector.
func (r *ASRCResampler) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.inner != nil {
		C.asrc_destroy(r.inner)
		r.inner = nil
	}
}

// SetRatio updates the resampling ratio. ratio = 1.0 + driftPPM * 1e-6.
func (r *ASRCResampler) SetRatio(ratio float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ratio = ratio
	if r.inner != nil {
		C.asrc_set_ratio(r.inner, C.double(ratio))
	}
}

// SetDriftPPM is a convenience: sets ratio from drift estimator output.
func (r *ASRCResampler) SetDriftPPM(ppm float64) {
	r.SetRatio(1.0 + ppm*1e-6)
}

// Process resamples interleaved float64 audio.
// Input: normalized float64 samples (interleaved stereo).
// Output: resampled samples.
func (r *ASRCResampler) Process(input []float64) []float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.inner == nil || len(input) == 0 {
		return nil
	}

	inputFrames := len(input) / r.channels

	// Estimate output size to preallocate buffer
	// Add some margin just in case.
	outputFrames := int(math.Ceil(float64(inputFrames)/r.ratio)) + 2
	outputCap := outputFrames * r.channels
	output := make([]float64, outputCap)

	inputPtr := (*C.double)(unsafe.Pointer(&input[0]))
	outputPtr := (*C.double)(unsafe.Pointer(&output[0]))

	framesProcessed := C.asrc_process(
		r.inner,
		inputPtr,
		C.uintptr_t(len(input)),
		outputPtr,
		C.uintptr_t(outputCap),
	)

	samplesProcessed := int(framesProcessed) * r.channels
	
	// Safe bounds check
	if samplesProcessed > outputCap {
		samplesProcessed = outputCap
	}

	return output[:samplesProcessed]
}

// Reset clears the resampler state (call at track boundaries).
func (r *ASRCResampler) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.inner != nil {
		C.asrc_reset(r.inner)
	}
}
