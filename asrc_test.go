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
	
	// Create resampler (1 channel for test)
	r := NewASRCResampler(ASRCHigh, 1)
	if r == nil {
		t.Fatalf("Failed to create resampler")
	}
	
	// Generate 1 second of 440 Hz sine wave
	input := generateSineWave(440.0, inRate, 1.0)
	
	// Set ratio
	r.SetRatio(inRate / outRate) // 48000 / 44100 = 1.088
	
	expectedLen := int(math.Ceil(float64(len(input)) / r.ratio))
	
	output := r.Process(input)
	if len(output) == 0 {
		t.Fatalf("Process failed, got empty output")
	}
	
	if math.Abs(float64(len(output) - expectedLen)) > 100 {
		t.Errorf("Expected length ~%d, got %d", expectedLen, len(output))
	}
}

func TestResampler_Upsample(t *testing.T) {
	inRate := 44100.0
	outRate := 96000.0
	
	r := NewASRCResampler(ASRCBalanced, 1)
	if r == nil {
		t.Fatalf("Failed to create resampler")
	}
	
	input := generateSineWave(1000.0, inRate, 0.5)
	
	r.SetRatio(inRate / outRate) 
	
	expectedLen := int(math.Ceil(float64(len(input)) / r.ratio))
	
	output := r.Process(input)
	if len(output) == 0 {
		t.Fatalf("Process failed, got empty output")
	}
	
	if math.Abs(float64(len(output) - expectedLen)) > 100 {
		t.Errorf("Expected length ~%d, got %d", expectedLen, len(output))
	}
}
