#ifndef WASMVM_SERVICE_SDK_COMMON_H
#define WASMVM_SERVICE_SDK_COMMON_H
#include "sdk.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static uint64_t sdk_now(void) {
    struct timespec value;
    if (clock_gettime(CLOCK_MONOTONIC, &value)) abort();
    return (uint64_t)value.tv_sec * UINT64_C(1000000000) + value.tv_nsec;
}
static void sdk_message(char *out, const char *data, size_t size) {
    size_t n = size < 511 ? size : 511;
    if (n) memcpy(out, data, n);
    out[n] = 0;
}
#define SDK_FAILURE (UINT64_C(1) << 32)
#endif
