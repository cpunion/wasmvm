#ifndef WASMVM_PHASE_COMMON_H
#define WASMVM_PHASE_COMMON_H

#include <errno.h>
#include <inttypes.h>
#include <limits.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define MAX_CALLS 64

typedef struct {
  const char *path;
  const char *backend;
  uint32_t seed, iterations, expected, samples, warmups;
} options;

typedef struct {
  uint64_t load, engine_setup, compile_validate, instantiate, export_lookup;
  uint64_t first_call, cleanup, parse, validate;
  uint64_t warm[MAX_CALLS], warmup[MAX_CALLS], timer_pair[7];
  uint32_t first_result, warm_results[MAX_CALLS], warmup_results[MAX_CALLS];
  bool separate_validation;
} measurements;

static void fail(const char *message) {
  fprintf(stderr, "%s\n", message);
  exit(EXIT_FAILURE);
}

static uint64_t now_ns(void) {
  struct timespec timestamp;
  if (clock_gettime(CLOCK_MONOTONIC, &timestamp) != 0)
    fail("clock_gettime failed");
  return (uint64_t)timestamp.tv_sec * UINT64_C(1000000000) +
         (uint64_t)timestamp.tv_nsec;
}

static uint32_t parse_u32(const char *text) {
  if (!*text || strspn(text, "0123456789") != strlen(text))
    fail("arguments require unsigned decimal integers");
  errno = 0;
  char *end = NULL;
  unsigned long long value = strtoull(text, &end, 10);
  if (errno || *end || value > UINT32_MAX)
    fail("integer argument outside uint32 range");
  return (uint32_t)value;
}

static options begin(int argc, char **argv, const char *backend) {
  puts("{\"ready\":true}");
  if (fflush(stdout) != 0)
    fail("cannot flush ready record");
  char acknowledgement[5];
  if (!fgets(acknowledgement, sizeof(acknowledgement), stdin) ||
      strcmp(acknowledgement, "run\n") != 0)
    fail("expected run acknowledgement after ready record");
  options result = {0};
  unsigned seen = 0;
  const char *names[] = {"--wasm", "--backend", "--iterations", "--seed",
                         "--expected", "--samples", "--warmups"};
  for (int i = 1; i < argc; i += 2) {
    if (i + 1 >= argc)
      fail("argument requires a value");
    unsigned key;
    for (key = 0; key < 7 && strcmp(argv[i], names[key]) != 0; ++key) {}
    if (key == 7 || (seen & (1u << key)))
      fail("unknown or duplicate argument");
    seen |= 1u << key;
    const char *value = argv[i + 1];
    switch (key) {
    case 0: result.path = value; break;
    case 1: result.backend = value; break;
    case 2: result.iterations = parse_u32(value); break;
    case 3: result.seed = parse_u32(value); break;
    case 4: result.expected = parse_u32(value); break;
    case 5: result.samples = parse_u32(value); break;
    case 6: result.warmups = parse_u32(value); break;
    }
  }
  if (seen != 127 || !*result.path || strcmp(result.backend, backend) != 0)
    fail("all arguments and the runner's supported backend are required");
  if (!result.iterations || result.iterations > INT32_MAX ||
      !result.samples || result.samples > MAX_CALLS || result.warmups > MAX_CALLS)
    fail("iterations must be positive i32; samples 1..64 and warmups 0..64");
  return result;
}

static uint8_t *read_module(const char *path, size_t *size) {
  FILE *file = fopen(path, "rb");
  if (!file || fseek(file, 0, SEEK_END) != 0)
    fail("cannot open or seek module");
  long length = ftell(file);
  if (length < 8 || (unsigned long)length > UINT32_MAX || fseek(file, 0, SEEK_SET) != 0)
    fail("invalid module size");
  *size = (size_t)length;
  uint8_t *bytes = malloc(*size);
  if (!bytes || fread(bytes, 1, *size, file) != *size || fclose(file) != 0)
    fail("cannot read module");
  return bytes;
}

static uint32_t read_uleb(const uint8_t *data, size_t *offset, size_t end) {
  uint32_t value = 0;
  for (unsigned shift = 0; shift < 35; shift += 7) {
    if (*offset >= end)
      fail("truncated section encoding");
    uint8_t byte = data[(*offset)++];
    if (shift == 28 && (byte & 0xf0))
      fail("section encoding overflows uint32");
    value |= (uint32_t)(byte & 0x7f) << shift;
    if (!(byte & 0x80))
      return value;
  }
  fail("invalid section encoding");
  return 0;
}

