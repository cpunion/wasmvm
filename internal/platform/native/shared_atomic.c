//go:build (cgo || llgo) && ((linux && amd64) || (darwin && arm64))

#include "adapter.h"
#include <string.h>

/* Guest bytes follow Wasm's mixed-size memory model, not the C/LLVM atomic
 * object model. Keep all guest data accesses in opaque target instructions.
 * The native size cell is separate and always accessed atomically as u64. */
#if defined(__x86_64__)
#define LOAD(T, S, M) do { T value; \
    __asm__ __volatile__("mov" S " (%1), " M "0" : "=r"(value) : "r"(address) : "memory"); \
    return value; } while (0)
#define STORE(T, S, M) do { T value = (T)bits; \
    __asm__ __volatile__("mov" S " " M "0, (%1)" : : "r"(value), "r"(address) : "memory"); \
    return; } while (0)
#define EXCHANGE(T, S, M) do { T value = (T)bits; \
    __asm__ __volatile__("xchg" S " " M "0, (%1)" : "+r"(value) : "r"(address) : "memory"); \
    return value; } while (0)
#define XADD(T, S, M) do { T value = (T)bits; \
    __asm__ __volatile__("lock; xadd" S " " M "0, (%1)" : "+r"(value) : "r"(address) : "memory", "cc"); \
    return value; } while (0)
#define COMPARE_EXCHANGE(T, S, M) do { T value = (T)expected, next = (T)desired; \
    __asm__ __volatile__("lock; cmpxchg" S " " M "2, (%1)" : "+a"(value) : "r"(address), "r"(next) : "memory", "cc"); \
    return (struct wasmvm_atomic_result){value, 0}; } while (0)

static uint64_t machine_load(const void *address, uint32_t width, int sequential) {
    (void)sequential; /* TSO: SC stores/fences supply the store-load ordering. */
    switch (width) {
    case 1: LOAD(uint8_t, "b", "%b");
    case 2: LOAD(uint16_t, "w", "%w");
    case 4: LOAD(uint32_t, "l", "%k");
    case 8: LOAD(uint64_t, "q", "%");
    }
    __builtin_unreachable();
}
static void machine_store(void *address, uint32_t width, uint64_t bits) {
    switch (width) {
    case 1: STORE(uint8_t, "b", "%b");
    case 2: STORE(uint16_t, "w", "%w");
    case 4: STORE(uint32_t, "l", "%k");
    case 8: STORE(uint64_t, "q", "%");
    }
    __builtin_unreachable();
}
static uint64_t machine_exchange(void *address, uint32_t width, uint64_t bits) {
    switch (width) {
    case 1: EXCHANGE(uint8_t, "b", "%b");
    case 2: EXCHANGE(uint16_t, "w", "%w");
    case 4: EXCHANGE(uint32_t, "l", "%k");
    case 8: EXCHANGE(uint64_t, "q", "%");
    }
    __builtin_unreachable();
}
static uint64_t machine_xadd(void *address, uint32_t width, uint64_t bits) {
    switch (width) {
    case 1: XADD(uint8_t, "b", "%b");
    case 2: XADD(uint16_t, "w", "%w");
    case 4: XADD(uint32_t, "l", "%k");
    case 8: XADD(uint64_t, "q", "%");
    }
    __builtin_unreachable();
}
static struct wasmvm_atomic_result machine_compare_exchange(void *address, uint32_t width, uint64_t expected, uint64_t desired) {
    switch (width) {
    case 1: COMPARE_EXCHANGE(uint8_t, "b", "%b");
    case 2: COMPARE_EXCHANGE(uint16_t, "w", "%w");
    case 4: COMPARE_EXCHANGE(uint32_t, "l", "%k");
    case 8: COMPARE_EXCHANGE(uint64_t, "q", "%");
    }
    __builtin_unreachable();
}
#elif defined(__aarch64__)
#define LOAD(T, I, M) do { T value; \
    __asm__ __volatile__(I " " M "0, [%1]" : "=r"(value) : "r"(address) : "memory"); \
    return value; } while (0)
