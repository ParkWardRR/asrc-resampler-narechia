use crate::ASRCResampler;

pub fn process_frame_scalar(
    resampler: &ASRCResampler,
    int_pos: isize,
    filter_offset: usize,
    input: &[f64],
    input_frames: isize,
    output: &mut [f64],
) {
    let ch = resampler.channels;
    let half_taps = resampler.half_taps as isize;

    for c in 0..ch {
        let mut sum = 0.0;

        for tap in -half_taps..half_taps {
            let input_idx = int_pos + tap;
            if input_idx < 0 || input_idx >= input_frames {
                continue;
            }

            let table_idx = ((tap + half_taps) as usize) * resampler.oversample + filter_offset;
            if table_idx >= resampler.table_len {
                continue;
            }

            sum += input[(input_idx as usize) * ch + c] * resampler.filter_table[table_idx];
        }

        output[c] = sum;
    }
}

// Future implementations for NEON and AVX-512 can be added here
// #[cfg(all(target_arch = "aarch64", target_feature = "neon"))]
// pub fn process_frame_neon(...)
//
// #[cfg(all(target_arch = "x86_64", target_feature = "avx512f"))]
// pub fn process_frame_avx512(...)
