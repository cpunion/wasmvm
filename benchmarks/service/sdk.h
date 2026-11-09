#ifndef WASMVM_SERVICE_SDK_H
#define WASMVM_SERVICE_SDK_H
#include <stddef.h>
#include <stdint.h>

/* Exactly one optional SDK translation unit supplies this opaque bridge.
 * Each worker owns its mutable store/executor/function and error storage. */
const char *sdk_name(void);
const char *sdk_version(void);
void *sdk_open(const uint8_t *bytes, size_t size, unsigned workers);
const char *sdk_error(void *owner, unsigned worker);
uint64_t sdk_phase(void *owner, unsigned phase);
uint64_t sdk_call(void *owner, unsigned worker, uint32_t seed, uint32_t count);
void sdk_close(void *owner);
#endif
