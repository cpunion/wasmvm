#include <wasmtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

static uint64_t now(void) {
  struct timespec t;
  clock_gettime(CLOCK_MONOTONIC, &t);
  return (uint64_t)t.tv_sec * 1000000000 + (uint64_t)t.tv_nsec;
}
static void check(wasmtime_error_t *e) {
  if (!e) return;
  wasm_byte_vec_t message;
  wasmtime_error_message(e, &message);
  fprintf(stderr, "%.*s\n", (int)message.size, message.data);
  wasm_byte_vec_delete(&message);
  wasmtime_error_delete(e);
  exit(1);
}
typedef struct {
  wasmtime_store_t *store;
  wasmtime_context_t *context;
  wasmtime_component_instance_t instance;
  wasmtime_component_func_t function;
  wasmtime_error_t *error;
} invocation;

static void prepare(invocation *call, wasm_engine_t *engine,
                    wasmtime_component_linker_t *linker,
                    wasmtime_component_t *component, const char *name,
                    uint64_t *setup, uint64_t *instantiate, uint64_t *lookup) {
  uint64_t t = now();
  call->store = wasmtime_store_new(engine, NULL, NULL);
  call->context = wasmtime_store_context(call->store);
  check(wasmtime_context_set_wasi(call->context, wasi_config_new()));
  *setup += now() - t;
  t = now();
  check(wasmtime_component_linker_instantiate(linker, call->context, component, &call->instance));
  *instantiate += now() - t;
  t = now();
  wasmtime_component_export_index_t *index = wasmtime_component_instance_get_export_index(&call->instance, call->context, NULL, name, strlen(name));
  if (!index || !wasmtime_component_instance_get_func(&call->instance, call->context, index, &call->function)) exit(5);
  wasmtime_component_export_index_delete(index);
  *lookup += now() - t;
}
static void verify(invocation *call, int expected) {
  int status = -1;
  if (!call->error || !wasmtime_error_exit_status(call->error, &status) || status != expected) {
    check(call->error);
    exit(6);
  }
  wasmtime_error_delete(call->error);
  wasmtime_store_delete(call->store);
}
int main(int argc, char **argv) {
  if (argc != 3) return 2;
  uint64_t count = strtoull(argv[2], NULL, 10), t = now();
  if (!count) return 2;
  FILE *file = fopen(argv[1], "rb");
  if (!file) return 3;
  fseek(file, 0, SEEK_END);
  long size = ftell(file);
  rewind(file);
  uint8_t *bytes = malloc((size_t)size);
  if (!bytes || fread(bytes, 1, (size_t)size, file) != (size_t)size) return 4;
  fclose(file);
  uint64_t load = now() - t;
  t = now();
  wasm_engine_t *engine = wasm_engine_new();
  uint64_t engine_time = now() - t;
  t = now();
  wasmtime_component_t *component = NULL;
  check(wasmtime_component_new(engine, bytes, (size_t)size, &component));
  uint64_t compile = now() - t;
  t = now();
  wasmtime_component_linker_t *linker = wasmtime_component_linker_new(engine);
  check(wasmtime_component_linker_add_wasip2(linker));
  uint64_t linker_time = now() - t, setup = 0, instantiate = 0, lookup = 0;
  /* Verify both interface operations and status extremes before timing. */
  uint64_t first = 0;
  for (int i = 0; i < 4; i++) {
    invocation call;
    prepare(&call, engine, linker, component, i < 2 ? (i ? "failure" : "success") : "with-code", &setup, &instantiate, &lookup);
    wasmtime_component_val_t arg = {.kind = WASMTIME_COMPONENT_U8, .of.u8 = i == 3 ? 255 : 42};
    t = now();
    call.error = wasmtime_component_func_call(&call.function, call.context, i < 2 ? NULL : &arg, i < 2 ? 0 : 1, NULL, 0);
    if (!i) first = now() - t;
    verify(&call, i < 2 ? i : arg.of.u8);
  }
  setup = instantiate = lookup = 0;
  uint64_t calls = 0, destroy = 0;
  invocation batch[256];
  wasmtime_component_val_t arg = {.kind = WASMTIME_COMPONENT_U8, .of.u8 = 42};
  for (uint64_t done = 0; done < count;) {
    size_t n = count - done < 256 ? (size_t)(count - done) : 256;
    for (size_t i = 0; i < n; i++) prepare(&batch[i], engine, linker, component, "with-code", &setup, &instantiate, &lookup);
    t = now();
    for (size_t i = 0; i < n; i++) batch[i].error = wasmtime_component_func_call(&batch[i].function, batch[i].context, &arg, 1, NULL, 0);
    calls += now() - t;
    t = now();
    for (size_t i = 0; i < n; i++) verify(&batch[i], 42);
    destroy += now() - t;
    done += n;
  }
  printf("{\"load_ns\":%llu,\"engine_ns\":%llu,\"compile_ns\":%llu,\"linker_ns\":%llu,\"first_ns\":%llu,\"store_setup_ns_each\":%.3f,\"instantiate_ns_each\":%.3f,\"lookup_ns_each\":%.3f,\"exit_call_ns_each\":%.3f,\"verify_and_destroy_ns_each\":%.3f,\"count\":%llu,\"checksum\":%llu}\n", (unsigned long long)load, (unsigned long long)engine_time, (unsigned long long)compile, (unsigned long long)linker_time, (unsigned long long)first, (double)setup/count, (double)instantiate/count, (double)lookup/count, (double)calls/count, (double)destroy/count, (unsigned long long)count, (unsigned long long)(42*count));
  wasmtime_component_linker_delete(linker);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(bytes);
  return 0;
}