#define STORE(T, I, M) do { T value = (T)bits; \
    __asm__ __volatile__(I " " M "0, [%1]" : : "r"(value), "r"(address) : "memory"); \
    return; } while (0)
#define COMPARE_EXCHANGE(T, L, S, M) do { T value; uint64_t before = (T)expected, next = (T)desired; uint32_t status, budget = 16; \
    __asm__ __volatile__("1: " L " " M "0, [%3]\n\t" \
        "cmp " M "0, " M "4\n\t" \
        "b.ne 2f\n\t" \
        S " %w1, " M "5, [%3]\n\t" \
        "cbz %w1, 3f\n\t" \
        "subs %w2, %w2, #1\n\t" \
        "b.ne 1b\n\t" \
        "clrex\n\t" \
        "b 3f\n" \
        "2: clrex\n\t" \
        "mov %w1, #0\n" \
        "3:" : "=&r"(value), "=&r"(status), "+&r"(budget) : "r"(address), "r"(before), "r"(next) : "memory", "cc"); \
    return (struct wasmvm_atomic_result){value, status}; } while (0)
static uint64_t machine_load(const void *address, uint32_t width, int sequential) {
    if (sequential) {
        switch (width) {
        case 1: LOAD(uint8_t, "ldarb", "%w");
        case 2: LOAD(uint16_t, "ldarh", "%w");
        case 4: LOAD(uint32_t, "ldar", "%w");
        case 8: LOAD(uint64_t, "ldar", "%");
        }
    } else {
        switch (width) {
        case 1: LOAD(uint8_t, "ldrb", "%w");
        case 2: LOAD(uint16_t, "ldrh", "%w");
        case 4: LOAD(uint32_t, "ldr", "%w");
        case 8: LOAD(uint64_t, "ldr", "%");
        }
    }
    __builtin_unreachable();
}
static void machine_store(void *address, uint32_t width, uint64_t bits) {
    switch (width) {
    case 1: STORE(uint8_t, "strb", "%w");
    case 2: STORE(uint16_t, "strh", "%w");
    case 4: STORE(uint32_t, "str", "%w");
    case 8: STORE(uint64_t, "str", "%");
    }
    __builtin_unreachable();
}
static void machine_store_sequential(void *address, uint32_t width, uint64_t bits) {
    switch (width) {
    case 1: STORE(uint8_t, "stlrb", "%w");
    case 2: STORE(uint16_t, "stlrh", "%w");
    case 4: STORE(uint32_t, "stlr", "%w");
    case 8: STORE(uint64_t, "stlr", "%");
    }
    __builtin_unreachable();
}
static struct wasmvm_atomic_result machine_compare_exchange(void *address, uint32_t width, uint64_t expected, uint64_t desired) {
    switch (width) {
    case 1: COMPARE_EXCHANGE(uint8_t, "ldaxrb", "stlxrb", "%w");
    case 2: COMPARE_EXCHANGE(uint16_t, "ldaxrh", "stlxrh", "%w");
    case 4: COMPARE_EXCHANGE(uint32_t, "ldaxr", "stlxr", "%w");
    case 8: COMPARE_EXCHANGE(uint64_t, "ldaxr", "stlxr", "%");
    }
    __builtin_unreachable();
}
#else
#error unsupported shared-memory target
#endif

