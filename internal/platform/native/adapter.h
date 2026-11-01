#ifndef WASMVM_NATIVE_ADAPTER_H
#define WASMVM_NATIVE_ADAPTER_H
#include <stdint.h>
#include <stddef.h>

struct wasmvm_record {
    uint64_t version, arena, capacity, max_depth;
    uint64_t used, depth, function, continuation;
    uint64_t source, root_map, fuel, budget;
    uint64_t result_lo, result_hi, payload[4];
    /* Adapter accesses use __atomic_*; generated code uses aligned atomic loads. */
    uint64_t control;
    uint64_t memory_base, memory_size;
};
_Static_assert(sizeof(void *) == 8, "native ABI requires 64-bit pointers");
_Static_assert(sizeof(struct wasmvm_record) == 168, "record size");
_Static_assert(_Alignof(struct wasmvm_record) == 8, "record alignment");
_Static_assert(offsetof(struct wasmvm_record, control) == 144, "control offset");
_Static_assert(__atomic_always_lock_free(8, 0), "control must be lock-free");

void *wasmvm_publish(const void *source, size_t size);
void wasmvm_unpublish(void *code, size_t size);
void *wasmvm_allocate(size_t size);
void wasmvm_deallocate(void *p);
struct wasmvm_atomic_result { uint64_t value, retry; };
struct wasmvm_atomic_result wasmvm_atomic_access(void *address, uint32_t width, uint32_t operation,
                              uint64_t a, uint64_t b);
uint64_t wasmvm_shared_load(const void *address, uint32_t width);
void wasmvm_shared_store(void *address, uint32_t width, uint64_t value);
void wasmvm_shared_transfer(void *destination, const void *source, size_t size, uint32_t mode);
void wasmvm_shared_fill(void *destination, size_t size, uint8_t byte);
void wasmvm_shared_fence(void);
void wasmvm_cancel(struct wasmvm_record *ctx);
uint32_t wasmvm_enter(const void *entry, struct wasmvm_record *ctx);
uint64_t wasmvm_layout(uint32_t field);
/* Raw CPU/OS capabilities; all zero on other targets. */
void wasmvm_cpu_probe(uint32_t *ecx1, uint32_t *ebx7, uint32_t *ecx7, uint64_t *xcr0);
/* ABI diagnostic: invoke an entry under a non-default host FP environment. */
uint32_t wasmvm_probe_fp(const void *entry, struct wasmvm_record *ctx);
uint32_t wasmvm_probe_registers(const void *entry, struct wasmvm_record *ctx);
#endif
