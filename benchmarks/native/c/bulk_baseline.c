// Native libc reference for runtime-fixtures/bulk-overlap.wat. Compile with
// -O3 -fno-builtin-memset -fno-builtin-memmove to retain the complete memory work.
// This measures the unchecked C loop separately from validated Wasm engines.
#if defined(__APPLE__) && !defined(_DARWIN_C_SOURCE)
#define _DARWIN_C_SOURCE
#endif

#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static uint64_t now_ns(void) {
    struct timespec ts;
#ifdef __APPLE__
    const clockid_t clock = CLOCK_MONOTONIC_RAW;
#else
    const clockid_t clock = CLOCK_MONOTONIC;
#endif
    if (clock_gettime(clock, &ts) != 0) abort();
    return (uint64_t)ts.tv_sec * UINT64_C(1000000000) + ts.tv_nsec;
}

static uint32_t run(unsigned char *memory, uint32_t seed, uint32_t count) {
    uint32_t x = seed;
    for (uint32_t i = 0; i < count; i++) {
        memset(memory + 256, (unsigned char)(x + i), 8192);
        memmove(memory + 272, memory + 256, 8192);
        x += memory[287];
    }
    return x;
}

int main(int argc, char **argv) {
    if (argc != 4) return 2;
    uint32_t seed = (uint32_t)strtoul(argv[1], NULL, 0);
    uint32_t count = (uint32_t)strtoul(argv[2], NULL, 0);
    uint32_t expected = (uint32_t)strtoul(argv[3], NULL, 0);
    int interleave = getenv("WASMVM_INTERLEAVE") != NULL;
    char token[16];
    if (interleave) {
        puts("{\"ready\":true}");
        fflush(stdout);
        if (!fgets(token, sizeof(token), stdin)) return 2;
    }
    unsigned char *memory = calloc(1, 65536);
    if (!memory || !count) return 2;
    uint64_t samples[7], first = 0;
    // First call, two warmups and seven measured calls all validate output.
    for (int sample = -3; sample < 7; sample++) {
        if (sample >= 0 && interleave) {
            printf("{\"sampleReady\":%d}\n", sample);
            fflush(stdout);
            if (!fgets(token, sizeof(token), stdin)) return 2;
        }
        uint64_t start = now_ns();
        uint32_t actual = run(memory, seed, count);
        uint64_t elapsed = now_ns() - start;
        if (actual != expected) {
            fprintf(stderr, "checksum: got %" PRIu32 ", expected %" PRIu32 "\n", actual, expected);
            free(memory);
            return 1;
        }
        if (sample == -3) first = elapsed;
        if (sample >= 0) samples[sample] = elapsed;
    }
    printf("{\"seed\":%" PRIu32 ",\"iterations\":%" PRIu32 ",\"checksum\":%" PRIu32 ",\"firstCallNs\":%" PRIu64 ",\"warmCallNs\":[", seed, count, expected, first);
    for (int i = 0; i < 7; i++) printf("%s%" PRIu64, i ? "," : "", samples[i]);
    puts("]}");
    free(memory);
    return 0;
}
