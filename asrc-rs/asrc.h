#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct ASRCResampler ASRCResampler;

struct ASRCResampler *asrc_create(int32_t quality, int32_t channels);

void asrc_set_ratio(struct ASRCResampler *resampler, double ratio);

void asrc_reset(struct ASRCResampler *resampler);

uintptr_t asrc_process(struct ASRCResampler *resampler,
                       const double *input_ptr,
                       uintptr_t input_len,
                       double *output_ptr,
                       uintptr_t output_cap);

void asrc_destroy(struct ASRCResampler *resampler);
