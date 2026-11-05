// Complete streaming SHA-256 over generated 64-byte message blocks.
// Every exported call resets the state and includes generation and padding.
#if defined(__APPLE__) && !defined(_DARWIN_C_SOURCE)
#define _DARWIN_C_SOURCE
#endif
#include <stdint.h>

static const uint32_t constants[64] = {
    0x428a2f98u, 0x71374491u, 0xb5c0fbcfu, 0xe9b5dba5u, 0x3956c25bu, 0x59f111f1u, 0x923f82a4u, 0xab1c5ed5u,
    0xd807aa98u, 0x12835b01u, 0x243185beu, 0x550c7dc3u, 0x72be5d74u, 0x80deb1feu, 0x9bdc06a7u, 0xc19bf174u,
    0xe49b69c1u, 0xefbe4786u, 0x0fc19dc6u, 0x240ca1ccu, 0x2de92c6fu, 0x4a7484aau, 0x5cb0a9dcu, 0x76f988dau,
    0x983e5152u, 0xa831c66du, 0xb00327c8u, 0xbf597fc7u, 0xc6e00bf3u, 0xd5a79147u, 0x06ca6351u, 0x14292967u,
    0x27b70a85u, 0x2e1b2138u, 0x4d2c6dfcu, 0x53380d13u, 0x650a7354u, 0x766a0abbu, 0x81c2c92eu, 0x92722c85u,
    0xa2bfe8a1u, 0xa81a664bu, 0xc24b8b70u, 0xc76c51a3u, 0xd192e819u, 0xd6990624u, 0xf40e3585u, 0x106aa070u,
    0x19a4c116u, 0x1e376c08u, 0x2748774cu, 0x34b0bcb5u, 0x391c0cb3u, 0x4ed8aa4au, 0x5b9cca4fu, 0x682e6ff3u,
    0x748f82eeu, 0x78a5636fu, 0x84c87814u, 0x8cc70208u, 0x90befffau, 0xa4506cebu, 0xbef9a3f7u, 0xc67178f2u
};
static uint32_t rotate(uint32_t x, unsigned n) { return x >> n | x << (32 - n); }

static void compress(uint32_t h[8], uint32_t w[64]) {
    for (unsigned i = 16; i < 64; ++i) {
        uint32_t a = w[i-15], b = w[i-2];
        w[i] = w[i-16] + (rotate(a,7) ^ rotate(a,18) ^ a>>3)
             + w[i-7] + (rotate(b,17) ^ rotate(b,19) ^ b>>10);
    }
    uint32_t a=h[0], b=h[1], c=h[2], d=h[3], e=h[4], f=h[5], g=h[6], z=h[7];
    for (unsigned i = 0; i < 64; ++i) {
        uint32_t t1 = z + (rotate(e,6) ^ rotate(e,11) ^ rotate(e,25))
                    + ((e&f) ^ (~e&g)) + constants[i] + w[i];
        uint32_t t2 = (rotate(a,2) ^ rotate(a,13) ^ rotate(a,22))
                    + ((a&b) ^ (a&c) ^ (b&c));
        z=g; g=f; f=e; e=d+t1; d=c; c=b; b=a; a=t1+t2;
    }
    h[0]+=a; h[1]+=b; h[2]+=c; h[3]+=d; h[4]+=e; h[5]+=f; h[6]+=g; h[7]+=z;
}

uint32_t run(uint32_t seed, uint32_t count) {
    if (!count || count > 65536) __builtin_trap();
    uint32_t h[8] = {0x6a09e667u, 0xbb67ae85u, 0x3c6ef372u, 0xa54ff53au, 0x510e527fu, 0x9b05688cu, 0x1f83d9abu, 0x5be0cd19u}, w[64];
    uint32_t x = seed;
    for (uint32_t block = 0; block < count; ++block) {
        for (unsigned i = 0; i < 16; ++i) {
            x = x * 1664525u + 1013904223u;
            w[i] = __builtin_bswap32(x);
        }
        compress(h,w);
    }
    w[0]=0x80000000u;
    for (unsigned i = 1; i < 15; ++i) w[i]=0;
    w[15]=count*512u;
    compress(h,w);
    return h[0];
}

#ifndef __wasm__
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

// Prevent hoisting/CSE of identical pure calls across sample clocks.
static uint32_t (*volatile invoke_kernel)(uint32_t,uint32_t) = run;
static uint64_t now_ns(void) {
    struct timespec t;
#ifdef __APPLE__
    const clockid_t clock = CLOCK_UPTIME_RAW;
#else
    const clockid_t clock = CLOCK_MONOTONIC;
#endif
    if (clock_gettime(clock,&t)) abort();
    return (uint64_t)t.tv_sec*1000000000u + (uint64_t)t.tv_nsec;
}
int main(int argc,char **argv) {
    if (argc!=4) return 2;
    uint32_t args[3];
    for (int i=0;i<3;++i) {
        char *end; unsigned long long n=strtoull(argv[i+1],&end,0);
        if (!*argv[i+1] || *end || n>UINT32_MAX) return 2;
        args[i]=(uint32_t)n;
    }
    if (!args[1] || args[1]>65536) return 2;
    int interleave=getenv("WASMVM_INTERLEAVE")!=NULL; char token[16];
    if (interleave) {puts("{\"ready\":true}");fflush(stdout);if(!fgets(token,sizeof(token),stdin))return 2;}
    uint64_t samples[7],warmups[2],first=0;
    for (int sample=-3;sample<7;++sample) {
        if (sample>=0 && interleave) {printf("{\"sampleReady\":%d}\n",sample);fflush(stdout);if(!fgets(token,sizeof(token),stdin))return 2;}
        uint64_t start=now_ns();uint32_t result=invoke_kernel(args[0],args[1]);uint64_t elapsed=now_ns()-start;
        if (result!=args[2]) {fprintf(stderr,"checksum mismatch: %"PRIu32" != %"PRIu32"\n",result,args[2]);return 1;}
        if (sample==-3)first=elapsed;else if(sample<0)warmups[sample+2]=elapsed;else samples[sample]=elapsed;
    }
    printf("{\"seed\":%"PRIu32",\"iterations\":%"PRIu32",\"checksum\":%"PRIu32",\"firstCallNs\":%"PRIu64",\"warmupCallNs\":[%"PRIu64",%"PRIu64"],\"warmCallNs\":[",args[0],args[1],args[2],first,warmups[0],warmups[1]);
    for(int i=0;i<7;++i)printf("%s%"PRIu64,i?",":"",samples[i]);puts("]}");return 0;
}
#endif
