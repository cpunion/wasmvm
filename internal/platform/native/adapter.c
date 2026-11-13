//go:build (cgo || llgo) && ((linux && amd64) || (darwin && arm64))

#define _GNU_SOURCE
#include "adapter.h"
#include <fenv.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>
#if defined(__APPLE__) && defined(__aarch64__)
#include <sys/sysctl.h>
#endif
#ifdef __x86_64__
#include <cpuid.h>
#endif

uint32_t wasmvm_arm64_features(void) {
#if defined(__APPLE__) && defined(__aarch64__)
    int value = 0;
    size_t size = sizeof(value);
    if (sysctlbyname("hw.optional.arm.FEAT_LSE", &value, &size, 0, 0) == 0 && value)
        return 1;
#endif
    return 0;
}

void wasmvm_cpu_probe(uint32_t *ecx1, uint32_t *ebx7, uint32_t *ecx7, uint64_t *xcr0) {
    *ecx1 = *ebx7 = *ecx7 = 0;
    *xcr0 = 0;
#ifdef __x86_64__
    unsigned int eax, ebx, ecx, edx;
    if (!__get_cpuid(1, &eax, &ebx, &ecx, &edx)) return;
    *ecx1 = ecx;
    const unsigned int avx_state = (1u << 26) | (1u << 27) | (1u << 28);
    if ((ecx & avx_state) == avx_state) {
        unsigned int low, high;
        __asm__ volatile ("xgetbv" : "=a"(low), "=d"(high) : "c"(0));
        *xcr0 = ((uint64_t)high << 32) | low;
    }
    if (__get_cpuid_max(0, NULL) >= 7) {
        __cpuid_count(7, 0, eax, ebx, ecx, edx);
        *ebx7 = ebx;
        *ecx7 = ecx;
    }
#endif
}
#ifdef __APPLE__
#include <pthread.h>
#include <libkern/OSCacheControl.h>

/* One bounded MAP_JIT arena also supports hardened processes which permit only
 * a single JIT mapping. All metadata lives outside the executable mapping.
 * Reserve 1 GiB of virtual address space; pages become resident only as code
 * is published. Per-module code quotas remain separate. Large real Go guests
 * exceed the former 64 MiB process arena even with an explicit larger quota.
 * The embedding application supplies its own entitlements. */
#define JIT_BYTES (1024u * 1024u * 1024u)
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

/* Reserve address space without touching uncommitted pages. Growth only changes
   permissions on newly exposed anonymous zero pages; existing addresses stay. */
