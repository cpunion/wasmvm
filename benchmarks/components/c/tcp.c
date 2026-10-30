#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <wasmtime.h>

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
// Diagnostic only: the pinned C API has no setter to enable TCP permission.
int main(int argc, char **argv) {
  if (argc != 2)
    return 2;
  FILE *file = fopen(argv[1], "rb");
  if (!file)
    return 3;
  fseek(file, 0, SEEK_END);
  long size = ftell(file);
  rewind(file);
  if (size < 0)
    return 4;
  uint8_t *bytes = malloc((size_t)size);
  if (!bytes || fread(bytes, 1, (size_t)size, file) != (size_t)size)
    return 4;
  fclose(file);
  wasm_engine_t *engine = wasm_engine_new();
  wasmtime_component_t *component = NULL;
  check(wasmtime_component_new(engine, bytes, (size_t)size, &component));
  wasmtime_store_t *store = wasmtime_store_new(engine, NULL, NULL);
  wasmtime_context_t *context = wasmtime_store_context(store);
  wasmtime_component_linker_t *linker = wasmtime_component_linker_new(engine);
  wasi_config_t *wasi = wasi_config_new();
  wasi_config_inherit_network(wasi);
  check(wasmtime_context_set_wasi(context, wasi));
  check(wasmtime_component_linker_add_wasip2(linker));
  wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker, context, component,
                                              &instance));
  wasmtime_component_func_t probe = named(&instance, context, "create-probe");
  wasmtime_component_val_t family = {.kind = WASMTIME_COMPONENT_U32,
                                     .of.u32 = 0};
  wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U32,
                                     .of.u32 = 0};
  check(wasmtime_component_func_call(&probe, context, &family, 1, &result, 1));
  if (result.kind != WASMTIME_COMPONENT_U32)
    return 5;
  int status = result.of.u32 ? 11 : 0;
  if (status)
    fprintf(stderr, "TCP creation error-code=%u\n", result.of.u32 - 1);
  wasmtime_component_linker_delete(linker);
  wasmtime_store_delete(store);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(bytes);
  return status;
}
