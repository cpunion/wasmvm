#include "common.h"
#include <wasmedge/wasmedge.h>
#ifdef WASMVM_WASMEDGE_AOT
#include <unistd.h>

static char temporary_directory[PATH_MAX];
static char compiled_path[PATH_MAX];

static void remove_artifact(void) {
  if (*compiled_path) { unlink(compiled_path); *compiled_path = '\0'; }
  if (*temporary_directory) { rmdir(temporary_directory); *temporary_directory = '\0'; }
}
#endif

typedef struct {
  WasmEdge_ExecutorContext *executor;
  const WasmEdge_FunctionInstanceContext *function;
} call_context;

static void check(WasmEdge_Result status) {
  if (!WasmEdge_ResultOK(status)) {
    fprintf(stderr, "wasmedge: %s\n", WasmEdge_ResultGetMessage(status));
    exit(EXIT_FAILURE);
  }
}

static uint32_t invoke(void *opaque, const options *opt, uint64_t *elapsed) {
  call_context *context = opaque;
  WasmEdge_Value arguments[2] = {WasmEdge_ValueGenI32((int32_t)opt->seed),
                                WasmEdge_ValueGenI32((int32_t)opt->iterations)};
  WasmEdge_Value returned;
  uint64_t start = now_ns();
  WasmEdge_Result status = WasmEdge_ExecutorInvoke(context->executor,
                                                  context->function, arguments,
                                                  2, &returned, 1);
  *elapsed = now_ns() - start;
  check(status);
  if (!WasmEdge_ValTypeIsI32(returned.Type)) fail("run returned a non-i32 value");
  return (uint32_t)WasmEdge_ValueGetI32(returned);
}

