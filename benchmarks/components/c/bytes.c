#include <wasmtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <stdint.h>
#include <string.h>

static uint64_t now(void) {
  struct timespec t;
  clock_gettime(CLOCK_MONOTONIC, &t);
  return (uint64_t)t.tv_sec * 1000000000 + (uint64_t)t.tv_nsec;
}
static void check(wasmtime_error_t *error) {
  if (!error) return;
  wasm_byte_vec_t message;
  wasmtime_error_message(error, &message);
  fprintf(stderr, "%.*s\n", (int)message.size, message.data);
  wasm_byte_vec_delete(&message);
  wasmtime_error_delete(error);
  exit(1);
}
static void verify(wasmtime_component_val_t *value, size_t size, int full) {
  if (value->kind != WASMTIME_COMPONENT_LIST || value->of.list.size != size) exit(6);
  if (value->of.list.data[0].kind != WASMTIME_COMPONENT_U8 || value->of.list.data[0].of.u8 != 0 || value->of.list.data[size-1].kind != WASMTIME_COMPONENT_U8 || value->of.list.data[size-1].of.u8 != 255) exit(7);
  if (full) {
    for (size_t i = 0; i < size; i++) if (value->of.list.data[i].kind != WASMTIME_COMPONENT_U8 || value->of.list.data[i].of.u8 != (uint8_t)i) exit(8);
  }
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
  wasmtime_store_t *store = wasmtime_store_new(engine, NULL, NULL);
  wasmtime_context_t *context = wasmtime_store_context(store);
  wasmtime_component_linker_t *linker = wasmtime_component_linker_new(engine);
  uint64_t setup = now() - t;
  t = now();
  wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker, context, component, &instance));
  uint64_t instantiate = now() - t;
  t = now();
  wasmtime_component_export_index_t *index = wasmtime_component_instance_get_export_index(&instance, context, NULL, "echo", 4);
  wasmtime_component_func_t function;
  if (!index || !wasmtime_component_instance_get_func(&instance, context, index, &function)) return 5;
  wasmtime_component_export_index_delete(index);
  uint64_t lookup = now() - t;
  const size_t n = 16384;
  wasmtime_component_val_t arg = {.kind = WASMTIME_COMPONENT_LIST};
  wasmtime_component_vallist_new_uninit(&arg.of.list, n);
  for (size_t i = 0; i < n; i++) arg.of.list.data[i] = (wasmtime_component_val_t){.kind = WASMTIME_COMPONENT_U8, .of.u8 = (uint8_t)i};
  /* v48.0.1 replaces and drops the old C result; keep its prior kind valid. */
  wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U8, .of.u8 = 0};
  t = now();
  check(wasmtime_component_func_call(&function, context, &arg, 1, &result, 1));
  uint64_t first = now() - t;
  verify(&result, n, 1);
  wasmtime_component_val_delete(&result);
  t = now();
  for (uint64_t i = 0; i < count; i++) {
    result = (wasmtime_component_val_t){.kind = WASMTIME_COMPONENT_U8, .of.u8 = 0};
    check(wasmtime_component_func_call(&function, context, &arg, 1, &result, 1));
    verify(&result, n, 0);
    wasmtime_component_val_delete(&result);
  }
  uint64_t warm = now() - t;
  printf("{\"load_ns\":%llu,\"engine_ns\":%llu,\"compile_ns\":%llu,\"store_linker_ns\":%llu,\"instantiate_ns\":%llu,\"lookup_ns\":%llu,\"first_ns\":%llu,\"warm_call_and_result_delete_ns\":%.3f,\"bytes\":%zu,\"count\":%llu,\"checksum\":%llu}\n", (unsigned long long)load, (unsigned long long)engine_time, (unsigned long long)compile, (unsigned long long)setup, (unsigned long long)instantiate, (unsigned long long)lookup, (unsigned long long)first, (double)warm/count, n, (unsigned long long)count, (unsigned long long)(n*count));
  wasmtime_component_val_delete(&arg);
  wasmtime_component_linker_delete(linker);
  wasmtime_store_delete(store);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(bytes);
  return 0;
}
