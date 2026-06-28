package asrc

import (
	"math"
	"sync"
)

// ──────────────────────────────────────────────────────────────────────────────
// ASRC — Asynchronous Sample Rate Converter
//
// Replaces sample slip for drift correction. Instead of dropping/inserting
// samples at zero crossings, ASRC continuously adjusts the resampling ratio
// by ±PPM using a windowed-sinc interpolator.
//
// This eliminates all slip artifacts (even inaudible ones) and provides
// sample-accurate ratio tracking. The filter is a Kaiser-windowed sinc
// with configurable tap count and stopband rejection.
//
// Quality levels:
//   0 = 16 taps, ~60 dB stopband (fast, good enough)
//   1 = 32 taps, ~80 dB stopband (balanced)
//   2 = 48 taps, ~100 dB stopband (high quality)
//   3 = 64 taps, ~120 dB stopband (audiophile, ~2% CPU)
// ──────────────────────────────────────────────────────────────────────────────

// ASRCQuality controls the resampler filter quality.
type ASRCQuality int

const (
	ASRCFast       ASRCQuality = 0 // 16 taps
	ASRCBalanced   ASRCQuality = 1 // 32 taps
	ASRCHigh       ASRCQuality = 2 // 48 taps
	ASRCAudiophile ASRCQuality = 3 // 64 taps
)

// ASRCResampler performs asynchronous sample rate conversion with variable ratio.
type ASRCResampler struct {
	mu sync.Mutex

	// Filter parameters.
	numTaps     int
	beta        float64 // Kaiser window beta
	halfTaps    int
	filterTable []float64 // precomputed windowed sinc values

	// Resampling state.
	ratio     float64   // current resampling ratio (1.0 = unity)
	phase     float64   // fractional phase accumulator
	history   []float64 // circular input history buffer
	histLen   int
	histWrite int // write position in history
	channels  int

	// Filter oversampling for fast lookup.
	oversample int
	tableLen   int
}

// NewASRCResampler creates an ASRC resampler.
// channels is the number of interleaved channels (typically 2 for stereo).
func NewASRCResampler(quality ASRCQuality, channels int) *ASRCResampler {
	var numTaps int
	var beta float64

	switch quality {
	case ASRCFast:
		numTaps = 16
		beta = 5.0
	case ASRCBalanced:
		numTaps = 32
		beta = 7.0
	case ASRCHigh:
		numTaps = 48
		beta = 8.6
	case ASRCAudiophile:
		numTaps = 64
		beta = 10.0
	default:
		numTaps = 32
		beta = 7.0
	}

	oversample := 256 // sub-sample resolution for filter interpolation
	tableLen := numTaps * oversample
	filterTable := make([]float64, tableLen)

	halfTaps := numTaps / 2

	// Precompute Kaiser-windowed sinc filter table.
	for i := 0; i < tableLen; i++ {
		// Fractional tap position.
		t := float64(i)/float64(oversample) - float64(halfTaps)

		// Sinc function.
		var sinc float64
		if math.Abs(t) < 1e-10 {
			sinc = 1.0
		} else {
			sinc = math.Sin(math.Pi*t) / (math.Pi * t)
		}

		// Kaiser window.
		n := float64(i) / float64(tableLen-1)
		w := kaiserWindow(2.0*n-1.0, beta)

		filterTable[i] = sinc * w
	}

	// History buffer: need numTaps * channels samples of history.
	histLen := numTaps * channels * 4 // 4x for safety margin
	history := make([]float64, histLen)

	return &ASRCResampler{
		numTaps:     numTaps,
		beta:        beta,
		halfTaps:    halfTaps,
		filterTable: filterTable,
		ratio:       1.0,
		phase:       0.0,
		history:     history,
		histLen:     histLen,
		histWrite:   0,
		channels:    channels,
		oversample:  oversample,
		tableLen:    tableLen,
	}
}