int main(int argc, char **argv) {
#ifdef WASMVM_WASMEDGE_AOT
  options opt = begin(argc, argv, "compiler");
  const enum WasmEdge_RunMode mode = WasmEdge_RunMode_AOT;
#else
  options opt = begin(argc, argv, "interpreter");
  const enum WasmEdge_RunMode mode = WasmEdge_RunMode_Interpreter;
#endif
  measurements result = {0};
  size_t size;
  uint64_t start = now_ns();
  uint8_t *bytes = read_module(opt.path, &size);
  result.load = now_ns() - start;
  require_pure_module(bytes, size);
  if (strcmp(WasmEdge_VersionGet(), "0.17.1") != 0) fail("expected WasmEdge 0.17.1");

  start = now_ns();
  /* Keep stdout exclusively for protocol records. Every API error is checked
   * and reported to stderr below, including guest traps. */
  WasmEdge_LogOff();
  WasmEdge_ConfigureContext *config = WasmEdge_ConfigureCreate();
  if (!config) fail("cannot create WasmEdge configuration");
  WasmEdge_ConfigureSetRunMode(config, WasmEdge_RunMode_Interpreter);
#ifdef WASMVM_WASMEDGE_AOT
  WasmEdge_ConfigureSetRunMode(config, mode);
  WasmEdge_ConfigureCompilerSetOptimizationLevel(config, WasmEdge_CompilerOptimizationLevel_O3);
  WasmEdge_ConfigureCompilerSetOutputFormat(config, WasmEdge_CompilerOutputFormat_Native);
  WasmEdge_CompilerContext *compiler = WasmEdge_CompilerCreate(config);
  if (!compiler) fail("WasmEdge AOT compiler is unavailable in this SDK");
#endif
  WasmEdge_LoaderContext *loader = WasmEdge_LoaderCreate(config);
  WasmEdge_ValidatorContext *validator = WasmEdge_ValidatorCreate(config);
  WasmEdge_ExecutorContext *executor = WasmEdge_ExecutorCreate(config, NULL);
  WasmEdge_StoreContext *store = WasmEdge_StoreCreate();
  result.engine_setup = now_ns() - start;
  if (!loader || !validator || !executor || !store) fail("cannot create WasmEdge contexts");
  if (WasmEdge_ConfigureGetRunMode(config) != mode)
    fail("WasmEdge execution-mode selection failed");

  WasmEdge_ASTModuleContext *module = NULL;
  WasmEdge_Bytes binary = WasmEdge_BytesWrap(bytes, (uint32_t)size);
#ifdef WASMVM_WASMEDGE_AOT
  /* A fresh native image for each process, with checked compiler and loader
   * results. AOT mode loads its native code rather than an interpreter copy.
   * This API also validates and writes an artifact; do not call it pure codegen. */
  const char *tmp = getenv("TMPDIR");
  if (!tmp || !*tmp) tmp = "/tmp";
  int written = snprintf(temporary_directory, sizeof(temporary_directory),
                         "%s/wasmvm-wasmedge-XXXXXX", tmp);
  if (written < 0 || (size_t)written >= sizeof(temporary_directory) ||
      !mkdtemp(temporary_directory)) fail("cannot create AOT artifact directory");
  if (atexit(remove_artifact) != 0) fail("cannot register AOT artifact cleanup");
  written = snprintf(compiled_path, sizeof(compiled_path), "%s/module.so", temporary_directory);
  if (written < 0 || (size_t)written >= sizeof(compiled_path)) fail("AOT artifact path exceeds limit");
  start = now_ns();
  WasmEdge_Result status = WasmEdge_CompilerCompileFromBytes(compiler, binary, compiled_path);
  result.native_compile = now_ns() - start;
  check(status);
  start = now_ns();
  status = WasmEdge_LoaderParseFromFile(loader, &module, compiled_path);
  result.artifact_load = result.parse = now_ns() - start;
  result.separate_native_compile = true;
#else
  start = now_ns();
  WasmEdge_Result status = WasmEdge_LoaderParseFromBytes(loader, &module, binary);
  result.parse = now_ns() - start;
#endif
  check(status);
  start = now_ns();
  status = WasmEdge_ValidatorValidate(validator, module);
  result.validate = now_ns() - start;
  check(status);
  result.separate_validation = true;
  result.compile_validate = result.native_compile + result.parse + result.validate;

  WasmEdge_ModuleInstanceContext *instance = NULL;
  start = now_ns();
  status = WasmEdge_ExecutorInstantiate(executor, &instance, store, module);
  result.instantiate = now_ns() - start;
  check(status);

  WasmEdge_String name = WasmEdge_StringWrap("run", 3);
  start = now_ns();
  const WasmEdge_FunctionInstanceContext *function =
      WasmEdge_ModuleInstanceFindFunction(instance, name);
  result.export_lookup = now_ns() - start;
  if (!function) fail("missing run function");
  const WasmEdge_FunctionTypeContext *type = WasmEdge_FunctionInstanceGetFunctionType(function);
  WasmEdge_ValType parameters[2], returns[1];
  if (WasmEdge_FunctionTypeGetParametersLength(type) != 2 ||
      WasmEdge_FunctionTypeGetReturnsLength(type) != 1)
    fail("run must have signature (i32,i32)->i32");
  WasmEdge_FunctionTypeGetParameters(type, parameters, 2);
  WasmEdge_FunctionTypeGetReturns(type, returns, 1);
  if (!WasmEdge_ValTypeIsI32(parameters[0]) || !WasmEdge_ValTypeIsI32(parameters[1]) ||
      !WasmEdge_ValTypeIsI32(returns[0]))
    fail("run must have signature (i32,i32)->i32");
  call_context call = {executor, function};
  calls(&opt, &result, &call, invoke);

  start = now_ns();
  WasmEdge_ModuleInstanceDelete(instance);
  WasmEdge_ASTModuleDelete(module);
  WasmEdge_StoreDelete(store);
  WasmEdge_ExecutorDelete(executor);
  WasmEdge_ValidatorDelete(validator);
  WasmEdge_LoaderDelete(loader);
#ifdef WASMVM_WASMEDGE_AOT
  WasmEdge_CompilerDelete(compiler);
  remove_artifact();
#endif
  WasmEdge_ConfigureDelete(config);
  free(bytes);
  result.cleanup = now_ns() - start;
  report(&opt, &result);
  return EXIT_SUCCESS;
}
