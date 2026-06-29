#[cfg(feature = "wasm")]
use wasm_bindgen::prelude::*;

#[cfg(feature = "wasm")]
#[wasm_bindgen]
pub fn resample_web(input: &[f64], ratio: f64) -> Vec<f64> {
    // Wrapper for WebWorkers
    crate::ASRCResampler::new(crate::ASRCQuality::Fast, 2).process_standalone(input, ratio)
}
