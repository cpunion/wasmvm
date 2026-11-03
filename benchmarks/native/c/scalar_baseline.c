// Unchecked native C references for six existing parameterized Wasm kernels.
// Clang may inline/vectorize; each call includes the fixture's memory reset.
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
    if (clock_gettime(clock, &ts)) abort();
    return (uint64_t)ts.tv_sec * UINT64_C(1000000000) + ts.tv_nsec;
}

static uint32_t rotate(uint32_t x, unsigned bits) {
    return (x << bits) | (x >> (32 - bits));
}

static uint32_t run(unsigned kind, unsigned char *memory, uint32_t seed, uint32_t count) {
    uint32_t sum = seed;
    switch (kind) {
    case 0:
        for (uint32_t i = 0; i < count; i++)
            sum = rotate(sum ^ i, 5) * UINT32_C(1664525) + UINT32_C(1013904223);
        break;
    case 1:
        for (uint32_t i = 0; i < count; i++)
            sum = rotate(sum + i, 7) ^ UINT32_C(0x9e3779b9);
        break;
    case 2:
        memset(memory, 0, 32768);
        for (uint32_t i = 0; i < count; i++) {
            uint32_t address = ((i * 17) & 8191) * 4, value;
            memcpy(&value, memory + address, sizeof(value));
            value = (value ^ i ^ seed) + UINT32_C(0x9e3779b9);
            memcpy(memory + address, &value, sizeof(value));
            sum += value;
        }
        break;
    case 3:
        memset(memory, 0, 2049 * 4);
        for (uint32_t i = 0; i < count; i++) {
            uint32_t address = ((i * 13) & 2047) * 4 + 1;
            int16_t old;
            memcpy(&old, memory + address, sizeof(old));
            uint16_t value = (uint16_t)(((uint32_t)(int32_t)old ^ seed ^ i) + 31337);
            memcpy(memory + address, &value, sizeof(value));
            sum += (uint32_t)value ^ memory[address + 1];
        }
        break;
    case 4: {
        for (uint32_t i = 0; i < 4096; i++) {
            uint32_t value = (i * 109 + seed) & 4095;
            memcpy(memory + i * 4, &value, sizeof(value));
        }
        uint32_t address = seed & 4095;
        for (uint32_t i = 0; i < count; i++) {
            memcpy(&address, memory + address * 4, sizeof(address));
            address &= 4095;
            sum += address;
        }
        break;
    }
    case 5: {
        double total = 0, initial = (double)(seed & 255);
        for (uint32_t i = 0; i < 2048; i++)
            memcpy(memory + i * 8 + 1, &initial, sizeof(initial));
        for (uint32_t i = 0; i < count; i++) {
            uint32_t address = ((i * 17) & 2047) * 8 + 1;
            double number;
            memcpy(&number, memory + address, sizeof(number));
            number += 0.125;
            memcpy(memory + address, &number, sizeof(number));
            total += number;
        }
        sum = (uint32_t)total;
        break;
    }
    default: abort();
    }
    return sum;
}

// Keep every invocation between its clocks even when the same pure scalar
// inputs repeat. The volatile indirect entry prevents hoisting/CSE across calls.
static uint32_t (*volatile kernel)(unsigned, unsigned char *, uint32_t, uint32_t) = run;

int main(int argc, char **argv) {
    static const char *names[] = {"integer-loop", "function-calls", "memory-load-store",
        "narrow-unaligned", "dependent-address", "float64-unaligned"};
    if (argc != 5) return 2;
    unsigned kind = 0;
    while (kind < sizeof(names) / sizeof(names[0]) && strcmp(names[kind], argv[1])) kind++;
    if (kind == sizeof(names) / sizeof(names[0])) return 2;
    uint32_t seed = (uint32_t)strtoul(argv[2], NULL, 0);
    uint32_t count = (uint32_t)strtoul(argv[3], NULL, 0);
    uint32_t expected = (uint32_t)strtoul(argv[4], NULL, 0);
    // The float fixture must return an in-range Wasm i32.trunc_f64_u result.
    if (!count || count > 1000000) return 2;
    unsigned char *memory = calloc(1, 65536);
    if (!memory) return 2;
    int interleave = getenv("WASMVM_INTERLEAVE") != NULL;
    char token[16];
    if (interleave) {
        puts("{\"ready\":true}"); fflush(stdout);
        if (!fgets(token, sizeof(token), stdin)) return 2;
    }
    uint64_t samples[7], first = 0;
    for (int sample = -3; sample < 7; sample++) {
        if (sample >= 0 && interleave) {
            printf("{\"sampleReady\":%d}\n", sample); fflush(stdout);
            if (!fgets(token, sizeof(token), stdin)) return 2;
        }
        uint64_t start = now_ns();
        uint32_t actual = kernel(kind, memory, seed, count);
        uint64_t elapsed = now_ns() - start;
        if (actual != expected) {
            fprintf(stderr, "checksum: got %" PRIu32 ", expected %" PRIu32 "\n", actual, expected);
            free(memory);
            return 1;
        }
        if (sample == -3) first = elapsed;
        if (sample >= 0) samples[sample] = elapsed;
    }
    printf("{\"seed\":%" PRIu32 ",\"iterations\":%" PRIu32 ",\"checksum\":%" PRIu32
        ",\"firstCallNs\":%" PRIu64 ",\"warmCallNs\":[", seed, count, expected, first);
    for (int i = 0; i < 7; i++) printf("%s%" PRIu64, i ? "," : "", samples[i]);
    puts("]}");
    free(memory);
    return 0;
}
