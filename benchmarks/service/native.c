/* Keep the original scalar source unchanged. Its standalone CLI is renamed
 * and never called; the Go/LLGo adapter invokes only run(seed, count). */
#define main wasmvm_reference_unused_main
#include "../applications/sha256.c"
