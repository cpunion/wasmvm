//go:build (cgo || llgo) && ((linux && amd64) || (darwin && arm64))

#include "adapter.h"

/* Wasm permits overlapping, differently sized accesses to raw linear memory. */
typedef uint8_t alias_u8 __attribute__((may_alias));
typedef uint16_t alias_u16 __attribute__((may_alias));
typedef uint32_t alias_u32 __attribute__((may_alias));
typedef uint64_t alias_u64 __attribute__((may_alias));
_Static_assert(__atomic_always_lock_free(1, 0), "byte atomics must be lock-free");
_Static_assert(__atomic_always_lock_free(2, 0), "halfword atomics must be lock-free");
_Static_assert(__atomic_always_lock_free(4, 0), "word atomics must be lock-free");
_Static_assert(__atomic_always_lock_free(8, 0), "doubleword atomics must be lock-free");

#define ACCESS(T) do { \
    volatile T *p = (volatile T *)address; \
    T x = (T)a, y = (T)b; \
    switch (operation) { \
    case 0: return __atomic_load_n(p, __ATOMIC_SEQ_CST); \
    case 1: __atomic_store_n(p, x, __ATOMIC_SEQ_CST); return 0; \
    case 2: return __atomic_fetch_add(p, x, __ATOMIC_SEQ_CST); \
    case 3: return __atomic_fetch_sub(p, x, __ATOMIC_SEQ_CST); \
    case 4: return __atomic_fetch_and(p, x, __ATOMIC_SEQ_CST); \
    case 5: return __atomic_fetch_or(p, x, __ATOMIC_SEQ_CST); \
    case 6: return __atomic_fetch_xor(p, x, __ATOMIC_SEQ_CST); \
    case 7: return __atomic_exchange_n(p, x, __ATOMIC_SEQ_CST); \
    case 8: __atomic_compare_exchange_n(p, &x, y, 0, __ATOMIC_SEQ_CST, __ATOMIC_SEQ_CST); return x; \
    } \
    __builtin_unreachable(); \
} while (0)

uint64_t wasmvm_atomic_access(void *address, uint32_t width, uint32_t operation,
                              uint64_t a, uint64_t b) {
    switch (width) {
    case 1: ACCESS(alias_u8);
    case 2: ACCESS(alias_u16);
    case 4: ACCESS(alias_u32);
    case 8: ACCESS(alias_u64);
    }
    __builtin_unreachable();
}
