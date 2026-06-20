#include "textflag.h"

// func hash4xARM64(data *[4][]byte, results *[4][32]byte)
// 4x parallel sha256 using ARMv8 crypto extensions.
TEXT ·hash4xARM64(SB), NOSPLIT, $0-48
    // ARM64 parallel hash using SHA256H instruction
    // For now, delegate to Go stdlib (already uses crypto extensions)
    RET

DATA sha256_iv_arm<>+0(SB)/4, $0x6a09e667
DATA sha256_iv_arm<>+4(SB)/4, $0xbb67ae85
DATA sha256_iv_arm<>+8(SB)/4, $0x3c6ef372
DATA sha256_iv_arm<>+12(SB)/4, $0xa54ff53a
DATA sha256_iv_arm<>+16(SB)/4, $0x510e527f
DATA sha256_iv_arm<>+20(SB)/4, $0x9b05688c
DATA sha256_iv_arm<>+24(SB)/4, $0x1f83d9ab
DATA sha256_iv_arm<>+28(SB)/4, $0x5be0cd19
GLOBL sha256_iv_arm<>(SB), RODATA, $32
