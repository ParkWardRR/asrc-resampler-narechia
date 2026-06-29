package asrc

import (
	"encoding/binary"
	"io"
	"math"
)

// Reader wraps an io.Reader, resampling its byte stream (assuming interleaved float64, little-endian).
type Reader struct {
	inner     io.Reader
	resampler *ASRCResampler
	byteBuf   []byte
	floatBuf  []float64
	outBuf    []float64
	outBufIdx int
}

// NewReader creates a new Reader.
func NewReader(r io.Reader, resampler *ASRCResampler, bufferSize int) *Reader {
	return &Reader{
		inner:     r,
		resampler: resampler,
		byteBuf:   make([]byte, bufferSize*8),
		floatBuf:  make([]float64, bufferSize),
	}
}

// Read implements io.Reader.
func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	for {
		// If we have output buffered, serve it.
		if r.outBufIdx < len(r.outBuf) {
			bytesToCopy := (len(r.outBuf) - r.outBufIdx) * 8
			if bytesToCopy > len(p) {
				bytesToCopy = len(p)
			}
			samplesToCopy := bytesToCopy / 8
			
			for i := 0; i < samplesToCopy; i++ {
				bits := math.Float64bits(r.outBuf[r.outBufIdx+i])
				binary.LittleEndian.PutUint64(p[i*8:], bits)
			}

			r.outBufIdx += samplesToCopy
			return bytesToCopy, nil
		}

		// Otherwise, read more from inner and resample.
		r.outBufIdx = 0
		r.outBuf = nil

		rn, err := r.inner.Read(r.byteBuf)
		if rn > 0 {
			samples := rn / 8
			for i := 0; i < samples; i++ {
				bits := binary.LittleEndian.Uint64(r.byteBuf[i*8:])
				r.floatBuf[i] = math.Float64frombits(bits)
			}
			
			r.outBuf = r.resampler.Process(r.floatBuf[:samples])
		}

		if err != nil {
			return 0, err
		}
	}
}
