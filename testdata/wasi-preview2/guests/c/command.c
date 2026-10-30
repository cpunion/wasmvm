#include "c_command.h"
#include <string.h>

uint32_t exports_wasmvm_world_tests_arithmetic_increment(uint32_t value) {
  return value + 1;
}

bool exports_wasi_cli_run_run(void) {
  c_command_list_string_t args;
  wasi_cli_environment_get_arguments(&args);
  bool ok = args.len == 2 && args.ptr[1].len == 5 &&
            memcmp(args.ptr[1].ptr, "world", 5) == 0;
  c_command_list_string_free(&args);
  if (!ok) return false;

  c_command_list_u8_t random;
  wasi_random_random_get_random_bytes(16, &random);
  ok = random.len == 16;
  c_command_list_u8_free(&random);
  if (!ok) return false;

  wasi_io_poll_own_pollable_t poll = wasi_clocks_monotonic_clock_subscribe_duration(0);
  wasi_io_poll_method_pollable_block(wasi_io_poll_borrow_pollable(poll));
  wasi_io_poll_pollable_drop_own(poll);

  wasi_filesystem_preopens_list_tuple2_own_descriptor_string_t dirs;
  wasi_filesystem_preopens_get_directories(&dirs);
  ok = dirs.len == 1;
  if (ok) {
    wasi_filesystem_types_descriptor_stat_t stat;
    wasi_filesystem_types_error_code_t error;
    ok = wasi_filesystem_types_method_descriptor_stat(
      wasi_filesystem_types_borrow_descriptor(dirs.ptr[0].f0), &stat, &error) &&
      stat.type == WASI_FILESYSTEM_TYPES_DESCRIPTOR_TYPE_DIRECTORY;
  }
  wasi_filesystem_preopens_list_tuple2_own_descriptor_string_free(&dirs);
  if (!ok) return false;

  wasi_io_streams_own_output_stream_t out = wasi_cli_stdout_get_stdout();
  c_command_list_u8_t message = {(uint8_t *)"c-command:ok\n", 13};
  wasi_io_streams_stream_error_t error;
  ok = wasi_io_streams_method_output_stream_blocking_write_and_flush(
    wasi_io_streams_borrow_output_stream(out), &message, &error);
  wasi_io_streams_output_stream_drop_own(out);
  return ok;
}