/* Reject imports and start before any API can instantiate the module. Runtime
 * validation remains authoritative for all other WebAssembly semantics. */
static void require_pure_module(const uint8_t *data, size_t size) {
  const uint8_t header[] = {0, 'a', 's', 'm', 1, 0, 0, 0};
  if (size < sizeof(header) || memcmp(data, header, sizeof(header)) != 0)
    fail("expected a Core WebAssembly v1 module");
  size_t offset = sizeof(header);
  while (offset < size) {
    uint8_t section = data[offset++];
    uint32_t length = read_uleb(data, &offset, size);
    if (length > size - offset)
      fail("section exceeds module size");
    size_t end = offset + length;
    if (section == 8)
      fail("start sections are excluded from phase benchmarks");
    if (section == 2 && read_uleb(data, &offset, end) != 0)
      fail("imports are excluded from phase benchmarks");
    offset = end;
  }
}

typedef uint32_t (*invocation)(void *, const options *, uint64_t *);

static void calls(const options *opt, measurements *result, void *context,
                  invocation invoke) {
  result->first_result = invoke(context, opt, &result->first_call);
  if (result->first_result != opt->expected)
    fail("first call checksum mismatch");
  for (uint32_t i = 0; i < opt->warmups; ++i) {
    result->warmup_results[i] = invoke(context, opt, &result->warmup[i]);
    if (result->warmup_results[i] != opt->expected)
      fail("warmup call checksum mismatch");
  }
  for (uint32_t i = 0; i < opt->samples; ++i) {
    const char *interleave = getenv("WASMVM_INTERLEAVE");
    if (interleave && strcmp(interleave, "1") == 0) {
      printf("{\"sampleReady\":%" PRIu32 "}\n", i);
      if (fflush(stdout) != 0)
        fail("cannot flush sample readiness");
      char acknowledgement[8];
      if (!fgets(acknowledgement, sizeof(acknowledgement), stdin) ||
          strcmp(acknowledgement, "sample\n") != 0)
        fail("expected sample acknowledgement");
    }
    result->warm_results[i] = invoke(context, opt, &result->warm[i]);
    if (result->warm_results[i] != opt->expected)
      fail("warm call checksum mismatch");
  }
}

static void ns_array(const uint64_t *values, uint32_t length) {
  putchar('[');
  for (uint32_t i = 0; i < length; ++i)
    printf("%s%" PRIu64, i ? "," : "", values[i]);
  putchar(']');
}

static void result_array(const uint32_t *values, uint32_t length) {
  putchar('[');
  for (uint32_t i = 0; i < length; ++i)
    printf("%s%" PRIu32, i ? "," : "", values[i]);
  putchar(']');
}

static void report(const options *opt, measurements *result) {
  for (unsigned i = 0; i < 7; ++i) {
    uint64_t start = now_ns();
    result->timer_pair[i] = now_ns() - start;
  }
  printf("{\"schemaVersion\":1,\"backend\":\"%s\",\"phasesNs\":{"
         "\"load\":%" PRIu64 ",\"engineSetup\":%" PRIu64
         ",\"compileValidate\":%" PRIu64 ",\"instantiate\":%" PRIu64
         ",\"exportLookup\":%" PRIu64 ",\"firstCall\":%" PRIu64
         ",\"cleanup\":%" PRIu64 "},\"warmupCallNs\":",
         opt->backend, result->load, result->engine_setup, result->compile_validate,
         result->instantiate, result->export_lookup, result->first_call, result->cleanup);
  ns_array(result->warmup, opt->warmups);
  printf(",\"warmCallNs\":"); ns_array(result->warm, opt->samples);
  printf(",\"timerPairNs\":"); ns_array(result->timer_pair, 7);
  printf(",\"firstResult\":%" PRIu32 ",\"result\":%" PRIu32,
         result->first_result, result->first_result);
  printf(",\"warmupResults\":"); result_array(result->warmup_results, opt->warmups);
  printf(",\"warmResults\":"); result_array(result->warm_results, opt->samples);
  printf(",\"iterations\":%" PRIu32 ",\"seed\":%" PRIu32
         ",\"samples\":%" PRIu32 ",\"warmups\":%" PRIu32,
         opt->iterations, opt->seed, opt->samples, opt->warmups);
  if (result->separate_validation)
    printf(",\"preparationPartsNs\":{\"parse\":%" PRIu64 ",\"validate\":%" PRIu64 "}",
           result->parse, result->validate);
  puts("}");
}

#endif
