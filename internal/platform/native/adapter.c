//go:build (cgo || llgo) && ((linux && amd64) || (darwin && arm64))

#define _GNU_SOURCE
#include "adapter.h"
#include <fenv.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#ifdef __APPLE__
#include <pthread.h>
#include <libkern/OSCacheControl.h>

/* One bounded MAP_JIT arena also supports hardened processes which permit only
 * a single JIT mapping. All metadata lives outside the executable mapping.
 * The embedding application supplies its own entitlements. */
#define JIT_BYTES (64u * 1024u * 1024u)
struct block { size_t start, size; int used; struct block *next; };
static pthread_mutex_t jit_lock = PTHREAD_MUTEX_INITIALIZER;
static void *jit_base;
static struct block *jit_blocks;

static void *jit_publish(const void *source, size_t size) {
    if (size > JIT_BYTES) return NULL;
    size_t needed = (size + 63u) & ~(size_t)63u;
    pthread_mutex_lock(&jit_lock);
    if (!jit_base) {
        struct block *first = calloc(1, sizeof(*first));
        if (!first) { pthread_mutex_unlock(&jit_lock); return NULL; }
        void *base = mmap(NULL, JIT_BYTES, PROT_READ | PROT_WRITE | PROT_EXEC,
                         MAP_PRIVATE | MAP_ANON | MAP_JIT, -1, 0);
        if (base == MAP_FAILED) { free(first); pthread_mutex_unlock(&jit_lock); return NULL; }
        first->size = JIT_BYTES;
        jit_base = base;
        jit_blocks = first;
    }
    for (struct block *b = jit_blocks; b; b = b->next) {
        if (b->used || b->size < needed) continue;
        if (b->size > needed) {
            struct block *tail = calloc(1, sizeof(*tail));
            if (!tail) { pthread_mutex_unlock(&jit_lock); return NULL; }
            tail->start = b->start + needed;
            tail->size = b->size - needed;
            tail->next = b->next;
            b->next = tail;
            b->size = needed;
        }
        b->used = 1;
        void *dest = (char *)jit_base + b->start;
        /* One C call keeps this entire transaction on one OS thread. Published
         * ranges are never patched; this range has no retained entry/lease. */
        pthread_jit_write_protect_np(0);
        memcpy(dest, source, size);
        sys_icache_invalidate(dest, size);
        pthread_jit_write_protect_np(1);
        pthread_mutex_unlock(&jit_lock);
        return dest;
    }
    pthread_mutex_unlock(&jit_lock);
    return NULL;
}
#endif

