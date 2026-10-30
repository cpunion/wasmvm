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
  wasm_byte_vec_t m;
  wasmtime_error_message(e, &m);
  fprintf(stderr, "%.*s\n", (int)m.size, m.data);
  wasm_byte_vec_delete(&m);
  wasmtime_error_delete(e);
  exit(1);
}
static uint8_t *read_file(const char *path, size_t *n) {
  FILE *f = fopen(path, "rb");
  if (!f)
    exit(3);
  fseek(f, 0, SEEK_END);
  long size = ftell(f);
  rewind(f);
  uint8_t *p = malloc((size_t)size);
  if (!p || fread(p, 1, (size_t)size, f) != (size_t)size)
    exit(4);
  fclose(f);
  *n = (size_t)size;
  return p;
}
static wasmtime_error_t *mark(void *data, wasmtime_context_t *ctx,
                              const wasmtime_component_func_type_t *type,
                              wasmtime_component_val_t *args, size_t argc,
                              wasmtime_component_val_t *results,
                              size_t resultc) {
  (void)ctx;
  (void)type;
  (void)args;
  (void)results;
  if (argc || resultc)
    exit(5);
  (*(unsigned *)data)++;
  return NULL;
}
int main(int argc, char **argv) {
  if (argc != 5)
    return 2;
  uint64_t count = strtoull(argv[4], NULL, 10);
  if (!count)
    return 2;
  uint64_t t = now();
  size_t component_size, module_size;
  uint8_t *component_bytes = read_file(argv[1], &component_size),
          *module_bytes = read_file(argv[2], &module_size);
  uint64_t load = now() - t;
  t = now();
  wasm_config_t *config = wasm_config_new();
  wasmtime_config_wasm_memory64_set(config, true);
  wasmtime_config_wasm_gc_set(config, true);
  wasm_engine_t *engine = wasm_engine_new_with_config(config);
  uint64_t engine_time = now() - t;
  t = now();
  wasmtime_module_t *module = NULL;
  check(wasmtime_module_new(engine, module_bytes, module_size, &module));
  uint64_t core_compile = now() - t;
  t = now();
  wasmtime_component_t *component = NULL;
  check(wasmtime_component_new(engine, component_bytes, component_size,
                               &component));
  uint64_t component_compile = now() - t;
  t = now();
  wasmtime_store_t *store = wasmtime_store_new(engine, NULL, NULL);
  wasmtime_context_t *ctx = wasmtime_store_context(store);
  wasmtime_component_linker_t *linker = wasmtime_component_linker_new(engine);
  wasmtime_component_linker_instance_t *root =
      wasmtime_component_linker_root(linker);
  unsigned marks = 0;
  int add = !strcmp(argv[3], "add"), get = !strcmp(argv[3], "get");
  const char *module_name = add ? "math" : "module";
  check(wasmtime_component_linker_instance_add_module(
      root, module_name, strlen(module_name), module));
  if (add)
    check(wasmtime_component_linker_instance_add_func(root, "mark", 4, mark,
                                                      &marks, NULL));
  wasmtime_component_linker_instance_delete(root);
  wasmtime_module_delete(module);
  uint64_t setup = now() - t;
  t = now();
  wasmtime_component_instance_t instance;
  check(
      wasmtime_component_linker_instantiate(linker, ctx, component, &instance));
  uint64_t instantiate = now() - t;
  if (add && marks != 1)
    return 6;
  t = now();
  wasmtime_component_export_index_t *index =
      wasmtime_component_instance_get_export_index(&instance, ctx, NULL,
                                                   argv[3], strlen(argv[3]));
  wasmtime_component_func_t function;
  if (!index ||
      !wasmtime_component_instance_get_func(&instance, ctx, index, &function))
    return 7;
  wasmtime_component_export_index_delete(index);
  uint64_t lookup = now() - t;
  wasmtime_component_val_t args[2] = {{.kind = WASMTIME_COMPONENT_U32,
                                       .of.u32 = add ? 37 : 41},
                                      {.kind = WASMTIME_COMPONENT_U32,
                                       .of.u32 = 5}},
                           result = {.kind = WASMTIME_COMPONENT_U32,
                                     .of.u32 = 0};
  size_t arguments = get ? 0 : add ? 2 : 1;
  t = now();
  check(wasmtime_component_func_call(&function, ctx, args, arguments, &result,
                                     1));
  uint64_t first = now() - t;
  if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != 42)
    return 8;
  uint64_t checksum = 0;
  t = now();
  for (uint64_t i = 0; i < count; i++) {
    args[0].of.u32 = (uint32_t)i;
    check(wasmtime_component_func_call(&function, ctx, args, arguments, &result,
                                       1));
    uint32_t want = get ? 42 : (uint32_t)i + (add ? 5 : 1);
    if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != want)
      return 9;
    checksum += result.of.u32;
  }
  uint64_t warm = now() - t;
  printf("{\"load_ns\":%llu,\"engine_ns\":%llu,\"core_compile_ns\":%llu,"
         "\"component_compile_ns\":%llu,\"store_linker_ns\":%llu,\"instantiate_"
         "ns\":%llu,\"lookup_ns\":%llu,\"first_ns\":%llu,\"warm_ns_per_call\":%"
         ".3f,\"count\":%llu,\"checksum\":%llu}\n",
         (unsigned long long)load, (unsigned long long)engine_time,
         (unsigned long long)core_compile,
         (unsigned long long)component_compile, (unsigned long long)setup,
         (unsigned long long)instantiate, (unsigned long long)lookup,
         (unsigned long long)first, (double)warm / count,
         (unsigned long long)count, (unsigned long long)checksum);
  wasmtime_component_linker_delete(linker);
  wasmtime_store_delete(store);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(component_bytes);
  free(module_bytes);
  return 0;
}
