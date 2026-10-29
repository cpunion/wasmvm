#include <pthread.h>
#include <sched.h>
#include <stdatomic.h>
#include <stdint.h>

static _Thread_local uint32_t local_value;
static pthread_key_t key;
static atomic_uint destructors;
static atomic_uint failures;

static void destroy(void *value) {
  if (local_value != (uint32_t)(uintptr_t)value)
    atomic_fetch_add(&failures, 1);
  atomic_fetch_add(&destructors, 1);
}

int32_t wasmvm_probe_tls(uint32_t value) {
  if (local_value != 0) return 1;
  local_value = value;
  for (int n = 0; n < 128; n++) {
    sched_yield();
    if (local_value != value) return 2;
  }
  return 0;
}

int32_t wasmvm_register_tsd(uint32_t value) {
  return pthread_setspecific(key, (void *)(uintptr_t)value);
}

static void *worker(void *arg) {
  uint32_t value = (uint32_t)(uintptr_t)arg;
  if (wasmvm_probe_tls(value) || wasmvm_register_tsd(value))
    atomic_fetch_add(&failures, 1);
  return arg;
}

int32_t wasmvm_probe_pthreads(void) {
  if (pthread_key_create(&key, destroy)) return 1;
  pthread_t threads[3];
  for (uintptr_t i = 0; i < 3; i++)
    if (pthread_create(&threads[i], 0, worker, (void *)(i + 1))) return 2;
  for (uintptr_t i = 0; i < 3; i++) {
    void *returned = 0;
    if (pthread_join(threads[i], &returned) || returned != (void *)(i + 1)) return 3;
  }
  return local_value != 0 || atomic_load(&failures) != 0;
}

uint32_t wasmvm_destructor_count(void) {
  return atomic_load(&failures) ? UINT32_MAX : atomic_load(&destructors);
}

uint32_t wasmvm_grow_memory(uint32_t pages) {
  return __builtin_wasm_memory_grow(0, pages);
}

uint32_t wasmvm_probe_simd(uint32_t seed) {
  typedef uint32_t lanes4 __attribute__((vector_size(16)));
  static volatile lanes4 stored;
  stored = (lanes4){seed, seed + 1, seed + 2, seed + 3};
  lanes4 values = stored;
  stored = values * (lanes4){2, 2, 2, 2};
  values = stored;
  return values[0] + values[1] + values[2] + values[3] + 6;
}
