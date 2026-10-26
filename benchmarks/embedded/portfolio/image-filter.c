/* Authored for wasmvm: seeded 32x32 image, 3x3 binomial filter and checksum.
 * Initialization is inside every timed run; each iteration filters 900 pixels.
 */
typedef unsigned int u32;
static u32 rot(u32 x) { return (x << 5) | (x >> 27); }
u32 run(u32 seed, u32 iterations) {
    unsigned char image[1024];
    u32 state = seed;
    for (u32 i = 0; i < 1024; i++) {
        state = state * 1664525u + 1013904223u;
        image[i] = state >> 24;
    }
    u32 hash = seed;
    for (u32 i = 0; i < iterations; i++) {
        for (u32 y = 1; y < 31; y++) {
            for (u32 x = 1; x < 31; x++) {
                u32 p = y * 32 + x;
                u32 pixel = (image[p-33] + 2*image[p-32] + image[p-31]
                           + 2*image[p-1] + 4*image[p] + 2*image[p+1]
                           + image[p+31] + 2*image[p+32] + image[p+33]) >> 4;
                hash = rot(hash) ^ (pixel + i + p);
            }
        }
    }
    return hash;
}