struct wasmvm_atomic_result wasmvm_atomic_access(void *address, uint32_t width,
                                                uint32_t operation, uint64_t a, uint64_t b) {
    if (operation == 0) return (struct wasmvm_atomic_result){machine_load(address, width, 1), 0};
    if (operation == 1) {
#if defined(__x86_64__)
        (void)machine_exchange(address, width, a);
#else
        machine_store_sequential(address, width, a);
#endif
        return (struct wasmvm_atomic_result){0, 0};
    }
    if (operation == 8) return machine_compare_exchange(address, width, a, b);
#if defined(__x86_64__)
    if (operation == 2) return (struct wasmvm_atomic_result){machine_xadd(address, width, a), 0};
    if (operation == 3) return (struct wasmvm_atomic_result){machine_xadd(address, width, -a), 0};
    if (operation == 7) return (struct wasmvm_atomic_result){machine_exchange(address, width, a), 0};
#endif
    uint64_t mask = UINT64_MAX >> (64 - width * 8);
    uint64_t old = machine_load(address, width, 1);
    for (unsigned attempt = 0; attempt < 16; ++attempt) {
        uint64_t next;
        switch (operation) {
        case 2: next = old + a; break;
        case 3: next = old - a; break;
        case 4: next = old & a; break;
        case 5: next = old | a; break;
        case 6: next = old ^ a; break;
        case 7: next = a; break;
        default: __builtin_unreachable();
        }
        struct wasmvm_atomic_result observed = machine_compare_exchange(address, width, old, next & mask);
        if (observed.retry) return observed;
        if (old == observed.value) return observed;
        old = observed.value;
    }
    /* No write committed. Let the Go caller poll cancellation before retry. */
    return (struct wasmvm_atomic_result){0, 1};
}

/* Nonatomic Wasm/host transfers allow byte tearing. Use relaxed atomics in C
 * so concurrent JIT/atomic access never races an ordinary C memory operation. */
uint64_t wasmvm_shared_load(const void *address, uint32_t width) {
    if (((uintptr_t)address & (width - 1)) == 0) {
        return machine_load(address, width, 0);
    }
    uint64_t value = 0;
    for (uint32_t i = 0; i < width; ++i)
        value |= machine_load((const uint8_t *)address + i, 1, 0) << (i * 8);
    return value;
}

void wasmvm_shared_store(void *address, uint32_t width, uint64_t value) {
    if (((uintptr_t)address & (width - 1)) == 0) {
        machine_store(address, width, value);
        return;
    }
    for (uint32_t i = 0; i < width; ++i)
        machine_store((uint8_t *)address + i, 1, (uint8_t)(value >> (i * 8)));
}

static uint64_t transfer_read(const uint8_t *p, uint32_t width, uint32_t mode) {
    if (mode & 2) return wasmvm_shared_load(p, width);
    uint64_t value = 0;
    memcpy(&value, p, width);
    return value;
}
static void transfer_write(uint8_t *p, uint32_t width, uint64_t value, uint32_t mode) {
    if (mode & 1) wasmvm_shared_store(p, width, value);
    else memcpy(p, &value, width);
}

void wasmvm_shared_transfer(void *destination, const void *source, size_t size, uint32_t mode) {
    uint8_t *dst = destination;
    const uint8_t *src = source;
    uintptr_t d = (uintptr_t)dst, s = (uintptr_t)src;
    if (d > s && d - s < size) {
        while (size >= 8) {
            size -= 8;
            transfer_write(dst + size, 8, transfer_read(src + size, 8, mode), mode);
        }
        while (size) {
            --size;
            transfer_write(dst + size, 1, transfer_read(src + size, 1, mode), mode);
        }
    } else {
        size_t at = 0;
        while (size - at >= 8) {
            transfer_write(dst + at, 8, transfer_read(src + at, 8, mode), mode);
            at += 8;
        }
        while (at < size) {
            transfer_write(dst + at, 1, transfer_read(src + at, 1, mode), mode);
            ++at;
        }
    }
}

void wasmvm_shared_fill(void *destination, size_t size, uint8_t byte) {
    uint8_t *dst = destination;
    while (size && ((uintptr_t)dst & 7)) {
        wasmvm_shared_store(dst++, 1, byte);
        --size;
    }
    uint64_t word = (uint64_t)byte * UINT64_C(0x0101010101010101);
    while (size >= 8) {
        wasmvm_shared_store(dst, 8, word);
        dst += 8;
        size -= 8;
    }
    while (size) {
        wasmvm_shared_store(dst++, 1, byte);
        --size;
    }
}

void wasmvm_shared_fence(void) {
#if defined(__x86_64__)
    __asm__ __volatile__("mfence" : : : "memory");
#else
    __asm__ __volatile__("dmb ish" : : : "memory");
#endif
}