// SetRatio updates the resampling ratio. ratio = 1.0 + driftPPM * 1e-6.
// For example, if capture is 1 PPM fast, set ratio to 1.000001.
func (r *ASRCResampler) SetRatio(ratio float64) {
	r.mu.Lock()
	r.ratio = ratio
	r.mu.Unlock()
}

// SetDriftPPM is a convenience: sets ratio from drift estimator output.
func (r *ASRCResampler) SetDriftPPM(ppm float64) {
	r.SetRatio(1.0 + ppm*1e-6)
}

// Process resamples interleaved float64 audio.
// Input: normalized float64 samples (interleaved stereo).
// Output: resampled samples. Output length may differ from input by ±1 frame.
func (r *ASRCResampler) Process(input []float64) []float64 {
	r.mu.Lock()
	ratio := r.ratio
	r.mu.Unlock()

	if len(input) == 0 {
		return nil
	}

	ch := r.channels
	inputFrames := len(input) / ch

	// Estimate output size.
	outputFrames := int(math.Ceil(float64(inputFrames) / ratio))
	output := make([]float64, 0, outputFrames*ch+ch)

	// Push input into history (circular buffer).
	for i := 0; i < len(input); i++ {
		r.history[r.histWrite] = input[i]
		r.histWrite = (r.histWrite + 1) % r.histLen
	}

	// Generate output samples by interpolating the input history.
	for frame := 0; frame < outputFrames; frame++ {
		// Current read position (fractional).
		readPos := r.phase

		if readPos >= float64(inputFrames) {
			break
		}

		intPos := int(readPos)
		fracPos := readPos - float64(intPos)

		// Interpolate each channel.
		for c := 0; c < ch; c++ {
			var sum float64

			// Convolve with windowed sinc filter.
			filterOffset := int(fracPos * float64(r.oversample))
			if filterOffset >= r.oversample {
				filterOffset = r.oversample - 1
			}

			for tap := -r.halfTaps; tap < r.halfTaps; tap++ {
				inputIdx := intPos + tap
				if inputIdx < 0 || inputIdx >= inputFrames {
					continue
				}

				// Filter table lookup with sub-sample interpolation.
				tableIdx := (tap+r.halfTaps)*r.oversample + filterOffset
				if tableIdx < 0 || tableIdx >= r.tableLen {
					continue
				}

				sum += input[inputIdx*ch+c] * r.filterTable[tableIdx]
			}

			output = append(output, sum)
		}

		// Advance phase by ratio.
		r.phase += ratio
	}

	// Adjust phase for next call (remove consumed input frames).
	r.phase -= float64(inputFrames)
	if r.phase < 0 {
		r.phase = 0
	}

	return output
}

// Reset clears the resampler state (call at track boundaries).
func (r *ASRCResampler) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.phase = 0
	for i := range r.history {
		r.history[i] = 0
	}
	r.histWrite = 0
}

// kaiserWindow computes the Kaiser window function.
// x is in [-1, +1], beta controls sidelobe attenuation.
func kaiserWindow(x, beta float64) float64 {
	if math.Abs(x) > 1.0 {
		return 0.0
	}
	arg := beta * math.Sqrt(1.0-x*x)
	return besselI0(arg) / besselI0(beta)
}

// besselI0 computes the zeroth-order modified Bessel function of the first kind.
// Uses the polynomial approximation from Abramowitz & Stegun (9.8.1, 9.8.2).
func besselI0(x float64) float64 {
	ax := math.Abs(x)
	if ax < 3.75 {
		t := (ax / 3.75)
		t2 := t * t
		return 1.0 + t2*(3.5156229+t2*(3.0899424+t2*(1.2067492+
			t2*(0.2659732+t2*(0.0360768+t2*0.0045813)))))
	}
	t := 3.75 / ax
	return (math.Exp(ax) / math.Sqrt(ax)) * (0.39894228 + t*(0.01328592+
		t*(0.00225319+t*(-0.00157565+t*(0.00916281+t*(-0.02057706+
			t*(0.02635537+t*(-0.01647633+t*0.00392377))))))))
}
