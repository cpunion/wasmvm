#include <wasmtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <stdint.h>

static uint64_t now(void) {
  struct timespec t;
  clock_gettime(CLOCK_MONOTONIC, &t);
  return (uint64_t)t.tv_sec * 1000000000 + (uint64_t)t.tv_nsec;
}
static void check(wasmtime_error_t *e) {
  if (!e) return;
  wasm_byte_vec_t msg;
  wasmtime_error_message(e, &msg);
  fprintf(stderr, "%.*s\n", (int)msg.size, msg.data);
  wasm_byte_vec_delete(&msg);
  wasmtime_error_delete(e);
  exit(1);
}
int main(int argc, char **argv) {
  if (argc != 3) return 2;
  uint64_t count = strtoull(argv[2], NULL, 10), t = now();
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
  wasmtime_store_t *store = wasmtime_store_new(engine, NULL, NULL);
  wasmtime_context_t *context = wasmtime_store_context(store);
  wasmtime_component_linker_t *linker = wasmtime_component_linker_new(engine);
  uint64_t setup = now() - t;
  t = now();
  wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker, context, component, &instance));
  uint64_t instantiate = now() - t;
  t = now();
  wasmtime_component_export_index_t *index = wasmtime_component_instance_get_export_index(&instance, context, NULL, "add", 3);
  wasmtime_component_func_t function;
  if (!index || !wasmtime_component_instance_get_func(&instance, context, index, &function)) return 5;
  uint64_t lookup = now() - t;
  wasmtime_component_val_t args[2] = {{.kind = WASMTIME_COMPONENT_U32, .of.u32 = 0}, {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 5}}, result = {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 0};
  t = now();
  check(wasmtime_component_func_call(&function, context, args, 2, &result, 1));
  uint64_t first = now() - t;
  if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != 5) return 6;
  uint64_t checksum = 0;
  t = now();
  for (uint64_t i = 0; i < count; i++) {
    args[0].of.u32 = (uint32_t)i;
    check(wasmtime_component_func_call(&function, context, args, 2, &result, 1));
    if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != (uint32_t)i + 5) return 7;
    checksum += result.of.u32;
  }
  uint64_t warm = now() - t;
  printf("{\"load_ns\":%llu,\"engine_ns\":%llu,\"compile_ns\":%llu,\"store_linker_ns\":%llu,\"instantiate_ns\":%llu,\"lookup_ns\":%llu,\"first_ns\":%llu,\"warm_ns_per_call\":%.3f,\"count\":%llu,\"checksum\":%llu}\n", (unsigned long long)load, (unsigned long long)engine_time, (unsigned long long)compile, (unsigned long long)setup, (unsigned long long)instantiate, (unsigned long long)lookup, (unsigned long long)first, (double)warm/count, (unsigned long long)count, (unsigned long long)checksum);
  wasmtime_component_export_index_delete(index);
  wasmtime_component_linker_delete(linker);
  wasmtime_store_delete(store);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(bytes);
  return 0;
}
