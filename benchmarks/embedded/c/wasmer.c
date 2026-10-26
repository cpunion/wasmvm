#include "common.h"
#include <wasmer.h>

static void last_error(const char *phase) {
  int size = wasmer_last_error_length();
  char *message = size > 0 ? malloc((size_t)size) : NULL;
  if (message) {
    wasmer_last_error_message(message, size);
    fprintf(stderr, "wasmer %s: %s\n", phase, message);
    free(message);
  }
  fail(phase);
}

static void check_trap(wasm_trap_t *trap) {
  if (!trap) return;
  wasm_message_t message;
  wasm_trap_message(trap, &message);
  fprintf(stderr, "wasmer trap: %.*s\n", (int)message.size, message.data);
  wasm_byte_vec_delete(&message);
  wasm_trap_delete(trap);
  exit(EXIT_FAILURE);
}

static uint32_t invoke(void *opaque, const options *opt, uint64_t *elapsed) {
  wasm_func_t *function = opaque;
  wasm_val_t arguments[2] = {{.kind = WASM_I32, .of.i32 = (int32_t)opt->seed},
                           {.kind = WASM_I32, .of.i32 = (int32_t)opt->iterations}};
  wasm_val_t returned = {.kind = WASM_I32};
  wasm_val_vec_t args = {2, arguments}, returns = {1, &returned};
  uint64_t start = now_ns();
  wasm_trap_t *trap = wasm_func_call(function, &args, &returns);
  *elapsed = now_ns() - start;
  check_trap(trap);
  if (returned.kind != WASM_I32) fail("run returned a non-i32 value");
  return (uint32_t)returned.of.i32;
}

int main(int argc, char **argv) {
  options opt = begin(argc, argv, "compiler");
  measurements result = {0};
  size_t size;
  uint64_t start = now_ns();
  uint8_t *bytes = read_module(opt.path, &size);
  result.load = now_ns() - start;
  require_pure_module(bytes, size);
  if (strcmp(wasmer_version(), "7.3.0") != 0) fail("expected Wasmer 7.3.0");

  start = now_ns();
  wasm_config_t *config = wasm_config_new();
  if (!config) last_error("create config");
  wasm_config_set_backend(config, CRANELIFT);
  wasm_engine_t *engine = wasm_engine_new_with_config(config);
  if (!engine) last_error("create engine");
  wasm_store_t *store = wasm_store_new(engine);
  result.engine_setup = now_ns() - start;
  if (!store) last_error("create store");

  wasm_byte_vec_t binary = {size, (char *)bytes};
  start = now_ns();
  wasm_module_t *module = wasm_module_new(store, &binary);
  result.compile_validate = now_ns() - start;
  if (!module) last_error("compile and validate");

  wasm_extern_vec_t imports = {0, NULL};
  wasm_trap_t *trap = NULL;
  start = now_ns();
  wasm_instance_t *instance = wasm_instance_new(store, module, &imports, &trap);
  result.instantiate = now_ns() - start;
  check_trap(trap);
  if (!instance) last_error("instantiate");

  wasm_exporttype_vec_t names;
  wasm_extern_vec_t exports;
  wasm_func_t *function = NULL;
  start = now_ns();
  wasm_module_exports(module, &names);
  wasm_instance_exports(instance, &exports);
  if (names.size != exports.size) fail("inconsistent Wasmer export vectors");
  for (size_t i = 0; i < names.size; ++i) {
    const wasm_name_t *name = wasm_exporttype_name(names.data[i]);
    if (name->size == 3 && memcmp(name->data, "run", 3) == 0)
      function = wasm_extern_as_func(exports.data[i]);
  }
  result.export_lookup = now_ns() - start;
  if (!function) fail("missing run function");
  wasm_functype_t *type = wasm_func_type(function);
  const wasm_valtype_vec_t *parameters = wasm_functype_params(type);
  const wasm_valtype_vec_t *returns = wasm_functype_results(type);
  if (parameters->size != 2 || returns->size != 1 ||
      wasm_valtype_kind(parameters->data[0]) != WASM_I32 ||
      wasm_valtype_kind(parameters->data[1]) != WASM_I32 ||
      wasm_valtype_kind(returns->data[0]) != WASM_I32)
    fail("run must have signature (i32,i32)->i32");
  wasm_functype_delete(type);
  calls(&opt, &result, function, invoke);

  start = now_ns();
  wasm_extern_vec_delete(&exports);
  wasm_exporttype_vec_delete(&names);
  wasm_instance_delete(instance);
  wasm_module_delete(module);
  wasm_store_delete(store);
  wasm_engine_delete(engine);
  free(bytes);
  result.cleanup = now_ns() - start;
  report(&opt, &result);
  return EXIT_SUCCESS;
}
