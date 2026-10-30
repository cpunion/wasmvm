#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <wasmtime.h>

static uint64_t now(void) {
  struct timespec t;
  clock_gettime(CLOCK_MONOTONIC, &t);
  return (uint64_t)t.tv_sec * 1000000000 + (uint64_t)t.tv_nsec;
}
static void check(wasmtime_error_t *e) {
  if (!e)
    return;
  wasm_byte_vec_t msg;
  wasmtime_error_message(e, &msg);
  fprintf(stderr, "%.*s\n", (int)msg.size, msg.data);
  wasm_byte_vec_delete(&msg);
  wasmtime_error_delete(e);
  exit(1);
}
static wasmtime_component_func_t named(wasmtime_component_instance_t *in,
                                       wasmtime_context_t *ctx,
                                       const char *name) {
  wasmtime_component_export_index_t *index =
      wasmtime_component_instance_get_export_index(in, ctx, NULL, name,
                                                   strlen(name));
  wasmtime_component_func_t f;
  if (!index || !wasmtime_component_instance_get_func(in, ctx, index, &f))
    exit(5);
  wasmtime_component_export_index_delete(index);
  return f;
}
static void verify_interfaces(wasmtime_component_instance_t *in,
                              wasmtime_context_t *ctx) {
  const char *names[] = {"now", "resolution"};
  for (size_t i = 0; i < 2; i++) {
    wasmtime_component_func_t f = named(in, ctx, names[i]);
    wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U8,
                                       .of.u8 = 0};
    check(wasmtime_component_func_call(&f, ctx, NULL, 0, &result, 1));
    if (result.kind != WASMTIME_COMPONENT_RECORD || result.of.record.size != 2)
      exit(6);
    wasmtime_component_val_t *sec = &result.of.record.data[0].val,
                             *nano = &result.of.record.data[1].val;
    if (sec->kind != WASMTIME_COMPONENT_U64 ||
        nano->kind != WASMTIME_COMPONENT_U32 || nano->of.u32 >= 1000000000)
      exit(7);
    if (i == 1 && sec->of.u64 == 0 && nano->of.u32 == 0)
      exit(8);
    wasmtime_component_val_delete(&result);
  }
}
int main(int argc, char **argv) {
  if (argc != 3)
    return 2;
  uint64_t count = strtoull(argv[2], NULL, 10), t = now();
  if (!count)
    return 2;
  FILE *file = fopen(argv[1], "rb");
  if (!file)
    return 3;
  fseek(file, 0, SEEK_END);
  long size = ftell(file);
  rewind(file);
  uint8_t *bytes = malloc((size_t)size);
  if (!bytes || fread(bytes, 1, (size_t)size, file) != (size_t)size)
    return 4;
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
  check(wasmtime_context_set_wasi(context, wasi_config_new()));
  check(wasmtime_component_linker_add_wasip2(linker));
  uint64_t setup = now() - t;
  t = now();
  wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker, context, component,
                                              &instance));
  uint64_t instantiate = now() - t;
  t = now();
  wasmtime_component_func_t function = named(&instance, context, "now-seconds");
  uint64_t lookup = now() - t;
  wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U64,
                                     .of.u64 = 0};
  t = now();
  check(wasmtime_component_func_call(&function, context, NULL, 0, &result, 1));
  uint64_t first = now() - t;
  if (result.kind != WASMTIME_COMPONENT_U64 || result.of.u64 == 0)
    return 10;
  verify_interfaces(&instance, context);
  uint64_t checksum = 0;
  t = now();
  for (uint64_t i = 0; i < count; i++) {
    check(
        wasmtime_component_func_call(&function, context, NULL, 0, &result, 1));
    if (result.kind != WASMTIME_COMPONENT_U64 || result.of.u64 == 0)
      return 11;
    checksum += result.of.u64;
  }
  uint64_t warm = now() - t;
  printf(
      "{\"load_ns\":%llu,\"engine_ns\":%llu,\"compile_ns\":%llu,\"store_linker_"
      "provider_ns\":%llu,\"instantiate_ns\":%llu,\"lookup_ns\":%llu,\"first_"
      "ns\":%llu,\"warm_ns_per_call\":%.3f,\"count\":%llu,\"checksum\":%llu}\n",
      (unsigned long long)load, (unsigned long long)engine_time,
      (unsigned long long)compile, (unsigned long long)setup,
      (unsigned long long)instantiate, (unsigned long long)lookup,
      (unsigned long long)first, (double)warm / count,
      (unsigned long long)count, (unsigned long long)checksum);
  wasmtime_component_linker_delete(linker);
  wasmtime_store_delete(store);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(bytes);
  return 0;
}
