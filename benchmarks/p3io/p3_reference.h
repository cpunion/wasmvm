#ifndef WASMVM_P3_REFERENCE_H
#define WASMVM_P3_REFERENCE_H
#include <stdint.h>
#include <stddef.h>
typedef struct wasmvm_p3_owner wasmvm_p3_owner;
typedef struct { const uint8_t *bytes; size_t length; uint32_t status,guest,trailer; } wasmvm_p3_reply;
wasmvm_p3_owner *wasmvm_p3_open(const uint8_t *,size_t,const uint8_t *,size_t,uint32_t);
uint64_t wasmvm_p3_phase(const wasmvm_p3_owner *,uint32_t);
const char *wasmvm_p3_error(const wasmvm_p3_owner *,uint32_t);
uint32_t wasmvm_p3_call(const wasmvm_p3_owner *,uint32_t,wasmvm_p3_reply *);
void wasmvm_p3_close(wasmvm_p3_owner *);
#endif