void *wasmvm_publish(const void *source, size_t size) {
#ifdef __APPLE__
    return jit_publish(source, size);
#else
    void *p = mmap(NULL, size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (p == MAP_FAILED) return NULL;
    memcpy(p, source, size);
    __builtin___clear_cache(p, (char *)p + size);
    if (mprotect(p, size, PROT_READ | PROT_EXEC) != 0) { munmap(p, size); return NULL; }
    return p;
#endif
}

void wasmvm_unpublish(void *code, size_t size) {
#ifdef __APPLE__
    (void)size;
    pthread_mutex_lock(&jit_lock);
    for (struct block *b = jit_blocks; b; b = b->next) {
        if ((char *)jit_base + b->start == code) { b->used = 0; break; }
    }
    for (struct block *b = jit_blocks; b && b->next;) {
        struct block *next = b->next;
        if (!b->used && !next->used) {
            b->size += next->size;
            b->next = next->next;
            free(next);
        } else { b = next; }
    }
    /* The bounded process arena is reserved for reuse, never repeatedly mapped. */
    pthread_mutex_unlock(&jit_lock);
#else
    munmap(code, size);
#endif
}

void *wasmvm_allocate(size_t size) { return calloc(1, size); }
void wasmvm_deallocate(void *p) { free(p); }
void wasmvm_cancel(struct wasmvm_record *ctx) { __atomic_store_n(&ctx->control, 1, __ATOMIC_RELEASE); }

uint32_t wasmvm_enter(const void *entry, struct wasmvm_record *ctx) {
    if (ctx->version != 1 || ctx->used > ctx->capacity || ctx->depth > ctx->max_depth) return 0;
    if (__atomic_load_n(&ctx->control, __ATOMIC_ACQUIRE)) return 6; /* Poll */
    fenv_t saved;
    if (fegetenv(&saved) != 0) return 0;
    if (fesetenv(FE_DFL_ENV) != 0) { fesetenv(&saved); return 0; }
#ifdef __APPLE__
    pthread_jit_write_protect_np(1);
#endif
    uint32_t result = ((uint32_t (*)(struct wasmvm_record *))entry)(ctx);
    fesetenv(&saved);
    return result;
}

uint64_t wasmvm_layout(uint32_t field) {
    const uint64_t offsets[] = {
        sizeof(struct wasmvm_record),
        offsetof(struct wasmvm_record, arena), offsetof(struct wasmvm_record, capacity),
        offsetof(struct wasmvm_record, max_depth), offsetof(struct wasmvm_record, used),
        offsetof(struct wasmvm_record, depth), offsetof(struct wasmvm_record, function),
        offsetof(struct wasmvm_record, continuation), offsetof(struct wasmvm_record, source),
        offsetof(struct wasmvm_record, root_map), offsetof(struct wasmvm_record, fuel),
        offsetof(struct wasmvm_record, budget), offsetof(struct wasmvm_record, result_lo),
        offsetof(struct wasmvm_record, result_hi), offsetof(struct wasmvm_record, payload),
        offsetof(struct wasmvm_record, control)
    };
    return field < sizeof(offsets)/sizeof(offsets[0]) ? offsets[field] : UINT64_MAX;
}

uint32_t wasmvm_probe_fp(const void *entry, struct wasmvm_record *ctx) {
    fenv_t original;
    if (fegetenv(&original) != 0 || fesetround(FE_DOWNWARD) != 0) return 0;
    uint32_t status = wasmvm_enter(entry, ctx);
    int restored = fegetround() == FE_DOWNWARD;
    fesetenv(&original);
    return status == 1 && restored;
}

uint32_t wasmvm_probe_registers(const void *entry, struct wasmvm_record *ctx) {
#ifdef __aarch64__
    register uint64_t a __asm__("x19") = 19, b __asm__("x20") = 20;
    register uint64_t c __asm__("x21") = 21, d __asm__("x22") = 22;
    register uint64_t e __asm__("x23") = 23, f __asm__("x24") = 24;
    register uint64_t g __asm__("x25") = 25, h __asm__("x26") = 26;
    register uint64_t i __asm__("x27") = 27, j __asm__("x28") = 28;
    register double v0 __asm__("d8") = 8, v1 __asm__("d9") = 9;
    register double v2 __asm__("d10") = 10, v3 __asm__("d11") = 11;
    register double v4 __asm__("d12") = 12, v5 __asm__("d13") = 13;
    register double v6 __asm__("d14") = 14, v7 __asm__("d15") = 15;
    __asm__ volatile("" : "+r"(a),"+r"(b),"+r"(c),"+r"(d),"+r"(e),"+r"(f),"+r"(g),"+r"(h),"+r"(i),"+r"(j));
    __asm__ volatile("" : "+w"(v0),"+w"(v1),"+w"(v2),"+w"(v3),"+w"(v4),"+w"(v5),"+w"(v6),"+w"(v7));
    uint32_t status = wasmvm_enter(entry,ctx);
    __asm__ volatile("" : "+r"(a),"+r"(b),"+r"(c),"+r"(d),"+r"(e),"+r"(f),"+r"(g),"+r"(h),"+r"(i),"+r"(j));
    __asm__ volatile("" : "+w"(v0),"+w"(v1),"+w"(v2),"+w"(v3),"+w"(v4),"+w"(v5),"+w"(v6),"+w"(v7));
    return status == 1 && a==19 && b==20 && c==21 && d==22 && e==23 && f==24 && g==25 && h==26 && i==27 && j==28
        && v0==8 && v1==9 && v2==10 && v3==11 && v4==12 && v5==13 && v6==14 && v7==15;
#else
    register uint64_t a __asm__("rbx") = 11, b __asm__("r12") = 12;
    register uint64_t c __asm__("r13") = 13, d __asm__("r14") = 14, e __asm__("r15") = 15;
    __asm__ volatile("" : "+r"(a),"+r"(b),"+r"(c),"+r"(d),"+r"(e));
    uint32_t status = wasmvm_enter(entry,ctx);
    __asm__ volatile("" : "+r"(a),"+r"(b),"+r"(c),"+r"(d),"+r"(e));
    return status == 1 && a==11 && b==12 && c==13 && d==14 && e==15;
#endif
}
