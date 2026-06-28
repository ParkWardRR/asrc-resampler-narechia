package asrc

import (
	"math"
	"testing"
)

func generateSineWave(freq, sampleRate float64, durationSec float64) []float64 {
	numSamples := int(sampleRate * durationSec)
	samples := make([]float64, numSamples)
	for i := range samples {
		t := float64(i) / sampleRate
		samples[i] = math.Sin(2.0 * math.Pi * freq * t)
	}
	return samples
}

func TestResampler_Downsample(t *testing.T) {
	inRate := 48000.0
	outRate := 44100.0
	
	// Create resampler
	r := NewResampler(inRate, outRate, QualityHigh)
	if r == nil {
		t.Fatalf("Failed to create resampler")
	}
	
	// Generate 1 second of 440 Hz sine wave
	input := generateSineWave(440.0, inRate, 1.0)
	
	// The output should be roughly (44100/48000) * input length
	expectedLen := int(math.Ceil(float64(len(input)) * outRate / inRate))
	
	output, err := r.ProcessF64(input)
	if err != nil {
		t.Fatalf("ProcessF64 failed: %v", err)
	}
	
	// Allow slight padding/delay from the polyphase filter bank
	if math.Abs(float64(len(output) - expectedLen)) > 100 {
		t.Errorf("Expected length ~%d, got %d", expectedLen, len(output))
	}
}

func TestResampler_Upsample(t *testing.T) {
	inRate := 44100.0
	outRate := 96000.0
	
	r := NewResampler(inRate, outRate, QualityMedium)
	if r == nil {
		t.Fatalf("Failed to create resampler")
	}
	
	input := generateSineWave(1000.0, inRate, 0.5)
	
	expectedLen := int(math.Ceil(float64(len(input)) * outRate / inRate))
	
	output, err := r.ProcessF64(input)
	if err != nil {
		t.Fatalf("ProcessF64 failed: %v", err)
	}
	
	if math.Abs(float64(len(output) - expectedLen)) > 100 {
		t.Errorf("Expected length ~%d, got %d", expectedLen, len(output))
	}
}
