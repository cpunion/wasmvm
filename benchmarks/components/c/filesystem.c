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
int main(int argc, char **argv) {
  if (argc != 4) return 2;
  uint64_t count = strtoull(argv[3], NULL, 10), t = now();
  if (!count) return 2;
  FILE *file = fopen(argv[1], "rb");
  if (!file) return 3;
  fseek(file, 0, SEEK_END);
  long size = ftell(file);
  rewind(file);
  uint8_t *bytes = malloc((size_t)size);
  if (!bytes || fread(bytes, 1, (size_t)size, file) != (size_t)size) return 4;
  fclose(file);
  uint64_t load = now()-t;
  t=now(); wasm_engine_t *engine=wasm_engine_new(); uint64_t engine_time=now()-t;
  t=now(); wasmtime_component_t *component=NULL;
  check(wasmtime_component_new(engine,bytes,(size_t)size,&component)); uint64_t compile=now()-t;
  t=now(); wasmtime_store_t *store=wasmtime_store_new(engine,NULL,NULL);
  wasmtime_context_t *context=wasmtime_store_context(store);
  wasmtime_component_linker_t *linker=wasmtime_component_linker_new(engine);
  wasi_config_t *wasi=wasi_config_new();
  if (!wasi_config_preopen_dir(wasi,argv[2],"/sandbox",true)) return 6;
  check(wasmtime_context_set_wasi(context,wasi));
  check(wasmtime_component_linker_add_wasip2(linker)); uint64_t setup=now()-t;
  t=now(); wasmtime_component_instance_t instance;
  check(wasmtime_component_linker_instantiate(linker,context,component,&instance)); uint64_t instantiate=now()-t;
  wasmtime_component_val_t result={.kind=WASMTIME_COMPONENT_U32,.of.u32=0};
  wasmtime_component_func_t exercise=named(&instance,context,"exercise");
  check(wasmtime_component_func_call(&exercise,context,NULL,0,&result,1));
  if (result.kind!=WASMTIME_COMPONENT_U32 || result.of.u32!=42) return 7;
  t=now(); wasmtime_component_func_t batch=named(&instance,context,"stat-batch"); uint64_t lookup=now()-t;
  wasmtime_component_val_t arg={.kind=WASMTIME_COMPONENT_U32,.of.u32=128};
  t=now(); check(wasmtime_component_func_call(&batch,context,&arg,1,&result,1)); uint64_t first=now()-t;
  if (result.kind!=WASMTIME_COMPONENT_U32 || result.of.u32!=128) return 8;
  for (int i=0;i<100;i++) {check(wasmtime_component_func_call(&batch,context,&arg,1,&result,1)); if(result.of.u32!=128) return 9;}
  uint64_t checksum=0;t=now();
  for (uint64_t i=0;i<count;i++) {check(wasmtime_component_func_call(&batch,context,&arg,1,&result,1));if(result.kind!=WASMTIME_COMPONENT_U32 || result.of.u32!=128)return 10;checksum+=result.of.u32;}
  uint64_t warm=now()-t;
  t=now(); wasmtime_component_linker_delete(linker);wasmtime_store_delete(store);wasmtime_component_delete(component);wasm_engine_delete(engine);free(bytes);uint64_t close=now()-t;
  printf("{\"runtime\":\"wasmtime-48.0.1-wasip2\",\"load_ns\":%llu,\"engine_ns\":%llu,\"compile_ns\":%llu,\"setup_ns\":%llu,\"instantiate_ns\":%llu,\"lookup_ns\":%llu,\"first_call_ns\":%llu,\"warm_call_ns\":%.1f,\"stats_per_call\":128,\"warm_stat_ns\":%.3f,\"close_ns\":%llu,\"checksum\":%llu}\n",(unsigned long long)load,(unsigned long long)engine_time,(unsigned long long)compile,(unsigned long long)setup,(unsigned long long)instantiate,(unsigned long long)lookup,(unsigned long long)first,(double)warm/count,(double)warm/count/128,(unsigned long long)close,(unsigned long long)checksum);
  return 0;
}
