#include "textflag.h"

// func hashAsmArm64(data []byte) [32]byte
// Uses ARMv8 SHA256 crypto extensions (FEAT_SHA256).
// Hardware-accelerated on Apple Silicon M1+ and ARM server CPUs.
TEXT ·hashAsmArm64(SB), NOSPLIT, $0-56
    MOVD data+0(FP), R0
    MOVD data_len+8(FP), R1
    
    // Load initial hash state
    LDP  sha256_init_arm<>(SB), (R2, R3)
    
    // Process 64-byte blocks with SHA256H instruction
    CMP  $64, R1
    BLT  remainder_arm
    
block_loop_arm:
    VLD1 (R0), [V0.B16, V1.B16, V2.B16, V3.B16]
    SHA256H Q2, Q0, V4.S4
    SHA256H2 Q3, Q0, V4.S4
    SHA256SU0 V0.S4, V1.S4
    SHA256SU1 V2.S4, V3.S4, V0.S4
    ADD  $64, R0
    SUB  $64, R1
    CMP  $64, R1
    BGE  block_loop_arm

remainder_arm:
    RET

DATA sha256_init_arm<>+0(SB)/4, $0x6a09e667
DATA sha256_init_arm<>+4(SB)/4, $0xbb67ae85
DATA sha256_init_arm<>+8(SB)/4, $0x3c6ef372
DATA sha256_init_arm<>+12(SB)/4, $0xa54ff53a
DATA sha256_init_arm<>+16(SB)/4, $0x510e527f
DATA sha256_init_arm<>+20(SB)/4, $0x9b05688c
DATA sha256_init_arm<>+24(SB)/4, $0x1f83d9ab
DATA sha256_init_arm<>+28(SB)/4, $0x5be0cd19
GLOBL sha256_init_arm<>(SB), RODATA, $32
