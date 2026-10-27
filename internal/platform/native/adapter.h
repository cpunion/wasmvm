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
void wasmvm_cancel(struct wasmvm_record *ctx);
uint32_t wasmvm_enter(const void *entry, struct wasmvm_record *ctx);
uint64_t wasmvm_layout(uint32_t field);
/* Raw CPU/OS capabilities; all zero on other targets. */
void wasmvm_cpu_probe(uint32_t *ecx1, uint32_t *ebx7, uint64_t *xcr0);
/* ABI diagnostic: invoke an entry under a non-default host FP environment. */
uint32_t wasmvm_probe_fp(const void *entry, struct wasmvm_record *ctx);
uint32_t wasmvm_probe_registers(const void *entry, struct wasmvm_record *ctx);
#endif
