#include <wasmtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <stdint.h>

static uint64_t now(void) {
  struct timespec t; clock_gettime(CLOCK_MONOTONIC, &t);
  return (uint64_t)t.tv_sec * 1000000000 + (uint64_t)t.tv_nsec;
}
static void check(wasmtime_error_t *e) {
  if (!e) return;
  wasm_byte_vec_t m; wasmtime_error_message(e, &m);
  fprintf(stderr, "%.*s\n", (int)m.size, m.data);
  wasm_byte_vec_delete(&m); wasmtime_error_delete(e); exit(1);
}
static wasmtime_component_func_t lookup(wasmtime_component_instance_t *instance, wasmtime_context_t *ctx, const char *name) {
  wasmtime_component_export_index_t *index = wasmtime_component_instance_get_export_index(instance, ctx, NULL, name, strlen(name));
  wasmtime_component_func_t f;
  if (!index || !wasmtime_component_instance_get_func(instance, ctx, index, &f)) exit(5);
  wasmtime_component_export_index_delete(index); return f;
}
static wasmtime_component_val_t call(wasmtime_component_func_t *f, wasmtime_context_t *ctx, wasmtime_component_val_t *args, size_t count) {
  wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 0};
  check(wasmtime_component_func_call(f, ctx, args, count, &result, 1)); return result;
}
static void scalar(wasmtime_component_val_t v, uint32_t expected) {
  if (v.kind != WASMTIME_COMPONENT_U32 || v.of.u32 != expected) exit(6);
}
static void resource(wasmtime_component_val_t v) {
  if (v.kind != WASMTIME_COMPONENT_RESOURCE || !wasmtime_component_resource_any_owned(v.of.resource)) exit(7);
}
int main(int argc, char **argv) {
  if (argc != 3) return 2;
  uint64_t count = strtoull(argv[2], NULL, 10); if (!count) return 2;
  FILE *file = fopen(argv[1], "rb"); if (!file) return 3;
  fseek(file, 0, SEEK_END); long size = ftell(file); rewind(file);
  uint8_t *bytes = malloc((size_t)size);
  if (!bytes || fread(bytes, 1, (size_t)size, file) != (size_t)size) return 4;
  fclose(file);
  wasm_engine_t *engine = wasm_engine_new();
  wasmtime_component_t *component = NULL;
  check(wasmtime_component_new(engine, bytes, (size_t)size, &component));
  wasmtime_store_t *store = wasmtime_store_new(engine, NULL, NULL);
  wasmtime_context_t *ctx = wasmtime_store_context(store);
  wasmtime_component_linker_t *linker = wasmtime_component_linker_new(engine);
  wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker, ctx, component, &instance));
  wasmtime_component_func_t make = lookup(&instance, ctx, "make"), get = lookup(&instance, ctx, "get"),
    pass = lookup(&instance, ctx, "pass"), consume = lookup(&instance, ctx, "consume"), total = lookup(&instance, ctx, "total");
  wasmtime_component_val_t n = {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 42};
  wasmtime_component_val_t value = call(&make, ctx, &n, 1); resource(value);
  scalar(call(&get, ctx, &value, 1), 42);
  wasmtime_component_val_t moved = call(&pass, ctx, &value, 1); resource(moved);
  scalar(call(&consume, ctx, &moved, 1), 42);
  wasmtime_component_resource_any_delete(value.of.resource);
  wasmtime_component_resource_any_delete(moved.of.resource);
  scalar(call(&total, ctx, NULL, 0), 42);
  n.of.u32 = 7; value = call(&make, ctx, &n, 1);
  check(wasmtime_component_resource_any_drop(ctx, value.of.resource));
  wasmtime_component_resource_any_delete(value.of.resource);
  scalar(call(&total, ctx, NULL, 0), 49);
  n.of.u32 = 1;
  uint64_t start = now();
  for (uint64_t i = 0; i < count; i++) {
    value = call(&make, ctx, &n, 1); resource(value);
    scalar(call(&get, ctx, &value, 1), 1);
    moved = call(&pass, ctx, &value, 1); resource(moved);
    scalar(call(&consume, ctx, &moved, 1), 1);
    wasmtime_component_resource_any_delete(value.of.resource);
    wasmtime_component_resource_any_delete(moved.of.resource);
  }
  uint64_t elapsed = now() - start;
  scalar(call(&total, ctx, NULL, 0), (uint32_t)(49 + count));
  printf("{\"semantic\":true,\"count\":%llu,\"warm_ns_per_lifetime\":%.3f,\"calls_per_lifetime\":4,\"checksum\":%u}\n", (unsigned long long)count, (double)elapsed/count, (uint32_t)(49+count));
  /* A failed canonical resource lower locks down the reference instance.
     Keep the negative check after every successful semantic/timing assertion. */
  n.of.u32 = 42; value = call(&make, ctx, &n, 1);
  moved = call(&pass, ctx, &value, 1);
  wasmtime_component_val_t ignored = {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 0};
  wasmtime_error_t *stale = wasmtime_component_func_call(&get, ctx, &value, 1, &ignored, 1);
  if (!stale) return 8; wasmtime_error_delete(stale);
  wasmtime_component_resource_any_delete(value.of.resource);
  wasmtime_component_resource_any_delete(moved.of.resource);
  wasmtime_component_linker_delete(linker); wasmtime_store_delete(store);
  wasmtime_component_delete(component); wasm_engine_delete(engine); free(bytes); return 0;
}
