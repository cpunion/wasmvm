//go:build !llgo && wasmer

#include "sdk_common.h"
#include <wasmer.h>

typedef struct {
    wasm_store_t *store;
    wasm_instance_t *instance;
    wasm_extern_vec_t exports;
    wasm_func_t *function;
    char error[512];
} worker;
typedef struct {
    wasm_engine_t *engine;
    wasm_module_t *module;
    worker *workers;
    unsigned count;
    uint64_t phases[4];
    char error[512];
} owner;

const char *sdk_name(void) { return "wasmer"; }
const char *sdk_version(void) { return wasmer_version(); }
static void failure(char *out, const char *phase) {
    int size = wasmer_last_error_length();
    char *message = size > 0 ? malloc((size_t)size) : NULL;
    if (message && wasmer_last_error_message(message, size) > 0)
        sdk_message(out, message, strnlen(message, (size_t)size));
    else sdk_message(out, phase, strlen(phase));
    free(message);
}
static int trap_failure(char *out, wasm_trap_t *trap) {
    if (!trap) return 0;
    wasm_message_t message;
    wasm_trap_message(trap, &message);
    sdk_message(out, message.data, message.size);
    wasm_byte_vec_delete(&message); wasm_trap_delete(trap);
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
        wasm_extern_vec_delete(&w->exports);
        if (w->instance) wasm_instance_delete(w->instance);
        if (w->store) wasm_store_delete(w->store);
    }
    if (o->module) wasm_module_delete(o->module);
    if (o->engine) wasm_engine_delete(o->engine);
    free(o->workers); free(o);
}
void *sdk_open(const uint8_t *bytes, size_t size, unsigned count) {
    owner *o = calloc(1, sizeof(*o));
    if (!o) return NULL;
    o->workers = calloc(count, sizeof(worker));
    if (!o->workers) { strcpy(o->error, "worker allocation failed"); return o; }
    o->count = count;
    uint64_t start = sdk_now();
    wasm_config_t *config = wasm_config_new();
    if (!config) { failure(o->error, "create config"); return o; }
    wasm_config_set_backend(config, CRANELIFT);
    o->engine = wasm_engine_new_with_config(config);
    if (!o->engine) { failure(o->error, "create engine"); return o; }
    for (unsigned i = 0; i < count; ++i) {
        o->workers[i].store = wasm_store_new(o->engine);
        if (!o->workers[i].store) { failure(o->error, "create store"); return o; }
    }
    o->phases[0] = sdk_now() - start;
    wasm_byte_vec_t binary = {size, (char *)bytes};
    start = sdk_now();
    o->module = wasm_module_new(o->workers[0].store, &binary);
    o->phases[1] = sdk_now() - start;
    if (!o->module) { failure(o->error, "compile module"); return o; }
    for (unsigned i = 0; i < count; ++i) {
        worker *w = &o->workers[i];
        wasm_extern_vec_t imports = {0, NULL};
        wasm_trap_t *trap = NULL;
        start = sdk_now();
        /* In pinned Wasmer 7.3.0 the module owns only immutable Module code.
         * wasm_instance_new uses its supplied store. The generic C API's
         * module_share/obtain declarations have no exported implementations. */
        w->instance = wasm_instance_new(w->store, o->module, &imports, &trap);
        o->phases[2] += sdk_now() - start;
        if (trap_failure(o->error, trap)) return o;
        if (!w->instance) { failure(o->error, "instantiate module"); return o; }
        wasm_exporttype_vec_t names;
        start = sdk_now();
        wasm_module_exports(o->module, &names);
        wasm_instance_exports(w->instance, &w->exports);
        if (names.size == w->exports.size) {
            for (size_t n = 0; n < names.size; ++n) {
                const wasm_name_t *name = wasm_exporttype_name(names.data[n]);
                if (name->size == 3 && !memcmp(name->data, "run", 3))
                    w->function = wasm_extern_as_func(w->exports.data[n]);
            }
        }
        o->phases[3] += sdk_now() - start;
        wasm_exporttype_vec_delete(&names);
        if (!w->function) { strcpy(o->error, "missing run function"); return o; }
        wasm_functype_t *type = wasm_func_type(w->function);
        const wasm_valtype_vec_t *params = wasm_functype_params(type), *results = wasm_functype_results(type);
        int valid = params->size == 2 && results->size == 1 && wasm_valtype_kind(params->data[0]) == WASM_I32 && wasm_valtype_kind(params->data[1]) == WASM_I32 && wasm_valtype_kind(results->data[0]) == WASM_I32;
        wasm_functype_delete(type);
        if (!valid) { strcpy(o->error, "run must have signature (i32,i32)->i32"); return o; }
    }
    return o;
}
uint64_t sdk_call(void *p, unsigned index, uint32_t seed, uint32_t count) {
    worker *w = &((owner *)p)->workers[index];
    wasm_val_t args[2] = {{.kind=WASM_I32, .of.i32=(int32_t)seed}, {.kind=WASM_I32, .of.i32=(int32_t)count}};
    wasm_val_t result = {.kind=WASM_I32};
    wasm_val_vec_t arguments = {2,args}, results = {1,&result};
    if (trap_failure(w->error, wasm_func_call(w->function, &arguments, &results))) return SDK_FAILURE;
    if (result.kind != WASM_I32) { strcpy(w->error, "non-i32 result"); return SDK_FAILURE; }
    return (uint32_t)result.of.i32;
}