static size_t wasmvm_page_round(size_t size, size_t page) {
    if (!page || size > SIZE_MAX - (page - 1)) return 0;
    return (size + page - 1) / page * page;
}
void *wasmvm_reserve(size_t maximum) {
    long page = sysconf(_SC_PAGESIZE);
    size_t rounded = page > 0 ? wasmvm_page_round(maximum ? maximum : 1, (size_t)page) : 0;
    if (!rounded) return NULL;
    void *p = mmap(NULL, rounded, PROT_NONE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    return p == MAP_FAILED ? NULL : p;
}
int wasmvm_commit(void *p, size_t previous, size_t size) {
    long page = sysconf(_SC_PAGESIZE);
    if (page <= 0 || size < previous) return 0;
    size_t from = wasmvm_page_round(previous, (size_t)page);
    size_t to = wasmvm_page_round(size, (size_t)page);
    if (size && !to) return 0;
    return to == from || mprotect((char *)p + from, to - from, PROT_READ | PROT_WRITE) == 0;
}
void wasmvm_unreserve(void *p, size_t maximum) {
    long page = sysconf(_SC_PAGESIZE);
    size_t rounded = page > 0 ? wasmvm_page_round(maximum ? maximum : 1, (size_t)page) : 0;
    if (rounded) munmap(p, rounded);
}
void wasmvm_cancel(struct wasmvm_record *ctx) { __atomic_store_n(&ctx->control, 1, __ATOMIC_RELEASE); }

#ifdef __aarch64__
struct wasmvm_fp_environment { uint64_t control, status; };
static inline struct wasmvm_fp_environment wasmvm_fp_read(void) {
    struct wasmvm_fp_environment env;
    __asm__ volatile ("mrs %0, fpcr\n\tmrs %1, fpsr"
                      : "=r"(env.control), "=r"(env.status) : : "memory");
    return env;
}
static inline void wasmvm_fp_write(struct wasmvm_fp_environment env) {
    __asm__ volatile ("msr fpcr, %0\n\tmsr fpsr, %1\n\tisb"
                      : : "r"(env.control), "r"(env.status) : "memory");
}
#else
/* FNSTENV/FNINIT are non-waiting: an unmasked pending host exception must
   survive entry rather than be raised by the environment transition. Keep
   the full x87 environment, even though generated floating code uses SSE. */
struct wasmvm_fp_environment {
    unsigned char x87[28];
    uint32_t mxcsr;
};
static inline struct wasmvm_fp_environment wasmvm_fp_begin(void) {
    struct wasmvm_fp_environment saved;
    const uint32_t default_mxcsr = 0x1f80;
    __asm__ volatile ("fnstenv %0\n\tstmxcsr %1\n\tfninit\n\tldmxcsr %2"
                      : "=m"(saved.x87), "=m"(saved.mxcsr)
                      : "m"(default_mxcsr) : "memory");
    return saved;
}
static inline void wasmvm_fp_write(struct wasmvm_fp_environment env) {
    __asm__ volatile ("fldenv %0\n\tldmxcsr %1"
                      : : "m"(env.x87), "m"(env.mxcsr) : "memory");
}
#endif

static uint32_t wasmvm_enter_impl(const void *entry, struct wasmvm_record *ctx, uint32_t integer_only) {
    if (ctx->version != 2 || ctx->used > ctx->capacity || ctx->depth > ctx->max_depth) return 0;
    if (__atomic_load_n(&ctx->control, __ATOMIC_ACQUIRE)) return 6; /* Poll */
    if (integer_only) {
#ifdef __APPLE__
        pthread_jit_write_protect_np(1);
#endif
        return ((uint32_t (*)(struct wasmvm_record *))entry)(ctx);
    }
#ifdef __aarch64__
    /* FPCR=0 gives nearest-even, gradual underflow and masked exceptions.
       Preserve FPSR too: guest exceptions and SIMD saturation cannot leak. */
    struct wasmvm_fp_environment saved = wasmvm_fp_read();
    wasmvm_fp_write((struct wasmvm_fp_environment){0, 0});
#else
    struct wasmvm_fp_environment saved = wasmvm_fp_begin();
#endif
#ifdef __APPLE__
    pthread_jit_write_protect_np(1);
#endif
    uint32_t result = ((uint32_t (*)(struct wasmvm_record *))entry)(ctx);
    wasmvm_fp_write(saved);
    return result;
}
uint32_t wasmvm_enter(const void *entry, struct wasmvm_record *ctx) {
    return wasmvm_enter_impl(entry, ctx, 0);
}
uint32_t wasmvm_enter_integer(const void *entry, struct wasmvm_record *ctx) {
    return wasmvm_enter_impl(entry, ctx, 1);
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
        offsetof(struct wasmvm_record, control),
        offsetof(struct wasmvm_record, memory_base), offsetof(struct wasmvm_record, memory_size)
    };
    return field < sizeof(offsets)/sizeof(offsets[0]) ? offsets[field] : UINT64_MAX;
}

uint32_t wasmvm_probe_fp(const void *entry, struct wasmvm_record *ctx, uint32_t expected_status, uint32_t integer_only) {
    fenv_t original;
    if (feholdexcept(&original) != 0) return 0;
    if (fesetround(FE_DOWNWARD) != 0 || feraiseexcept(FE_INVALID | FE_INEXACT) != 0) {
        fesetenv(&original);
        return 0;
    }
#ifdef __aarch64__
    struct wasmvm_fp_environment expected = wasmvm_fp_read();
    /* Include flush-to-zero, default NaNs and cumulative SIMD saturation. */
    expected.control |= (1u << 24) | (1u << 25);
    expected.status |= 1u << 27;
    wasmvm_fp_write(expected);
#else
    uint32_t expected_mxcsr, supported;
    struct { unsigned char bytes[512]; } __attribute__((aligned(16))) fx;
    __asm__ volatile ("fxsave %0" : "=m"(fx));
    memcpy(&supported, fx.bytes + 28, sizeof(supported));
    if (!supported) supported = 0xffbf;
    __asm__ volatile ("stmxcsr %0" : "=m"(expected_mxcsr));
    expected_mxcsr |= supported & 0x8040; /* FZ and optional DAZ. */
    __asm__ volatile ("ldmxcsr %0" : : "m"(expected_mxcsr) : "memory");
    unsigned char expected_x87[28];
    __asm__ volatile ("fnstenv %0\n\tfldenv %0" : "=m"(expected_x87) : : "memory");
#endif
    uint32_t status = wasmvm_enter_impl(entry, ctx, integer_only);
    int restored = fegetround() == FE_DOWNWARD &&
                   fetestexcept(FE_ALL_EXCEPT) == (FE_INVALID | FE_INEXACT);
#ifdef __aarch64__
    struct wasmvm_fp_environment actual = wasmvm_fp_read();
    restored = restored && actual.control == expected.control && actual.status == expected.status;
#else
    uint32_t actual_mxcsr;
    __asm__ volatile ("stmxcsr %0" : "=m"(actual_mxcsr));
    unsigned char actual_x87[28];
    __asm__ volatile ("fnstenv %0\n\tfldenv %0" : "=m"(actual_x87) : : "memory");
    restored = restored && actual_mxcsr == expected_mxcsr &&
               memcmp(actual_x87, expected_x87, sizeof(actual_x87)) == 0;
#endif
    fesetenv(&original);
    return status == expected_status && restored;
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
