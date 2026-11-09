//go:build !llgo && wasmtime

#include "sdk_common.h"
#include <wasmtime.h>

typedef struct {
    wasmtime_store_t *store;
    wasmtime_func_t function;
    char error[512];
} worker;
typedef struct {
    wasm_engine_t *engine;
    wasmtime_module_t *module;
    worker *workers;
    unsigned count;
    uint64_t phases[4];
    char error[512];
} owner;

const char *sdk_name(void) { return "wasmtime"; }
const char *sdk_version(void) { return WASMTIME_VERSION; }
static int failure(char *out, wasmtime_error_t *error, wasm_trap_t *trap) {
    if (!error && !trap) return 0;
    wasm_byte_vec_t message;
    if (error) wasmtime_error_message(error, &message); else wasm_trap_message(trap, &message);
    sdk_message(out, message.data, message.size);
    wasm_byte_vec_delete(&message);
    if (error) wasmtime_error_delete(error);
    if (trap) wasm_trap_delete(trap);
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
    for (unsigned i = 0; i < o->count; ++i) if (o->workers[i].store) wasmtime_store_delete(o->workers[i].store);
    if (o->module) wasmtime_module_delete(o->module);
    if (o->engine) wasm_engine_delete(o->engine);
    free(o->workers); free(o);
}
void *sdk_open(const uint8_t *bytes, size_t size, unsigned count) {
    owner *o = calloc(1, sizeof(*o));
    if (!o) return NULL;
    o->workers = calloc(count, sizeof(worker));
    if (!o->workers) { strcpy(o->error,"worker allocation failed"); return o; }
    o->count = count;
    uint64_t start = sdk_now();
    wasm_config_t *config = wasm_config_new();
    if (!config) { strcpy(o->error,"config allocation failed"); return o; }
    wasmtime_config_strategy_set(config, WASMTIME_STRATEGY_CRANELIFT);
    wasmtime_config_cranelift_opt_level_set(config, WASMTIME_OPT_LEVEL_SPEED);
    o->engine = wasm_engine_new_with_config(config);
    if (!o->engine) { strcpy(o->error,"engine allocation failed"); return o; }
    for (unsigned i=0;i<count;++i) {
        o->workers[i].store = wasmtime_store_new(o->engine,NULL,NULL);
        if (!o->workers[i].store) { strcpy(o->error,"store allocation failed"); return o; }
    }
    o->phases[0] = sdk_now()-start;
    start = sdk_now();
    wasmtime_error_t *error = wasmtime_module_new(o->engine,bytes,size,&o->module);
    o->phases[1] = sdk_now()-start;
    if (failure(o->error,error,NULL)) return o;
    for (unsigned i=0;i<count;++i) {
        wasmtime_context_t *context = wasmtime_store_context(o->workers[i].store);
        wasmtime_instance_t instance;
        wasm_trap_t *trap = NULL;
        start = sdk_now();
        error = wasmtime_instance_new(context,o->module,NULL,0,&instance,&trap);
        o->phases[2] += sdk_now()-start;
        if (failure(o->error,error,trap)) return o;
        wasmtime_extern_t item;
        start = sdk_now();
        int found = wasmtime_instance_export_get(context,&instance,"run",3,&item);
        o->phases[3] += sdk_now()-start;
        if (!found || item.kind != WASMTIME_EXTERN_FUNC) {
            if (found) wasmtime_extern_delete(&item);
            strcpy(o->error,"missing run function"); return o;
        }
        wasm_functype_t *type = wasmtime_func_type(context,&item.of.func);
        const wasm_valtype_vec_t *params=wasm_functype_params(type),*results=wasm_functype_results(type);
        int valid=params->size==2 && results->size==1 && wasm_valtype_kind(params->data[0])==WASM_I32 && wasm_valtype_kind(params->data[1])==WASM_I32 && wasm_valtype_kind(results->data[0])==WASM_I32;
        o->workers[i].function=item.of.func;
        wasm_functype_delete(type); wasmtime_extern_delete(&item);
        if (!valid) { strcpy(o->error,"run must have signature (i32,i32)->i32"); return o; }
    }
    return o;
}
uint64_t sdk_call(void *p,unsigned index,uint32_t seed,uint32_t count) {
    owner *o=p; worker *w=&o->workers[index];
    wasmtime_val_t args[2]={{.kind=WASMTIME_I32,.of.i32=(int32_t)seed},{.kind=WASMTIME_I32,.of.i32=(int32_t)count}},result;
    wasm_trap_t *trap=NULL;
    wasmtime_error_t *error=wasmtime_func_call(wasmtime_store_context(w->store),&w->function,args,2,&result,1,&trap);
    if (failure(w->error,error,trap)) return SDK_FAILURE;
    if (result.kind != WASMTIME_I32) { strcpy(w->error,"non-i32 result"); return SDK_FAILURE; }
    return (uint32_t)result.of.i32;
}
