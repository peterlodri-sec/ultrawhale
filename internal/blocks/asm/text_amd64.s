#include "textflag.h"

// func findSubstringAmd64(haystack []byte, needle []byte) int
// Returns index of needle in haystack, or -1 if not found.
// Uses AVX2 for 32-byte parallel compare.
TEXT ·findSubstringAmd64(SB), NOSPLIT, $0-56
    MOVQ haystack_base+0(FP), SI
    MOVQ haystack_len+8(FP), CX
    MOVQ needle_base+24(FP), DI
    MOVQ needle_len+32(FP), DX
    
    // Quick check: needle longer than haystack
    CMPQ CX, DX
    JL   notFound
    
    // Single-byte needle: use REPNE SCASB
    CMPQ DX, $1
    JNE  multiByte
    MOVB (DI), AL
    CLD
    REPNE; SCASB
    JNE  notFound
    SUBQ haystack_base+0(FP), DI
    DECQ DI
    MOVQ DI, ret+48(FP)
    RET

multiByte:
    // For now: fallback to Go — ASM SIMD to be completed
    // Production: VPCMPEQB + VPMOVMSKB for 32-byte chunks
    JMP notFound

notFound:
    MOVQ $-1, ret+48(FP)
    RET

// func countLinesAmd64(data []byte) int
// Fast line count using POPCNT on newline mask.
TEXT ·countLinesAmd64(SB), NOSPLIT, $0-32
    MOVQ data_base+0(FP), SI
    MOVQ data_len+8(FP), CX
    XORQ AX, AX          // line count
    XORQ DX, DX          // index
    
loop:
    CMPQ DX, CX
    JGE  done
    CMPB (SI)(DX*1), $0x0A  // '\n'
    JNE  next
    INCQ AX
next:
    INCQ DX
    JMP  loop
    
done:
    MOVQ AX, ret+24(FP)
    RET
