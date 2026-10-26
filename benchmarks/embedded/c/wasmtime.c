#include "common.h"
#include <wasmtime.h>

typedef struct { wasmtime_context_t *store; wasmtime_func_t function; } call_context;

static void check(wasmtime_error_t *error, wasm_trap_t *trap) {
  if (!error && !trap) return;
  wasm_byte_vec_t message;
  if (error) wasmtime_error_message(error, &message);
  else wasm_trap_message(trap, &message);
  fprintf(stderr, "wasmtime: %.*s\n", (int)message.size, message.data);
  wasm_byte_vec_delete(&message);
  if (error) wasmtime_error_delete(error);
  if (trap) wasm_trap_delete(trap);
  exit(EXIT_FAILURE);
}

static uint32_t invoke(void *opaque, const options *opt, uint64_t *elapsed) {
  call_context *context = opaque;
  wasmtime_val_t arguments[2] = {{.kind = WASMTIME_I32, .of.i32 = (int32_t)opt->seed},
                               {.kind = WASMTIME_I32, .of.i32 = (int32_t)opt->iterations}};
  wasmtime_val_t result;
  wasm_trap_t *trap = NULL;
  uint64_t start = now_ns();
  wasmtime_error_t *error = wasmtime_func_call(context->store, &context->function,
                                               arguments, 2, &result, 1, &trap);
  *elapsed = now_ns() - start;
  check(error, trap);
  if (result.kind != WASMTIME_I32) fail("run returned a non-i32 value");
  return (uint32_t)result.of.i32;
}

int main(int argc, char **argv) {
  options opt = begin(argc, argv, "compiler");
  measurements result = {0};
  size_t size;
  uint64_t start = now_ns();
  uint8_t *bytes = read_module(opt.path, &size);
  result.load = now_ns() - start;
  require_pure_module(bytes, size);

  start = now_ns();
  wasm_config_t *config = wasm_config_new();
  if (!config) fail("cannot create Wasmtime configuration");
  wasmtime_config_strategy_set(config, WASMTIME_STRATEGY_CRANELIFT);
  wasmtime_config_cranelift_opt_level_set(config, WASMTIME_OPT_LEVEL_SPEED);
  wasm_engine_t *engine = wasm_engine_new_with_config(config);
  if (!engine) fail("cannot create Wasmtime engine");
  wasmtime_store_t *store = wasmtime_store_new(engine, NULL, NULL);
  result.engine_setup = now_ns() - start;
  if (!store) fail("cannot create Wasmtime store");
  wasmtime_context_t *context = wasmtime_store_context(store);

  wasmtime_module_t *module = NULL;
  start = now_ns();
  wasmtime_error_t *error = wasmtime_module_new(engine, bytes, size, &module);
  result.compile_validate = now_ns() - start;
  check(error, NULL);

  wasmtime_instance_t instance;
  wasm_trap_t *trap = NULL;
  start = now_ns();
  error = wasmtime_instance_new(context, module, NULL, 0, &instance, &trap);
  result.instantiate = now_ns() - start;
  check(error, trap);

  wasmtime_extern_t item;
  start = now_ns();
  bool found = wasmtime_instance_export_get(context, &instance, "run", 3, &item);
  result.export_lookup = now_ns() - start;
  if (!found || item.kind != WASMTIME_EXTERN_FUNC) fail("missing run function");
  wasm_functype_t *type = wasmtime_func_type(context, &item.of.func);
  const wasm_valtype_vec_t *parameters = wasm_functype_params(type);
  const wasm_valtype_vec_t *returns = wasm_functype_results(type);
  if (parameters->size != 2 || returns->size != 1 ||
      wasm_valtype_kind(parameters->data[0]) != WASM_I32 ||
      wasm_valtype_kind(parameters->data[1]) != WASM_I32 ||
      wasm_valtype_kind(returns->data[0]) != WASM_I32)
    fail("run must have signature (i32,i32)->i32");
  wasm_functype_delete(type);
  call_context call = {context, item.of.func};
  calls(&opt, &result, &call, invoke);

  start = now_ns();
  wasmtime_extern_delete(&item);
  wasmtime_store_delete(store);
  wasmtime_module_delete(module);
  wasm_engine_delete(engine);
  free(bytes);
  result.cleanup = now_ns() - start;
  report(&opt, &result);
  return EXIT_SUCCESS;
}
