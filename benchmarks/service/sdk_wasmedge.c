//go:build !llgo && wasmedge

#include "sdk_common.h"
#include <wasmedge/wasmedge.h>

typedef struct {
    WasmEdge_ExecutorContext *executor;
    WasmEdge_StoreContext *store;
    WasmEdge_ModuleInstanceContext *instance;
    const WasmEdge_FunctionInstanceContext *function;
    char error[512];
} worker;
typedef struct {
    WasmEdge_ConfigureContext *config;
    WasmEdge_LoaderContext *loader;
    WasmEdge_ValidatorContext *validator;
    WasmEdge_ASTModuleContext *module;
    worker *workers;
    unsigned count;
    uint64_t phases[4];
    char error[512];
} owner;

const char *sdk_name(void) { return "wasmedge"; }
const char *sdk_version(void) { return WasmEdge_VersionGet(); }
static int failure(char *out, WasmEdge_Result status) {
    if (WasmEdge_ResultOK(status)) return 0;
    const char *message = WasmEdge_ResultGetMessage(status);
    sdk_message(out, message, strlen(message));
    return 1;
}
const char *sdk_error(void *p, unsigned index) {
    owner *o = p;
    return index < o->count ? o->workers[index].error : o->error;
}
uint64_t sdk_phase(void *p, unsigned phase) { return ((owner *)p)->phases[phase]; }
void sdk_close(void *p) {
    owner *o = p;
    if (!o) return;
    for (unsigned i = 0; i < o->count; ++i) {
        worker *w = &o->workers[i];
        WasmEdge_ModuleInstanceDelete(w->instance);
        WasmEdge_StoreDelete(w->store);
        WasmEdge_ExecutorDelete(w->executor);
    }
    WasmEdge_ASTModuleDelete(o->module);
    WasmEdge_ValidatorDelete(o->validator);
    WasmEdge_LoaderDelete(o->loader);
    WasmEdge_ConfigureDelete(o->config);
    free(o->workers); free(o);
}
void *sdk_open(const uint8_t *bytes, size_t size, unsigned count) {
    owner *o = calloc(1, sizeof(*o));
    if (!o) return NULL;
    if (size > UINT32_MAX) { strcpy(o->error,"module exceeds SDK byte limit"); return o; }
    o->workers = calloc(count, sizeof(worker));
    if (!o->workers) { strcpy(o->error,"worker allocation failed"); return o; }
    o->count = count;
    uint64_t start = sdk_now();
    WasmEdge_LogOff();
    o->config = WasmEdge_ConfigureCreate();
    if (!o->config) { strcpy(o->error,"config allocation failed"); return o; }
    WasmEdge_ConfigureSetRunMode(o->config, WasmEdge_RunMode_Interpreter);
    o->loader = WasmEdge_LoaderCreate(o->config);
    o->validator = WasmEdge_ValidatorCreate(o->config);
    if (!o->loader || !o->validator) { strcpy(o->error,"loader/validator allocation failed"); return o; }
    for (unsigned i = 0; i < count; ++i) {
        worker *w = &o->workers[i];
        w->executor = WasmEdge_ExecutorCreate(o->config, NULL);
        w->store = WasmEdge_StoreCreate();
        if (!w->executor || !w->store) { strcpy(o->error,"worker context allocation failed"); return o; }
    }
    o->phases[0] = sdk_now() - start;
    start = sdk_now();
    WasmEdge_Result status = WasmEdge_LoaderParseFromBytes(o->loader, &o->module, WasmEdge_BytesWrap(bytes, (uint32_t)size));
    if (WasmEdge_ResultOK(status)) status = WasmEdge_ValidatorValidate(o->validator, o->module);
    o->phases[1] = sdk_now() - start;
    if (failure(o->error, status)) return o;
    for (unsigned i = 0; i < count; ++i) {
        worker *w = &o->workers[i];
        start = sdk_now();
        status = WasmEdge_ExecutorInstantiate(w->executor, &w->instance, w->store, o->module);
        o->phases[2] += sdk_now() - start;
        if (failure(o->error, status)) return o;
        start = sdk_now();
        w->function = WasmEdge_ModuleInstanceFindFunction(w->instance, WasmEdge_StringWrap("run", 3));
        o->phases[3] += sdk_now() - start;
        if (!w->function) { strcpy(o->error,"missing run function"); return o; }
        const WasmEdge_FunctionTypeContext *type = WasmEdge_FunctionInstanceGetFunctionType(w->function);
        if (WasmEdge_FunctionTypeGetParametersLength(type) != 2 || WasmEdge_FunctionTypeGetReturnsLength(type) != 1) {
            strcpy(o->error,"run must have signature (i32,i32)->i32"); return o;
        }
        WasmEdge_ValType params[2], results[1];
        WasmEdge_FunctionTypeGetParameters(type, params, 2);
        WasmEdge_FunctionTypeGetReturns(type, results, 1);
        if (!WasmEdge_ValTypeIsI32(params[0]) || !WasmEdge_ValTypeIsI32(params[1]) || !WasmEdge_ValTypeIsI32(results[0])) {
            strcpy(o->error,"run must have signature (i32,i32)->i32"); return o;
        }
    }
    return o;
}
uint64_t sdk_call(void *p, unsigned index, uint32_t seed, uint32_t count) {
    worker *w = &((owner *)p)->workers[index];
    WasmEdge_Value args[2] = {WasmEdge_ValueGenI32((int32_t)seed), WasmEdge_ValueGenI32((int32_t)count)}, result;
    if (failure(w->error, WasmEdge_ExecutorInvoke(w->executor, w->function, args, 2, &result, 1))) return SDK_FAILURE;
    if (!WasmEdge_ValTypeIsI32(result.Type)) { strcpy(w->error,"non-i32 result"); return SDK_FAILURE; }
    return (uint32_t)WasmEdge_ValueGetI32(result);
}
