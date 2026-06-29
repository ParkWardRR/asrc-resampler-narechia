use crate::{ASRCQuality, ASRCResampler};

#[cfg(not(feature = "std"))]
use core::slice;
#[cfg(feature = "std")]
use std::slice;

#[cfg(not(feature = "std"))]
use alloc::boxed::Box;
#[cfg(feature = "std")]
use std::boxed::Box;

#[no_mangle]
pub extern "C" fn asrc_create(quality: i32, channels: i32) -> *mut ASRCResampler {
    let q = match quality {
        0 => ASRCQuality::Fast,
        1 => ASRCQuality::Balanced,
        2 => ASRCQuality::High,
        3 => ASRCQuality::Audiophile,
        _ => ASRCQuality::Balanced,
    };
    
    let resampler = ASRCResampler::new(q, channels as usize);
    let b = Box::new(resampler);
    Box::into_raw(b)
}

#[no_mangle]
pub extern "C" fn asrc_set_ratio(resampler: *mut ASRCResampler, ratio: f64) {
    if let Some(r) = unsafe { resampler.as_mut() } {
        r.set_ratio(ratio);
    }
}

#[no_mangle]
pub extern "C" fn asrc_reset(resampler: *mut ASRCResampler) {
    if let Some(r) = unsafe { resampler.as_mut() } {
        r.reset();
    }
}

#[no_mangle]
pub extern "C" fn asrc_process(
    resampler: *mut ASRCResampler,
    input_ptr: *const f64,
    input_len: usize,
    output_ptr: *mut f64,
    output_cap: usize,
) -> usize {
    let r = match unsafe { resampler.as_mut() } {
        Some(r) => r,
        None => return 0,
    };

    let input = unsafe { slice::from_raw_parts(input_ptr, input_len) };
    let output = unsafe { slice::from_raw_parts_mut(output_ptr, output_cap) };

    r.process(input, output)
}

#[no_mangle]
pub extern "C" fn asrc_destroy(resampler: *mut ASRCResampler) {
    if !resampler.is_null() {
        unsafe {
            let _ = Box::from_raw(resampler);
        }
    }
}
