// Native C11 reference for the parameterized atomic-rmw guest kernel.
// This reference has no VM cancellation, bounds, validation or fuel contract.
#include <inttypes.h>
#include <stdint.h>
#include <stdatomic.h>
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

static uint32_t run(unsigned kind, unsigned char *memory, uint32_t seed, uint32_t count) {
    (void)kind;
    _Atomic uint32_t *word = (_Atomic uint32_t *)memory;
    atomic_store_explicit(word, seed, memory_order_seq_cst);
    for (uint32_t i = 0; i < count; ++i) {
        atomic_fetch_add_explicit(word, 1, memory_order_seq_cst);
    }
    return atomic_load_explicit(word, memory_order_seq_cst);
}

// Keep every invocation between its clocks even when the same pure scalar
// inputs repeat. The volatile indirect entry prevents hoisting/CSE across calls.
static uint32_t (*volatile kernel)(unsigned, unsigned char *, uint32_t, uint32_t) = run;

int main(int argc, char **argv) {
    static const char *names[] = {"atomic-rmw"};
    if (argc != 5) return 2;
    unsigned kind = 0;
    while (kind < sizeof(names) / sizeof(names[0]) && strcmp(names[kind], argv[1])) kind++;
    if (kind == sizeof(names) / sizeof(names[0])) return 2;
    uint32_t seed = (uint32_t)strtoul(argv[2], NULL, 0);
    uint32_t count = (uint32_t)strtoul(argv[3], NULL, 0);
    uint32_t expected = (uint32_t)strtoul(argv[4], NULL, 0);
    // Match the guest signature's positive signed-I32 loop bound.
    if (!count || count > INT32_MAX) return 2;
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
