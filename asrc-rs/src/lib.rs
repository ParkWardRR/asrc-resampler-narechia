#![cfg_attr(not(feature = "std"), no_std)]

#[cfg(not(feature = "std"))]
extern crate core;

#[cfg(not(feature = "std"))]
#[macro_use]
extern crate alloc;

#[cfg(not(feature = "std"))]
use alloc::vec::Vec;
#[cfg(feature = "std")]
use std::vec::Vec;

use libm::{fabs, exp, sqrt, sin};

pub mod ffi;
pub mod simd;
pub mod gpu;
pub mod ane_tensor;
pub mod crypto;
pub mod ebpf;
pub mod wasm;

const PI: f64 = 3.14159265358979323846;

#[derive(Clone, Copy, Debug)]
#[repr(i32)]
pub enum ASRCQuality {
    Fast = 0,
    Balanced = 1,
    High = 2,
    Audiophile = 3,
}

pub struct ASRCResampler {
    num_taps: usize,
    half_taps: usize,
    filter_table: Vec<f64>,
    ratio: f64,
    phase: f64,
    history: Vec<f64>,
    hist_len: usize,
    hist_write: usize,
    channels: usize,
    oversample: usize,
    table_len: usize,
}

impl ASRCResampler {
    pub fn new(quality: ASRCQuality, channels: usize) -> Self {
        let (num_taps, beta) = match quality {
            ASRCQuality::Fast => (16, 5.0),
            ASRCQuality::Balanced => (32, 7.0),
            ASRCQuality::High => (48, 8.6),
            ASRCQuality::Audiophile => (64, 10.0),
        };

        let oversample = 256;
        let table_len = num_taps * oversample;
        let mut filter_table = Vec::with_capacity(table_len);
        let half_taps = num_taps / 2;

        for i in 0..table_len {
            let t = (i as f64) / (oversample as f64) - (half_taps as f64);
            let sinc_val = if fabs(t) < 1e-10 {
                1.0
            } else {
                sin(PI * t) / (PI * t)
            };

            let n = (i as f64) / ((table_len - 1) as f64);
            let w = kaiser_window(2.0 * n - 1.0, beta);

            filter_table.push(sinc_val * w);
        }

        let hist_len = num_taps * channels * 4;
        let history = vec![0.0; hist_len];

        Self {
            num_taps,
            half_taps,
            filter_table,
            ratio: 1.0,
            phase: 0.0,
            history,
            hist_len,
            hist_write: 0,
            channels,
            oversample,
            table_len,
        }
    }

    pub fn set_ratio(&mut self, ratio: f64) {
        self.ratio = ratio;
    }

    pub fn reset(&mut self) {
        self.phase = 0.0;
        self.hist_write = 0;
        for val in self.history.iter_mut() {
            *val = 0.0;
        }
    }

    pub fn process(&mut self, input: &[f64], output: &mut [f64]) -> usize {
        if input.is_empty() {
            return 0;
        }

        let ch = self.channels;
        let input_frames = input.len() / ch;

        // Push input into history
        for &sample in input.iter() {
            self.history[self.hist_write] = sample;
            self.hist_write = (self.hist_write + 1) % self.hist_len;
        }

        let mut out_idx = 0;

        loop {
            let read_pos = self.phase;
            if read_pos >= (input_frames as f64) {
                break;
            }
            if out_idx + ch > output.len() {
                // Not enough space in output buffer.
                // In a real application, we might want to return here and save state,
                // but for this zero-copy pipeline, we assume output is large enough.
                break;
            }

            let int_pos = read_pos as isize;
            let frac_pos = read_pos - (int_pos as f64);
            
            let mut filter_offset = (frac_pos * (self.oversample as f64)) as usize;
            if filter_offset >= self.oversample {
                filter_offset = self.oversample - 1;
            }

            // SIMD optimization dispatch can go here (see simd.rs)
            // For now, fallback to scalar if SIMD isn't available/enabled
            simd::process_frame_scalar(
                self,
                int_pos,
                filter_offset,
                input,
                input_frames as isize,
                &mut output[out_idx..out_idx + ch],
            );

            out_idx += ch;
            self.phase += self.ratio;
        }

        self.phase -= input_frames as f64;
        if self.phase < 0.0 {
            self.phase = 0.0;
        }

        out_idx / ch
    }

    #[cfg(feature = "wasm")]
    pub fn process_standalone(&mut self, input: &[f64], ratio: f64) -> Vec<f64> {
        self.set_ratio(ratio);
        let output_frames = (input.len() / self.channels) as f64 / ratio;
        let output_cap = (output_frames.ceil() as usize + 2) * self.channels;
        let mut output = vec![0.0; output_cap];
        let processed = self.process(input, &mut output);
        output.truncate(processed * self.channels);
        output
    }
}

fn kaiser_window(x: f64, beta: f64) -> f64 {
    if fabs(x) > 1.0 {
        return 0.0;
    }
    let arg = beta * sqrt(1.0 - x * x);
    bessel_i0(arg) / bessel_i0(beta)
}

fn bessel_i0(x: f64) -> f64 {
    let ax = fabs(x);
    if ax < 3.75 {
        let t = ax / 3.75;
        let t2 = t * t;
        1.0 + t2 * (3.5156229 + t2 * (3.0899424 + t2 * (1.2067492 +
            t2 * (0.2659732 + t2 * (0.0360768 + t2 * 0.0045813)))))
    } else {
        let t = 3.75 / ax;
        (exp(ax) / sqrt(ax)) * (0.39894228 + t * (0.01328592 +
            t * (0.00225319 + t * (-0.00157565 + t * (0.00916281 + t * (-0.02057706 +
                t * (0.02635537 + t * (-0.01647633 + t * 0.00392377))))))))
    }
}
