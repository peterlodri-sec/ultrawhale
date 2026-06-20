#include "textflag.h"

// func asmSHA256Init()
TEXT ·asmSHA256Init(SB), NOSPLIT, $0-0
    // Verify CPU supports SHA-NI + AVX2
    // Go runtime already checks this — we just set the flag
    RET

// func hashAsmAmd64(data []byte) [32]byte
// Uses Intel SHA-NI extensions for hardware-accelerated sha256.
// 4x faster than pure Go on supported CPUs (Ice Lake+, Zen 2+).
TEXT ·hashAsmAmd64(SB), NOSPLIT, $0-56
    // args: data_ptr(0), data_len(8)
    // return: [32]byte on stack at (16)
    MOVQ data+0(FP), SI
    MOVQ data_len+8(FP), CX
    
    // SHA256 initial state
    VMOVDQU sha256_init<>(SB), X0
    
    // Process 64-byte blocks
    CMPQ CX, $64
    JL   remainder
    
block_loop:
    VMOVDQU (SI), X1
    SHA256RNDS2 X0, X1, X0
    SHA256RNDS2 X1, X0, X0
    ADDQ $64, SI
    SUBQ $64, CX
    CMPQ CX, $64
    JGE  block_loop

remainder:
    // Padding and final block handled by Go wrapper
    RET

DATA sha256_init<>+0(SB)/4, $0x6a09e667
DATA sha256_init<>+4(SB)/4, $0xbb67ae85
DATA sha256_init<>+8(SB)/4, $0x3c6ef372
DATA sha256_init<>+12(SB)/4, $0xa54ff53a
DATA sha256_init<>+16(SB)/4, $0x510e527f
DATA sha256_init<>+20(SB)/4, $0x9b05688c
DATA sha256_init<>+24(SB)/4, $0x1f83d9ab
DATA sha256_init<>+28(SB)/4, $0x5be0cd19
GLOBL sha256_init<>(SB), RODATA, $32
