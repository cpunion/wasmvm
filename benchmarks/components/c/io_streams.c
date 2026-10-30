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
typedef struct {
  uint64_t bytes;
  int semantic;
  int fail;
  unsigned char text[64];
} output_state;
static ptrdiff_t output(void *data, const unsigned char *bytes, size_t size) {
  output_state *state = data;
  if (state->fail)
    return -5;
  if (state->semantic == 1) {
    if (state->bytes + size > sizeof(state->text))
      exit(6);
    memcpy(state->text + state->bytes, bytes, size);
  } else if (state->semantic == 0 && size) {
    const char *pattern = "0123456789abcdef";
    if (bytes[0] != (unsigned char)pattern[state->bytes % 16] ||
        bytes[size - 1] !=
            (unsigned char)pattern[(state->bytes + size - 1) % 16])
      exit(7);
  }
  state->bytes += size;
  return (ptrdiff_t)size;
}
static void verify_interfaces(wasmtime_component_instance_t *in,
                              wasmtime_context_t *ctx, output_state *state) {
  wasmtime_component_func_t f = named(in, ctx, "exercise");
  wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U32,
                                     .of.u32 = 0};
  state->semantic = 1;
  check(wasmtime_component_func_call(&f, ctx, NULL, 0, &result, 1));
  const unsigned char expected[] = {88, 89, 90, 0,   0,   0,
                                    0,  0,  33, 101, 102, 103};
  if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != 42 ||
      state->bytes != sizeof(expected) ||
      memcmp(state->text, expected, sizeof(expected)))
    exit(8);
  f = named(in, ctx, "terminal-probe");
  check(wasmtime_component_func_call(&f, ctx, NULL, 0, &result, 1));
  if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != 0)
    exit(9);
  state->bytes = 0;
  state->semantic = 0;
}
int main(int argc, char **argv) {
  if (argc != 3 && argc != 4)
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
  wasi_config_t *wasi = wasi_config_new();
  wasm_byte_vec_t input;
  wasm_byte_vec_new(&input, 8, "abcdefgh");
  wasi_config_set_stdin_bytes(wasi, &input);
  output_state state = {0};
  wasi_config_set_stdout_custom(wasi, output, &state, NULL);
  check(wasmtime_context_set_wasi(context, wasi));
  check(wasmtime_component_linker_add_wasip2(linker));
  uint64_t setup = now() - t;
  t = now();
  wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker, context, component,
                                              &instance));
  uint64_t instantiate = now() - t;
  if (argc == 4 && !strcmp(argv[3], "error")) {
    state.fail = 1;
    wasmtime_component_func_t f = named(&instance, context, "error-probe");
    wasmtime_component_val_t v = {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 0};
    check(wasmtime_component_func_call(&f, context, NULL, 0, &v, 1));
    if (v.kind != WASMTIME_COMPONENT_U32 || v.of.u32 != 1)
      return 13;
    puts("{\"error_resource_semantics\":true}");
    wasmtime_component_linker_delete(linker);
    wasmtime_store_delete(store);
    wasmtime_component_delete(component);
    wasm_engine_delete(engine);
    free(bytes);
    return 0;
  }
  if (argc == 4) {
    state.semantic = 2;
    wasmtime_component_func_t f = named(&instance, context, argv[3]);
    wasmtime_component_val_t v = {.kind = WASMTIME_COMPONENT_U32, .of.u32 = 0};
    wasmtime_error_t *e =
        wasmtime_component_func_call(&f, context, NULL, 0, &v, 1);
    if (!e) {
      fprintf(stderr, "negative probe %s did not trap\n", argv[3]);
      return 14;
    }
    wasm_byte_vec_t message;
    wasmtime_error_message(e, &message);
    fprintf(stderr, "%.*s\n", (int)message.size, message.data);
    wasm_byte_vec_delete(&message);
    wasmtime_error_delete(e);
    printf("{\"negative\":\"%s\",\"trapped\":true}\n", argv[3]);
    wasmtime_component_linker_delete(linker);
    wasmtime_store_delete(store);
    wasmtime_component_delete(component);
    wasm_engine_delete(engine);
    free(bytes);
    return 0;
  }
  verify_interfaces(&instance, context, &state);
  t = now();
  wasmtime_component_func_t function = named(&instance, context, "write-batch");
  uint64_t lookup = now() - t;
  wasmtime_component_val_t arg = {.kind = WASMTIME_COMPONENT_U32,
                                  .of.u32 = 128};
  wasmtime_component_val_t result = {.kind = WASMTIME_COMPONENT_U32,
                                     .of.u32 = 0};
  t = now();
  check(wasmtime_component_func_call(&function, context, &arg, 1, &result, 1));
  uint64_t first = now() - t;
  if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != 128)
    return 10;
  state.bytes = 0;
  t = now();
  for (uint64_t i = 0; i < count; i++) {
    check(
        wasmtime_component_func_call(&function, context, &arg, 1, &result, 1));
    if (result.kind != WASMTIME_COMPONENT_U32 || result.of.u32 != 128)
      return 11;
  }
  uint64_t warm = now() - t;
  if (state.bytes != count * 128 * 4096)
    return 12;
  printf("{\"load_ns\":%llu,\"engine_ns\":%llu,\"compile_ns\":%llu,\"store_"
         "linker_provider_ns\":%llu,\"instantiate_ns\":%llu,\"lookup_ns\":%llu,"
         "\"first_ns\":%llu,\"warm_ns_per_call\":%.3f,\"writes_per_call\":128,"
         "\"bytes_per_call\":524288,\"count\":%llu,\"checksum\":%llu}\n",
         (unsigned long long)load, (unsigned long long)engine_time,
         (unsigned long long)compile, (unsigned long long)setup,
         (unsigned long long)instantiate, (unsigned long long)lookup,
         (unsigned long long)first, (double)warm / count,
         (unsigned long long)count, (unsigned long long)state.bytes);
  wasmtime_component_linker_delete(linker);
  wasmtime_store_delete(store);
  wasmtime_component_delete(component);
  wasm_engine_delete(engine);
  free(bytes);
  return 0;
}
