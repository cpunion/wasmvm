// Authored for wasmvm: generate 256-byte packets and compute IEEE CRC-32.
// Each iteration processes one packet; state generation is inside timed run.
#![no_std]
use core::panic::PanicInfo;
#[panic_handler]
fn panic(_: &PanicInfo) -> ! { core::arch::wasm32::unreachable() }
#[unsafe(no_mangle)]
pub extern "C" fn run(seed: u32, iterations: u32) -> u32 {
    let mut hash = seed;
    for i in 0..iterations {
        let mut state = seed.wrapping_add(i);
        let mut crc = u32::MAX;
        for _ in 0..256 {
            state = state.wrapping_mul(1664525).wrapping_add(1013904223);
            crc ^= state >> 24;
            for _ in 0..8 {
                crc = (crc >> 1) ^ (0xedb88320u32.wrapping_mul(crc & 1));
            }
        }
        hash = hash.rotate_left(3) ^ !crc;
    }
    hash
}
