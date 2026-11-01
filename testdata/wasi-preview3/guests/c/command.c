#include "c_command.h"
#include <string.h>

static c_command_waitable_status_t wait_for(uint32_t handle, c_command_waitable_status_t status) {
  if (status != C_COMMAND_WAITABLE_STATUS_BLOCKED) return status;
  c_command_waitable_set_t set = c_command_waitable_set_new();
  c_command_waitable_join(handle, set);
  c_command_event_t event;
  c_command_waitable_set_wait(set, &event);
  c_command_waitable_join(handle, 0);
  c_command_waitable_set_drop(set);
  return event.code;
}

bool exports_wasi_cli_run_run(void) {
  c_command_list_string_t args;
  wasi_cli_environment_get_arguments(&args);
  bool ok = args.len == 2 && args.ptr[1].len == 5 && memcmp(args.ptr[1].ptr, "world", 5) == 0;
  c_command_list_string_free(&args);
  if (!ok) return false;
  c_command_list_u8_t random;
  wasi_random_random_get_random_bytes(16, &random);
  ok = random.len == 16;
  c_command_list_u8_free(&random);
  if (!ok) return false;
  wasi_clocks_monotonic_clock_wait_for(0);

  wasi_filesystem_preopens_list_tuple2_own_descriptor_string_t dirs;
  wasi_filesystem_preopens_get_directories(&dirs);
  ok = dirs.len == 1;
  if (ok) {
    wasi_filesystem_types_descriptor_stat_t stat;
    wasi_filesystem_types_error_code_t error;
    ok = wasi_filesystem_types_method_descriptor_stat(wasi_filesystem_types_borrow_descriptor(dirs.ptr[0].f0), &stat, &error)
      && stat.type.tag == WASI_FILESYSTEM_TYPES_DESCRIPTOR_TYPE_DIRECTORY;
  }
  for (size_t i = 0; i < dirs.len; ++i) wasi_filesystem_types_descriptor_drop_own(dirs.ptr[i].f0);
  wasi_filesystem_preopens_list_tuple2_own_descriptor_string_free(&dirs);
  if (!ok) return false;

  wasi_cli_stdin_stream_u8_writer_t writer;
  wasi_cli_stdin_stream_u8_t reader = wasi_cli_stdin_stream_u8_new(&writer);
  wasi_cli_stdout_future_result_void_error_code_t complete = wasi_cli_stdout_write_via_stream(reader);
  const uint8_t message[] = "c-p3-command:ok\n";
  size_t offset = 0;
  while (offset < sizeof(message) - 1) {
    c_command_waitable_status_t status = wait_for(writer, wasi_cli_stdin_stream_u8_write(writer, message + offset, sizeof(message) - 1 - offset));
    offset += C_COMMAND_WAITABLE_COUNT(status);
    if (C_COMMAND_WAITABLE_STATE(status) != C_COMMAND_WAITABLE_COMPLETED) { ok = false; break; }
  }
  wasi_cli_stdin_stream_u8_drop_writable(writer);
  wasi_cli_stdout_result_void_error_code_t result;
  c_command_waitable_status_t status = wait_for(complete, wasi_cli_stdout_future_result_void_error_code_read(complete, &result));
  wasi_cli_stdout_future_result_void_error_code_drop_readable(complete);
  return ok && C_COMMAND_WAITABLE_STATE(status) == C_COMMAND_WAITABLE_COMPLETED && !result.is_err;
}
